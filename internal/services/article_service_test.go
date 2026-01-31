package services

import (
	"testing"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/alfafaa/alfafaa-blog/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ArticleServiceTestSuite struct {
	suite.Suite
	articleRepo  *mocks.MockArticleRepository
	categoryRepo *mocks.MockCategoryRepository
	tagRepo      *mocks.MockTagRepository
	service      ArticleService
}

func (suite *ArticleServiceTestSuite) SetupTest() {
	suite.articleRepo = new(mocks.MockArticleRepository)
	suite.categoryRepo = new(mocks.MockCategoryRepository)
	suite.tagRepo = new(mocks.MockTagRepository)
	// Pass nil for db in unit tests - service handles nil db gracefully
	// Transaction behavior is tested in integration tests
	suite.service = NewArticleService(nil, suite.articleRepo, suite.categoryRepo, suite.tagRepo)
}

func TestArticleServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ArticleServiceTestSuite))
}

// CreateArticle Tests

func (suite *ArticleServiceTestSuite) TestCreateArticle_Success() {
	authorID := uuid.New()
	categoryID := uuid.New()
	tagID := uuid.New()

	categories := []models.Category{
		{ID: categoryID, Name: "Test Category", Slug: "test-category"},
	}
	tags := []models.Tag{
		{ID: tagID, Name: "Test Tag", Slug: "test-tag"},
	}

	req := &dto.CreateArticleRequest{
		Title:       "Test Article Title",
		Content:     "This is the test article content with enough words to be meaningful.",
		Excerpt:     "Test excerpt",
		CategoryIDs: []string{categoryID.String()},
		TagIDs:      []string{tagID.String()},
	}

	suite.articleRepo.On("ExistsBySlug", "test-article-title").Return(false, nil)
	suite.categoryRepo.On("FindByIDs", []uuid.UUID{categoryID}).Return(categories, nil)
	suite.tagRepo.On("FindByIDs", []uuid.UUID{tagID}).Return(tags, nil)
	suite.articleRepo.On("Create", mock.AnythingOfType("*models.Article")).Return(nil)
	suite.tagRepo.On("IncrementUsage", tagID).Return(nil)
	suite.articleRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(&models.Article{
		ID:         uuid.New(),
		Title:      req.Title,
		Slug:       "test-article-title",
		Content:    req.Content,
		AuthorID:   authorID,
		Status:     models.StatusDraft,
		Categories: categories,
		Tags:       tags,
	}, nil)

	result, err := suite.service.CreateArticle(req, authorID.String())

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), req.Title, result.Title)
	suite.articleRepo.AssertExpectations(suite.T())
	suite.categoryRepo.AssertExpectations(suite.T())
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestCreateArticle_InvalidAuthorID() {
	req := &dto.CreateArticleRequest{
		Title:   "Test Article",
		Content: "Content",
	}

	result, err := suite.service.CreateArticle(req, "invalid-uuid")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *ArticleServiceTestSuite) TestCreateArticle_DuplicateSlug() {
	authorID := uuid.New()
	categoryID := uuid.New()

	categories := []models.Category{
		{ID: categoryID, Name: "Test Category", Slug: "test-category"},
	}

	req := &dto.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "Content",
		CategoryIDs: []string{categoryID.String()},
	}

	// First slug exists, second doesn't
	suite.articleRepo.On("ExistsBySlug", "test-article").Return(true, nil)
	suite.articleRepo.On("ExistsBySlug", "test-article-1").Return(false, nil)
	suite.categoryRepo.On("FindByIDs", []uuid.UUID{categoryID}).Return(categories, nil)
	suite.tagRepo.On("FindByIDs", []uuid.UUID{}).Return([]models.Tag{}, nil)
	suite.articleRepo.On("Create", mock.AnythingOfType("*models.Article")).Return(nil)
	suite.articleRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(&models.Article{
		ID:       uuid.New(),
		Title:    req.Title,
		Slug:     "test-article-1",
		AuthorID: authorID,
		Status:   models.StatusDraft,
	}, nil)

	result, err := suite.service.CreateArticle(req, authorID.String())

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestCreateArticle_InvalidCategoryID() {
	authorID := uuid.New()

	req := &dto.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "Content",
		CategoryIDs: []string{"invalid-uuid"},
	}

	suite.articleRepo.On("ExistsBySlug", "test-article").Return(false, nil)

	result, err := suite.service.CreateArticle(req, authorID.String())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "INVALID_CATEGORY_ID", appErr.Code)
}

