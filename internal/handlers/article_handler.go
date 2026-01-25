package handlers

import (
	"net/http"
	"strconv"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/middlewares"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// ArticleHandler handles article-related HTTP requests
type ArticleHandler struct {
	articleService services.ArticleService
}

// NewArticleHandler creates a new article handler
func NewArticleHandler(articleService services.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
	}
}

// GetArticles returns a list of articles
// @Summary List articles
// @Description Get a paginated list of published articles (editors can see all)
// @Tags articles
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param category query string false "Filter by category slug"
// @Param tag query string false "Filter by tag slug"
// @Param author query string false "Filter by author ID"
// @Param search query string false "Search in title and content"
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.ArticleListResponse} "Articles retrieved successfully"
// @Router /articles [get]
func (h *ArticleHandler) GetArticles(c *gin.Context) {
	var query dto.ArticleListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	// Check if user can see unpublished articles
	includeUnpublished := middlewares.IsEditor(c)

	articles, total, err := h.articleService.GetArticles(&query, includeUnpublished)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Articles retrieved successfully", articles, meta)
}

// GetArticle returns a single article by slug
// @Summary Get article by slug
// @Description Get a single article by its slug (increments view count for non-authenticated users)
// @Tags articles
// @Produce json
// @Param slug path string true "Article slug"
// @Success 200 {object} utils.Response{data=dto.ArticleDetailResponse} "Article retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid slug"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug} [get]
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_SLUG", "Article slug is required", nil)
		return
	}

	// Increment view count for public access
	incrementView := !middlewares.IsAuthenticated(c)

	article, err := h.articleService.GetArticle(slug, incrementView)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Article retrieved successfully", article)
}

// CreateArticle creates a new article
// @Summary Create article
// @Description Create a new article (requires author role or higher)
// @Tags articles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateArticleRequest true "Article data"
// @Success 201 {object} utils.Response{data=dto.ArticleDetailResponse} "Article created successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires author role"
// @Router /articles [post]
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req dto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	authorID := middlewares.GetUserID(c)

	article, err := h.articleService.CreateArticle(&req, authorID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Article created successfully", article)
}

// UpdateArticle updates an article
// @Summary Update article
// @Description Update an article (authors can only update their own, editors can update any)
// @Tags articles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Article ID (UUID)"
// @Param request body dto.UpdateArticleRequest true "Article update data"
// @Success 200 {object} utils.Response{data=dto.ArticleDetailResponse} "Article updated successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - not the author"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{id} [put]
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Article ID is required", nil)
		return
	}

	var req dto.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	userID := middlewares.GetUserID(c)
	isEditor := middlewares.IsEditor(c)

	article, err := h.articleService.UpdateArticle(id, &req, userID, isEditor)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Article updated successfully", article)
}

// DeleteArticle deletes an article
// @Summary Delete article
// @Description Delete an article (authors can only delete their own, editors can delete any)
// @Tags articles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Article ID (UUID)"
// @Success 200 {object} utils.Response "Article deleted successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - not the author"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{id} [delete]
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Article ID is required", nil)
		return
	}

	userID := middlewares.GetUserID(c)
	isEditor := middlewares.IsEditor(c)

	if err := h.articleService.DeleteArticle(id, userID, isEditor); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Article deleted successfully", nil)
}

// PublishArticle publishes an article
// @Summary Publish article
// @Description Publish an article (requires editor role)
// @Tags articles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Article ID (UUID)"
// @Success 200 {object} utils.Response{data=dto.ArticleDetailResponse} "Article published successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{id}/publish [patch]
func (h *ArticleHandler) PublishArticle(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Article ID is required", nil)
		return
	}

	article, err := h.articleService.PublishArticle(id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Article published successfully", article)
}

// UnpublishArticle unpublishes an article
// @Summary Unpublish article
// @Description Unpublish an article (requires editor role)
// @Tags articles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Article ID (UUID)"
// @Success 200 {object} utils.Response{data=dto.ArticleDetailResponse} "Article unpublished successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{id}/unpublish [patch]
func (h *ArticleHandler) UnpublishArticle(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Article ID is required", nil)
		return
	}

	article, err := h.articleService.UnpublishArticle(id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Article unpublished successfully", article)
}

// GetTrendingArticles returns trending articles
// @Summary Get trending articles
// @Description Get articles sorted by view count
// @Tags articles
// @Produce json
// @Param limit query int false "Number of articles to return" default(10)
// @Success 200 {object} utils.Response{data=[]dto.ArticleListResponse} "Trending articles retrieved successfully"
// @Router /articles/trending [get]
func (h *ArticleHandler) GetTrendingArticles(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	articles, err := h.articleService.GetTrendingArticles(limit)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Trending articles retrieved successfully", articles)
}

// GetRecentArticles returns recent articles
// @Summary Get recent articles
// @Description Get the most recently published articles
// @Tags articles
// @Produce json
// @Param limit query int false "Number of articles to return" default(10)
// @Success 200 {object} utils.Response{data=[]dto.ArticleListResponse} "Recent articles retrieved successfully"
// @Router /articles/recent [get]
func (h *ArticleHandler) GetRecentArticles(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	articles, err := h.articleService.GetRecentArticles(limit)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Recent articles retrieved successfully", articles)
}

// GetRelatedArticles returns related articles
// @Summary Get related articles
// @Description Get articles related to a specific article based on categories and tags
// @Tags articles
// @Produce json
// @Param slug path string true "Article slug"
// @Param limit query int false "Number of articles to return" default(5)
// @Success 200 {object} utils.Response{data=[]dto.ArticleListResponse} "Related articles retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid slug"
// @Failure 404 {object} utils.Response "Article not found"
// @Router /articles/{slug}/related [get]
func (h *ArticleHandler) GetRelatedArticles(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_SLUG", "Article slug is required", nil)
		return
	}

	limit := 5
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	articles, err := h.articleService.GetRelatedArticles(slug, limit)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Related articles retrieved successfully", articles)
}
