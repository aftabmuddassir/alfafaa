package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag represents a blog tag for categorizing articles
type Tag struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Slug        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	UsageCount  int       `gorm:"default:0;index" json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Articles []Article `gorm:"many2many:article_tags;" json:"articles,omitempty"`
}

// TableName returns the table name for the Tag model
func (Tag) TableName() string {
	return "tags"
}

// BeforeCreate is a GORM hook that runs before creating a tag
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// IncrementUsage increments the usage count of the tag
func (t *Tag) IncrementUsage() {
	t.UsageCount++
}

// DecrementUsage decrements the usage count of the tag
func (t *Tag) DecrementUsage() {
	if t.UsageCount > 0 {
		t.UsageCount--
	}
}
