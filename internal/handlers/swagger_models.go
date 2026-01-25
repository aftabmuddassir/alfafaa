package handlers

import (
	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
)

// Swagger model definitions - these are used to help swag discover types
// DO NOT use these directly in code, they're only for documentation

// swaggerResponseArticleList is for swagger documentation
// @Description Response with article list and pagination
type swaggerResponseArticleList struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Data    []dto.ArticleListResponse `json:"data"`
	Meta    *utils.Meta               `json:"meta"`
}

// swaggerResponseArticleDetail is for swagger documentation
type swaggerResponseArticleDetail struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Data    dto.ArticleDetailResponse `json:"data"`
}

// swaggerResponseUserList is for swagger documentation
type swaggerResponseUserList struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    []dto.UserResponse `json:"data"`
	Meta    *utils.Meta        `json:"meta"`
}

// swaggerResponseUser is for swagger documentation
type swaggerResponseUser struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    dto.UserResponse `json:"data"`
}

// swaggerResponsePublicUser is for swagger documentation
type swaggerResponsePublicUser struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    dto.PublicUserResponse `json:"data"`
}

// swaggerResponseCategoryList is for swagger documentation
type swaggerResponseCategoryList struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    []dto.CategoryResponse `json:"data"`
}

// swaggerResponseCategory is for swagger documentation
type swaggerResponseCategory struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    dto.CategoryResponse `json:"data"`
}

// swaggerResponseTagList is for swagger documentation
type swaggerResponseTagList struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    []dto.TagResponse `json:"data"`
	Meta    *utils.Meta       `json:"meta"`
}

// swaggerResponseTag is for swagger documentation
type swaggerResponseTag struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    dto.TagResponse `json:"data"`
}

// swaggerResponseMediaList is for swagger documentation
type swaggerResponseMediaList struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    []dto.MediaResponse `json:"data"`
	Meta    *utils.Meta         `json:"meta"`
}

// swaggerResponseMedia is for swagger documentation
type swaggerResponseMedia struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    dto.MediaResponse `json:"data"`
}

// swaggerResponseAuth is for swagger documentation
type swaggerResponseAuth struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    dto.AuthResponse `json:"data"`
}

// swaggerResponseToken is for swagger documentation
type swaggerResponseToken struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    dto.TokenResponse `json:"data"`
}

// swaggerResponseSearch is for swagger documentation
type swaggerResponseSearch struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    dto.SearchResponse `json:"data"`
}

// swaggerResponseSuccess is for swagger documentation
type swaggerResponseSuccess struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
