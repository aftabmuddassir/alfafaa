package handlers

import (
	"net/http"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/middlewares"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// UserActionHandler handles user social actions (follow, interests, feed)
type UserActionHandler struct {
	userService services.UserService
}

// NewUserActionHandler creates a new user action handler
func NewUserActionHandler(userService services.UserService) *UserActionHandler {
	return &UserActionHandler{
		userService: userService,
	}
}

// FollowUser handles following a user
// @Summary Follow a user
// @Description Follow another user to see their articles in your feed
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID to follow"
// @Success 200 {object} utils.Response{data=dto.FollowResponse} "Followed successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id}/follow [post]
func (h *UserActionHandler) FollowUser(c *gin.Context) {
	currentUserID := middlewares.GetUserID(c)
	if currentUserID == "" {
		utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "BAD_REQUEST", "User ID is required", nil)
		return
	}

	response, err := h.userService.FollowUser(currentUserID, targetUserID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Followed successfully", response)
}

// UnfollowUser handles unfollowing a user
// @Summary Unfollow a user
// @Description Unfollow a user to stop seeing their articles in your feed
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID to unfollow"
// @Success 200 {object} utils.Response{data=dto.FollowResponse} "Unfollowed successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /users/{id}/unfollow [post]
func (h *UserActionHandler) UnfollowUser(c *gin.Context) {
	currentUserID := middlewares.GetUserID(c)
	if currentUserID == "" {
		utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	targetUserID := c.Param("id")
	if targetUserID == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "BAD_REQUEST", "User ID is required", nil)
		return
	}

	response, err := h.userService.UnfollowUser(currentUserID, targetUserID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Unfollowed successfully", response)
}

// GetFollowers handles getting a user's followers
// @Summary Get user followers
// @Description Get a list of users who follow the specified user
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} utils.Response{data=dto.FollowListResponse} "Followers retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id}/followers [get]
func (h *UserActionHandler) GetFollowers(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "BAD_REQUEST", "User ID is required", nil)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	response, err := h.userService.GetFollowers(userID, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Followers retrieved successfully", response)
}

// GetFollowing handles getting users that a user follows
// @Summary Get users followed by a user
// @Description Get a list of users that the specified user follows
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} utils.Response{data=dto.FollowListResponse} "Following list retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id}/following [get]
func (h *UserActionHandler) GetFollowing(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "BAD_REQUEST", "User ID is required", nil)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	response, err := h.userService.GetFollowing(userID, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Following list retrieved successfully", response)
}

// SetInterests handles setting user interests (categories for onboarding)
// @Summary Set user interests
// @Description Set the user's interests (categories) for personalized feed
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SetInterestsRequest true "Category IDs"
// @Success 200 {object} utils.Response{data=[]dto.CategoryResponse} "Interests set successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /users/interests [post]
func (h *UserActionHandler) SetInterests(c *gin.Context) {
	currentUserID := middlewares.GetUserID(c)
	if currentUserID == "" {
		utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req dto.SetInterestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	response, err := h.userService.SetInterests(currentUserID, &req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Interests set successfully", response)
}

// GetInterests handles getting user interests
// @Summary Get user interests
// @Description Get the user's interests (categories)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]dto.CategoryResponse} "Interests retrieved successfully"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /users/interests [get]
func (h *UserActionHandler) GetInterests(c *gin.Context) {
	currentUserID := middlewares.GetUserID(c)
	if currentUserID == "" {
		utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	response, err := h.userService.GetInterests(currentUserID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Interests retrieved successfully", response)
}

// GetPersonalizedFeed handles getting personalized article feed
// @Summary Get personalized feed
// @Description Get articles from followed authors and interested categories
// @Tags articles
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param sort query string false "Sort by: newest, oldest, popular" default(newest)
// @Success 200 {object} utils.Response{data=[]dto.ArticleListItemResponse} "Feed retrieved successfully"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /articles/feed [get]
func (h *UserActionHandler) GetPersonalizedFeed(c *gin.Context) {
	currentUserID := middlewares.GetUserID(c)
	if currentUserID == "" {
		utils.ErrorResponseJSON(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	articles, total, err := h.userService.GetPersonalizedFeed(currentUserID, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Feed retrieved successfully", articles, &utils.Meta{
		Page:       query.GetPage(),
		PerPage:    query.GetPerPage(),
		Total:      total,
		TotalPages: utils.CalculateTotalPages(total, query.GetPerPage()),
	})
}

// GetStaffPicks handles getting staff-picked articles
// @Summary Get staff picks
// @Description Get articles marked as staff picks
// @Tags articles
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Param sort query string false "Sort by: newest, oldest, popular" default(newest)
// @Success 200 {object} utils.Response{data=[]dto.ArticleListItemResponse} "Staff picks retrieved successfully"
// @Router /articles/staff-picks [get]
func (h *UserActionHandler) GetStaffPicks(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	articles, total, err := h.userService.GetStaffPicks(&query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponseWithMeta(c, http.StatusOK, "Staff picks retrieved successfully", articles, &utils.Meta{
		Page:       query.GetPage(),
		PerPage:    query.GetPerPage(),
		Total:      total,
		TotalPages: utils.CalculateTotalPages(total, query.GetPerPage()),
	})
}

// GetUserProfile handles getting a user's public profile with social info
// @Summary Get user profile
// @Description Get a user's public profile including follower count, following count, etc.
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} utils.Response{data=dto.UserProfileResponse} "Profile retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid request"
// @Failure 404 {object} utils.Response "User not found"
// @Router /users/{id}/profile [get]
func (h *UserActionHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "BAD_REQUEST", "User ID is required", nil)
		return
	}

	// Get current user ID for checking if they're following this user
	currentUserID := middlewares.GetUserID(c)

	response, err := h.userService.GetUserProfile(userID, currentUserID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", response)
}