func (suite *ArticleServiceTestSuite) TestCreateArticle_CategoryNotFound() {
	authorID := uuid.New()
	categoryID := uuid.New()

	req := &dto.CreateArticleRequest{
		Title:       "Test Article",
		Content:     "Content",
		CategoryIDs: []string{categoryID.String()},
	}

	suite.articleRepo.On("ExistsBySlug", "test-article").Return(false, nil)
	suite.categoryRepo.On("FindByIDs", []uuid.UUID{categoryID}).Return([]models.Category{}, nil)

	result, err := suite.service.CreateArticle(req, authorID.String())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "CATEGORY_NOT_FOUND", appErr.Code)
}

// GetArticle Tests

func (suite *ArticleServiceTestSuite) TestGetArticle_Success() {
	articleID := uuid.New()
	article := &models.Article{
		ID:        articleID,
		Title:     "Test Article",
		Slug:      "test-article",
		Content:   "Content",
		Status:    models.StatusPublished,
		ViewCount: 10,
	}

	suite.articleRepo.On("FindBySlug", "test-article").Return(article, nil)

	result, err := suite.service.GetArticle("test-article", false)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), article.Title, result.Title)
	assert.Equal(suite.T(), 10, result.ViewCount)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetArticle_WithViewIncrement() {
	articleID := uuid.New()
	article := &models.Article{
		ID:        articleID,
		Title:     "Test Article",
		Slug:      "test-article",
		Content:   "Content",
		Status:    models.StatusPublished,
		ViewCount: 10,
	}

	suite.articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	suite.articleRepo.On("IncrementViewCount", articleID).Return(nil)

	result, err := suite.service.GetArticle("test-article", true)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), 11, result.ViewCount) // Incremented locally
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetArticle_NotFound() {
	suite.articleRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetArticle("nonexistent", false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.articleRepo.AssertExpectations(suite.T())
}

// GetArticles Tests

func (suite *ArticleServiceTestSuite) TestGetArticles_Success() {
	articles := []models.Article{
		{ID: uuid.New(), Title: "Article 1", Slug: "article-1", Status: models.StatusPublished},
		{ID: uuid.New(), Title: "Article 2", Slug: "article-2", Status: models.StatusPublished},
	}

	query := &dto.ArticleListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
	}

	suite.articleRepo.On("FindAll", mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(2), nil)

	result, total, err := suite.service.GetArticles(query, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetArticles_WithCategoryFilter() {
	categoryID := uuid.New()
	category := &models.Category{ID: categoryID, Name: "News", Slug: "news"}
	articles := []models.Article{
		{ID: uuid.New(), Title: "News Article", Slug: "news-article", Status: models.StatusPublished},
	}

	query := &dto.ArticleListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		CategorySlug:    "news",
	}

	suite.categoryRepo.On("FindBySlug", "news").Return(category, nil)
	suite.articleRepo.On("FindByCategory", categoryID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(1), nil)

	result, total, err := suite.service.GetArticles(query, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	suite.categoryRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetArticles_CategoryNotFound() {
	query := &dto.ArticleListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		CategorySlug:    "nonexistent",
	}

	suite.categoryRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, total, err := suite.service.GetArticles(query, false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), int64(0), total)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "CATEGORY_NOT_FOUND", appErr.Code)
}

func (suite *ArticleServiceTestSuite) TestGetArticles_WithTagFilter() {
	tagID := uuid.New()
	tag := &models.Tag{ID: tagID, Name: "Ramadan", Slug: "ramadan"}
	articles := []models.Article{
		{ID: uuid.New(), Title: "Ramadan Article", Slug: "ramadan-article", Status: models.StatusPublished},
	}

	query := &dto.ArticleListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		TagSlug:         "ramadan",
	}

	suite.tagRepo.On("FindBySlug", "ramadan").Return(tag, nil)
	suite.articleRepo.On("FindByTag", tagID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(1), nil)

	result, total, err := suite.service.GetArticles(query, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	suite.tagRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetArticles_WithSearch() {
	articles := []models.Article{
		{ID: uuid.New(), Title: "Search Result", Slug: "search-result", Status: models.StatusPublished},
	}

	query := &dto.ArticleListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		Search:          "search",
	}

	suite.articleRepo.On("Search", "search", mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(1), nil)

	result, total, err := suite.service.GetArticles(query, false)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	suite.articleRepo.AssertExpectations(suite.T())
}

// UpdateArticle Tests

