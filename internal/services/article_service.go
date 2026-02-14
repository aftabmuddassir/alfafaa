package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArticleService defines the interface for article operations
type ArticleService interface {
	CreateArticle(req *dto.CreateArticleRequest, authorID string) (*dto.ArticleDetailResponse, error)
	GetArticle(slug string, incrementView bool) (*dto.ArticleDetailResponse, error)
	GetArticles(query *dto.ArticleListQuery, includeUnpublished bool) ([]dto.ArticleListItemResponse, int64, error)
	UpdateArticle(id string, req *dto.UpdateArticleRequest, userID string, isEditor bool) (*dto.ArticleDetailResponse, error)
	DeleteArticle(id string, userID string, isEditor bool) error
	PublishArticle(id string) (*dto.ArticleDetailResponse, error)
	UnpublishArticle(id string) (*dto.ArticleDetailResponse, error)
	GetTrendingArticles(limit int) ([]dto.ArticleListItemResponse, error)
	GetRecentArticles(limit int) ([]dto.ArticleListItemResponse, error)
	GetRelatedArticles(slug string, limit int) ([]dto.ArticleListItemResponse, error)
	SearchArticles(query string, filters *dto.ArticleListQuery) ([]dto.ArticleListItemResponse, int64, error)
}

type articleService struct {
	db             *gorm.DB
	articleRepo    repositories.ArticleRepository
	categoryRepo   repositories.CategoryRepository
	tagRepo        repositories.TagRepository
	engagementRepo repositories.EngagementRepository
	userRepo       repositories.UserRepository
}

