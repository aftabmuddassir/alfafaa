package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeLike    NotificationType = "like"
	NotificationTypeComment NotificationType = "comment"
	NotificationTypeFollow  NotificationType = "follow"
	NotificationTypeArticle NotificationType = "article"
)

// Notification represents a notification to a user
type Notification struct {
	ID        uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;index" json:"user_id"`
	ActorID   uuid.UUID        `gorm:"type:uuid;not null;index" json:"actor_id"`
	Type      NotificationType `gorm:"type:varchar(20);not null;index" json:"type"`
	Message   string           `gorm:"type:text;not null" json:"message"`
	ArticleID *uuid.UUID       `gorm:"type:uuid;index" json:"article_id"`
	Read      bool             `gorm:"default:false;index" json:"read"`
	CreatedAt time.Time        `json:"created_at"`

	// Relationships
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Actor   *User    `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
	Article *Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
}

// TableName returns the table name for the Notification model
func (Notification) TableName() string {
	return "notifications"
}

// BeforeCreate is a GORM hook that runs before creating a notification
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
