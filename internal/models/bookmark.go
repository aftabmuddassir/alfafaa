package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Bookmark represents a user's bookmark on an article
type Bookmark struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ArticleID uuid.UUID `gorm:"type:uuid;not null;index" json:"article_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// TableName returns the table name for the Bookmark model
func (Bookmark) TableName() string {
	return "bookmarks"
}

// BeforeCreate is a GORM hook that runs before creating a bookmark
func (b *Bookmark) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
