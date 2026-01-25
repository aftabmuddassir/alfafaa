package repositories

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	Create(media *models.Media) error
	FindByID(id uuid.UUID) (*models.Media, error)
	FindAll(filters MediaFilters) ([]models.Media, int64, error)
	FindByUploader(uploaderID uuid.UUID, filters MediaFilters) ([]models.Media, int64, error)
	Delete(id uuid.UUID) error
}

// MediaFilters contains filter options for querying media
type MediaFilters struct {
	MimeType string
	Limit    int
	Offset   int
}

type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

// Create creates a new media record
func (r *mediaRepository) Create(media *models.Media) error {
	return r.db.Create(media).Error
}

// FindByID finds a media record by ID
func (r *mediaRepository) FindByID(id uuid.UUID) (*models.Media, error) {
	var media models.Media
	err := r.db.Preload("Uploader").First(&media, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// FindAll finds all media with filters
func (r *mediaRepository) FindAll(filters MediaFilters) ([]models.Media, int64, error) {
	var media []models.Media
	var total int64

	query := r.db.Model(&models.Media{})

	// Apply filters
	if filters.MimeType != "" {
		query = query.Where("mime_type = ?", filters.MimeType)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and order
	query = query.Order("created_at DESC")
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err := query.Preload("Uploader").Find(&media).Error
	return media, total, err
}

// FindByUploader finds all media uploaded by a specific user
func (r *mediaRepository) FindByUploader(uploaderID uuid.UUID, filters MediaFilters) ([]models.Media, int64, error) {
	var media []models.Media
	var total int64

	query := r.db.Model(&models.Media{}).Where("uploaded_by = ?", uploaderID)

	// Apply filters
	if filters.MimeType != "" {
		query = query.Where("mime_type = ?", filters.MimeType)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and order
	query = query.Order("created_at DESC")
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err := query.Find(&media).Error
	return media, total, err
}

// Delete deletes a media record
func (r *mediaRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Media{}, "id = ?", id).Error
}
