package services

import (
	"testing"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/config"
	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/alfafaa/alfafaa-blog/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	suite.Suite
	userRepo  *mocks.MockUserRepository
	jwtConfig config.JWTConfig
	service   AuthService
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.jwtConfig = config.JWTConfig{
		Secret:            "test-secret-key-for-testing",
		Expiration:        24 * time.Hour,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	suite.service = NewAuthService(suite.userRepo, suite.jwtConfig)
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// Register Tests

func (suite *AuthServiceTestSuite) TestRegister_Success() {
	req := &dto.RegisterRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "User",
	}

	suite.userRepo.On("ExistsByEmail", req.Email).Return(false, nil)
	suite.userRepo.On("ExistsByUsername", req.Username).Return(false, nil)
	suite.userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	result, err := suite.service.Register(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), req.Email, result.User.Email)
	assert.Equal(suite.T(), req.Username, result.User.Username)
	assert.Equal(suite.T(), "reader", result.User.Role)
	assert.NotEmpty(suite.T(), result.AccessToken)
	assert.NotEmpty(suite.T(), result.RefreshToken)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRegister_WeakPassword() {
	req := &dto.RegisterRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "weak", // No uppercase, no number, too short
		FirstName: "Test",
		LastName:  "User",
	}

	result, err := suite.service.Register(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "WEAK_PASSWORD", appErr.Code)
}

func (suite *AuthServiceTestSuite) TestRegister_DuplicateEmail() {
	req := &dto.RegisterRequest{
		Username:  "testuser",
		Email:     "existing@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "User",
	}

	suite.userRepo.On("ExistsByEmail", req.Email).Return(true, nil)

	result, err := suite.service.Register(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrEmailExists, err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRegister_DuplicateUsername() {
	req := &dto.RegisterRequest{
		Username:  "existinguser",
		Email:     "test@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "User",
	}

	suite.userRepo.On("ExistsByEmail", req.Email).Return(false, nil)
	suite.userRepo.On("ExistsByUsername", req.Username).Return(true, nil)

	result, err := suite.service.Register(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrUsernameExists, err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRegister_CreateFails() {
	req := &dto.RegisterRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "Password123!",
		FirstName: "Test",
		LastName:  "User",
	}

	suite.userRepo.On("ExistsByEmail", req.Email).Return(false, nil)
	suite.userRepo.On("ExistsByUsername", req.Username).Return(false, nil)
	suite.userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(gorm.ErrInvalidDB)

	result, err := suite.service.Register(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	suite.userRepo.AssertExpectations(suite.T())
}

// Login Tests

func (suite *AuthServiceTestSuite) TestLogin_Success() {
	hashedPassword, _ := utils.HashPassword("Password123!")
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         models.RoleAuthor,
		IsActive:     true,
	}

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	suite.userRepo.On("FindByEmail", req.Email).Return(user, nil)
	suite.userRepo.On("UpdateLastLogin", user.ID).Return(nil)

	result, err := suite.service.Login(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), user.Email, result.User.Email)
	assert.NotEmpty(suite.T(), result.AccessToken)
	assert.NotEmpty(suite.T(), result.RefreshToken)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidEmail() {
	req := &dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "Password123!",
	}

	suite.userRepo.On("FindByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.Login(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrInvalidCredentials, err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidPassword() {
	hashedPassword, _ := utils.HashPassword("CorrectPassword123!")
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         models.RoleReader,
		IsActive:     true,
	}

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	}

	suite.userRepo.On("FindByEmail", req.Email).Return(user, nil)

	result, err := suite.service.Login(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrInvalidCredentials, err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_InactiveUser() {
	hashedPassword, _ := utils.HashPassword("Password123!")
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         models.RoleReader,
		IsActive:     false, // Inactive
	}

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	suite.userRepo.On("FindByEmail", req.Email).Return(user, nil)

	result, err := suite.service.Login(req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "ACCOUNT_DISABLED", appErr.Code)
	suite.userRepo.AssertExpectations(suite.T())
}

// RefreshToken Tests

func (suite *AuthServiceTestSuite) TestRefreshToken_Success() {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		Role:     models.RoleAuthor,
		IsActive: true,
	}

	// Generate a valid refresh token
	refreshToken, _, _ := utils.GenerateToken(
		userID,
		user.Email,
		string(user.Role),
		suite.jwtConfig.Secret,
		suite.jwtConfig.RefreshExpiration,
		utils.RefreshToken,
	)

	suite.userRepo.On("FindByID", userID).Return(user, nil)

	result, err := suite.service.RefreshToken(refreshToken)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.NotEmpty(suite.T(), result.AccessToken)
	assert.NotEmpty(suite.T(), result.RefreshToken)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_InvalidToken() {
	result, err := suite.service.RefreshToken("invalid-token")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrInvalidToken, err)
}

func (suite *AuthServiceTestSuite) TestRefreshToken_ExpiredToken() {
	userID := uuid.New()

	// Generate an expired refresh token
	refreshToken, _, _ := utils.GenerateToken(
		userID,
		"test@example.com",
		"reader",
		suite.jwtConfig.Secret,
		-time.Hour, // Expired
		utils.RefreshToken,
	)

	result, err := suite.service.RefreshToken(refreshToken)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrInvalidToken, err)
}

