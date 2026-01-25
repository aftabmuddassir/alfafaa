package repositories

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentRepository defines the interface for comment data access
type CommentRepository interface {
	Create(comment *models.Comment) error
	FindByID(id uuid.UUID) (*models.Comment, error)
	FindByArticle(articleID uuid.UUID, filters CommentFilters) ([]models.Comment, int64, error)
	Update(comment *models.Comment) error
	Delete(id uuid.UUID) error
}

// CommentFilters contains filter options for querying comments
type CommentFilters struct {
	IncludeUnapproved bool
	ParentOnly        bool
	Limit             int
	Offset            int
}

type commentRepository struct {
	db *gorm.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// Create creates a new comment
func (r *commentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// FindByID finds a comment by ID
func (r *commentRepository) FindByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.
		Preload("User").
		Preload("Replies").
		Preload("Replies.User").
		First(&comment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// FindByArticle finds comments for an article
func (r *commentRepository) FindByArticle(articleID uuid.UUID, filters CommentFilters) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := r.db.Model(&models.Comment{}).
		Where("article_id = ?", articleID).
		Where("parent_id IS NULL") // Only get top-level comments

	// Apply filters
	if !filters.IncludeUnapproved {
		query = query.Where("is_approved = ?", true)
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

	// Preload user and replies
	preloadReplies := func(db *gorm.DB) *gorm.DB {
		q := db.Order("created_at ASC")
		if !filters.IncludeUnapproved {
			q = q.Where("is_approved = ?", true)
		}
		return q
	}

	err := query.
		Preload("User").
		Preload("Replies", preloadReplies).
		Preload("Replies.User").
		Find(&comments).Error

	return comments, total, err
}

// Update updates a comment
func (r *commentRepository) Update(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

// Delete soft deletes a comment
func (r *commentRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, "id = ?", id).Error
}
