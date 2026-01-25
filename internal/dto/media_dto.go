package dto

import "time"

// MediaListQuery represents query parameters for listing media
type MediaListQuery struct {
	PaginationQuery
	MimeType string `form:"mime_type" binding:"omitempty"`
	UserID   string `form:"user_id" binding:"omitempty,uuid"`
}

// MediaResponse represents a media item in API responses
type MediaResponse struct {
	ID               string             `json:"id"`
	Filename         string             `json:"filename"`
	OriginalFilename string             `json:"original_filename"`
	FilePath         string             `json:"file_path"`
	FileSize         int64              `json:"file_size"`
	FileSizeFormatted string            `json:"file_size_formatted"`
	MimeType         string             `json:"mime_type"`
	UploadedBy       string             `json:"uploaded_by"`
	Uploader         *PublicUserResponse `json:"uploader,omitempty"`
	IsFeatured       bool               `json:"is_featured"`
	AltText          string             `json:"alt_text"`
	URL              string             `json:"url"`
	CreatedAt        time.Time          `json:"created_at"`
}

// MediaUploadResponse represents a successful upload response
type MediaUploadResponse struct {
	ID               string    `json:"id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	FilePath         string    `json:"file_path"`
	FileSize         int64     `json:"file_size"`
	MimeType         string    `json:"mime_type"`
	URL              string    `json:"url"`
	CreatedAt        time.Time `json:"created_at"`
}

// UpdateMediaRequest represents a media update request
type UpdateMediaRequest struct {
	AltText    *string `json:"alt_text" binding:"omitempty,max=255"`
	IsFeatured *bool   `json:"is_featured"`
}
