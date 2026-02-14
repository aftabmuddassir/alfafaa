package repositories

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EngagementRepository defines the interface for engagement data access (likes, bookmarks, notifications)
type EngagementRepository interface {
	// Likes
	CreateLike(like *models.Like) error
	DeleteLike(userID, articleID uuid.UUID) error
	HasLiked(userID, articleID uuid.UUID) (bool, error)
	GetLikesCount(articleID uuid.UUID) (int64, error)

	// Bookmarks
	CreateBookmark(bookmark *models.Bookmark) error
	DeleteBookmark(userID, articleID uuid.UUID) error
	HasBookmarked(userID, articleID uuid.UUID) (bool, error)
	GetBookmarkedArticles(userID uuid.UUID, limit, offset int) ([]models.Article, int64, error)

	// Notifications
	CreateNotification(notification *models.Notification) error
	GetNotifications(userID uuid.UUID, limit, offset int) ([]models.Notification, int64, error)
	GetUnreadCount(userID uuid.UUID) (int64, error)
	MarkAsRead(notificationID, userID uuid.UUID) error
	MarkAllAsRead(userID uuid.UUID) error

	// Comment counts
	GetCommentsCount(articleID uuid.UUID) (int64, error)
}

type engagementRepository struct {
	db *gorm.DB
}

// NewEngagementRepository creates a new engagement repository
func NewEngagementRepository(db *gorm.DB) EngagementRepository {
	return &engagementRepository{db: db}
}

// --- Likes ---

// CreateLike creates a new like
func (r *engagementRepository) CreateLike(like *models.Like) error {
	return r.db.Create(like).Error
}

// DeleteLike removes a like
func (r *engagementRepository) DeleteLike(userID, articleID uuid.UUID) error {
	return r.db.Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(&models.Like{}).Error
}

// HasLiked checks if a user has liked an article
func (r *engagementRepository) HasLiked(userID, articleID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Count(&count).Error
	return count > 0, err
}

// GetLikesCount returns the number of likes for an article
func (r *engagementRepository) GetLikesCount(articleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Like{}).
		Where("article_id = ?", articleID).
		Count(&count).Error
	return count, err
}

// --- Bookmarks ---

// CreateBookmark creates a new bookmark
func (r *engagementRepository) CreateBookmark(bookmark *models.Bookmark) error {
	return r.db.Create(bookmark).Error
}

// DeleteBookmark removes a bookmark
func (r *engagementRepository) DeleteBookmark(userID, articleID uuid.UUID) error {
	return r.db.Where("user_id = ? AND article_id = ?", userID, articleID).
		Delete(&models.Bookmark{}).Error
}

// HasBookmarked checks if a user has bookmarked an article
func (r *engagementRepository) HasBookmarked(userID, articleID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.Bookmark{}).
		Where("user_id = ? AND article_id = ?", userID, articleID).
		Count(&count).Error
	return count > 0, err
}

// GetBookmarkedArticles returns articles bookmarked by a user
func (r *engagementRepository) GetBookmarkedArticles(userID uuid.UUID, limit, offset int) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).
		Joins("JOIN bookmarks ON bookmarks.article_id = articles.id").
		Where("bookmarks.user_id = ?", userID).
		Where("articles.deleted_at IS NULL")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("bookmarks.created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, total, err
}

// --- Notifications ---

// CreateNotification creates a new notification
func (r *engagementRepository) CreateNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

// GetNotifications returns notifications for a user, newest first
func (r *engagementRepository) GetNotifications(userID uuid.UUID, limit, offset int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{}).
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.
		Preload("Actor").
		Preload("Article").
		Find(&notifications).Error

	return notifications, total, err
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *engagementRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a single notification as read
func (r *engagementRepository) MarkAsRead(notificationID, userID uuid.UUID) error {
	result := r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("read", true)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// MarkAllAsRead marks all notifications as read for a user
func (r *engagementRepository) MarkAllAsRead(userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Update("read", true).Error
}

// --- Comments ---

// GetCommentsCount returns the number of comments for an article
func (r *engagementRepository) GetCommentsCount(articleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).
		Where("article_id = ? AND deleted_at IS NULL", articleID).
		Count(&count).Error
	return count, err
}