// NewArticleService creates a new article service
func NewArticleService(
	db *gorm.DB,
	articleRepo repositories.ArticleRepository,
	categoryRepo repositories.CategoryRepository,
	tagRepo repositories.TagRepository,
	opts ...ArticleServiceOption,
) ArticleService {
	svc := &articleService{
		db:           db,
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// ArticleServiceOption is a functional option for configuring the article service
type ArticleServiceOption func(*articleService)

// WithEngagementRepo sets the engagement repository on the article service
func WithEngagementRepo(repo repositories.EngagementRepository) ArticleServiceOption {
	return func(s *articleService) {
		s.engagementRepo = repo
	}
}

// WithUserRepo sets the user repository on the article service (for publish notifications)
func WithUserRepo(repo repositories.UserRepository) ArticleServiceOption {
	return func(s *articleService) {
		s.userRepo = repo
	}
}

// CreateArticle creates a new article
func (s *articleService) CreateArticle(req *dto.CreateArticleRequest, authorID string) (*dto.ArticleDetailResponse, error) {
	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	// Generate slug from title
	slug := utils.GenerateSlug(req.Title)
	slug = utils.TruncateSlug(slug, 200)

	// Check if slug already exists
	exists, err := s.articleRepo.ExistsBySlug(slug)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check slug")
	}
	if exists {
		// Append a suffix to make it unique
		for i := 1; exists; i++ {
			slug = utils.GenerateUniqueSlug(utils.TruncateSlug(utils.GenerateSlug(req.Title), 190), fmt.Sprintf("%d", i))
			exists, err = s.articleRepo.ExistsBySlug(slug)
			if err != nil {
				return nil, utils.WrapError(err, "failed to check slug")
			}
		}
	}

	// Get categories
	categoryIDs := make([]uuid.UUID, len(req.CategoryIDs))
	for i, id := range req.CategoryIDs {
		catID, err := uuid.Parse(id)
		if err != nil {
			return nil, utils.NewAppError("INVALID_CATEGORY_ID", "Invalid category ID: "+id, 400)
		}
		categoryIDs[i] = catID
	}

	categories, err := s.categoryRepo.FindByIDs(categoryIDs)
	if err != nil {
		return nil, utils.WrapError(err, "failed to find categories")
	}
	if len(categories) != len(categoryIDs) {
		return nil, utils.NewAppError("CATEGORY_NOT_FOUND", "One or more categories not found", 404)
	}

	// Get tags
	var tags []models.Tag
	if len(req.TagIDs) > 0 {
		tagIDs := make([]uuid.UUID, len(req.TagIDs))
		for i, id := range req.TagIDs {
			tagID, err := uuid.Parse(id)
			if err != nil {
				return nil, utils.NewAppError("INVALID_TAG_ID", "Invalid tag ID: "+id, 400)
			}
			tagIDs[i] = tagID
		}

		tags, err = s.tagRepo.FindByIDs(tagIDs)
		if err != nil {
			return nil, utils.WrapError(err, "failed to find tags")
		}
	}

	// Determine status
	status := models.StatusDraft
	if req.Status == string(models.StatusPublished) {
		status = models.StatusPublished
	}

	// Generate excerpt if not provided
	excerpt := req.Excerpt
	if excerpt == "" && len(req.Content) > 200 {
		excerpt = req.Content[:200] + "..."
	}

	article := &models.Article{
		Title:            req.Title,
		Slug:             slug,
		Content:          req.Content,
		Excerpt:          excerpt,
		FeaturedImageURL: req.FeaturedImageURL,
		AuthorID:         authorUUID,
		Status:           status,
		MetaTitle:        req.MetaTitle,
		MetaDescription:  req.MetaDescription,
		MetaKeywords:     req.MetaKeywords,
		Categories:       categories,
		Tags:             tags,
	}

	if status == models.StatusPublished {
		now := time.Now()
		article.PublishedAt = &now
	}

	// Use transaction to ensure atomicity of article creation and tag updates
	if s.db != nil {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			// Create repositories scoped to this transaction
			txArticleRepo := s.articleRepo.WithTx(tx)
			txTagRepo := s.tagRepo.WithTx(tx)

			if err := txArticleRepo.Create(article); err != nil {
				return err
			}

			// Update tag usage counts within the same transaction
			for _, tag := range tags {
				if err := txTagRepo.IncrementUsage(tag.ID); err != nil {
					return err
				}
			}

			return nil
		})
	} else {
		// Fallback for unit tests without db - run without transaction
		if err := s.articleRepo.Create(article); err != nil {
			return nil, utils.WrapError(err, "failed to create article")
		}
		for _, tag := range tags {
			if err := s.tagRepo.IncrementUsage(tag.ID); err != nil {
				return nil, utils.WrapError(err, "failed to increment tag usage")
			}
		}
	}

	if err != nil {
		return nil, utils.WrapError(err, "failed to create article")
	}

	// Fetch the created article with relationships
	createdArticle, err := s.articleRepo.FindByID(article.ID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to fetch created article")
	}

	return s.toDetailResponse(createdArticle), nil
}

// GetArticle retrieves an article by slug
func (s *articleService) GetArticle(slug string, incrementView bool) (*dto.ArticleDetailResponse, error) {
	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	if incrementView {
		_ = s.articleRepo.IncrementViewCount(article.ID)
		article.ViewCount++
	}

	return s.toDetailResponse(article), nil
}

