package handlers

import (
	"net/http"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/services"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/gin-gonic/gin"
)

// SearchHandler handles search-related HTTP requests
type SearchHandler struct {
	searchService services.SearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search performs a search across articles, categories, and tags
// @Summary Search
// @Description Search across articles, categories, and tags
// @Tags search
// @Produce json
// @Param q query string true "Search query"
// @Param type query string false "Filter by type (articles, categories, tags)" Enums(articles, categories, tags)
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=dto.SearchResponse} "Search completed successfully"
// @Failure 400 {object} utils.Response "Search query is required"
// @Router /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	var query dto.SearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.HandleValidationError(c, utils.ParseValidationErrors(err))
		return
	}

	if query.Query == "" {
		utils.ErrorResponseJSON(c, http.StatusBadRequest, "MISSING_QUERY", "Search query is required", nil)
		return
	}

	results, err := h.searchService.Search(&query)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Search completed successfully", results)
}
