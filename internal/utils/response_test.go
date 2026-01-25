package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccessResponse_Format(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"key": "value"}
	SuccessResponse(c, http.StatusOK, "Success message", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Success message", response.Message)
	assert.NotNil(t, response.Data)
}

func TestSuccessResponse_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SuccessResponse(c, http.StatusOK, "No data", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestSuccessResponseWithMeta_Format(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := []string{"item1", "item2"}
	meta := NewMeta(1, 10, 100)

	SuccessResponseWithMeta(c, http.StatusOK, "Paginated data", data, meta)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Meta)
	assert.Equal(t, 1, response.Meta.Page)
	assert.Equal(t, 10, response.Meta.PerPage)
	assert.Equal(t, int64(100), response.Meta.Total)
	assert.Equal(t, 10, response.Meta.TotalPages)
}

func TestErrorResponseJSON_Format(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ErrorResponseJSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Equal(t, "Validation failed", response.Error.Message)
}

func TestErrorResponseJSON_WithDetails(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	details := []ValidationError{
		{Field: "email", Message: "Invalid email format"},
		{Field: "password", Message: "Password too short"},
	}

	ErrorResponseJSON(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Error.Details, 2)
}

func TestNewMeta_Calculation(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		perPage    int
		total      int64
		wantPages  int
	}{
		{
			name:      "exact pages",
			page:      1,
			perPage:   10,
			total:     100,
			wantPages: 10,
		},
		{
			name:      "partial last page",
			page:      1,
			perPage:   10,
			total:     95,
			wantPages: 10,
		},
		{
			name:      "single page",
			page:      1,
			perPage:   10,
			total:     5,
			wantPages: 1,
		},
		{
			name:      "empty results",
			page:      1,
			perPage:   10,
			total:     0,
			wantPages: 0,
		},
		{
			name:      "one item per page",
			page:      1,
			perPage:   1,
			total:     5,
			wantPages: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewMeta(tt.page, tt.perPage, tt.total)
			assert.Equal(t, tt.page, meta.Page)
			assert.Equal(t, tt.perPage, meta.PerPage)
			assert.Equal(t, tt.total, meta.Total)
			assert.Equal(t, tt.wantPages, meta.TotalPages)
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		total    int64
		perPage  int
		expected int
	}{
		{100, 10, 10},
		{95, 10, 10},
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{100, 1, 100},
	}

	for _, tt := range tests {
		result := CalculateTotalPages(tt.total, tt.perPage)
		assert.Equal(t, tt.expected, result, "CalculateTotalPages(%d, %d) should be %d", tt.total, tt.perPage, tt.expected)
	}
}

func TestHandleError_AppError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleError(c, ErrNotFound)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
}

func TestHandleError_GenericError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleError(c, assert.AnError)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errors := []ValidationError{
		{Field: "email", Message: "Invalid email"},
	}

	HandleValidationError(c, errors)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
}
