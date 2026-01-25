package dto

import "time"

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	FirstName       *string `json:"first_name" binding:"omitempty,max=100"`
	LastName        *string `json:"last_name" binding:"omitempty,max=100"`
	Bio             *string `json:"bio" binding:"omitempty,max=1000"`
	ProfileImageURL *string `json:"profile_image_url" binding:"omitempty,url"`
}

// AdminUpdateUserRequest represents an admin user update request
type AdminUpdateUserRequest struct {
	FirstName       *string `json:"first_name" binding:"omitempty,max=100"`
	LastName        *string `json:"last_name" binding:"omitempty,max=100"`
	Bio             *string `json:"bio" binding:"omitempty,max=1000"`
	ProfileImageURL *string `json:"profile_image_url" binding:"omitempty,url"`
	Role            *string `json:"role" binding:"omitempty,oneof=reader author editor admin"`
	IsActive        *bool   `json:"is_active"`
	IsVerified      *bool   `json:"is_verified"`
}

// UserListQuery represents query parameters for listing users
type UserListQuery struct {
	PaginationQuery
	Role     string `form:"role" binding:"omitempty,oneof=reader author editor admin"`
	IsActive *bool  `form:"is_active"`
	Search   string `form:"search" binding:"omitempty,max=100"`
}

// UserDetailResponse represents detailed user information
type UserDetailResponse struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Bio             string    `json:"bio"`
	ProfileImageURL *string   `json:"profile_image_url"`
	Role            string    `json:"role"`
	IsVerified      bool      `json:"is_verified"`
	IsActive        bool      `json:"is_active"`
	ArticleCount    int       `json:"article_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserListItemResponse represents a user item in a list
type UserListItemResponse struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	ProfileImageURL *string   `json:"profile_image_url"`
	Role            string    `json:"role"`
	IsVerified      bool      `json:"is_verified"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}
