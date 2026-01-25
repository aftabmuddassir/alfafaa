package services

import (
	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
)

// SearchService defines the interface for search operations
type SearchService interface {
	Search(query *dto.SearchQuery) (*dto.SearchResponse, error)
}

type searchService struct {
	articleRepo  repositories.ArticleRepository
	categoryRepo repositories.CategoryRepository
	tagRepo      repositories.TagRepository
}

// NewSearchService creates a new search service
func NewSearchService(
	articleRepo repositories.ArticleRepository,
	categoryRepo repositories.CategoryRepository,
	tagRepo repositories.TagRepository,
) SearchService {
	return &searchService{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}
}

// Search performs a search across articles, categories, and tags
func (s *searchService) Search(query *dto.SearchQuery) (*dto.SearchResponse, error) {
	response := &dto.SearchResponse{}
	searchType := query.GetType()

	// Search articles
	if searchType == "all" || searchType == "articles" {
		filters := repositories.ArticleFilters{
			Status: string(models.StatusPublished),
			Limit:  query.GetPerPage(),
			Offset: query.GetOffset(),
		}

		articles, _, err := s.articleRepo.Search(query.Query, filters)
		if err != nil {
			return nil, utils.WrapError(err, "failed to search articles")
		}

		for _, article := range articles {
			response.Articles = append(response.Articles, toArticleListItemResponse(&article))
		}
	}

	// Search categories
	if searchType == "all" || searchType == "categories" {
		categoryFilters := repositories.CategoryFilters{
			Search: query.Query,
		}

		categories, err := s.categoryRepo.FindAll(categoryFilters)
		if err != nil {
			return nil, utils.WrapError(err, "failed to search categories")
		}

		for _, category := range categories {
			var parentID *string
			if category.ParentID != nil {
				pid := category.ParentID.String()
				parentID = &pid
			}
			response.Categories = append(response.Categories, dto.CategoryResponse{
				ID:           category.ID.String(),
				Name:         category.Name,
				Slug:         category.Slug,
				Description:  category.Description,
				ParentID:     parentID,
				DisplayOrder: category.DisplayOrder,
				IsActive:     category.IsActive,
				CreatedAt:    category.CreatedAt,
				UpdatedAt:    category.UpdatedAt,
			})
		}
	}

	// Search tags
	if searchType == "all" || searchType == "tags" {
		tagFilters := repositories.TagFilters{
			Search: query.Query,
			Limit:  20,
		}

		tags, _, err := s.tagRepo.FindAll(tagFilters)
		if err != nil {
			return nil, utils.WrapError(err, "failed to search tags")
		}

		for _, tag := range tags {
			response.Tags = append(response.Tags, dto.TagResponse{
				ID:          tag.ID.String(),
				Name:        tag.Name,
				Slug:        tag.Slug,
				Description: tag.Description,
				UsageCount:  tag.UsageCount,
				CreatedAt:   tag.CreatedAt,
				UpdatedAt:   tag.UpdatedAt,
			})
		}
	}

	return response, nil
}
