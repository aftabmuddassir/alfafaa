package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Like represents a user's like on an article
type Like struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	ArticleID uuid.UUID `gorm:"type:uuid;not null;index" json:"article_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// TableName returns the table name for the Like model
func (Like) TableName() string {
	return "likes"
}

// BeforeCreate is a GORM hook that runs before creating a like
func (l *Like) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
