package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(token string) (*dto.TokenResponse, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TokenResponse), args.Error(1)
}

func (m *MockAuthService) GetCurrentUser(userID string) (*dto.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockAuthService) ChangePassword(userID string, req *dto.ChangePasswordRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

type AuthHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	mockService *MockAuthService
	handler     *AuthHandler
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.mockService = new(MockAuthService)
	suite.handler = NewAuthHandler(suite.mockService)
	suite.router = gin.New()

	// Setup routes
	auth := suite.router.Group("/api/v1/auth")
	{
		auth.POST("/register", suite.handler.Register)
		auth.POST("/login", suite.handler.Login)
		auth.POST("/refresh-token", suite.handler.RefreshToken)
		auth.GET("/me", suite.withAuth(), suite.handler.GetMe)
		auth.POST("/change-password", suite.withAuth(), suite.handler.ChangePassword)
		auth.POST("/logout", suite.handler.Logout)
	}
}

// withAuth is a test middleware that sets user context
func (suite *AuthHandlerTestSuite) withAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from header for testing
		userID := c.GetHeader("X-User-ID")
		if userID != "" {
			c.Set("userID", userID)
			c.Set("userEmail", "test@example.com")
			c.Set("userRole", "reader")
		}
		c.Next()
	}
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

// Register Tests

func (suite *AuthHandlerTestSuite) TestRegister_Success() {
	reqBody := dto.RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "Password123!",
	}

	authResponse := &dto.AuthResponse{
		User: dto.UserResponse{
			ID:       uuid.New().String(),
			Username: "newuser",
			Email:    "new@example.com",
			Role:     "reader",
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	suite.mockService.On("Register", mock.AnythingOfType("*dto.RegisterRequest")).Return(authResponse, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), true, response["success"])
	assert.Equal(suite.T(), "Registration successful", response["message"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestRegister_ValidationError() {
	reqBody := map[string]string{
		"username": "",
		"email":    "invalid-email",
		"password": "weak",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), false, response["success"])
}

func (suite *AuthHandlerTestSuite) TestRegister_DuplicateEmail() {
	reqBody := dto.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "Password123!",
	}

	suite.mockService.On("Register", mock.AnythingOfType("*dto.RegisterRequest")).
		Return(nil, utils.ErrConflict)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)
	suite.mockService.AssertExpectations(suite.T())
}

// Login Tests

func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}

	authResponse := &dto.AuthResponse{
		User: dto.UserResponse{
			ID:       uuid.New().String(),
			Username: "testuser",
			Email:    "test@example.com",
			Role:     "reader",
		},
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	suite.mockService.On("Login", mock.AnythingOfType("*dto.LoginRequest")).Return(authResponse, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), true, response["success"])
	assert.Equal(suite.T(), "Login successful", response["message"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogin_InvalidCredentials() {
	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	suite.mockService.On("Login", mock.AnythingOfType("*dto.LoginRequest")).
		Return(nil, utils.ErrUnauthorized)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogin_ValidationError() {
	reqBody := map[string]string{
		"email":    "",
		"password": "",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// RefreshToken Tests

func (suite *AuthHandlerTestSuite) TestRefreshToken_Success() {
	reqBody := dto.RefreshTokenRequest{
		RefreshToken: "valid-refresh-token",
	}

	tokenResponse := &dto.TokenResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresAt:    1706200000,
	}

	suite.mockService.On("RefreshToken", "valid-refresh-token").Return(tokenResponse, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), true, response["success"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestRefreshToken_InvalidToken() {
	reqBody := dto.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	}

	suite.mockService.On("RefreshToken", "invalid-token").
		Return(nil, utils.ErrUnauthorized)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/refresh-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	suite.mockService.AssertExpectations(suite.T())
}

// GetMe Tests

func (suite *AuthHandlerTestSuite) TestGetMe_Success() {
	userID := uuid.New().String()
	userResponse := &dto.UserResponse{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "reader",
	}

	suite.mockService.On("GetCurrentUser", userID).Return(userResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("X-User-ID", userID)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), true, response["success"])
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestGetMe_Unauthorized() {
	req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
	// No X-User-ID header

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// ChangePassword Tests

func (suite *AuthHandlerTestSuite) TestChangePassword_Success() {
	userID := uuid.New().String()
	reqBody := dto.ChangePasswordRequest{
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword123!",
	}

	suite.mockService.On("ChangePassword", userID, mock.AnythingOfType("*dto.ChangePasswordRequest")).Return(nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	suite.mockService.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestChangePassword_Unauthorized() {
	reqBody := dto.ChangePasswordRequest{
		CurrentPassword: "OldPassword123!",
		NewPassword:     "NewPassword123!",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-User-ID header

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthHandlerTestSuite) TestChangePassword_WrongPassword() {
	userID := uuid.New().String()
	reqBody := dto.ChangePasswordRequest{
		CurrentPassword: "WrongPassword!",
		NewPassword:     "NewPassword123!",
	}

	suite.mockService.On("ChangePassword", userID, mock.AnythingOfType("*dto.ChangePasswordRequest")).
		Return(utils.ErrUnauthorized)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	suite.mockService.AssertExpectations(suite.T())
}

// Logout Tests

func (suite *AuthHandlerTestSuite) TestLogout_Success() {
	req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), true, response["success"])
	assert.Equal(suite.T(), "Logout successful", response["message"])
}

