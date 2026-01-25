package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alfafaa/alfafaa-blog/internal/config"
	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaService defines the interface for media operations
type MediaService interface {
	UploadMedia(media *models.Media) (*dto.MediaUploadResponse, error)
	GetMedia(id string) (*dto.MediaResponse, error)
	GetAllMedia(query *dto.MediaListQuery) ([]dto.MediaResponse, int64, error)
	DeleteMedia(id string, userID string, isAdmin bool) error
}

type mediaService struct {
	mediaRepo    repositories.MediaRepository
	uploadConfig config.UploadConfig
}

// NewMediaService creates a new media service
func NewMediaService(mediaRepo repositories.MediaRepository, uploadConfig config.UploadConfig) MediaService {
	return &mediaService{
		mediaRepo:    mediaRepo,
		uploadConfig: uploadConfig,
	}
}

// UploadMedia saves a media record
func (s *mediaService) UploadMedia(media *models.Media) (*dto.MediaUploadResponse, error) {
	if err := s.mediaRepo.Create(media); err != nil {
		return nil, utils.WrapError(err, "failed to save media record")
	}

	return &dto.MediaUploadResponse{
		ID:               media.ID.String(),
		Filename:         media.Filename,
		OriginalFilename: media.OriginalFilename,
		FilePath:         media.FilePath,
		FileSize:         media.FileSize,
		MimeType:         media.MimeType,
		URL:              "/" + media.FilePath,
		CreatedAt:        media.CreatedAt,
	}, nil
}

// GetMedia retrieves a media record by ID
func (s *mediaService) GetMedia(id string) (*dto.MediaResponse, error) {
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	media, err := s.mediaRepo.FindByID(mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find media")
	}

	return s.toResponse(media), nil
}

// GetAllMedia retrieves all media with filters
func (s *mediaService) GetAllMedia(query *dto.MediaListQuery) ([]dto.MediaResponse, int64, error) {
	filters := repositories.MediaFilters{
		MimeType: query.MimeType,
		Limit:    query.GetPerPage(),
		Offset:   query.GetOffset(),
	}

	media, total, err := s.mediaRepo.FindAll(filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to find media")
	}

	responses := make([]dto.MediaResponse, len(media))
	for i, m := range media {
		responses[i] = *s.toResponse(&m)
	}

	return responses, total, nil
}

// DeleteMedia deletes a media record and its file
func (s *mediaService) DeleteMedia(id string, userID string, isAdmin bool) error {
	mediaID, err := uuid.Parse(id)
	if err != nil {
		return utils.ErrBadRequest
	}

	media, err := s.mediaRepo.FindByID(mediaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to find media")
	}

	// Check permissions
	if media.UploadedBy.String() != userID && !isAdmin {
		return utils.ErrForbidden
	}

	// Delete the file
	if err := os.Remove(media.FilePath); err != nil && !os.IsNotExist(err) {
		// Log error but continue with database deletion
		fmt.Printf("Warning: failed to delete file %s: %v\n", media.FilePath, err)
	}

	if err := s.mediaRepo.Delete(mediaID); err != nil {
		return utils.WrapError(err, "failed to delete media")
	}

	return nil
}

// toResponse converts a media model to a response DTO
func (s *mediaService) toResponse(media *models.Media) *dto.MediaResponse {
	response := &dto.MediaResponse{
		ID:                media.ID.String(),
		Filename:          media.Filename,
		OriginalFilename:  media.OriginalFilename,
		FilePath:          media.FilePath,
		FileSize:          media.FileSize,
		FileSizeFormatted: media.GetFileSizeFormatted(),
		MimeType:          media.MimeType,
		UploadedBy:        media.UploadedBy.String(),
		IsFeatured:        media.IsFeatured,
		AltText:           media.AltText,
		URL:               "/" + media.FilePath,
		CreatedAt:         media.CreatedAt,
	}

	if media.Uploader != nil {
		response.Uploader = &dto.PublicUserResponse{
			ID:              media.Uploader.ID.String(),
			Username:        media.Uploader.Username,
			FirstName:       media.Uploader.FirstName,
			LastName:        media.Uploader.LastName,
			ProfileImageURL: media.Uploader.ProfileImageURL,
		}
	}

	return response
}

// GenerateUniqueFilename generates a unique filename for uploads
func GenerateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return uuid.New().String() + ext
}
