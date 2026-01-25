package dto

import "time"

// CreateCategoryRequest represents a category creation request
type CreateCategoryRequest struct {
	Name         string  `json:"name" binding:"required,min=2,max=100"`
	Description  string  `json:"description" binding:"omitempty,max=500"`
	ParentID     *string `json:"parent_id" binding:"omitempty,uuid"`
	DisplayOrder int     `json:"display_order" binding:"omitempty,min=0"`
	IsActive     *bool   `json:"is_active"`
}

// UpdateCategoryRequest represents a category update request
type UpdateCategoryRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description  *string `json:"description" binding:"omitempty,max=500"`
	ParentID     *string `json:"parent_id" binding:"omitempty,uuid"`
	DisplayOrder *int    `json:"display_order" binding:"omitempty,min=0"`
	IsActive     *bool   `json:"is_active"`
}

// CategoryListQuery represents query parameters for listing categories
type CategoryListQuery struct {
	IncludeInactive bool   `form:"include_inactive"`
	ParentOnly      bool   `form:"parent_only"`
	Search          string `form:"search" binding:"omitempty,max=100"`
}

// CategoryResponse represents a category in API responses
type CategoryResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Description  string     `json:"description"`
	ParentID     *string    `json:"parent_id"`
	DisplayOrder int        `json:"display_order"`
	IsActive     bool       `json:"is_active"`
	ArticleCount int        `json:"article_count,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CategoryTreeResponse represents a category with its children
type CategoryTreeResponse struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Slug         string                  `json:"slug"`
	Description  string                  `json:"description"`
	DisplayOrder int                     `json:"display_order"`
	IsActive     bool                    `json:"is_active"`
	ArticleCount int                     `json:"article_count,omitempty"`
	Children     []*CategoryTreeResponse `json:"children,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
}

// CategoryDetailResponse represents detailed category information
type CategoryDetailResponse struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description"`
	ParentID     *string            `json:"parent_id"`
	Parent       *CategoryResponse  `json:"parent,omitempty"`
	DisplayOrder int                `json:"display_order"`
	IsActive     bool               `json:"is_active"`
	ArticleCount int                `json:"article_count"`
	Children     []CategoryResponse `json:"children,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}
