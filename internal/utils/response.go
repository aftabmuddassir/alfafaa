package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ResponseWithMeta represents a paginated API response (for swagger documentation)
type ResponseWithMeta struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool         `json:"success"`
	Error   ErrorDetails `json:"error"`
}

// ErrorDetails contains detailed error information
type ErrorDetails struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Meta contains pagination metadata
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessResponseWithMeta sends a success response with pagination metadata
func SuccessResponseWithMeta(c *gin.Context, status int, message string, data interface{}, meta *Meta) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// ErrorResponseJSON sends an error response
func ErrorResponseJSON(c *gin.Context, status int, code, message string, details []ValidationError) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// HandleError handles an error and sends appropriate response
func HandleError(c *gin.Context, err error) {
	if appErr, ok := IsAppError(err); ok {
		ErrorResponseJSON(c, appErr.Status, appErr.Code, appErr.Message, nil)
		return
	}
	// Log the actual error for debugging (in production, use proper logging)
	ErrorResponseJSON(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", nil)
}

// HandleValidationError handles validation errors
func HandleValidationError(c *gin.Context, errors []ValidationError) {
	ErrorResponseJSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", errors)
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(total int64, perPage int) int {
	if perPage <= 0 {
		return 0
	}
	pages := int(total) / perPage
	if int(total)%perPage > 0 {
		pages++
	}
	return pages
}

// NewMeta creates a new Meta object for pagination
func NewMeta(page, perPage int, total int64) *Meta {
	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: CalculateTotalPages(total, perPage),
	}
}
