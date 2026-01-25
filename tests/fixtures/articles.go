package fixtures

import (
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
)

// ValidArticleID is a valid UUID for testing
var ValidArticleID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

// TestArticles contains sample article data for testing
var TestArticles = struct {
	Published models.Article
	Draft     models.Article
	Archived  models.Article
}{
	Published: models.Article{
		ID:                 uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Title:              "Test Published Article",
		Slug:               "test-published-article",
		Excerpt:            "This is a test excerpt for the published article.",
		Content:            "This is the content of the test published article. It contains enough words to calculate a reasonable reading time.",
		AuthorID:           TestUsers.Author.ID,
		Status:             models.StatusPublished,
		PublishedAt:        func() *time.Time { t := time.Now(); return &t }(),
		ViewCount:          100,
		ReadingTimeMinutes: 5,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	},
	Draft: models.Article{
		ID:                 uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Title:              "Test Draft Article",
		Slug:               "test-draft-article",
		Excerpt:            "This is a test excerpt for the draft article.",
		Content:            "This is the content of the test draft article.",
		AuthorID:           TestUsers.Author.ID,
		Status:             models.StatusDraft,
		ViewCount:          0,
		ReadingTimeMinutes: 1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	},
	Archived: models.Article{
		ID:                 uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
		Title:              "Test Archived Article",
		Slug:               "test-archived-article",
		Excerpt:            "This is a test excerpt for the archived article.",
		Content:            "This is the content of the test archived article.",
		AuthorID:           TestUsers.Author.ID,
		Status:             models.StatusArchived,
		ViewCount:          50,
		ReadingTimeMinutes: 2,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	},
}

// NewTestArticle creates a new article with default values that can be overridden
func NewTestArticle(authorID uuid.UUID, opts ...func(*models.Article)) *models.Article {
	article := &models.Article{
		ID:                 uuid.New(),
		Title:              "Test Article " + uuid.New().String()[:8],
		Slug:               "test-article-" + uuid.New().String()[:8],
		Excerpt:            "This is a test excerpt.",
		Content:            "This is the test article content with enough words to make it meaningful for testing purposes.",
		AuthorID:           authorID,
		Status:             models.StatusDraft,
		ViewCount:          0,
		ReadingTimeMinutes: 1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	for _, opt := range opts {
		opt(article)
	}

	return article
}

// WithStatus sets the article status
func WithStatus(status models.ArticleStatus) func(*models.Article) {
	return func(a *models.Article) {
		a.Status = status
		if status == models.StatusPublished {
			now := time.Now()
			a.PublishedAt = &now
		}
	}
}

// WithTitle sets the article title
func WithTitle(title string) func(*models.Article) {
	return func(a *models.Article) {
		a.Title = title
	}
}

// WithSlug sets the article slug
func WithSlug(slug string) func(*models.Article) {
	return func(a *models.Article) {
		a.Slug = slug
	}
}

// WithContent sets the article content
func WithContent(content string) func(*models.Article) {
	return func(a *models.Article) {
		a.Content = content
	}
}

// WithViewCount sets the view count
func WithViewCount(count int) func(*models.Article) {
	return func(a *models.Article) {
		a.ViewCount = count
	}
}
