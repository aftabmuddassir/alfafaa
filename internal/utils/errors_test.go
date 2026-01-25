package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
		Status:  400,
	}

	assert.Equal(t, "Test error message", err.Error())
}

func TestNewAppError(t *testing.T) {
	err := NewAppError("CUSTOM_ERROR", "Custom message", 422)

	assert.Equal(t, "CUSTOM_ERROR", err.Code)
	assert.Equal(t, "Custom message", err.Message)
	assert.Equal(t, 422, err.Status)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		code     string
		status   int
	}{
		{"ErrNotFound", ErrNotFound, "NOT_FOUND", 404},
		{"ErrUnauthorized", ErrUnauthorized, "UNAUTHORIZED", 401},
		{"ErrForbidden", ErrForbidden, "FORBIDDEN", 403},
		{"ErrValidation", ErrValidation, "VALIDATION_ERROR", 400},
		{"ErrBadRequest", ErrBadRequest, "BAD_REQUEST", 400},
		{"ErrInternal", ErrInternal, "INTERNAL_ERROR", 500},
		{"ErrConflict", ErrConflict, "CONFLICT", 409},
		{"ErrInvalidToken", ErrInvalidToken, "INVALID_TOKEN", 401},
		{"ErrInvalidCredentials", ErrInvalidCredentials, "INVALID_CREDENTIALS", 401},
		{"ErrEmailExists", ErrEmailExists, "EMAIL_EXISTS", 409},
		{"ErrUsernameExists", ErrUsernameExists, "USERNAME_EXISTS", 409},
		{"ErrFileTooLarge", ErrFileTooLarge, "FILE_TOO_LARGE", 400},
		{"ErrInvalidFileType", ErrInvalidFileType, "INVALID_FILE_TYPE", 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, tt.status, tt.err.Status)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, "additional context")

	assert.Contains(t, wrappedErr.Error(), "additional context")
	assert.Contains(t, wrappedErr.Error(), "original error")

	// Should be able to unwrap
	assert.True(t, errors.Is(wrappedErr, originalErr))
}

func TestIsAppError_WithAppError(t *testing.T) {
	err := ErrNotFound

	appErr, ok := IsAppError(err)

	assert.True(t, ok)
	assert.NotNil(t, appErr)
	assert.Equal(t, "NOT_FOUND", appErr.Code)
}

func TestIsAppError_WithGenericError(t *testing.T) {
	err := errors.New("generic error")

	appErr, ok := IsAppError(err)

	assert.False(t, ok)
	assert.Nil(t, appErr)
}

func TestIsAppError_WithWrappedAppError(t *testing.T) {
	originalErr := ErrNotFound
	wrappedErr := WrapError(originalErr, "context")

	appErr, ok := IsAppError(wrappedErr)

	assert.True(t, ok)
	assert.NotNil(t, appErr)
	assert.Equal(t, "NOT_FOUND", appErr.Code)
}

func TestGetStatusCode_AppError(t *testing.T) {
	tests := []struct {
		err            error
		expectedStatus int
	}{
		{ErrNotFound, 404},
		{ErrUnauthorized, 401},
		{ErrForbidden, 403},
		{ErrBadRequest, 400},
		{ErrInternal, 500},
	}

	for _, tt := range tests {
		status := GetStatusCode(tt.err)
		assert.Equal(t, tt.expectedStatus, status)
	}
}

func TestGetStatusCode_GenericError(t *testing.T) {
	err := errors.New("generic error")

	status := GetStatusCode(err)

	assert.Equal(t, 500, status) // Default to 500
}

func TestGetStatusCode_WrappedAppError(t *testing.T) {
	wrappedErr := WrapError(ErrNotFound, "context")

	status := GetStatusCode(wrappedErr)

	assert.Equal(t, 404, status)
}

func TestAppError_IsError(t *testing.T) {
	var err error = ErrNotFound

	// Should be compatible with error interface
	assert.NotNil(t, err)
	assert.NotEmpty(t, err.Error())
}

func TestNewAppError_CustomStatus(t *testing.T) {
	err := NewAppError("CUSTOM", "Custom error", 418) // I'm a teapot

	assert.Equal(t, 418, err.Status)
	assert.Equal(t, "CUSTOM", err.Code)
}
