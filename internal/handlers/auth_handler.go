package handlers

import (
	"net/http"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/middlewares"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} utils.Response{data=dto.AuthResponse} "Registration successful"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 409 {object} utils.Response "Email or username already exists"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Registration successful", response)
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} utils.Response{data=dto.AuthResponse} "Login successful"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} utils.Response{data=dto.TokenResponse} "Token refreshed successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Invalid or expired refresh token"
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", response)
}

// GetMe returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's profile
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.UserResponse} "User retrieved successfully"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	if userID == "" {
		utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	response, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", response)
}

// ChangePassword handles password change
// @Summary Change password
// @Description Change the current user's password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} utils.Response "Password changed successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized or incorrect current password"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	if userID == "" {
		utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	if err := h.authService.ChangePassword(userID, &req); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}

// Logout handles user logout
// @Summary Logout user
// @Description Logout the current user (client should discard tokens)
// @Tags auth
// @Produce json
// @Success 200 {object} utils.Response "Logout successful"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is typically handled client-side
	// by removing the token from storage. This endpoint exists for
	// API completeness and could be extended to blacklist tokens.
	utils.SuccessResponse(c, http.StatusOK, "Logout successful", nil)
}
