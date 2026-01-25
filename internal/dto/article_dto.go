package dto

import "time"

// CreateArticleRequest represents an article creation request
type CreateArticleRequest struct {
	Title            string   `json:"title" binding:"required,min=5,max=255"`
	Content          string   `json:"content" binding:"required,min=50"`
	Excerpt          string   `json:"excerpt" binding:"omitempty,max=500"`
	FeaturedImageURL *string  `json:"featured_image_url" binding:"omitempty,url"`
	CategoryIDs      []string `json:"category_ids" binding:"required,min=1,dive,uuid"`
	TagIDs           []string `json:"tag_ids" binding:"omitempty,dive,uuid"`
	Status           string   `json:"status" binding:"omitempty,oneof=draft published"`
	MetaTitle        string   `json:"meta_title" binding:"omitempty,max=70"`
	MetaDescription  string   `json:"meta_description" binding:"omitempty,max=160"`
	MetaKeywords     string   `json:"meta_keywords" binding:"omitempty,max=255"`
}

// UpdateArticleRequest represents an article update request
type UpdateArticleRequest struct {
	Title            *string  `json:"title" binding:"omitempty,min=5,max=255"`
	Content          *string  `json:"content" binding:"omitempty,min=50"`
	Excerpt          *string  `json:"excerpt" binding:"omitempty,max=500"`
	FeaturedImageURL *string  `json:"featured_image_url" binding:"omitempty,url"`
	CategoryIDs      []string `json:"category_ids" binding:"omitempty,min=1,dive,uuid"`
	TagIDs           []string `json:"tag_ids" binding:"omitempty,dive,uuid"`
	MetaTitle        *string  `json:"meta_title" binding:"omitempty,max=70"`
	MetaDescription  *string  `json:"meta_description" binding:"omitempty,max=160"`
	MetaKeywords     *string  `json:"meta_keywords" binding:"omitempty,max=255"`
}

// ArticleListQuery represents query parameters for listing articles
type ArticleListQuery struct {
	PaginationQuery
	CategorySlug string `form:"category" binding:"omitempty"`
	TagSlug      string `form:"tag" binding:"omitempty"`
	AuthorID     string `form:"author_id" binding:"omitempty,uuid"`
	Status       string `form:"status" binding:"omitempty,oneof=draft published archived"`
	Search       string `form:"search" binding:"omitempty,max=100"`
	FromDate     string `form:"from_date" binding:"omitempty"`
	ToDate       string `form:"to_date" binding:"omitempty"`
}

// ArticleResponse represents an article in API responses
type ArticleResponse struct {
	ID                 string               `json:"id"`
	Title              string               `json:"title"`
	Slug               string               `json:"slug"`
	Excerpt            string               `json:"excerpt"`
	FeaturedImageURL   *string              `json:"featured_image_url"`
	Author             PublicUserResponse   `json:"author"`
	Status             string               `json:"status"`
	PublishedAt        *time.Time           `json:"published_at"`
	ViewCount          int                  `json:"view_count"`
	ReadingTimeMinutes int                  `json:"reading_time_minutes"`
	Categories         []CategoryResponse   `json:"categories"`
	Tags               []TagResponse        `json:"tags"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// ArticleDetailResponse represents detailed article information
type ArticleDetailResponse struct {
	ID                 string             `json:"id"`
	Title              string             `json:"title"`
	Slug               string             `json:"slug"`
	Excerpt            string             `json:"excerpt"`
	Content            string             `json:"content"`
	FeaturedImageURL   *string            `json:"featured_image_url"`
	Author             PublicUserResponse `json:"author"`
	Status             string             `json:"status"`
	PublishedAt        *time.Time         `json:"published_at"`
	ViewCount          int                `json:"view_count"`
	ReadingTimeMinutes int                `json:"reading_time_minutes"`
	MetaTitle          string             `json:"meta_title"`
	MetaDescription    string             `json:"meta_description"`
	MetaKeywords       string             `json:"meta_keywords"`
	Categories         []CategoryResponse `json:"categories"`
	Tags               []TagResponse      `json:"tags"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// ArticleListItemResponse represents an article item in a list
type ArticleListItemResponse struct {
	ID                 string             `json:"id"`
	Title              string             `json:"title"`
	Slug               string             `json:"slug"`
	Excerpt            string             `json:"excerpt"`
	FeaturedImageURL   *string            `json:"featured_image_url"`
	Author             PublicUserResponse `json:"author"`
	Status             string             `json:"status"`
	PublishedAt        *time.Time         `json:"published_at"`
	ViewCount          int                `json:"view_count"`
	ReadingTimeMinutes int                `json:"reading_time_minutes"`
	Categories         []CategoryResponse `json:"categories"`
	Tags               []TagResponse      `json:"tags"`
	CreatedAt          time.Time          `json:"created_at"`
}

// ArticleListResponse is an alias for ArticleListItemResponse (used in swagger docs)
type ArticleListResponse = ArticleListItemResponse
