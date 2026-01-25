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

// TagService defines the interface for tag operations
type TagService interface {
	CreateTag(req *dto.CreateTagRequest) (*dto.TagResponse, error)
	GetTag(slug string) (*dto.TagDetailResponse, error)
	GetTags(query *dto.TagListQuery) ([]dto.TagResponse, int64, error)
	GetPopularTags(limit int) ([]dto.PopularTagResponse, error)
	UpdateTag(id string, req *dto.UpdateTagRequest) (*dto.TagResponse, error)
	DeleteTag(id string) error
	GetTagArticles(slug string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error)
}

type tagService struct {
	tagRepo     repositories.TagRepository
	articleRepo repositories.ArticleRepository
}

// NewTagService creates a new tag service
func NewTagService(tagRepo repositories.TagRepository, articleRepo repositories.ArticleRepository) TagService {
	return &tagService{
		tagRepo:     tagRepo,
		articleRepo: articleRepo,
	}
}

// CreateTag creates a new tag
func (s *tagService) CreateTag(req *dto.CreateTagRequest) (*dto.TagResponse, error) {
	// Generate slug from name
	slug := utils.GenerateSlug(req.Name)

	// Check if slug already exists
	exists, err := s.tagRepo.ExistsBySlug(slug)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check slug")
	}
	if exists {
		// Append a suffix to make it unique
		for i := 1; exists; i++ {
			slug = utils.GenerateUniqueSlug(utils.GenerateSlug(req.Name), fmt.Sprintf("%d", i))
			exists, err = s.tagRepo.ExistsBySlug(slug)
			if err != nil {
				return nil, utils.WrapError(err, "failed to check slug")
			}
		}
	}

	tag := &models.Tag{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		UsageCount:  0,
	}

	if err := s.tagRepo.Create(tag); err != nil {
		return nil, utils.WrapError(err, "failed to create tag")
	}

	return s.toResponse(tag), nil
}

// GetTag retrieves a tag by slug
func (s *tagService) GetTag(slug string) (*dto.TagDetailResponse, error) {
	tag, err := s.tagRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find tag")
	}

	return s.toDetailResponse(tag), nil
}

// GetTags retrieves all tags
func (s *tagService) GetTags(query *dto.TagListQuery) ([]dto.TagResponse, int64, error) {
	filters := repositories.TagFilters{
		Search: query.Search,
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
	}

	tags, total, err := s.tagRepo.FindAll(filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to find tags")
	}

	responses := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		responses[i] = *s.toResponse(&tag)
	}

	return responses, total, nil
}

// GetPopularTags retrieves the most popular tags
func (s *tagService) GetPopularTags(limit int) ([]dto.PopularTagResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	tags, err := s.tagRepo.FindPopular(limit)
	if err != nil {
		return nil, utils.WrapError(err, "failed to find popular tags")
	}

	responses := make([]dto.PopularTagResponse, len(tags))
	for i, tag := range tags {
		responses[i] = dto.PopularTagResponse{
			ID:         tag.ID.String(),
			Name:       tag.Name,
			Slug:       tag.Slug,
			UsageCount: tag.UsageCount,
		}
	}

	return responses, nil
}

// UpdateTag updates a tag
func (s *tagService) UpdateTag(id string, req *dto.UpdateTagRequest) (*dto.TagResponse, error) {
	tagID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	tag, err := s.tagRepo.FindByID(tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find tag")
	}

	// Update fields
	if req.Name != nil {
		tag.Name = *req.Name
		// Update slug if name changed
		newSlug := utils.GenerateSlug(*req.Name)
		if newSlug != tag.Slug {
			exists, err := s.tagRepo.ExistsBySlug(newSlug)
			if err != nil {
				return nil, utils.WrapError(err, "failed to check slug")
			}
			if !exists {
				tag.Slug = newSlug
			}
		}
	}
	if req.Description != nil {
		tag.Description = *req.Description
	}

	if err := s.tagRepo.Update(tag); err != nil {
		return nil, utils.WrapError(err, "failed to update tag")
	}

	return s.toResponse(tag), nil
}

// DeleteTag deletes a tag
func (s *tagService) DeleteTag(id string) error {
	tagID, err := uuid.Parse(id)
	if err != nil {
		return utils.ErrBadRequest
	}

	tag, err := s.tagRepo.FindByID(tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to find tag")
	}

	// Check if tag is in use
	if tag.UsageCount > 0 {
		return utils.NewAppError("TAG_IN_USE", "Cannot delete tag that is in use", 400)
	}

	if err := s.tagRepo.Delete(tagID); err != nil {
		return utils.WrapError(err, "failed to delete tag")
	}

	return nil
}

// GetTagArticles retrieves articles with a tag
func (s *tagService) GetTagArticles(slug string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error) {
	tag, err := s.tagRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, utils.ErrNotFound
		}
		return nil, 0, utils.WrapError(err, "failed to find tag")
	}

	filters := repositories.ArticleFilters{
		Status: string(models.StatusPublished),
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
		Sort:   query.GetSort(),
	}

	articles, total, err := s.articleRepo.FindByTag(tag.ID, filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to find articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}

// toResponse converts a tag model to a response DTO
func (s *tagService) toResponse(tag *models.Tag) *dto.TagResponse {
	return &dto.TagResponse{
		ID:          tag.ID.String(),
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		UsageCount:  tag.UsageCount,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}

// toDetailResponse converts a tag model to a detail response DTO
func (s *tagService) toDetailResponse(tag *models.Tag) *dto.TagDetailResponse {
	return &dto.TagDetailResponse{
		ID:          tag.ID.String(),
		Name:        tag.Name,
		Slug:        tag.Slug,
		Description: tag.Description,
		UsageCount:  tag.UsageCount,
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
	}
}
