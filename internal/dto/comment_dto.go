package dto

import "time"

// CreateCommentRequest represents a comment creation request
type CreateCommentRequest struct {
	Content  string  `json:"content" binding:"required,min=1,max=2000"`
	ParentID *string `json:"parent_id" binding:"omitempty,uuid"`
}

// UpdateCommentRequest represents a comment update request
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// CommentListQuery represents query parameters for listing comments
type CommentListQuery struct {
	PaginationQuery
	IncludeUnapproved bool `form:"include_unapproved"`
}

// CommentResponse represents a comment in API responses
type CommentResponse struct {
	ID         string              `json:"id"`
	ArticleID  string              `json:"article_id"`
	User       PublicUserResponse  `json:"user"`
	ParentID   *string             `json:"parent_id"`
	Content    string              `json:"content"`
	IsApproved bool                `json:"is_approved"`
	Replies    []CommentResponse   `json:"replies,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// CommentListItemResponse represents a comment item in a list
type CommentListItemResponse struct {
	ID         string             `json:"id"`
	ArticleID  string             `json:"article_id"`
	User       PublicUserResponse `json:"user"`
	ParentID   *string            `json:"parent_id"`
	Content    string             `json:"content"`
	IsApproved bool               `json:"is_approved"`
	ReplyCount int                `json:"reply_count"`
	CreatedAt  time.Time          `json:"created_at"`
}