// GetArticles retrieves articles with filters
func (s *articleService) GetArticles(query *dto.ArticleListQuery, includeUnpublished bool) ([]dto.ArticleListItemResponse, int64, error) {
	filters := repositories.ArticleFilters{
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
		Sort:   query.GetSort(),
	}

	// Only show published articles unless user can see unpublished
	if !includeUnpublished {
		filters.Status = string(models.StatusPublished)
	} else if query.Status != "" {
		filters.Status = query.Status
	}

	if query.AuthorID != "" {
		authorID, err := uuid.Parse(query.AuthorID)
		if err != nil {
			return nil, 0, utils.NewAppError("INVALID_AUTHOR_ID", "Invalid author ID", 400)
		}
		filters.AuthorID = &authorID
	}

	// Handle category filter
	var articles []models.Article
	var total int64
	var err error

	if query.CategorySlug != "" {
		category, err := s.categoryRepo.FindBySlug(query.CategorySlug)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, utils.NewAppError("CATEGORY_NOT_FOUND", "Category not found", 404)
			}
			return nil, 0, utils.WrapError(err, "failed to find category")
		}
		articles, total, err = s.articleRepo.FindByCategory(category.ID, filters)
		if err != nil {
			return nil, 0, utils.WrapError(err, "failed to find articles")
		}
	} else if query.TagSlug != "" {
		tag, err := s.tagRepo.FindBySlug(query.TagSlug)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, utils.NewAppError("TAG_NOT_FOUND", "Tag not found", 404)
			}
			return nil, 0, utils.WrapError(err, "failed to find tag")
		}
		articles, total, err = s.articleRepo.FindByTag(tag.ID, filters)
		if err != nil {
			return nil, 0, utils.WrapError(err, "failed to find articles")
		}
	} else if query.Search != "" {
		articles, total, err = s.articleRepo.Search(query.Search, filters)
		if err != nil {
			return nil, 0, utils.WrapError(err, "failed to search articles")
		}
	} else {
		articles, total, err = s.articleRepo.FindAll(filters)
		if err != nil {
			return nil, 0, utils.WrapError(err, "failed to find articles")
		}
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}

// UpdateArticle updates an article
func (s *articleService) UpdateArticle(id string, req *dto.UpdateArticleRequest, userID string, isEditor bool) (*dto.ArticleDetailResponse, error) {
	articleID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	// Check permissions
	if article.AuthorID.String() != userID && !isEditor {
		return nil, utils.ErrForbidden
	}

	// Update fields
	if req.Title != nil {
		article.Title = *req.Title
		// Update slug if title changed
		newSlug := utils.GenerateSlug(*req.Title)
		newSlug = utils.TruncateSlug(newSlug, 200)
		if newSlug != article.Slug {
			exists, err := s.articleRepo.ExistsBySlug(newSlug)
			if err != nil {
				return nil, utils.WrapError(err, "failed to check slug")
			}
			if !exists {
				article.Slug = newSlug
			}
		}
	}
	if req.Content != nil {
		article.Content = *req.Content
	}
	if req.Excerpt != nil {
		article.Excerpt = *req.Excerpt
	}
	if req.FeaturedImageURL != nil {
		article.FeaturedImageURL = req.FeaturedImageURL
	}
	if req.MetaTitle != nil {
		article.MetaTitle = *req.MetaTitle
	}
	if req.MetaDescription != nil {
		article.MetaDescription = *req.MetaDescription
	}
	if req.MetaKeywords != nil {
		article.MetaKeywords = *req.MetaKeywords
	}

	// Update categories
	if len(req.CategoryIDs) > 0 {
		categoryIDs := make([]uuid.UUID, len(req.CategoryIDs))
		for i, id := range req.CategoryIDs {
			catID, err := uuid.Parse(id)
			if err != nil {
				return nil, utils.NewAppError("INVALID_CATEGORY_ID", "Invalid category ID: "+id, 400)
			}
			categoryIDs[i] = catID
		}

		categories, err := s.categoryRepo.FindByIDs(categoryIDs)
		if err != nil {
			return nil, utils.WrapError(err, "failed to find categories")
		}
		if len(categories) != len(categoryIDs) {
			return nil, utils.NewAppError("CATEGORY_NOT_FOUND", "One or more categories not found", 404)
		}

		if err := s.articleRepo.UpdateCategories(article, categories); err != nil {
			return nil, utils.WrapError(err, "failed to update categories")
		}
	}

	// Update tags within a transaction
	if req.TagIDs != nil {
		var newTags []models.Tag
		if len(req.TagIDs) > 0 {
			tagIDs := make([]uuid.UUID, len(req.TagIDs))
			for i, id := range req.TagIDs {
				tagID, err := uuid.Parse(id)
				if err != nil {
					return nil, utils.NewAppError("INVALID_TAG_ID", "Invalid tag ID: "+id, 400)
				}
				tagIDs[i] = tagID
			}

			newTags, err = s.tagRepo.FindByIDs(tagIDs)
			if err != nil {
				return nil, utils.WrapError(err, "failed to find tags")
			}
		}

		oldTags := article.Tags

		// Use transaction for tag updates
		if s.db != nil {
			err = s.db.Transaction(func(tx *gorm.DB) error {
				txArticleRepo := s.articleRepo.WithTx(tx)
				txTagRepo := s.tagRepo.WithTx(tx)

				// Decrement old tag usage
				for _, tag := range oldTags {
					if err := txTagRepo.DecrementUsage(tag.ID); err != nil {
						return err
					}
				}

				// Update article tags
				if err := txArticleRepo.UpdateTags(article, newTags); err != nil {
					return err
				}

				// Increment new tag usage
				for _, tag := range newTags {
					if err := txTagRepo.IncrementUsage(tag.ID); err != nil {
						return err
					}
				}

				return nil
			})
		} else {
			// Fallback for unit tests without db - run without transaction
			for _, tag := range oldTags {
				if err := s.tagRepo.DecrementUsage(tag.ID); err != nil {
					return nil, utils.WrapError(err, "failed to decrement tag usage")
				}
			}
			if err := s.articleRepo.UpdateTags(article, newTags); err != nil {
				return nil, utils.WrapError(err, "failed to update tags")
			}
			for _, tag := range newTags {
				if err := s.tagRepo.IncrementUsage(tag.ID); err != nil {
					return nil, utils.WrapError(err, "failed to increment tag usage")
				}
			}
		}

		if err != nil {
			return nil, utils.WrapError(err, "failed to update tags")
		}
	}

	if err := s.articleRepo.Update(article); err != nil {
		return nil, utils.WrapError(err, "failed to update article")
	}

	// Fetch updated article
	updatedArticle, err := s.articleRepo.FindByID(article.ID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to fetch updated article")
	}

	return s.toDetailResponse(updatedArticle), nil
}

