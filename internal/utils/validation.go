package utils

import (
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	// emailRegex is a simple email validation regex
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// usernameRegex allows alphanumeric characters and underscores
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

	// AllowedImageTypes contains the allowed image MIME types
	AllowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}

	// AllowedImageExtensions contains the allowed image file extensions
	AllowedImageExtensions = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
	}
)

// RegisterCustomValidators registers custom validators with gin's validator
func RegisterCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("uuid", validateUUID)
		v.RegisterValidation("username", validateUsername)
		v.RegisterValidation("strongpassword", validateStrongPassword)
	}
}

// validateUUID validates that a string is a valid UUID
func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	return usernameRegex.MatchString(username)
}

// validateStrongPassword validates password strength
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	errors := ValidatePasswordStrength(password)
	return len(errors) == 0
}

// IsValidEmail validates an email address format
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidUsername validates a username format
func IsValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	return usernameRegex.MatchString(username)
}

// IsValidUUID validates a UUID string
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// IsValidImageType checks if a MIME type is an allowed image type
func IsValidImageType(mimeType string) bool {
	return AllowedImageTypes[mimeType]
}

// IsValidImageExtension checks if a file extension is an allowed image extension
func IsValidImageExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return AllowedImageExtensions[ext]
}

// ValidateImageFile validates an uploaded image file
func ValidateImageFile(file *multipart.FileHeader, maxSize int64) *AppError {
	// Check file size
	if file.Size > maxSize {
		return ErrFileTooLarge
	}

	// Check file extension
	if !IsValidImageExtension(file.Filename) {
		return ErrInvalidFileType
	}

	// Check MIME type
	contentType := file.Header.Get("Content-Type")
	if !IsValidImageType(contentType) {
		return ErrInvalidFileType
	}

	return nil
}

// ParseValidationErrors converts validator errors to ValidationError slice
func ParseValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   toSnakeCase(e.Field()),
				Message: getValidationMessage(e),
			})
		}
	}

	return errors
}

// getValidationMessage returns a user-friendly validation message
func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "uuid":
		return "Invalid UUID format"
	case "username":
		return "Username must be 3-30 characters, alphanumeric and underscores only"
	case "strongpassword":
		return "Password must be at least 8 characters with uppercase, lowercase, and number"
	case "oneof":
		return "Invalid value"
	default:
		return "Invalid value"
	}
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// CalculateReadingTime calculates reading time in minutes based on content length
// Assumes average reading speed of 200 words per minute
func CalculateReadingTime(content string) int {
	words := len(strings.Fields(content))
	readingTime := words / 200
	if readingTime < 1 {
		readingTime = 1
	}
	return readingTime
}
