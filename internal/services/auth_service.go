package services

import (
	"errors"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/config"
	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	GetCurrentUser(userID string) (*dto.UserResponse, error)
	ChangePassword(userID string, req *dto.ChangePasswordRequest) error
}

type authService struct {
	userRepo  repositories.UserRepository
	jwtConfig config.JWTConfig
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repositories.UserRepository, jwtConfig config.JWTConfig) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// Register creates a new user account
func (s *authService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Validate password strength
	passwordErrors := utils.ValidatePasswordStrength(req.Password)
	if len(passwordErrors) > 0 {
		return nil, utils.NewAppError("WEAK_PASSWORD", passwordErrors[0], 400)
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check email")
	}
	if exists {
		return nil, utils.ErrEmailExists
	}

	// Check if username already exists
	exists, err = s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check username")
	}
	if exists {
		return nil, utils.ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, utils.WrapError(err, "failed to hash password")
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.RoleReader, // Default role
		IsActive:     true,
		IsVerified:   false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, utils.WrapError(err, "failed to create user")
	}

	// Generate tokens
	tokens, err := utils.GenerateTokenPair(
		user.ID,
		user.Email,
		string(user.Role),
		s.jwtConfig.Secret,
		s.jwtConfig.Expiration,
		s.jwtConfig.RefreshExpiration,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to generate tokens")
	}

	return &dto.AuthResponse{
		User:         s.toUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// Login authenticates a user
func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInvalidCredentials
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, utils.NewAppError("ACCOUNT_DISABLED", "Your account has been disabled", 403)
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, utils.ErrInvalidCredentials
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(user.ID)

	// Generate tokens
	tokens, err := utils.GenerateTokenPair(
		user.ID,
		user.Email,
		string(user.Role),
		s.jwtConfig.Secret,
		s.jwtConfig.Expiration,
		s.jwtConfig.RefreshExpiration,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to generate tokens")
	}

	// Get updated user
	user.LastLoginAt = &[]time.Time{time.Now()}[0]

	return &dto.AuthResponse{
		User:         s.toUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// RefreshToken generates new tokens from a refresh token
func (s *authService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(refreshToken, s.jwtConfig.Secret)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	// Get user to ensure they still exist and are active
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInvalidToken
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	if !user.IsActive {
		return nil, utils.NewAppError("ACCOUNT_DISABLED", "Your account has been disabled", 403)
	}

	// Generate new tokens
	tokens, err := utils.GenerateTokenPair(
		user.ID,
		user.Email,
		string(user.Role),
		s.jwtConfig.Secret,
		s.jwtConfig.Expiration,
		s.jwtConfig.RefreshExpiration,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to generate tokens")
	}

	return &dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// GetCurrentUser retrieves the current user's information
func (s *authService) GetCurrentUser(userID string) (*dto.UserResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	response := s.toUserResponse(user)
	return &response, nil
}

// ChangePassword changes the user's password
func (s *authService) ChangePassword(userID string, req *dto.ChangePasswordRequest) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return utils.ErrBadRequest
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to find user")
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return utils.NewAppError("INVALID_PASSWORD", "Current password is incorrect", 400)
	}

	// Validate new password strength
	passwordErrors := utils.ValidatePasswordStrength(req.NewPassword)
	if len(passwordErrors) > 0 {
		return utils.NewAppError("WEAK_PASSWORD", passwordErrors[0], 400)
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return utils.WrapError(err, "failed to hash password")
	}

	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return utils.WrapError(err, "failed to update password")
	}

	return nil
}

// toUserResponse converts a user model to a response DTO
func (s *authService) toUserResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:              user.ID.String(),
		Username:        user.Username,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Bio:             user.Bio,
		ProfileImageURL: user.ProfileImageURL,
		Role:            string(user.Role),
		IsVerified:      user.IsVerified,
		IsActive:        user.IsActive,
		CreatedAt:       user.CreatedAt,
		LastLoginAt:     user.LastLoginAt,
	}
}