func (suite *AuthServiceTestSuite) TestRefreshToken_UserNotFound() {
	userID := uuid.New()

	// Generate a valid refresh token
	refreshToken, _, _ := utils.GenerateToken(
		userID,
		"test@example.com",
		"reader",
		suite.jwtConfig.Secret,
		suite.jwtConfig.RefreshExpiration,
		utils.RefreshToken,
	)

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.RefreshToken(refreshToken)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrInvalidToken, err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestRefreshToken_UserInactive() {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		Role:     models.RoleReader,
		IsActive: false, // Inactive
	}

	// Generate a valid refresh token
	refreshToken, _, _ := utils.GenerateToken(
		userID,
		user.Email,
		string(user.Role),
		suite.jwtConfig.Secret,
		suite.jwtConfig.RefreshExpiration,
		utils.RefreshToken,
	)

	suite.userRepo.On("FindByID", userID).Return(user, nil)

	result, err := suite.service.RefreshToken(refreshToken)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "ACCOUNT_DISABLED", appErr.Code)
	suite.userRepo.AssertExpectations(suite.T())
}

// GetCurrentUser Tests

func (suite *AuthServiceTestSuite) TestGetCurrentUser_Success() {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleAuthor,
		IsActive: true,
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)

	result, err := suite.service.GetCurrentUser(userID.String())

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), user.Email, result.Email)
	assert.Equal(suite.T(), user.Username, result.Username)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestGetCurrentUser_InvalidUUID() {
	result, err := suite.service.GetCurrentUser("not-a-valid-uuid")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *AuthServiceTestSuite) TestGetCurrentUser_NotFound() {
	userID := uuid.New()

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetCurrentUser(userID.String())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.userRepo.AssertExpectations(suite.T())
}

// ChangePassword Tests

func (suite *AuthServiceTestSuite) TestChangePassword_Success() {
	userID := uuid.New()
	currentPassword := "CurrentPass123!"
	hashedPassword, _ := utils.HashPassword(currentPassword)

	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	req := &dto.ChangePasswordRequest{
		CurrentPassword: currentPassword,
		NewPassword:     "NewPassword456!",
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	err := suite.service.ChangePassword(userID.String(), req)

	assert.NoError(suite.T(), err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestChangePassword_WrongCurrentPassword() {
	userID := uuid.New()
	hashedPassword, _ := utils.HashPassword("CorrectPassword123!")

	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	req := &dto.ChangePasswordRequest{
		CurrentPassword: "WrongPassword123!",
		NewPassword:     "NewPassword456!",
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)

	err := suite.service.ChangePassword(userID.String(), req)

	assert.Error(suite.T(), err)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "INVALID_PASSWORD", appErr.Code)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestChangePassword_WeakNewPassword() {
	userID := uuid.New()
	currentPassword := "CurrentPass123!"
	hashedPassword, _ := utils.HashPassword(currentPassword)

	user := &models.User{
		ID:           userID,
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	req := &dto.ChangePasswordRequest{
		CurrentPassword: currentPassword,
		NewPassword:     "weak", // Weak password
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)

	err := suite.service.ChangePassword(userID.String(), req)

	assert.Error(suite.T(), err)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "WEAK_PASSWORD", appErr.Code)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestChangePassword_UserNotFound() {
	userID := uuid.New()

	req := &dto.ChangePasswordRequest{
		CurrentPassword: "CurrentPass123!",
		NewPassword:     "NewPassword456!",
	}

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.ChangePassword(userID.String(), req)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestChangePassword_InvalidUUID() {
	req := &dto.ChangePasswordRequest{
		CurrentPassword: "CurrentPass123!",
		NewPassword:     "NewPassword456!",
	}

	err := suite.service.ChangePassword("not-a-valid-uuid", req)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}
