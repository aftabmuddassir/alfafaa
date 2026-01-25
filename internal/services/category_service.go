package services

import (
	"errors"
	"fmt"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryService defines the interface for category operations
type CategoryService interface {
	CreateCategory(req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetCategory(slug string) (*dto.CategoryDetailResponse, error)
	GetCategories(query *dto.CategoryListQuery) ([]dto.CategoryResponse, error)
	GetCategoriesHierarchical() ([]*dto.CategoryTreeResponse, error)
	UpdateCategory(id string, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	DeleteCategory(id string) error
	GetCategoryArticles(slug string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error)
}

type categoryService struct {
	categoryRepo repositories.CategoryRepository
	articleRepo  repositories.ArticleRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo repositories.CategoryRepository, articleRepo repositories.ArticleRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		articleRepo:  articleRepo,
	}
}

// CreateCategory creates a new category
func (s *categoryService) CreateCategory(req *dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	// Generate slug from name
	slug := utils.GenerateSlug(req.Name)

	// Check if slug already exists
	exists, err := s.categoryRepo.ExistsBySlug(slug)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check slug")
	}
	if exists {
		// Append a suffix to make it unique
		for i := 1; exists; i++ {
			slug = utils.GenerateUniqueSlug(utils.GenerateSlug(req.Name), fmt.Sprintf("%d", i))
			exists, err = s.categoryRepo.ExistsBySlug(slug)
			if err != nil {
				return nil, utils.WrapError(err, "failed to check slug")
			}
		}
	}

	category := &models.Category{
		Name:         req.Name,
		Slug:         slug,
		Description:  req.Description,
		DisplayOrder: req.DisplayOrder,
		IsActive:     true,
	}

	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if req.ParentID != nil {
		parentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, utils.NewAppError("INVALID_PARENT_ID", "Invalid parent ID", 400)
		}
		category.ParentID = &parentID

		// Verify parent exists
		_, err = s.categoryRepo.FindByID(parentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, utils.NewAppError("PARENT_NOT_FOUND", "Parent category not found", 404)
			}
			return nil, utils.WrapError(err, "failed to find parent category")
		}
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, utils.WrapError(err, "failed to create category")
	}

	return s.toResponse(category), nil
}

// GetCategory retrieves a category by slug
func (s *categoryService) GetCategory(slug string) (*dto.CategoryDetailResponse, error) {
	category, err := s.categoryRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find category")
	}

	// Get article count
	count, _ := s.categoryRepo.GetArticleCount(category.ID)

	return s.toDetailResponse(category, int(count)), nil
}

// GetCategories retrieves all categories
func (s *categoryService) GetCategories(query *dto.CategoryListQuery) ([]dto.CategoryResponse, error) {
	filters := repositories.CategoryFilters{
		IncludeInactive: query.IncludeInactive,
		ParentOnly:      query.ParentOnly,
		Search:          query.Search,
	}

	categories, err := s.categoryRepo.FindAll(filters)
	if err != nil {
		return nil, utils.WrapError(err, "failed to find categories")
	}

	responses := make([]dto.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = *s.toResponse(&category)
	}

	return responses, nil
}

// GetCategoriesHierarchical retrieves categories in a hierarchical structure
func (s *categoryService) GetCategoriesHierarchical() ([]*dto.CategoryTreeResponse, error) {
	categories, err := s.categoryRepo.FindAllHierarchical()
	if err != nil {
		return nil, utils.WrapError(err, "failed to find categories")
	}

	responses := make([]*dto.CategoryTreeResponse, len(categories))
	for i, category := range categories {
		responses[i] = s.toTreeResponse(&category)
	}

	return responses, nil
}