// DeleteArticle deletes an article
func (s *articleService) DeleteArticle(id string, userID string, isEditor bool) error {
	articleID, err := uuid.Parse(id)
	if err != nil {
		return utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to find article")
	}

	// Check permissions
	if article.AuthorID.String() != userID && !isEditor {
		return utils.ErrForbidden
	}

	// Use transaction for deletion and tag updates
	if s.db != nil {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			txArticleRepo := s.articleRepo.WithTx(tx)
			txTagRepo := s.tagRepo.WithTx(tx)

			// Decrement tag usage
			for _, tag := range article.Tags {
				if err := txTagRepo.DecrementUsage(tag.ID); err != nil {
					return err
				}
			}

			if err := txArticleRepo.Delete(articleID); err != nil {
				return err
			}

			return nil
		})
	} else {
		// Fallback for unit tests without db - run without transaction
		for _, tag := range article.Tags {
			if err := s.tagRepo.DecrementUsage(tag.ID); err != nil {
				return utils.WrapError(err, "failed to decrement tag usage")
			}
		}
		if err := s.articleRepo.Delete(articleID); err != nil {
			return utils.WrapError(err, "failed to delete article")
		}
	}

	if err != nil {
		return utils.WrapError(err, "failed to delete article")
	}

	return nil
}

// PublishArticle publishes an article
func (s *articleService) PublishArticle(id string) (*dto.ArticleDetailResponse, error) {
	articleID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	if article.Status == models.StatusPublished {
		return nil, utils.NewAppError("ALREADY_PUBLISHED", "Article is already published", 400)
	}

	article.Publish()

	if err := s.articleRepo.Update(article); err != nil {
		return nil, utils.WrapError(err, "failed to publish article")
	}

	// Create publish notification for the author's followers
	if s.engagementRepo != nil && s.userRepo != nil && article.Author != nil {
		actorName := article.Author.GetFullName()
		message := fmt.Sprintf("%s published a new article", actorName)
		followers, _, _ := s.userRepo.GetFollowers(article.AuthorID, 0, 0)
		for _, follower := range followers {
			if follower.ID != article.AuthorID {
				notification := &models.Notification{
					UserID:    follower.ID,
					ActorID:   article.AuthorID,
					Type:      models.NotificationTypeArticle,
					Message:   message,
					ArticleID: &article.ID,
				}
				_ = s.engagementRepo.CreateNotification(notification)
			}
		}
	}

	return s.toDetailResponse(article), nil
}

