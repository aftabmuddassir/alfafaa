package mocks

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockCommentRepository is a mock implementation of CommentRepository
type MockCommentRepository struct {
	mock.Mock
}

// Ensure MockCommentRepository implements CommentRepository
var _ repositories.CommentRepository = (*MockCommentRepository)(nil)

func (m *MockCommentRepository) Create(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) FindByID(id uuid.UUID) (*models.Comment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentRepository) FindByArticle(articleID uuid.UUID, filters repositories.CommentFilters) ([]models.Comment, int64, error) {
	args := m.Called(articleID, filters)
	return args.Get(0).([]models.Comment), args.Get(1).(int64), args.Error(2)
}

func (m *MockCommentRepository) Update(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
