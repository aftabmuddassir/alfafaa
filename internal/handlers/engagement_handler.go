package handlers

import (
	"net/http"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/middlewares"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// EngagementHandler handles engagement-related HTTP requests (likes, bookmarks, comments, notifications)
type EngagementHandler struct {
	engagementService services.EngagementService
}

// NewEngagementHandler creates a new engagement handler
func NewEngagementHandler(engagementService services.EngagementService) *EngagementHandler {
	return &EngagementHandler{
		engagementService: engagementService,
	}
}

// --- Likes ---

// LikeArticle handles liking an article
// @Summary Like an article
// @Description Like an article (idempotent - if already liked, returns current status)
// @Tags engagement
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Success 200 {object} utils.Response{data=dto.LikeResponse} "Article liked"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/like [post]
func (h *EngagementHandler) LikeArticle(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")

	response, err := h.engagementService.LikeArticle(userID, slug)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Article liked", response)
}

// UnlikeArticle handles unliking an article
// @Summary Unlike an article
// @Description Remove like from an article
// @Tags engagement
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Success 200 {object} utils.Response{data=dto.LikeResponse} "Like removed"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/like [delete]
func (h *EngagementHandler) UnlikeArticle(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")

	response, err := h.engagementService.UnlikeArticle(userID, slug)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Like removed", response)
}

// GetLikeStatus handles getting the like status of an article
// @Summary Get like status
// @Description Get current user's like status for an article
// @Tags engagement
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Success 200 {object} utils.Response{data=dto.LikeResponse} "Like status retrieved"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/like [get]
func (h *EngagementHandler) GetLikeStatus(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")

	response, err := h.engagementService.GetLikeStatus(userID, slug)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Like status retrieved", response)
}

// --- Bookmarks ---

// BookmarkArticle handles bookmarking an article
// @Summary Bookmark an article
// @Description Bookmark an article (idempotent)
// @Tags engagement
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Success 200 {object} utils.Response{data=dto.BookmarkResponse} "Article bookmarked"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/bookmark [post]
func (h *EngagementHandler) BookmarkArticle(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")

	response, err := h.engagementService.BookmarkArticle(userID, slug)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Article bookmarked", response)
}

// UnbookmarkArticle handles removing a bookmark
// @Summary Remove bookmark
// @Description Remove bookmark from an article
// @Tags engagement
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Success 200 {object} utils.Response{data=dto.BookmarkResponse} "Bookmark removed"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/bookmark [delete]
func (h *EngagementHandler) UnbookmarkArticle(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")

	response, err := h.engagementService.UnbookmarkArticle(userID, slug)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Bookmark removed", response)
}

// GetBookmarkedArticles handles getting bookmarked articles
// @Summary Get bookmarked articles
// @Description Get current user's bookmarked articles
// @Tags engagement
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.ArticleListItemResponse} "Bookmarked articles retrieved"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /users/bookmarks [get]
func (h *EngagementHandler) GetBookmarkedArticles(c *gin.Context) {
	userID := middlewares.GetUserID(c)

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	articles, total, err := h.engagementService.GetBookmarkedArticles(userID, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Bookmarked articles retrieved", articles, meta)
}

// --- Comments ---

// GetComments handles getting comments for an article
// @Summary Get article comments
// @Description Get comments for an article with nested replies
// @Tags engagement
// @Produce json
// @Param slug path string true "Article slug"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.EngagementCommentResponse} "Comments retrieved"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/comments [get]
func (h *EngagementHandler) GetComments(c *gin.Context) {
	slug := c.Param("slug")

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	comments, total, err := h.engagementService.GetComments(slug, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Comments retrieved", comments, meta)
}

// CreateComment handles creating a comment
// @Summary Create a comment
// @Description Create a comment on an article (optionally as a reply)
// @Tags engagement
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Param request body dto.CreateCommentRequest true "Comment data"
// @Success 201 {object} utils.Response{data=dto.EngagementCommentResponse} "Comment created"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/comments [post]
func (h *EngagementHandler) CreateComment(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	comment, err := h.engagementService.CreateComment(userID, slug, &req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Comment created", comment)
}

// UpdateComment handles updating a comment
// @Summary Update a comment
// @Description Update a comment (owner only)
// @Tags engagement
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Param id path string true "Comment ID"
// @Param request body dto.UpdateCommentRequest true "Updated comment data"
// @Success 200 {object} utils.Response{data=dto.EngagementCommentResponse} "Comment updated"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden"
// @Failure 404 {object} utils.Response "Comment not found"
// @Router /articles/{slug}/comments/{id} [put]
func (h *EngagementHandler) UpdateComment(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")
	commentID := c.Param("id")

	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	comment, err := h.engagementService.UpdateComment(userID, slug, commentID, &req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment updated", comment)
}

// DeleteComment handles deleting a comment
// @Summary Delete a comment
// @Description Delete a comment (owner or admin)
// @Tags engagement
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Article slug"
// @Param id path string true "Comment ID"
// @Success 200 {object} utils.Response "Comment deleted"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden"
// @Failure 404 {object} utils.Response "Comment not found"
// @Router /articles/{slug}/comments/{id} [delete]
func (h *EngagementHandler) DeleteComment(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	slug := c.Param("slug")
	commentID := c.Param("id")
	isAdmin := middlewares.GetUserRole(c) == "admin"

	if err := h.engagementService.DeleteComment(userID, slug, commentID, isAdmin); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Comment deleted", nil)
}

// --- Notifications ---

// GetNotifications handles getting notifications
// @Summary Get notifications
// @Description Get current user's notifications, newest first
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.NotificationResponse} "Notifications retrieved"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /notifications [get]
func (h *EngagementHandler) GetNotifications(c *gin.Context) {
	userID := middlewares.GetUserID(c)

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	notifications, total, err := h.engagementService.GetNotifications(userID, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Notifications retrieved", notifications, meta)
}

// GetUnreadCount handles getting unread notification count
// @Summary Get unread notification count
// @Description Get count of unread notifications
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=dto.UnreadCountResponse} "Unread count retrieved"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /notifications/unread-count [get]
func (h *EngagementHandler) GetUnreadCount(c *gin.Context) {
	userID := middlewares.GetUserID(c)

	response, err := h.engagementService.GetUnreadCount(userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Unread count retrieved", response)
}

// MarkNotificationAsRead handles marking a notification as read
// @Summary Mark notification as read
// @Description Mark a single notification as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} utils.Response "Notification marked as read"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "Notification not found"
// @Router /notifications/{id}/read [put]
func (h *EngagementHandler) MarkNotificationAsRead(c *gin.Context) {
	userID := middlewares.GetUserID(c)
	notificationID := c.Param("id")

	if err := h.engagementService.MarkNotificationAsRead(userID, notificationID); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification marked as read", nil)
}

// MarkAllNotificationsAsRead handles marking all notifications as read
// @Summary Mark all notifications as read
// @Description Mark all current user's notifications as read
// @Tags notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response "All notifications marked as read"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Router /notifications/read-all [put]
func (h *EngagementHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	userID := middlewares.GetUserID(c)

	if err := h.engagementService.MarkAllNotificationsAsRead(userID); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "All notifications marked as read", nil)
}
