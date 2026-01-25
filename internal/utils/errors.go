package utils

import (
	"errors"
	"fmt"
)

// AppError represents a custom application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// Common application errors
var (
	ErrNotFound         = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: 404}
	ErrUnauthorized     = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized access", Status: 401}
	ErrForbidden        = &AppError{Code: "FORBIDDEN", Message: "Access forbidden", Status: 403}
	ErrValidation       = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed", Status: 400}
	ErrBadRequest       = &AppError{Code: "BAD_REQUEST", Message: "Bad request", Status: 400}
	ErrInternal         = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", Status: 500}
	ErrConflict         = &AppError{Code: "CONFLICT", Message: "Resource already exists", Status: 409}
	ErrInvalidToken     = &AppError{Code: "INVALID_TOKEN", Message: "Invalid or expired token", Status: 401}
	ErrInvalidCredentials = &AppError{Code: "INVALID_CREDENTIALS", Message: "Invalid email or password", Status: 401}
	ErrEmailExists      = &AppError{Code: "EMAIL_EXISTS", Message: "Email already registered", Status: 409}
	ErrUsernameExists   = &AppError{Code: "USERNAME_EXISTS", Message: "Username already taken", Status: 409}
	ErrFileTooLarge     = &AppError{Code: "FILE_TOO_LARGE", Message: "File size exceeds limit", Status: 400}
	ErrInvalidFileType  = &AppError{Code: "INVALID_FILE_TYPE", Message: "Invalid file type", Status: 400}
)

// NewAppError creates a new application error with a custom message
func NewAppError(code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetStatusCode returns the HTTP status code for an error
func GetStatusCode(err error) int {
	if appErr, ok := IsAppError(err); ok {
		return appErr.Status
	}
	return 500
}
