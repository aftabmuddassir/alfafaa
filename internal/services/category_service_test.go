package services

import (
	"testing"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/alfafaa/alfafaa-blog/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type CategoryServiceTestSuite struct {
	suite.Suite
	categoryRepo *mocks.MockCategoryRepository
	articleRepo  *mocks.MockArticleRepository
	service      CategoryService
}

func (suite *CategoryServiceTestSuite) SetupTest() {
	suite.categoryRepo = new(mocks.MockCategoryRepository)
	suite.articleRepo = new(mocks.MockArticleRepository)
	suite.service = NewCategoryService(suite.categoryRepo, suite.articleRepo)
}

func TestCategoryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryServiceTestSuite))
}

// CreateCategory Tests

func (suite *CategoryServiceTestSuite) TestCreateCategory_Success() {
	req := &dto.CreateCategoryRequest{
		Name:         "Technology",
		Description:  "Tech articles",
		DisplayOrder: 1,
	}

	suite.categoryRepo.On("ExistsBySlug", "technology").Return(false, nil)
	suite.categoryRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(nil)

	result, err := suite.service.CreateCategory(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Technology", result.Name)
	assert.Equal(suite.T(), "technology", result.Slug)
	assert.True(suite.T(), result.IsActive)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestCreateCategory_WithParent_Success() {
	parentID := uuid.New()
	parentIDStr := parentID.String()
	parentCategory := &models.Category{
		ID:       parentID,
		Name:     "Parent Category",
		Slug:     "parent-category",
		IsActive: true,
	}

	req := &dto.CreateCategoryRequest{
		Name:     "Child Category",
		ParentID: &parentIDStr,
	}

	suite.categoryRepo.On("ExistsBySlug", "child-category").Return(false, nil)
	suite.categoryRepo.On("FindByID", parentID).Return(parentCategory, nil)
	suite.categoryRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(nil)

	result, err := suite.service.CreateCategory(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Child Category", result.Name)
	assert.NotNil(suite.T(), result.ParentID)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestCreateCategory_InvalidParentID() {
	invalidID := "not-a-uuid"
	req := &dto.CreateCategoryRequest{
		Name:     "Category",
		ParentID: &invalidID,
	}

	suite.categoryRepo.On("ExistsBySlug", "category").Return(false, nil)

	result, err := suite.service.CreateCategory(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "INVALID_PARENT_ID", appErr.Code)
}

func (suite *CategoryServiceTestSuite) TestCreateCategory_ParentNotFound() {
	parentID := uuid.New()
	parentIDStr := parentID.String()

	req := &dto.CreateCategoryRequest{
		Name:     "Category",
		ParentID: &parentIDStr,
	}

	suite.categoryRepo.On("ExistsBySlug", "category").Return(false, nil)
	suite.categoryRepo.On("FindByID", parentID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.CreateCategory(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "PARENT_NOT_FOUND", appErr.Code)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestCreateCategory_DuplicateSlug_GeneratesUnique() {
	req := &dto.CreateCategoryRequest{
		Name: "Technology",
	}

	// First slug exists, second one doesn't
	suite.categoryRepo.On("ExistsBySlug", "technology").Return(true, nil).Once()
	suite.categoryRepo.On("ExistsBySlug", "technology-1").Return(false, nil).Once()
	suite.categoryRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(nil)

	result, err := suite.service.CreateCategory(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "technology-1", result.Slug)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestCreateCategory_WithIsActiveFalse() {
	isActive := false
	req := &dto.CreateCategoryRequest{
		Name:     "Inactive Category",
		IsActive: &isActive,
	}

	suite.categoryRepo.On("ExistsBySlug", "inactive-category").Return(false, nil)
	suite.categoryRepo.On("Create", mock.AnythingOfType("*models.Category")).Return(nil)

	result, err := suite.service.CreateCategory(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.False(suite.T(), result.IsActive)
	suite.categoryRepo.AssertExpectations(suite.T())
}

// GetCategory Tests

func (suite *CategoryServiceTestSuite) TestGetCategory_Success() {
	categoryID := uuid.New()
	category := &models.Category{
		ID:          categoryID,
		Name:        "Technology",
		Slug:        "technology",
		Description: "Tech articles",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.categoryRepo.On("FindBySlug", "technology").Return(category, nil)
	suite.categoryRepo.On("GetArticleCount", categoryID).Return(int64(5), nil)

	result, err := suite.service.GetCategory("technology")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Technology", result.Name)
	assert.Equal(suite.T(), 5, result.ArticleCount)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestGetCategory_NotFound() {
	suite.categoryRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetCategory("nonexistent")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.categoryRepo.AssertExpectations(suite.T())
}

// GetCategories Tests

func (suite *CategoryServiceTestSuite) TestGetCategories_Success() {
	categories := []models.Category{
		{
			ID:       uuid.New(),
			Name:     "Technology",
			Slug:     "technology",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Name:     "Science",
			Slug:     "science",
			IsActive: true,
		},
	}

	query := &dto.CategoryListQuery{}

	suite.categoryRepo.On("FindAll", mock.AnythingOfType("repositories.CategoryFilters")).
		Return(categories, nil)

	result, err := suite.service.GetCategories(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestGetCategories_WithFilters() {
	categories := []models.Category{
		{
			ID:       uuid.New(),
			Name:     "Technology",
			Slug:     "technology",
			IsActive: true,
		},
	}

	query := &dto.CategoryListQuery{
		Search:          "tech",
		IncludeInactive: true,
		ParentOnly:      true,
	}

	suite.categoryRepo.On("FindAll", mock.MatchedBy(func(filters repositories.CategoryFilters) bool {
		return filters.Search == "tech" && filters.IncludeInactive && filters.ParentOnly
	})).Return(categories, nil)

	result, err := suite.service.GetCategories(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestGetCategories_Empty() {
	query := &dto.CategoryListQuery{}

	suite.categoryRepo.On("FindAll", mock.AnythingOfType("repositories.CategoryFilters")).
		Return([]models.Category{}, nil)

	result, err := suite.service.GetCategories(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
	suite.categoryRepo.AssertExpectations(suite.T())
}

// GetCategoriesHierarchical Tests

func (suite *CategoryServiceTestSuite) TestGetCategoriesHierarchical_Success() {
	categories := []models.Category{
		{
			ID:       uuid.New(),
			Name:     "Technology",
			Slug:     "technology",
			IsActive: true,
			Children: []models.Category{
				{
					ID:       uuid.New(),
					Name:     "Programming",
					Slug:     "programming",
					IsActive: true,
				},
			},
		},
	}

	suite.categoryRepo.On("FindAllHierarchical").Return(categories, nil)

	result, err := suite.service.GetCategoriesHierarchical()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Len(suite.T(), result[0].Children, 1)
	assert.Equal(suite.T(), "Programming", result[0].Children[0].Name)
	suite.categoryRepo.AssertExpectations(suite.T())
}

// UpdateCategory Tests

func (suite *CategoryServiceTestSuite) TestUpdateCategory_Success() {
	categoryID := uuid.New()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Technology",
		Slug:     "technology",
		IsActive: true,
	}

	newName := "Updated Technology"
	newDescription := "Updated description"
	req := &dto.UpdateCategoryRequest{
		Name:        &newName,
		Description: &newDescription,
	}

	suite.categoryRepo.On("FindByID", categoryID).Return(category, nil)
	suite.categoryRepo.On("ExistsBySlug", "updated-technology").Return(false, nil)
	suite.categoryRepo.On("Update", mock.AnythingOfType("*models.Category")).Return(nil)

	result, err := suite.service.UpdateCategory(categoryID.String(), req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Updated Technology", result.Name)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestUpdateCategory_InvalidUUID() {
	req := &dto.UpdateCategoryRequest{}

	result, err := suite.service.UpdateCategory("not-a-uuid", req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *CategoryServiceTestSuite) TestUpdateCategory_NotFound() {
	categoryID := uuid.New()
	req := &dto.UpdateCategoryRequest{}

	suite.categoryRepo.On("FindByID", categoryID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateCategory(categoryID.String(), req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestUpdateCategory_SelfReference_Prevented() {
	categoryID := uuid.New()
	categoryIDStr := categoryID.String()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Category",
		Slug:     "category",
		IsActive: true,
	}

	req := &dto.UpdateCategoryRequest{
		ParentID: &categoryIDStr,
	}

	suite.categoryRepo.On("FindByID", categoryID).Return(category, nil)

	result, err := suite.service.UpdateCategory(categoryID.String(), req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "INVALID_PARENT", appErr.Code)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestUpdateCategory_RemoveParent() {
	categoryID := uuid.New()
	parentID := uuid.New()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Category",
		Slug:     "category",
		ParentID: &parentID,
		IsActive: true,
	}

	emptyParent := ""
	req := &dto.UpdateCategoryRequest{
		ParentID: &emptyParent,
	}

	suite.categoryRepo.On("FindByID", categoryID).Return(category, nil)
	suite.categoryRepo.On("Update", mock.AnythingOfType("*models.Category")).Return(nil)

	result, err := suite.service.UpdateCategory(categoryID.String(), req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Nil(suite.T(), result.ParentID)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestUpdateCategory_UpdateDisplayOrder() {
	categoryID := uuid.New()
	category := &models.Category{
		ID:           categoryID,
		Name:         "Category",
		Slug:         "category",
		DisplayOrder: 1,
		IsActive:     true,
	}

	newOrder := 5
	req := &dto.UpdateCategoryRequest{
		DisplayOrder: &newOrder,
	}

	suite.categoryRepo.On("FindByID", categoryID).Return(category, nil)
	suite.categoryRepo.On("Update", mock.AnythingOfType("*models.Category")).Return(nil)

	result, err := suite.service.UpdateCategory(categoryID.String(), req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 5, result.DisplayOrder)
	suite.categoryRepo.AssertExpectations(suite.T())
}

// DeleteCategory Tests

func (suite *CategoryServiceTestSuite) TestDeleteCategory_Success() {
	categoryID := uuid.New()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Category",
		Slug:     "category",
		IsActive: true,
		Children: []models.Category{},
	}

	suite.categoryRepo.On("FindByID", categoryID).Return(category, nil)
	suite.categoryRepo.On("GetArticleCount", categoryID).Return(int64(0), nil)
	suite.categoryRepo.On("Delete", categoryID).Return(nil)

	err := suite.service.DeleteCategory(categoryID.String())

	assert.NoError(suite.T(), err)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestDeleteCategory_InvalidUUID() {
	err := suite.service.DeleteCategory("not-a-uuid")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *CategoryServiceTestSuite) TestDeleteCategory_NotFound() {
	categoryID := uuid.New()

	suite.categoryRepo.On("FindByID", categoryID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.DeleteCategory(categoryID.String())

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestDeleteCategory_HasChildren() {
	categoryID := uuid.New()
	childID := uuid.New()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Category",
		Slug:     "category",
		IsActive: true,
		Children: []models.Category{
			{
				ID:       childID,
				Name:     "Child",
				Slug:     "child",
				ParentID: &categoryID,
			},
		},
	}

	suite.categoryRepo.On("FindByID", categoryID).Return(category, nil)

	err := suite.service.DeleteCategory(categoryID.String())

	assert.Error(suite.T(), err)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "HAS_CHILDREN", appErr.Code)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestDeleteCategory_HasArticles() {
	categoryID := uuid.New()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Category",
		Slug:     "category",
		IsActive: true,
		Children: []models.Category{},
	}

	suite.categoryRepo.On("FindByID", categoryID).Return(category, nil)
	suite.categoryRepo.On("GetArticleCount", categoryID).Return(int64(5), nil)

	err := suite.service.DeleteCategory(categoryID.String())

	assert.Error(suite.T(), err)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "HAS_ARTICLES", appErr.Code)
	suite.categoryRepo.AssertExpectations(suite.T())
}

// GetCategoryArticles Tests

func (suite *CategoryServiceTestSuite) TestGetCategoryArticles_Success() {
	categoryID := uuid.New()
	authorID := uuid.New()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Technology",
		Slug:     "technology",
		IsActive: true,
	}

	articles := []models.Article{
		{
			ID:       uuid.New(),
			Title:    "Article 1",
			Slug:     "article-1",
			Status:   models.StatusPublished,
			AuthorID: authorID,
			Author: &models.User{
				ID:       authorID,
				Username: "author1",
			},
		},
		{
			ID:       uuid.New(),
			Title:    "Article 2",
			Slug:     "article-2",
			Status:   models.StatusPublished,
			AuthorID: authorID,
			Author: &models.User{
				ID:       authorID,
				Username: "author1",
			},
		},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.categoryRepo.On("FindBySlug", "technology").Return(category, nil)
	suite.articleRepo.On("FindByCategory", categoryID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(2), nil)

	result, total, err := suite.service.GetCategoryArticles("technology", query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
	suite.categoryRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestGetCategoryArticles_CategoryNotFound() {
	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.categoryRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, total, err := suite.service.GetCategoryArticles("nonexistent", query)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), int64(0), total)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.categoryRepo.AssertExpectations(suite.T())
}

func (suite *CategoryServiceTestSuite) TestGetCategoryArticles_NoArticles() {
	categoryID := uuid.New()
	category := &models.Category{
		ID:       categoryID,
		Name:     "Technology",
		Slug:     "technology",
		IsActive: true,
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.categoryRepo.On("FindBySlug", "technology").Return(category, nil)
	suite.articleRepo.On("FindByCategory", categoryID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return([]models.Article{}, int64(0), nil)

	result, total, err := suite.service.GetCategoryArticles("technology", query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
	assert.Equal(suite.T(), int64(0), total)
	suite.categoryRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}
