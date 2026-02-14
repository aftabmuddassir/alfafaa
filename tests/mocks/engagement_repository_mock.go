package mocks

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockEngagementRepository is a mock implementation of EngagementRepository
type MockEngagementRepository struct {
	mock.Mock
}

// Ensure MockEngagementRepository implements EngagementRepository
var _ repositories.EngagementRepository = (*MockEngagementRepository)(nil)

// --- Likes ---

func (m *MockEngagementRepository) CreateLike(like *models.Like) error {
	args := m.Called(like)
	return args.Error(0)
}

func (m *MockEngagementRepository) DeleteLike(userID, articleID uuid.UUID) error {
	args := m.Called(userID, articleID)
	return args.Error(0)
}

func (m *MockEngagementRepository) HasLiked(userID, articleID uuid.UUID) (bool, error) {
	args := m.Called(userID, articleID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEngagementRepository) GetLikesCount(articleID uuid.UUID) (int64, error) {
	args := m.Called(articleID)
	return args.Get(0).(int64), args.Error(1)
}

// --- Bookmarks ---

func (m *MockEngagementRepository) CreateBookmark(bookmark *models.Bookmark) error {
	args := m.Called(bookmark)
	return args.Error(0)
}

func (m *MockEngagementRepository) DeleteBookmark(userID, articleID uuid.UUID) error {
	args := m.Called(userID, articleID)
	return args.Error(0)
}

func (m *MockEngagementRepository) HasBookmarked(userID, articleID uuid.UUID) (bool, error) {
	args := m.Called(userID, articleID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEngagementRepository) GetBookmarkedArticles(userID uuid.UUID, limit, offset int) ([]models.Article, int64, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.Article), args.Get(1).(int64), args.Error(2)
}

// --- Notifications ---

func (m *MockEngagementRepository) CreateNotification(notification *models.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}

func (m *MockEngagementRepository) GetNotifications(userID uuid.UUID, limit, offset int) ([]models.Notification, int64, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.Notification), args.Get(1).(int64), args.Error(2)
}

func (m *MockEngagementRepository) GetUnreadCount(userID uuid.UUID) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEngagementRepository) MarkAsRead(notificationID, userID uuid.UUID) error {
	args := m.Called(notificationID, userID)
	return args.Error(0)
}

func (m *MockEngagementRepository) MarkAllAsRead(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// --- Comment counts ---

func (m *MockEngagementRepository) GetCommentsCount(articleID uuid.UUID) (int64, error) {
	args := m.Called(articleID)
	return args.Get(0).(int64), args.Error(1)
}
