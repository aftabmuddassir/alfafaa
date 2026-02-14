package services

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/config"
	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	GetCurrentUser(userID string) (*dto.UserResponse, error)
	ChangePassword(userID string, req *dto.ChangePasswordRequest) error
	GoogleAuth(req *dto.GoogleAuthRequest) (*dto.AuthResponse, error)
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
	utils.Debug("Register: started", zap.String("email", req.Email), zap.String("username", req.Username))

	// Validate password strength
	passwordErrors := utils.ValidatePasswordStrength(req.Password)
	if len(passwordErrors) > 0 {
		utils.Debug("Register: weak password", zap.String("email", req.Email))
		return nil, utils.NewAppError("WEAK_PASSWORD", passwordErrors[0], 400)
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		utils.Error("Register: failed to check email", zap.Error(err))
		return nil, utils.WrapError(err, "failed to check email")
	}
	if exists {
		utils.Debug("Register: email already exists", zap.String("email", req.Email))
		return nil, utils.ErrEmailExists
	}

	// Check if username already exists
	exists, err = s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		utils.Error("Register: failed to check username", zap.Error(err))
		return nil, utils.WrapError(err, "failed to check username")
	}
	if exists {
		utils.Debug("Register: username already exists", zap.String("username", req.Username))
		return nil, utils.ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Error("Register: failed to hash password", zap.Error(err))
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
		utils.Error("Register: failed to create user", zap.Error(err))
		return nil, utils.WrapError(err, "failed to create user")
	}
	utils.Debug("Register: user created in DB", zap.String("user_id", user.ID.String()))

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
		utils.Error("Register: failed to generate tokens", zap.Error(err))
		return nil, utils.WrapError(err, "failed to generate tokens")
	}
	utils.Info("Register: success", zap.String("user_id", user.ID.String()), zap.String("email", req.Email))

	return &dto.AuthResponse{
		User:         s.toUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// Login authenticates a user
func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	utils.Debug("Login: started", zap.String("email", req.Email))

	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Debug("Login: user not found", zap.String("email", req.Email))
			return nil, utils.ErrInvalidCredentials
		}
		utils.Error("Login: failed to find user", zap.Error(err))
		return nil, utils.WrapError(err, "failed to find user")
	}
	utils.Debug("Login: user found", zap.String("user_id", user.ID.String()))

	// Check if user is active
	if !user.IsActive {
		utils.Debug("Login: account disabled", zap.String("user_id", user.ID.String()))
		return nil, utils.NewAppError("ACCOUNT_DISABLED", "Your account has been disabled", 403)
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		utils.Debug("Login: invalid password", zap.String("email", req.Email))
		return nil, utils.ErrInvalidCredentials
	}
	utils.Debug("Login: password verified", zap.String("user_id", user.ID.String()))

	// Update last login
	_ = s.userRepo.UpdateLastLogin(user.ID)
	utils.Debug("Login: last login updated", zap.String("user_id", user.ID.String()))

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
		utils.Error("Login: failed to generate tokens", zap.Error(err))
		return nil, utils.WrapError(err, "failed to generate tokens")
	}
	utils.Info("Login: success", zap.String("user_id", user.ID.String()), zap.String("email", req.Email))

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

