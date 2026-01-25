package helpers

import (
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// TestJWTSecret is the secret used for testing JWT tokens
	TestJWTSecret = "test-jwt-secret-key-for-testing-only"
	// TestPassword is the default password used in tests
	TestPassword = "TestPassword123!"
)

// CreateTestUser creates a test user with the specified role and returns the user and JWT token
func CreateTestUser(db *gorm.DB, role models.UserRole) (*models.User, string) {
	hashedPassword, _ := utils.HashPassword(TestPassword)

	user := &models.User{
		ID:           uuid.New(),
		Username:     "test_" + string(role) + "_" + uuid.New().String()[:8],
		Email:        string(role) + "_" + uuid.New().String()[:8] + "@test.com",
		PasswordHash: hashedPassword,
		FirstName:    "Test",
		LastName:     string(role),
		Role:         role,
		IsActive:     true,
		IsVerified:   true,
	}

	db.Create(user)

	token, _, _ := utils.GenerateToken(
		user.ID,
		user.Email,
		string(user.Role),
		TestJWTSecret,
		24*time.Hour,
		utils.AccessToken,
	)

	return user, token
}

// CreateTestUserWithDetails creates a test user with custom details
func CreateTestUserWithDetails(db *gorm.DB, username, email string, role models.UserRole, isActive bool) (*models.User, string) {
	hashedPassword, _ := utils.HashPassword(TestPassword)

	user := &models.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    "Test",
		LastName:     "User",
		Role:         role,
		IsActive:     isActive,
		IsVerified:   true,
	}

	db.Create(user)

	token, _, _ := utils.GenerateToken(
		user.ID,
		user.Email,
		string(user.Role),
		TestJWTSecret,
		24*time.Hour,
		utils.AccessToken,
	)

	return user, token
}

// GenerateTestToken generates a JWT token for testing
func GenerateTestToken(userID uuid.UUID, email, role string) string {
	token, _, _ := utils.GenerateToken(
		userID,
		email,
		role,
		TestJWTSecret,
		24*time.Hour,
		utils.AccessToken,
	)
	return token
}

// GenerateExpiredToken generates an expired JWT token for testing
func GenerateExpiredToken(userID uuid.UUID, email, role string) string {
	token, _, _ := utils.GenerateToken(
		userID,
		email,
		role,
		TestJWTSecret,
		-1*time.Hour, // Expired
		utils.AccessToken,
	)
	return token
}

// GenerateTestRefreshToken generates a refresh token for testing
func GenerateTestRefreshToken(userID uuid.UUID, email, role string) string {
	token, _, _ := utils.GenerateToken(
		userID,
		email,
		role,
		TestJWTSecret,
		7*24*time.Hour,
		utils.RefreshToken,
	)
	return token
}
