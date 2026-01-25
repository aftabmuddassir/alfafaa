package handlers

import (
	"net/http"
	"strconv"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	tagService services.TagService
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagService services.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// GetTags returns all tags
// @Summary List tags
// @Description Get a paginated list of all tags
// @Tags tags
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Param search query string false "Search by tag name"
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.TagResponse} "Tags retrieved successfully"
// @Router /tags [get]
func (h *TagHandler) GetTags(c *gin.Context) {
	var query dto.TagListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	tags, total, err := h.tagService.GetTags(&query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Tags retrieved successfully", tags, meta)
}

// GetTag returns a single tag by slug
// @Summary Get tag by slug
// @Description Get a single tag by its slug
// @Tags tags
// @Produce json
// @Param slug path string true "Tag slug"
// @Success 200 {object} utils.Response{data=dto.TagResponse} "Tag retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid slug"
// @Failure 404 {object} utils.Response "Tag not found"
// @Router /tags/{slug} [get]
func (h *TagHandler) GetTag(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_SLUG", "Tag slug is required", nil)
		return
	}

	tag, err := h.tagService.GetTag(slug)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag retrieved successfully", tag)
}

// CreateTag creates a new tag
// @Summary Create tag
// @Description Create a new tag (requires editor role)
// @Tags tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTagRequest true "Tag data"
// @Success 201 {object} utils.Response{data=dto.TagResponse} "Tag created successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 409 {object} utils.Response "Tag already exists"
// @Router /tags [post]
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	tag, err := h.tagService.CreateTag(&req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Tag created successfully", tag)
}

// UpdateTag updates a tag
// @Summary Update tag
// @Description Update a tag (requires editor role)
// @Tags tags
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tag ID (UUID)"
// @Param request body dto.UpdateTagRequest true "Tag update data"
// @Success 200 {object} utils.Response{data=dto.TagResponse} "Tag updated successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 404 {object} utils.Response "Tag not found"
// @Router /tags/{id} [put]
func (h *TagHandler) UpdateTag(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Tag ID is required", nil)
		return
	}

	var req dto.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	tag, err := h.tagService.UpdateTag(id, &req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag updated successfully", tag)
}

// DeleteTag deletes a tag
// @Summary Delete tag
// @Description Delete a tag (requires editor role)
// @Tags tags
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tag ID (UUID)"
// @Success 200 {object} utils.Response "Tag deleted successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 404 {object} utils.Response "Tag not found"
// @Router /tags/{id} [delete]
func (h *TagHandler) DeleteTag(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Tag ID is required", nil)
		return
	}

	if err := h.tagService.DeleteTag(id); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Tag deleted successfully", nil)
}

// GetTagArticles returns articles with a tag
// @Summary Get tag articles
// @Description Get a paginated list of articles with a specific tag
// @Tags tags
// @Produce json
// @Param slug path string true "Tag slug"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.ArticleListResponse} "Articles retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid slug"
// @Failure 404 {object} utils.Response "Tag not found"
// @Router /tags/{slug}/articles [get]
func (h *TagHandler) GetTagArticles(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_SLUG", "Tag slug is required", nil)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	articles, total, err := h.tagService.GetTagArticles(slug, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Articles retrieved successfully", articles, meta)
}

// GetPopularTags returns popular tags
// @Summary Get popular tags
// @Description Get the most used tags sorted by usage count
// @Tags tags
// @Produce json
// @Param limit query int false "Number of tags to return" default(10)
// @Success 200 {object} utils.Response{data=[]dto.TagResponse} "Popular tags retrieved successfully"
// @Router /tags/popular [get]
func (h *TagHandler) GetPopularTags(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	tags, err := h.tagService.GetPopularTags(limit)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Popular tags retrieved successfully", tags)
}
