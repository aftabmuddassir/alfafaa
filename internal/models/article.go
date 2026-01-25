package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArticleStatus represents the publication status of an article
type ArticleStatus string

const (
	StatusDraft     ArticleStatus = "draft"
	StatusPublished ArticleStatus = "published"
	StatusArchived  ArticleStatus = "archived"
)

// IsValid checks if the status is valid
func (s ArticleStatus) IsValid() bool {
	switch s {
	case StatusDraft, StatusPublished, StatusArchived:
		return true
	}
	return false
}

// Article represents a blog article/post
type Article struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title              string         `gorm:"type:varchar(255);not null;index" json:"title"`
	Slug               string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Excerpt            string         `gorm:"type:text" json:"excerpt"`
	Content            string         `gorm:"type:text;not null" json:"content"`
	FeaturedImageURL   *string        `gorm:"type:varchar(500)" json:"featured_image_url"`
	AuthorID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"author_id"`
	Status             ArticleStatus  `gorm:"type:varchar(20);not null;default:'draft';index" json:"status"`
	PublishedAt        *time.Time     `gorm:"index" json:"published_at"`
	ViewCount          int            `gorm:"default:0" json:"view_count"`
	ReadingTimeMinutes int            `gorm:"default:1" json:"reading_time_minutes"`
	MetaTitle          string         `gorm:"type:varchar(70)" json:"meta_title"`
	MetaDescription    string         `gorm:"type:varchar(160)" json:"meta_description"`
	MetaKeywords       string         `gorm:"type:varchar(255)" json:"meta_keywords"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Author     *User      `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Categories []Category `gorm:"many2many:article_categories;" json:"categories,omitempty"`
	Tags       []Tag      `gorm:"many2many:article_tags;" json:"tags,omitempty"`
	Comments   []Comment  `gorm:"foreignKey:ArticleID" json:"comments,omitempty"`
}

// TableName returns the table name for the Article model
func (Article) TableName() string {
	return "articles"
}

// BeforeCreate is a GORM hook that runs before creating an article
func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// BeforeSave is a GORM hook that runs before saving an article
func (a *Article) BeforeSave(tx *gorm.DB) error {
	// Calculate reading time based on content length
	a.ReadingTimeMinutes = a.calculateReadingTime()
	return nil
}

// calculateReadingTime calculates the reading time in minutes
// Based on average reading speed of 200 words per minute
func (a *Article) calculateReadingTime() int {
	words := len(a.Content) / 5 // Rough estimate: average word length is 5 characters
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

// IsPublished checks if the article is published
func (a *Article) IsPublished() bool {
	return a.Status == StatusPublished && a.PublishedAt != nil
}

// IsDraft checks if the article is a draft
func (a *Article) IsDraft() bool {
	return a.Status == StatusDraft
}

// IsArchived checks if the article is archived
func (a *Article) IsArchived() bool {
	return a.Status == StatusArchived
}

// Publish publishes the article
func (a *Article) Publish() {
	a.Status = StatusPublished
	now := time.Now()
	a.PublishedAt = &now
}

// Unpublish unpublishes the article (moves to draft)
func (a *Article) Unpublish() {
	a.Status = StatusDraft
}

// Archive archives the article
func (a *Article) Archive() {
	a.Status = StatusArchived
}

// IncrementViewCount increments the view count
func (a *Article) IncrementViewCount() {
	a.ViewCount++
}

// GetCategoryIDs returns a slice of category IDs
func (a *Article) GetCategoryIDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(a.Categories))
	for i, cat := range a.Categories {
		ids[i] = cat.ID
	}
	return ids
}

// GetTagIDs returns a slice of tag IDs
func (a *Article) GetTagIDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(a.Tags))
	for i, tag := range a.Tags {
		ids[i] = tag.ID
	}
	return ids
}

// HasCategory checks if the article has a specific category
func (a *Article) HasCategory(categoryID uuid.UUID) bool {
	for _, cat := range a.Categories {
		if cat.ID == categoryID {
			return true
		}
	}
	return false
}

// HasTag checks if the article has a specific tag
func (a *Article) HasTag(tagID uuid.UUID) bool {
	for _, tag := range a.Tags {
		if tag.ID == tagID {
			return true
		}
	}
	return false
}
