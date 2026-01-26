package repositories

import (
	"testing"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/tests/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ArticleRepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	repo        ArticleRepository
	userRepo    UserRepository
	categoryRepo CategoryRepository
	tagRepo     TagRepository
	testUser    *models.User
}

func (suite *ArticleRepositoryTestSuite) SetupSuite() {
	suite.db = helpers.SetupTestDB()
	suite.repo = NewArticleRepository(suite.db)
	suite.userRepo = NewUserRepository(suite.db)
	suite.categoryRepo = NewCategoryRepository(suite.db)
	suite.tagRepo = NewTagRepository(suite.db)
}

func (suite *ArticleRepositoryTestSuite) SetupTest() {
	helpers.CleanupTestDB(suite.db)
	// Create a test user for articles
	suite.testUser = &models.User{
		ID:           uuid.New(),
		Username:     "testauthor",
		Email:        "author@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleAuthor,
		IsActive:     true,
	}
	suite.userRepo.Create(suite.testUser)
}

func TestArticleRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ArticleRepositoryTestSuite))
}

// Create Tests

func (suite *ArticleRepositoryTestSuite) TestCreate_Success() {
	article := &models.Article{
		ID:       uuid.New(),
		Title:    "Test Article",
		Slug:     "test-article",
		Content:  "This is the content",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusDraft,
	}

	err := suite.repo.Create(article)

	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), uuid.Nil, article.ID)
}

func (suite *ArticleRepositoryTestSuite) TestCreate_DuplicateSlug() {
	article1 := &models.Article{
		ID:       uuid.New(),
		Title:    "Article 1",
		Slug:     "same-slug",
		Content:  "Content 1",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusDraft,
	}
	article2 := &models.Article{
		ID:       uuid.New(),
		Title:    "Article 2",
		Slug:     "same-slug",
		Content:  "Content 2",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusDraft,
	}

	err := suite.repo.Create(article1)
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(article2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

// FindByID Tests

func (suite *ArticleRepositoryTestSuite) TestFindByID_Success() {
	article := &models.Article{
		ID:       uuid.New(),
		Title:    "Find Me",
		Slug:     "find-me",
		Content:  "Content here",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusPublished,
	}
	suite.repo.Create(article)

	found, err := suite.repo.FindByID(article.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), article.Title, found.Title)
	assert.Equal(suite.T(), article.Slug, found.Slug)
}

func (suite *ArticleRepositoryTestSuite) TestFindByID_NotFound() {
	found, err := suite.repo.FindByID(uuid.New())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

// FindBySlug Tests

func (suite *ArticleRepositoryTestSuite) TestFindBySlug_Success() {
	article := &models.Article{
		ID:       uuid.New(),
		Title:    "Slugged Article",
		Slug:     "slugged-article",
		Content:  "Content here",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusPublished,
	}
	suite.repo.Create(article)

	found, err := suite.repo.FindBySlug("slugged-article")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), article.Title, found.Title)
}

func (suite *ArticleRepositoryTestSuite) TestFindBySlug_NotFound() {
	found, err := suite.repo.FindBySlug("nonexistent-slug")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// FindAll Tests

func (suite *ArticleRepositoryTestSuite) TestFindAll_Success() {
	articles := []*models.Article{
		{
			ID:       uuid.New(),
			Title:    "Article 1",
			Slug:     "article-1",
			Content:  "Content 1",
			AuthorID: suite.testUser.ID,
			Status:   models.StatusPublished,
		},
		{
			ID:       uuid.New(),
			Title:    "Article 2",
			Slug:     "article-2",
			Content:  "Content 2",
			AuthorID: suite.testUser.ID,
			Status:   models.StatusDraft,
		},
	}

	for _, a := range articles {
		suite.repo.Create(a)
	}

	filters := ArticleFilters{Limit: 10}
	result, total, err := suite.repo.FindAll(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
}

func (suite *ArticleRepositoryTestSuite) TestFindAll_WithStatusFilter() {
	articles := []*models.Article{
		{
			ID:       uuid.New(),
			Title:    "Published Article",
			Slug:     "published-article",
			Content:  "Content",
			AuthorID: suite.testUser.ID,
			Status:   models.StatusPublished,
		},
		{
			ID:       uuid.New(),
			Title:    "Draft Article",
			Slug:     "draft-article",
			Content:  "Content",
			AuthorID: suite.testUser.ID,
			Status:   models.StatusDraft,
		},
	}

	for _, a := range articles {
		suite.repo.Create(a)
	}

	filters := ArticleFilters{Status: string(models.StatusPublished), Limit: 10}
	result, total, err := suite.repo.FindAll(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), models.StatusPublished, result[0].Status)
}

func (suite *ArticleRepositoryTestSuite) TestFindAll_WithPagination() {
	for i := 0; i < 5; i++ {
		suite.repo.Create(&models.Article{
			ID:       uuid.New(),
			Title:    "Article " + string(rune('0'+i)),
			Slug:     "article-" + string(rune('0'+i)),
			Content:  "Content",
			AuthorID: suite.testUser.ID,
			Status:   models.StatusPublished,
		})
	}

	filters := ArticleFilters{Limit: 2, Offset: 2}
	result, total, err := suite.repo.FindAll(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(5), total)
}

// FindByAuthor Tests

func (suite *ArticleRepositoryTestSuite) TestFindByAuthor_Success() {
	// Create another user
	otherUser := &models.User{
		ID:           uuid.New(),
		Username:     "otherauthor",
		Email:        "other@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleAuthor,
		IsActive:     true,
	}
	suite.userRepo.Create(otherUser)

	// Create articles for both users
	suite.repo.Create(&models.Article{
		ID:       uuid.New(),
		Title:    "My Article",
		Slug:     "my-article",
		Content:  "Content",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusPublished,
	})
	suite.repo.Create(&models.Article{
		ID:       uuid.New(),
		Title:    "Other Article",
		Slug:     "other-article",
		Content:  "Content",
		AuthorID: otherUser.ID,
		Status:   models.StatusPublished,
	})

	filters := ArticleFilters{Limit: 10}
	result, total, err := suite.repo.FindByAuthor(suite.testUser.ID, filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), "My Article", result[0].Title)
}

