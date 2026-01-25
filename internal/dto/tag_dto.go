package dto

import "time"

// CreateTagRequest represents a tag creation request
type CreateTagRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"omitempty,max=500"`
}

// UpdateTagRequest represents a tag update request
type UpdateTagRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=50"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

// TagListQuery represents query parameters for listing tags
type TagListQuery struct {
	PaginationQuery
	Search string `form:"search" binding:"omitempty,max=100"`
}

// TagResponse represents a tag in API responses
type TagResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TagDetailResponse represents detailed tag information
type TagDetailResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PopularTagResponse represents a popular tag
type PopularTagResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	UsageCount int    `json:"usage_count"`
}
