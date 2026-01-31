package repositories

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TagRepository defines the interface for tag data access
type TagRepository interface {
	Create(tag *models.Tag) error
	FindByID(id uuid.UUID) (*models.Tag, error)
	FindBySlug(slug string) (*models.Tag, error)
	FindAll(filters TagFilters) ([]models.Tag, int64, error)
	FindPopular(limit int) ([]models.Tag, error)
	Update(tag *models.Tag) error
	Delete(id uuid.UUID) error
	ExistsByName(name string) (bool, error)
	ExistsBySlug(slug string) (bool, error)
	IncrementUsage(id uuid.UUID) error
	DecrementUsage(id uuid.UUID) error
	FindByIDs(ids []uuid.UUID) ([]models.Tag, error)
	// WithTx returns a new repository instance using the provided transaction
	WithTx(tx *gorm.DB) TagRepository
}

// TagFilters contains filter options for querying tags
type TagFilters struct {
	Search string
	Limit  int
	Offset int
}

type tagRepository struct {
	db *gorm.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

// WithTx returns a new repository instance using the provided transaction
func (r *tagRepository) WithTx(tx *gorm.DB) TagRepository {
	return &tagRepository{db: tx}
}

// Create creates a new tag
func (r *tagRepository) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

// FindByID finds a tag by ID
func (r *tagRepository) FindByID(id uuid.UUID) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.First(&tag, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindBySlug finds a tag by slug
func (r *tagRepository) FindBySlug(slug string) (*models.Tag, error) {
	var tag models.Tag
	err := r.db.First(&tag, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// FindAll finds all tags with filters
func (r *tagRepository) FindAll(filters TagFilters) ([]models.Tag, int64, error) {
	var tags []models.Tag
	var total int64

	query := r.db.Model(&models.Tag{})

	// Apply filters
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and order
	query = query.Order("name ASC")
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// FindPopular finds the most popular tags by usage count
func (r *tagRepository) FindPopular(limit int) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Order("usage_count DESC").Limit(limit).Find(&tags).Error
	return tags, err
}

// Update updates a tag
func (r *tagRepository) Update(tag *models.Tag) error {
	return r.db.Save(tag).Error
}

// Delete deletes a tag
func (r *tagRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Tag{}, "id = ?", id).Error
}

// ExistsByName checks if a tag exists with the given name
func (r *tagRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Tag{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsBySlug checks if a tag exists with the given slug
func (r *tagRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Tag{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

// IncrementUsage increments the usage count of a tag
func (r *tagRepository) IncrementUsage(id uuid.UUID) error {
	return r.db.Model(&models.Tag{}).Where("id = ?", id).Update("usage_count", gorm.Expr("usage_count + 1")).Error
}

// DecrementUsage decrements the usage count of a tag
func (r *tagRepository) DecrementUsage(id uuid.UUID) error {
	return r.db.Model(&models.Tag{}).Where("id = ?", id).Where("usage_count > 0").Update("usage_count", gorm.Expr("usage_count - 1")).Error
}

// FindByIDs finds tags by their IDs
func (r *tagRepository) FindByIDs(ids []uuid.UUID) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.db.Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}