// FindTrending Tests

func (suite *ArticleRepositoryTestSuite) TestFindTrending_Success() {
	// Create articles with different view counts
	lowViews := &models.Article{
		ID:        uuid.New(),
		Title:     "Low Views",
		Slug:      "low-views",
		Content:   "Content",
		AuthorID:  suite.testUser.ID,
		Status:    models.StatusPublished,
		ViewCount: 10,
	}
	highViews := &models.Article{
		ID:        uuid.New(),
		Title:     "High Views",
		Slug:      "high-views",
		Content:   "Content",
		AuthorID:  suite.testUser.ID,
		Status:    models.StatusPublished,
		ViewCount: 100,
	}

	suite.repo.Create(lowViews)
	suite.repo.Create(highViews)

	result, err := suite.repo.FindTrending(10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	// High views should be first
	assert.Equal(suite.T(), "High Views", result[0].Title)
}

// FindRecent Tests

func (suite *ArticleRepositoryTestSuite) TestFindRecent_Success() {
	now := time.Now()
	older := now.Add(-24 * time.Hour)
	newer := now.Add(-1 * time.Hour)

	olderArticle := &models.Article{
		ID:          uuid.New(),
		Title:       "Older Article",
		Slug:        "older-article",
		Content:     "Content",
		AuthorID:    suite.testUser.ID,
		Status:      models.StatusPublished,
		PublishedAt: &older,
	}
	newerArticle := &models.Article{
		ID:          uuid.New(),
		Title:       "Newer Article",
		Slug:        "newer-article",
		Content:     "Content",
		AuthorID:    suite.testUser.ID,
		Status:      models.StatusPublished,
		PublishedAt: &newer,
	}

	suite.repo.Create(olderArticle)
	suite.repo.Create(newerArticle)

	result, err := suite.repo.FindRecent(10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	// Newer should be first
	assert.Equal(suite.T(), "Newer Article", result[0].Title)
}

// Update Tests

func (suite *ArticleRepositoryTestSuite) TestUpdate_Success() {
	article := &models.Article{
		ID:       uuid.New(),
		Title:    "Original Title",
		Slug:     "original-slug",
		Content:  "Original content",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusDraft,
	}
	suite.repo.Create(article)

	article.Title = "Updated Title"
	article.Content = "Updated content"
	err := suite.repo.Update(article)

	assert.NoError(suite.T(), err)

	found, _ := suite.repo.FindByID(article.ID)
	assert.Equal(suite.T(), "Updated Title", found.Title)
	assert.Equal(suite.T(), "Updated content", found.Content)
}

// Delete Tests

func (suite *ArticleRepositoryTestSuite) TestDelete_Success() {
	article := &models.Article{
		ID:       uuid.New(),
		Title:    "To Delete",
		Slug:     "to-delete",
		Content:  "Content",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusDraft,
	}
	suite.repo.Create(article)

	err := suite.repo.Delete(article.ID)

	assert.NoError(suite.T(), err)

	// Should not find deleted article
	found, err := suite.repo.FindByID(article.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// ExistsBySlug Tests

func (suite *ArticleRepositoryTestSuite) TestExistsBySlug_Exists() {
	article := &models.Article{
		ID:       uuid.New(),
		Title:    "Existing Article",
		Slug:     "existing-slug",
		Content:  "Content",
		AuthorID: suite.testUser.ID,
		Status:   models.StatusPublished,
	}
	suite.repo.Create(article)

	exists, err := suite.repo.ExistsBySlug("existing-slug")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *ArticleRepositoryTestSuite) TestExistsBySlug_NotExists() {
	exists, err := suite.repo.ExistsBySlug("nonexistent-slug")

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// IncrementViewCount Tests

func (suite *ArticleRepositoryTestSuite) TestIncrementViewCount_Success() {
	article := &models.Article{
		ID:        uuid.New(),
		Title:     "View Count Test",
		Slug:      "view-count-test",
		Content:   "Content",
		AuthorID:  suite.testUser.ID,
		Status:    models.StatusPublished,
		ViewCount: 5,
	}
	suite.repo.Create(article)

	err := suite.repo.IncrementViewCount(article.ID)

	assert.NoError(suite.T(), err)

	found, _ := suite.repo.FindByID(article.ID)
	assert.Equal(suite.T(), 6, found.ViewCount)
}