func (suite *ArticleServiceTestSuite) TestUpdateArticle_Success_Owner() {
	articleID := uuid.New()
	authorID := uuid.New()
	article := &models.Article{
		ID:       articleID,
		Title:    "Original Title",
		Slug:     "original-title",
		Content:  "Original content",
		AuthorID: authorID,
		Status:   models.StatusDraft,
	}

	newTitle := "Updated Title"
	req := &dto.UpdateArticleRequest{
		Title: &newTitle,
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil).Once()
	suite.articleRepo.On("ExistsBySlug", "updated-title").Return(false, nil)
	suite.articleRepo.On("Update", mock.AnythingOfType("*models.Article")).Return(nil)
	suite.articleRepo.On("FindByID", articleID).Return(&models.Article{
		ID:       articleID,
		Title:    newTitle,
		Slug:     "updated-title",
		AuthorID: authorID,
	}, nil).Once()

	result, err := suite.service.UpdateArticle(articleID.String(), req, authorID.String(), false)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), newTitle, result.Title)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestUpdateArticle_Forbidden_NotOwner() {
	articleID := uuid.New()
	authorID := uuid.New()
	otherUserID := uuid.New()
	article := &models.Article{
		ID:       articleID,
		Title:    "Title",
		AuthorID: authorID,
	}

	req := &dto.UpdateArticleRequest{}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil)

	result, err := suite.service.UpdateArticle(articleID.String(), req, otherUserID.String(), false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrForbidden, err)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestUpdateArticle_Success_Editor() {
	articleID := uuid.New()
	authorID := uuid.New()
	editorID := uuid.New()
	article := &models.Article{
		ID:       articleID,
		Title:    "Title",
		AuthorID: authorID,
		Status:   models.StatusDraft,
	}

	newContent := "Updated content"
	req := &dto.UpdateArticleRequest{
		Content: &newContent,
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil).Once()
	suite.articleRepo.On("Update", mock.AnythingOfType("*models.Article")).Return(nil)
	suite.articleRepo.On("FindByID", articleID).Return(&models.Article{
		ID:       articleID,
		Title:    article.Title,
		Content:  newContent,
		AuthorID: authorID,
	}, nil).Once()

	result, err := suite.service.UpdateArticle(articleID.String(), req, editorID.String(), true)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestUpdateArticle_NotFound() {
	articleID := uuid.New()

	req := &dto.UpdateArticleRequest{}

	suite.articleRepo.On("FindByID", articleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateArticle(articleID.String(), req, uuid.New().String(), false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.articleRepo.AssertExpectations(suite.T())
}

// DeleteArticle Tests

func (suite *ArticleServiceTestSuite) TestDeleteArticle_Success_Owner() {
	articleID := uuid.New()
	authorID := uuid.New()
	article := &models.Article{
		ID:       articleID,
		AuthorID: authorID,
		Tags:     []models.Tag{{ID: uuid.New()}},
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil)
	suite.tagRepo.On("DecrementUsage", mock.AnythingOfType("uuid.UUID")).Return(nil)
	suite.articleRepo.On("Delete", articleID).Return(nil)

	err := suite.service.DeleteArticle(articleID.String(), authorID.String(), false)

	assert.NoError(suite.T(), err)
	suite.articleRepo.AssertExpectations(suite.T())
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestDeleteArticle_Forbidden() {
	articleID := uuid.New()
	authorID := uuid.New()
	otherUserID := uuid.New()
	article := &models.Article{
		ID:       articleID,
		AuthorID: authorID,
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil)

	err := suite.service.DeleteArticle(articleID.String(), otherUserID.String(), false)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrForbidden, err)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestDeleteArticle_NotFound() {
	articleID := uuid.New()

	suite.articleRepo.On("FindByID", articleID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.DeleteArticle(articleID.String(), uuid.New().String(), false)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.articleRepo.AssertExpectations(suite.T())
}

// PublishArticle Tests

func (suite *ArticleServiceTestSuite) TestPublishArticle_Success() {
	articleID := uuid.New()
	article := &models.Article{
		ID:     articleID,
		Title:  "Draft Article",
		Status: models.StatusDraft,
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil)
	suite.articleRepo.On("Update", mock.AnythingOfType("*models.Article")).Return(nil)

	result, err := suite.service.PublishArticle(articleID.String())

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), string(models.StatusPublished), result.Status)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestPublishArticle_AlreadyPublished() {
	articleID := uuid.New()
	now := time.Now()
	article := &models.Article{
		ID:          articleID,
		Title:       "Published Article",
		Status:      models.StatusPublished,
		PublishedAt: &now,
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil)

	result, err := suite.service.PublishArticle(articleID.String())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "ALREADY_PUBLISHED", appErr.Code)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestPublishArticle_NotFound() {
	articleID := uuid.New()

	suite.articleRepo.On("FindByID", articleID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.PublishArticle(articleID.String())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
}

// UnpublishArticle Tests

func (suite *ArticleServiceTestSuite) TestUnpublishArticle_Success() {
	articleID := uuid.New()
	now := time.Now()
	article := &models.Article{
		ID:          articleID,
		Title:       "Published Article",
		Status:      models.StatusPublished,
		PublishedAt: &now,
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil)
	suite.articleRepo.On("Update", mock.AnythingOfType("*models.Article")).Return(nil)

	result, err := suite.service.UnpublishArticle(articleID.String())

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), string(models.StatusDraft), result.Status)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestUnpublishArticle_NotPublished() {
	articleID := uuid.New()
	article := &models.Article{
		ID:     articleID,
		Title:  "Draft Article",
		Status: models.StatusDraft,
	}

	suite.articleRepo.On("FindByID", articleID).Return(article, nil)

	result, err := suite.service.UnpublishArticle(articleID.String())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "NOT_PUBLISHED", appErr.Code)
	suite.articleRepo.AssertExpectations(suite.T())
}

// GetTrendingArticles Tests

func (suite *ArticleServiceTestSuite) TestGetTrendingArticles_Success() {
	articles := []models.Article{
		{ID: uuid.New(), Title: "Trending 1", ViewCount: 1000},
		{ID: uuid.New(), Title: "Trending 2", ViewCount: 500},
	}

	suite.articleRepo.On("FindTrending", 10).Return(articles, nil)

	result, err := suite.service.GetTrendingArticles(10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetTrendingArticles_DefaultLimit() {
	articles := []models.Article{}

	suite.articleRepo.On("FindTrending", 10).Return(articles, nil)

	result, err := suite.service.GetTrendingArticles(0)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetTrendingArticles_MaxLimit() {
	articles := []models.Article{}

	suite.articleRepo.On("FindTrending", 50).Return(articles, nil)

	result, err := suite.service.GetTrendingArticles(100) // Should be capped at 50

	assert.NoError(suite.T(), err)
	suite.articleRepo.AssertExpectations(suite.T())
	_ = result
}

// GetRecentArticles Tests

func (suite *ArticleServiceTestSuite) TestGetRecentArticles_Success() {
	articles := []models.Article{
		{ID: uuid.New(), Title: "Recent 1"},
		{ID: uuid.New(), Title: "Recent 2"},
	}

	suite.articleRepo.On("FindRecent", 10).Return(articles, nil)

	result, err := suite.service.GetRecentArticles(10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	suite.articleRepo.AssertExpectations(suite.T())
}

// GetRelatedArticles Tests

func (suite *ArticleServiceTestSuite) TestGetRelatedArticles_Success() {
	articleID := uuid.New()
	categoryID := uuid.New()
	tagID := uuid.New()

	article := &models.Article{
		ID:         articleID,
		Title:      "Main Article",
		Slug:       "main-article",
		Categories: []models.Category{{ID: categoryID}},
		Tags:       []models.Tag{{ID: tagID}},
	}

	relatedArticles := []models.Article{
		{ID: uuid.New(), Title: "Related 1"},
		{ID: uuid.New(), Title: "Related 2"},
	}

	suite.articleRepo.On("FindBySlug", "main-article").Return(article, nil)
	suite.articleRepo.On("FindRelated", articleID, []uuid.UUID{categoryID}, []uuid.UUID{tagID}, 5).
		Return(relatedArticles, nil)

	result, err := suite.service.GetRelatedArticles("main-article", 5)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *ArticleServiceTestSuite) TestGetRelatedArticles_ArticleNotFound() {
	suite.articleRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetRelatedArticles("nonexistent", 5)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.articleRepo.AssertExpectations(suite.T())
}

// SearchArticles Tests

func (suite *ArticleServiceTestSuite) TestSearchArticles_Success() {
	articles := []models.Article{
		{ID: uuid.New(), Title: "Search Result 1", Status: models.StatusPublished},
		{ID: uuid.New(), Title: "Search Result 2", Status: models.StatusPublished},
	}

	filters := &dto.ArticleListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
	}

	suite.articleRepo.On("Search", "test query", mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(2), nil)

	result, total, err := suite.service.SearchArticles("test query", filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
	suite.articleRepo.AssertExpectations(suite.T())
}