// UpdateCategory updates a category
func (s *categoryService) UpdateCategory(id string, req *dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	category, err := s.categoryRepo.FindByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find category")
	}

	// Update fields
	if req.Name != nil {
		category.Name = *req.Name
		// Update slug if name changed
		newSlug := utils.GenerateSlug(*req.Name)
		if newSlug != category.Slug {
			exists, err := s.categoryRepo.ExistsBySlug(newSlug)
			if err != nil {
				return nil, utils.WrapError(err, "failed to check slug")
			}
			if !exists {
				category.Slug = newSlug
			}
		}
	}
	if req.Description != nil {
		category.Description = *req.Description
	}
	if req.DisplayOrder != nil {
		category.DisplayOrder = *req.DisplayOrder
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.ParentID != nil {
		if *req.ParentID == "" {
			category.ParentID = nil
		} else {
			parentID, err := uuid.Parse(*req.ParentID)
			if err != nil {
				return nil, utils.NewAppError("INVALID_PARENT_ID", "Invalid parent ID", 400)
			}
			// Prevent self-reference
			if parentID == category.ID {
				return nil, utils.NewAppError("INVALID_PARENT", "Category cannot be its own parent", 400)
			}
			category.ParentID = &parentID
		}
	}

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, utils.WrapError(err, "failed to update category")
	}

	return s.toResponse(category), nil
}

// DeleteCategory deletes a category
func (s *categoryService) DeleteCategory(id string) error {
	categoryID, err := uuid.Parse(id)
	if err != nil {
		return utils.ErrBadRequest
	}

	category, err := s.categoryRepo.FindByID(categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to find category")
	}

	// Check if category has children
	if len(category.Children) > 0 {
		return utils.NewAppError("HAS_CHILDREN", "Cannot delete category with subcategories", 400)
	}

	// Check if category has articles
	count, _ := s.categoryRepo.GetArticleCount(categoryID)
	if count > 0 {
		return utils.NewAppError("HAS_ARTICLES", "Cannot delete category with articles", 400)
	}

	if err := s.categoryRepo.Delete(categoryID); err != nil {
		return utils.WrapError(err, "failed to delete category")
	}

	return nil
}

// GetCategoryArticles retrieves articles in a category
func (s *categoryService) GetCategoryArticles(slug string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error) {
	category, err := s.categoryRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, utils.ErrNotFound
		}
		return nil, 0, utils.WrapError(err, "failed to find category")
	}

	filters := repositories.ArticleFilters{
		Status: string(models.StatusPublished),
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
		Sort:   query.GetSort(),
	}

	articles, total, err := s.articleRepo.FindByCategory(category.ID, filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to find articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}

// toResponse converts a category model to a response DTO
func (s *categoryService) toResponse(category *models.Category) *dto.CategoryResponse {
	response := &dto.CategoryResponse{
		ID:           category.ID.String(),
		Name:         category.Name,
		Slug:         category.Slug,
		Description:  category.Description,
		DisplayOrder: category.DisplayOrder,
		IsActive:     category.IsActive,
		CreatedAt:    category.CreatedAt,
		UpdatedAt:    category.UpdatedAt,
	}

	if category.ParentID != nil {
		parentIDStr := category.ParentID.String()
		response.ParentID = &parentIDStr
	}

	return response
}

// toDetailResponse converts a category model to a detail response DTO
func (s *categoryService) toDetailResponse(category *models.Category, articleCount int) *dto.CategoryDetailResponse {
	response := &dto.CategoryDetailResponse{
		ID:           category.ID.String(),
		Name:         category.Name,
		Slug:         category.Slug,
		Description:  category.Description,
		DisplayOrder: category.DisplayOrder,
		IsActive:     category.IsActive,
		ArticleCount: articleCount,
		CreatedAt:    category.CreatedAt,
		UpdatedAt:    category.UpdatedAt,
	}

	if category.ParentID != nil {
		parentIDStr := category.ParentID.String()
		response.ParentID = &parentIDStr
	}

	if category.Parent != nil {
		response.Parent = s.toResponse(category.Parent)
	}

	for _, child := range category.Children {
		response.Children = append(response.Children, *s.toResponse(&child))
	}

	return response
}

// toTreeResponse converts a category model to a tree response DTO
func (s *categoryService) toTreeResponse(category *models.Category) *dto.CategoryTreeResponse {
	response := &dto.CategoryTreeResponse{
		ID:           category.ID.String(),
		Name:         category.Name,
		Slug:         category.Slug,
		Description:  category.Description,
		DisplayOrder: category.DisplayOrder,
		IsActive:     category.IsActive,
		CreatedAt:    category.CreatedAt,
	}

	for _, child := range category.Children {
		response.Children = append(response.Children, s.toTreeResponse(&child))
	}

	return response
}