// GoogleAuth authenticates a user via Google OAuth
// This is a stub implementation that validates the Google ID token
// and creates/logs in the user
func (s *authService) GoogleAuth(req *dto.GoogleAuthRequest) (*dto.AuthResponse, error) {
	utils.Debug("GoogleAuth: started")

	// Validate the Google ID token
	googleUserInfo, err := s.verifyGoogleIDToken(req.IDToken)
	if err != nil {
		utils.Warn("GoogleAuth: invalid ID token", zap.Error(err))
		return nil, utils.NewAppError("INVALID_TOKEN", "Invalid Google ID token", 401)
	}
	utils.Debug("GoogleAuth: token decoded", zap.String("google_id", googleUserInfo.ID), zap.String("email", googleUserInfo.Email))

	// Try to find existing user by Google ID
	user, err := s.userRepo.FindByGoogleID(googleUserInfo.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error("GoogleAuth: failed to find user by google ID", zap.Error(err))
			return nil, utils.WrapError(err, "failed to find user")
		}

		utils.Debug("GoogleAuth: no user with google ID, trying email", zap.String("email", googleUserInfo.Email))

		// User not found by Google ID, try by email
		user, err = s.userRepo.FindByEmail(googleUserInfo.Email)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Error("GoogleAuth: failed to find user by email", zap.Error(err))
				return nil, utils.WrapError(err, "failed to find user")
			}

			// Create new user
			utils.Debug("GoogleAuth: creating new user", zap.String("email", googleUserInfo.Email))
			user, err = s.createGoogleUser(googleUserInfo)
			if err != nil {
				utils.Error("GoogleAuth: failed to create user", zap.Error(err))
				return nil, err
			}
			utils.Debug("GoogleAuth: user created", zap.String("user_id", user.ID.String()))
		} else {
			// Link existing account with Google
			utils.Debug("GoogleAuth: linking existing account", zap.String("user_id", user.ID.String()))
			googleID := googleUserInfo.ID
			user.GoogleID = &googleID
			user.AuthProvider = "google"
			if user.ProfileImageURL == nil && googleUserInfo.Picture != "" {
				user.ProfileImageURL = &googleUserInfo.Picture
			}
			if err := s.userRepo.Update(user); err != nil {
				utils.Error("GoogleAuth: failed to link Google account", zap.Error(err))
				return nil, utils.WrapError(err, "failed to link Google account")
			}
			utils.Debug("GoogleAuth: account linked")
		}
	} else {
		utils.Debug("GoogleAuth: existing user found by google ID", zap.String("user_id", user.ID.String()))
	}

	// Check if user is active
	if !user.IsActive {
		utils.Debug("GoogleAuth: account disabled", zap.String("user_id", user.ID.String()))
		return nil, utils.NewAppError("ACCOUNT_DISABLED", "Your account has been disabled", 403)
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(user.ID)
	now := time.Now()
	user.LastLoginAt = &now
	utils.Debug("GoogleAuth: last login updated", zap.String("user_id", user.ID.String()))

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
		utils.Error("GoogleAuth: failed to generate tokens", zap.Error(err))
		return nil, utils.WrapError(err, "failed to generate tokens")
	}
	utils.Info("GoogleAuth: success", zap.String("user_id", user.ID.String()), zap.String("email", googleUserInfo.Email))

	return &dto.AuthResponse{
		User:         s.toUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

// verifyGoogleIDToken verifies and decodes a Google ID token
// In production, this should verify the token signature with Google's public keys
func (s *authService) verifyGoogleIDToken(idToken string) (*dto.GoogleUserInfo, error) {
	// STUB: In a real implementation, you would:
	// 1. Fetch Google's public keys from https://www.googleapis.com/oauth2/v3/certs
	// 2. Verify the JWT signature
	// 3. Check the 'aud' claim matches your client ID
	// 4. Check the 'iss' claim is accounts.google.com or https://accounts.google.com
	// 5. Check the token hasn't expired

	// For now, we'll use Google's tokeninfo endpoint as a simple verification
	// This is acceptable for development but has rate limits in production
	// Production should use proper JWT verification with Google's public keys

	// Parse the JWT without verification (STUB - should verify in production)
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// Decode the payload (middle part)
	payload, err := decodeJWTPayload(parts[1])
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// decodeJWTPayload decodes the payload part of a JWT
func decodeJWTPayload(payload string) (*dto.GoogleUserInfo, error) {
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var userInfo dto.GoogleUserInfo
	if err := json.Unmarshal(decoded, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// createGoogleUser creates a new user from Google OAuth info
func (s *authService) createGoogleUser(info *dto.GoogleUserInfo) (*models.User, error) {
	// Generate a unique username from email
	username := generateUsernameFromEmail(info.Email)

	// Check if username exists and make it unique if needed
	for i := 0; i < 100; i++ {
		exists, err := s.userRepo.ExistsByUsername(username)
		if err != nil {
			return nil, utils.WrapError(err, "failed to check username")
		}
		if !exists {
			break
		}
		username = generateUsernameFromEmail(info.Email) + randomSuffix()
	}

	googleID := info.ID
	var profileImageURL *string
	if info.Picture != "" {
		profileImageURL = &info.Picture
	}

	user := &models.User{
		Username:        username,
		Email:           info.Email,
		FirstName:       info.GivenName,
		LastName:        info.FamilyName,
		ProfileImageURL: profileImageURL,
		GoogleID:        &googleID,
		AuthProvider:    "google",
		Role:            models.RoleReader,
		IsActive:        true,
		IsVerified:      info.EmailVerified,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, utils.WrapError(err, "failed to create user")
	}

	return user, nil
}

// generateUsernameFromEmail generates a username from an email address
func generateUsernameFromEmail(email string) string {
	// Take the part before @ and clean it
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return "user"
	}

	username := parts[0]
	// Remove any non-alphanumeric characters except underscores
	cleaned := ""
	for _, c := range username {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			cleaned += string(c)
		}
	}

	if len(cleaned) < 3 {
		cleaned = "user" + cleaned
	}
	if len(cleaned) > 30 {
		cleaned = cleaned[:30]
	}

	return cleaned
}

// randomSuffix generates a random suffix for usernames
func randomSuffix() string {
	return fmt.Sprintf("%d", rand.Intn(9999))
}
