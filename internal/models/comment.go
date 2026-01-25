package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Comment represents a comment on an article
type Comment struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ArticleID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"article_id"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	ParentID   *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	IsApproved bool           `gorm:"default:false;index" json:"is_approved"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Article  *Article  `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent   *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies  []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName returns the table name for the Comment model
func (Comment) TableName() string {
	return "comments"
}

// BeforeCreate is a GORM hook that runs before creating a comment
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsReply checks if the comment is a reply to another comment
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}

// HasReplies checks if the comment has any replies
func (c *Comment) HasReplies() bool {
	return len(c.Replies) > 0
}

// Approve approves the comment
func (c *Comment) Approve() {
	c.IsApproved = true
}

// Unapprove unapproves the comment
func (c *Comment) Unapprove() {
	c.IsApproved = false
}

// GetDepth returns the depth of the comment in the reply chain
func (c *Comment) GetDepth() int {
	if c.Parent == nil {
		return 0
	}
	return c.Parent.GetDepth() + 1
}
