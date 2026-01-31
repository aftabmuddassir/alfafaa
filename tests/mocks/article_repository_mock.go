package mocks

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockArticleRepository is a mock implementation of ArticleRepository
type MockArticleRepository struct {
	mock.Mock
}

// Ensure MockArticleRepository implements ArticleRepository
var _ repositories.ArticleRepository = (*MockArticleRepository)(nil)

// Create mocks the Create method
func (m *MockArticleRepository) Create(article *models.Article) error {
	args := m.Called(article)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockArticleRepository) FindByID(id uuid.UUID) (*models.Article, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Article), args.Error(1)
}

// FindBySlug mocks the FindBySlug method
func (m *MockArticleRepository) FindBySlug(slug string) (*models.Article, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Article), args.Error(1)
}

// FindAll mocks the FindAll method
func (m *MockArticleRepository) FindAll(filters repositories.ArticleFilters) ([]models.Article, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.Article), args.Get(1).(int64), args.Error(2)
}

// FindByAuthor mocks the FindByAuthor method
func (m *MockArticleRepository) FindByAuthor(authorID uuid.UUID, filters repositories.ArticleFilters) ([]models.Article, int64, error) {
	args := m.Called(authorID, filters)
	return args.Get(0).([]models.Article), args.Get(1).(int64), args.Error(2)
}

// FindByCategory mocks the FindByCategory method
func (m *MockArticleRepository) FindByCategory(categoryID uuid.UUID, filters repositories.ArticleFilters) ([]models.Article, int64, error) {
	args := m.Called(categoryID, filters)
	return args.Get(0).([]models.Article), args.Get(1).(int64), args.Error(2)
}

// FindByTag mocks the FindByTag method
func (m *MockArticleRepository) FindByTag(tagID uuid.UUID, filters repositories.ArticleFilters) ([]models.Article, int64, error) {
	args := m.Called(tagID, filters)
	return args.Get(0).([]models.Article), args.Get(1).(int64), args.Error(2)
}

// FindTrending mocks the FindTrending method
func (m *MockArticleRepository) FindTrending(limit int) ([]models.Article, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Article), args.Error(1)
}

// FindRecent mocks the FindRecent method
func (m *MockArticleRepository) FindRecent(limit int) ([]models.Article, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Article), args.Error(1)
}

// FindRelated mocks the FindRelated method
func (m *MockArticleRepository) FindRelated(articleID uuid.UUID, categoryIDs, tagIDs []uuid.UUID, limit int) ([]models.Article, error) {
	args := m.Called(articleID, categoryIDs, tagIDs, limit)
	return args.Get(0).([]models.Article), args.Error(1)
}

// Update mocks the Update method
func (m *MockArticleRepository) Update(article *models.Article) error {
	args := m.Called(article)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockArticleRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// ExistsBySlug mocks the ExistsBySlug method
func (m *MockArticleRepository) ExistsBySlug(slug string) (bool, error) {
	args := m.Called(slug)
	return args.Bool(0), args.Error(1)
}

// IncrementViewCount mocks the IncrementViewCount method
func (m *MockArticleRepository) IncrementViewCount(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// UpdateCategories mocks the UpdateCategories method
func (m *MockArticleRepository) UpdateCategories(article *models.Article, categories []models.Category) error {
	args := m.Called(article, categories)
	return args.Error(0)
}

// UpdateTags mocks the UpdateTags method
func (m *MockArticleRepository) UpdateTags(article *models.Article, tags []models.Tag) error {
	args := m.Called(article, tags)
	return args.Error(0)
}

// Search mocks the Search method
func (m *MockArticleRepository) Search(query string, filters repositories.ArticleFilters) ([]models.Article, int64, error) {
	args := m.Called(query, filters)
	return args.Get(0).([]models.Article), args.Get(1).(int64), args.Error(2)
}

// WithTx mocks the WithTx method - returns itself for testing
func (m *MockArticleRepository) WithTx(tx *gorm.DB) repositories.ArticleRepository {
	m.Called(tx)
	return m // Return self to allow chaining in tests
}