// UnpublishArticle unpublishes an article
func (s *articleService) UnpublishArticle(id string) (*dto.ArticleDetailResponse, error) {
	articleID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindByID(articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	if article.Status != models.StatusPublished {
		return nil, utils.NewAppError("NOT_PUBLISHED", "Article is not published", 400)
	}

	article.Unpublish()

	if err := s.articleRepo.Update(article); err != nil {
		return nil, utils.WrapError(err, "failed to unpublish article")
	}

	return s.toDetailResponse(article), nil
}

// GetTrendingArticles retrieves trending articles
func (s *articleService) GetTrendingArticles(limit int) ([]dto.ArticleListItemResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	articles, err := s.articleRepo.FindTrending(limit)
	if err != nil {
		return nil, utils.WrapError(err, "failed to find trending articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, nil
}

// GetRecentArticles retrieves recent articles
func (s *articleService) GetRecentArticles(limit int) ([]dto.ArticleListItemResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	articles, err := s.articleRepo.FindRecent(limit)
	if err != nil {
		return nil, utils.WrapError(err, "failed to find recent articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, nil
}

// GetRelatedArticles retrieves related articles
func (s *articleService) GetRelatedArticles(slug string, limit int) ([]dto.ArticleListItemResponse, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	categoryIDs := article.GetCategoryIDs()
	tagIDs := article.GetTagIDs()

	articles, err := s.articleRepo.FindRelated(article.ID, categoryIDs, tagIDs, limit)
	if err != nil {
		return nil, utils.WrapError(err, "failed to find related articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, nil
}

// SearchArticles searches articles
func (s *articleService) SearchArticles(query string, filters *dto.ArticleListQuery) ([]dto.ArticleListItemResponse, int64, error) {
	articleFilters := repositories.ArticleFilters{
		Status: string(models.StatusPublished),
		Limit:  filters.GetPerPage(),
		Offset: filters.GetOffset(),
		Sort:   filters.GetSort(),
	}

	articles, total, err := s.articleRepo.Search(query, articleFilters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to search articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}

// toDetailResponse converts an article model to a detail response DTO
func (s *articleService) toDetailResponse(article *models.Article) *dto.ArticleDetailResponse {
	response := &dto.ArticleDetailResponse{
		ID:                 article.ID.String(),
		Title:              article.Title,
		Slug:               article.Slug,
		Excerpt:            article.Excerpt,
		Content:            article.Content,
		FeaturedImageURL:   article.FeaturedImageURL,
		Status:             string(article.Status),
		PublishedAt:        article.PublishedAt,
		ViewCount:          article.ViewCount,
		ReadingTimeMinutes: article.ReadingTimeMinutes,
		MetaTitle:          article.MetaTitle,
		MetaDescription:    article.MetaDescription,
		MetaKeywords:       article.MetaKeywords,
		CreatedAt:          article.CreatedAt,
		UpdatedAt:          article.UpdatedAt,
	}

	if article.Author != nil {
		response.Author = dto.PublicUserResponse{
			ID:              article.Author.ID.String(),
			Username:        article.Author.Username,
			FirstName:       article.Author.FirstName,
			LastName:        article.Author.LastName,
			Bio:             article.Author.Bio,
			ProfileImageURL: article.Author.ProfileImageURL,
		}
	}

	for _, cat := range article.Categories {
		response.Categories = append(response.Categories, dto.CategoryResponse{
			ID:   cat.ID.String(),
			Name: cat.Name,
			Slug: cat.Slug,
		})
	}

	for _, tag := range article.Tags {
		response.Tags = append(response.Tags, dto.TagResponse{
			ID:   tag.ID.String(),
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	return response
}
