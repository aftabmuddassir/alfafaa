package repositories

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	Create(category *models.Category) error
	FindByID(id uuid.UUID) (*models.Category, error)
	FindBySlug(slug string) (*models.Category, error)
	FindAll(filters CategoryFilters) ([]models.Category, error)
	FindAllHierarchical() ([]models.Category, error)
	Update(category *models.Category) error
	Delete(id uuid.UUID) error
	ExistsByName(name string) (bool, error)
	ExistsBySlug(slug string) (bool, error)
	GetArticleCount(id uuid.UUID) (int64, error)
	FindByIDs(ids []uuid.UUID) ([]models.Category, error)
}

// CategoryFilters contains filter options for querying categories
type CategoryFilters struct {
	IncludeInactive bool
	ParentOnly      bool
	Search          string
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// Create creates a new category
func (r *categoryRepository) Create(category *models.Category) error {
	return r.db.Create(category).Error
}

// FindByID finds a category by ID
func (r *categoryRepository) FindByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Parent").Preload("Children").First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// FindBySlug finds a category by slug
func (r *categoryRepository) FindBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Parent").Preload("Children").First(&category, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// FindAll finds all categories with filters
func (r *categoryRepository) FindAll(filters CategoryFilters) ([]models.Category, error) {
	var categories []models.Category

	query := r.db.Model(&models.Category{})

	// Apply filters
	if !filters.IncludeInactive {
		query = query.Where("is_active = ?", true)
	}
	if filters.ParentOnly {
		query = query.Where("parent_id IS NULL")
	}
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", search, search)
	}

	err := query.Order("display_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

// FindAllHierarchical finds all categories with parent-child relationships
func (r *categoryRepository) FindAllHierarchical() ([]models.Category, error) {
	var categories []models.Category

	// Get root categories with their children
	err := r.db.
		Where("parent_id IS NULL").
		Where("is_active = ?", true).
		Preload("Children", "is_active = ?", true).
		Order("display_order ASC, name ASC").
		Find(&categories).Error

	return categories, err
}

// Update updates a category
func (r *categoryRepository) Update(category *models.Category) error {
	return r.db.Save(category).Error
}

// Delete deletes a category (hard delete since categories don't have soft delete)
func (r *categoryRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Category{}, "id = ?", id).Error
}

// ExistsByName checks if a category exists with the given name
func (r *categoryRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Category{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// ExistsBySlug checks if a category exists with the given slug
func (r *categoryRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Category{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

// GetArticleCount returns the number of articles in a category
func (r *categoryRepository) GetArticleCount(id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Table("article_categories").Where("category_id = ?", id).Count(&count).Error
	return count, err
}

// FindByIDs finds categories by their IDs
func (r *categoryRepository) FindByIDs(ids []uuid.UUID) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Where("id IN ?", ids).Find(&categories).Error
	return categories, err
}
