package handlers

import (
	"net/http"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	categoryService services.CategoryService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories returns all categories
// @Summary List categories
// @Description Get all categories (optionally in hierarchical structure)
// @Tags categories
// @Produce json
// @Param hierarchical query bool false "Return categories in hierarchical structure" default(false)
// @Success 200 {object} utils.Response{data=[]dto.CategoryResponse} "Categories retrieved successfully"
// @Router /categories [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	var query dto.CategoryListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	// Check if hierarchical response is requested
	if c.Query("hierarchical") == "true" {
		categories, err := h.categoryService.GetCategoriesHierarchical()
		if err != nil {
			utils.HandleError(c, err)
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "Categories retrieved successfully", categories)
		return
	}

	categories, err := h.categoryService.GetCategories(&query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Categories retrieved successfully", categories)
}

// GetCategory returns a single category by slug
// @Summary Get category by slug
// @Description Get a single category by its slug
// @Tags categories
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} utils.Response{data=dto.CategoryResponse} "Category retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid slug"
// @Failure 404 {object} utils.Response "Category not found"
// @Router /categories/{slug} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_SLUG", "Category slug is required", nil)
		return
	}

	category, err := h.categoryService.GetCategory(slug)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category retrieved successfully", category)
}

// CreateCategory creates a new category
// @Summary Create category
// @Description Create a new category (requires editor role)
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCategoryRequest true "Category data"
// @Success 201 {object} utils.Response{data=dto.CategoryResponse} "Category created successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 409 {object} utils.Response "Category already exists"
// @Router /categories [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Category created successfully", category)
}

// UpdateCategory updates a category
// @Summary Update category
// @Description Update a category (requires editor role)
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID (UUID)"
// @Param request body dto.UpdateCategoryRequest true "Category update data"
// @Success 200 {object} utils.Response{data=dto.CategoryResponse} "Category updated successfully"
// @Failure 400 {object} utils.Response "Validation error"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 404 {object} utils.Response "Category not found"
// @Router /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Category ID is required", nil)
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	category, err := h.categoryService.UpdateCategory(id, &req)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category updated successfully", category)
}

// DeleteCategory deletes a category
// @Summary Delete category
// @Description Delete a category (requires editor role)
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID (UUID)"
// @Success 200 {object} utils.Response "Category deleted successfully"
// @Failure 400 {object} utils.Response "Invalid ID"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 403 {object} utils.Response "Forbidden - requires editor role"
// @Failure 404 {object} utils.Response "Category not found"
// @Router /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_ID", "Category ID is required", nil)
		return
	}

	if err := h.categoryService.DeleteCategory(id); err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Category deleted successfully", nil)
}

// GetCategoryArticles returns articles in a category
// @Summary Get category articles
// @Description Get a paginated list of articles in a category
// @Tags categories
// @Produce json
// @Param slug path string true "Category slug"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.ResponseWithMeta{data=[]dto.ArticleListResponse} "Articles retrieved successfully"
// @Failure 400 {object} utils.Response "Invalid slug"
// @Failure 404 {object} utils.Response "Category not found"
// @Router /categories/{slug}/articles [get]
func (h *CategoryHandler) GetCategoryArticles(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "INVALID_SLUG", "Category slug is required", nil)
		return
	}

	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	articles, total, err := h.categoryService.GetCategoryArticles(slug, &query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	meta := utils.NewMeta(query.GetPage(), query.GetPerPage(), total)
	utils.SuccessResponseWithMeta(c, http.StatusOK, "Articles retrieved successfully", articles, meta)
}
