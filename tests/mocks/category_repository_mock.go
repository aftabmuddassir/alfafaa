package mocks

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockCategoryRepository is a mock implementation of CategoryRepository
type MockCategoryRepository struct {
	mock.Mock
}

// Ensure MockCategoryRepository implements CategoryRepository
var _ repositories.CategoryRepository = (*MockCategoryRepository)(nil)

// Create mocks the Create method
func (m *MockCategoryRepository) Create(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockCategoryRepository) FindByID(id uuid.UUID) (*models.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

// FindBySlug mocks the FindBySlug method
func (m *MockCategoryRepository) FindBySlug(slug string) (*models.Category, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

// FindAll mocks the FindAll method
func (m *MockCategoryRepository) FindAll(filters repositories.CategoryFilters) ([]models.Category, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Category), args.Error(1)
}

// FindAllHierarchical mocks the FindAllHierarchical method
func (m *MockCategoryRepository) FindAllHierarchical() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

// Update mocks the Update method
func (m *MockCategoryRepository) Update(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockCategoryRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// ExistsByName mocks the ExistsByName method
func (m *MockCategoryRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

// ExistsBySlug mocks the ExistsBySlug method
func (m *MockCategoryRepository) ExistsBySlug(slug string) (bool, error) {
	args := m.Called(slug)
	return args.Bool(0), args.Error(1)
}

// GetArticleCount mocks the GetArticleCount method
func (m *MockCategoryRepository) GetArticleCount(id uuid.UUID) (int64, error) {
	args := m.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

// FindByIDs mocks the FindByIDs method
func (m *MockCategoryRepository) FindByIDs(ids []uuid.UUID) ([]models.Category, error) {
	args := m.Called(ids)
	return args.Get(0).([]models.Category), args.Error(1)
}
