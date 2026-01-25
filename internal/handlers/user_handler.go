package handlers

import (
	"net/http"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/middlewares"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUsers returns a list of users
// @Summary List users
// @Description Get a paginated list of users (requires editor or admin role)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param search query string false "Search by username or email"
// @Param role query string false "Filter by role"
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.UserResponse} "Users retrieved successfully"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Router /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	var query dto.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	users, total, err := h.userService.GetUsers(&query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Users retrieved successfully", users, meta)
}

// GetUser returns a single user by ID
// @Summary Get user by ID
// @Description Get a user's public profile by their ID
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} utils.Response{data=dto.PublicUserResponse} "User retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "User ID is required", nil)
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User retrieved successfully", user)
}

// UpdateUser updates a user
// @Summary Update user
// @Description Update a user's profile (users can only update their own profile unless admin)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} utils.Response{data=dto.UserResponse} "User updated successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - can only update own profile"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "User ID is required", nil)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	currentUserID := middlewares.GetUserID(c)
	isAdmin := middlewares.IsAdmin(c)

	user, err := h.userService.UpdateUser(id, &req, currentUserID, isAdmin)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", user)
}

// AdminUpdateUser updates a user with admin privileges
// @Summary Admin update user
// @Description Update any user's profile including role and status (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body dto.AdminUpdateUserRequest true "Admin user update data"
// @Success 200 {object} utils.Response{data=dto.UserResponse} "User updated successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires admin role"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id}/admin [put]
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "User ID is required", nil)
		return
	}

	var req dto.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	user, err := h.userService.AdminUpdateUser(id, &req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User updated successfully", user)
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Delete a user account (admin only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} utils.Response "User deleted successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires admin role"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "User ID is required", nil)
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}

// GetUserArticles returns articles by a user
// @Summary Get user's articles
// @Description Get a paginated list of published articles by a user
// @Tags users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.ArticleListResponse} "Articles retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id}/articles [get]
func (h *UserHandler) GetUserArticles(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "User ID is required", nil)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	articles, total, err := h.userService.GetUserArticles(id, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Articles retrieved successfully", articles, meta)
}
