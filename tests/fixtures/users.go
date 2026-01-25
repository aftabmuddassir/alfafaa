package fixtures

import (
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
)

// ValidUserID is a valid UUID for testing
var ValidUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

// TestUsers contains sample user data for testing
var TestUsers = struct {
	Admin  models.User
	Editor models.User
	Author models.User
	Reader models.User
}{
	Admin: models.User{
		ID:           uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Username:     "admin_test",
		Email:        "admin@test.com",
		PasswordHash: "$2a$10$dummy.hash.for.testing.purposes.only",
		FirstName:    "Admin",
		LastName:     "User",
		Bio:          "Admin user for testing",
		Role:         models.RoleAdmin,
		IsActive:     true,
		IsVerified:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	},
	Editor: models.User{
		ID:           uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Username:     "editor_test",
		Email:        "editor@test.com",
		PasswordHash: "$2a$10$dummy.hash.for.testing.purposes.only",
		FirstName:    "Editor",
		LastName:     "User",
		Bio:          "Editor user for testing",
		Role:         models.RoleEditor,
		IsActive:     true,
		IsVerified:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	},
	Author: models.User{
		ID:           uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Username:     "author_test",
		Email:        "author@test.com",
		PasswordHash: "$2a$10$dummy.hash.for.testing.purposes.only",
		FirstName:    "Author",
		LastName:     "User",
		Bio:          "Author user for testing",
		Role:         models.RoleAuthor,
		IsActive:     true,
		IsVerified:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	},
	Reader: models.User{
		ID:           uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		Username:     "reader_test",
		Email:        "reader@test.com",
		PasswordHash: "$2a$10$dummy.hash.for.testing.purposes.only",
		FirstName:    "Reader",
		LastName:     "User",
		Bio:          "Reader user for testing",
		Role:         models.RoleReader,
		IsActive:     true,
		IsVerified:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	},
}

// NewTestUser creates a new user with default values that can be overridden
func NewTestUser(opts ...func(*models.User)) *models.User {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser_" + uuid.New().String()[:8],
		Email:        "testuser_" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: "$2a$10$dummy.hash.for.testing.purposes.only",
		FirstName:    "Test",
		LastName:     "User",
		Role:         models.RoleReader,
		IsActive:     true,
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	for _, opt := range opts {
		opt(user)
	}

	return user
}

// WithRole sets the user role
func WithRole(role models.UserRole) func(*models.User) {
	return func(u *models.User) {
		u.Role = role
	}
}

// WithEmail sets the user email
func WithEmail(email string) func(*models.User) {
	return func(u *models.User) {
		u.Email = email
	}
}

// WithUsername sets the username
func WithUsername(username string) func(*models.User) {
	return func(u *models.User) {
		u.Username = username
	}
}

// WithActive sets the user active status
func WithActive(active bool) func(*models.User) {
	return func(u *models.User) {
		u.IsActive = active
	}
}

// WithVerified sets the user verified status
func WithVerified(verified bool) func(*models.User) {
	return func(u *models.User) {
		u.IsVerified = verified
	}
}
