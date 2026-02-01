package mocks

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Ensure MockUserRepository implements UserRepository
var _ repositories.UserRepository = (*MockUserRepository)(nil)

// Create mocks the Create method
func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// FindByEmail mocks the FindByEmail method
func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// FindByUsername mocks the FindByUsername method
func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// FindAll mocks the FindAll method
func (m *MockUserRepository) FindAll(filters repositories.UserFilters) ([]models.User, int64, error) {
	args := m.Called(filters)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

// Update mocks the Update method
func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// ExistsByEmail mocks the ExistsByEmail method
func (m *MockUserRepository) ExistsByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

// ExistsByUsername mocks the ExistsByUsername method
func (m *MockUserRepository) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

// UpdateLastLogin mocks the UpdateLastLogin method
func (m *MockUserRepository) UpdateLastLogin(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

// FindByIDWithRelations mocks the FindByIDWithRelations method
func (m *MockUserRepository) FindByIDWithRelations(id uuid.UUID) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// FindByGoogleID mocks the FindByGoogleID method
func (m *MockUserRepository) FindByGoogleID(googleID string) (*models.User, error) {
	args := m.Called(googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// FollowUser mocks the FollowUser method
func (m *MockUserRepository) FollowUser(followerID, followingID uuid.UUID) error {
	args := m.Called(followerID, followingID)
	return args.Error(0)
}

// UnfollowUser mocks the UnfollowUser method
func (m *MockUserRepository) UnfollowUser(followerID, followingID uuid.UUID) error {
	args := m.Called(followerID, followingID)
	return args.Error(0)
}

// IsFollowing mocks the IsFollowing method
func (m *MockUserRepository) IsFollowing(followerID, followingID uuid.UUID) (bool, error) {
	args := m.Called(followerID, followingID)
	return args.Bool(0), args.Error(1)
}

// GetFollowers mocks the GetFollowers method
func (m *MockUserRepository) GetFollowers(userID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

// GetFollowing mocks the GetFollowing method
func (m *MockUserRepository) GetFollowing(userID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}

// GetFollowingIDs mocks the GetFollowingIDs method
func (m *MockUserRepository) GetFollowingIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(userID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

// SetInterests mocks the SetInterests method
func (m *MockUserRepository) SetInterests(userID uuid.UUID, categoryIDs []uuid.UUID) error {
	args := m.Called(userID, categoryIDs)
	return args.Error(0)
}

// GetInterests mocks the GetInterests method
func (m *MockUserRepository) GetInterests(userID uuid.UUID) ([]models.Category, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Category), args.Error(1)
}

// GetInterestIDs mocks the GetInterestIDs method
func (m *MockUserRepository) GetInterestIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(userID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
