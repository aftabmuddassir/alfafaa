package mocks

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockTagRepository is a mock implementation of TagRepository
type MockTagRepository struct {
	mock.Mock
}

// Ensure MockTagRepository implements TagRepository
var _ repositories.TagRepository = (*MockTagRepository)(nil)

// Create mocks the Create method
func (m *MockTagRepository) Create(tag *models.Tag) error {
	args := m.Called(tag)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockTagRepository) FindByID(id uuid.UUID) (*models.Tag, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}

// FindBySlug mocks the FindBySlug method
func (m *MockTagRepository) FindBySlug(slug string) (*models.Tag, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tag), args.Error(1)
}

// FindAll mocks the FindAll method
func (m *MockTagRepository) FindAll(filters repositories.TagFilters) ([]models.Tag, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Tag), args.Get(1).(int64), args.Error(2)
}

// FindPopular mocks the FindPopular method
func (m *MockTagRepository) FindPopular(limit int) ([]models.Tag, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Tag), args.Error(1)
}

// Update mocks the Update method
func (m *MockTagRepository) Update(tag *models.Tag) error {
	args := m.Called(tag)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockTagRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// ExistsByName mocks the ExistsByName method
func (m *MockTagRepository) ExistsByName(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

// ExistsBySlug mocks the ExistsBySlug method
func (m *MockTagRepository) ExistsBySlug(slug string) (bool, error) {
	args := m.Called(slug)
	return args.Bool(0), args.Error(1)
}

// IncrementUsage mocks the IncrementUsage method
func (m *MockTagRepository) IncrementUsage(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// DecrementUsage mocks the DecrementUsage method
func (m *MockTagRepository) DecrementUsage(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// FindByIDs mocks the FindByIDs method
func (m *MockTagRepository) FindByIDs(ids []uuid.UUID) ([]models.Tag, error) {
	args := m.Called(ids)
	return args.Get(0).([]models.Tag), args.Error(1)
}

// WithTx mocks the WithTx method - returns itself for testing
func (m *MockTagRepository) WithTx(tx *gorm.DB) repositories.TagRepository {
	m.Called(tx)
	return m // Return self to allow chaining in tests
}
