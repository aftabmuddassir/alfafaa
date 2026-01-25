package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail_ValidEmails(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.com",
		"user+tag@example.org",
		"user123@test.co.uk",
		"a@b.cd",
		"test.user@subdomain.example.com",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			assert.True(t, IsValidEmail(email), "Expected %s to be valid", email)
		})
	}
}

func TestIsValidEmail_InvalidEmails(t *testing.T) {
	invalidEmails := []struct {
		email  string
		reason string
	}{
		{"", "empty string"},
		{"notanemail", "no @ symbol"},
		{"@nodomain.com", "no local part"},
		{"noat.domain.com", "no @ symbol"},
		{"user@", "no domain"},
		{"user@.com", "domain starts with dot"},
		{"user@domain", "no TLD"},
		{"user name@domain.com", "contains space"},
	}

	for _, tc := range invalidEmails {
		t.Run(tc.reason, func(t *testing.T) {
			assert.False(t, IsValidEmail(tc.email), "Expected %s to be invalid: %s", tc.email, tc.reason)
		})
	}
}

func TestIsValidUsername_ValidUsernames(t *testing.T) {
	validUsernames := []string{
		"john",
		"john_doe",
		"JohnDoe123",
		"user_123",
		"abc",
		"user_name_test_12345678901", // 30 chars max
	}

	for _, username := range validUsernames {
		t.Run(username, func(t *testing.T) {
			assert.True(t, IsValidUsername(username), "Expected %s to be valid", username)
		})
	}
}

func TestIsValidUsername_InvalidUsernames(t *testing.T) {
	invalidUsernames := []struct {
		username string
		reason   string
	}{
		{"", "empty string"},
		{"ab", "too short (< 3)"},
		{"user name", "contains space"},
		{"user-name", "contains hyphen"},
		{"user.name", "contains period"},
		{"user@name", "contains special char"},
		{"a_very_long_username_that_exceeds_thirty_characters", "too long (> 30)"},
	}

	for _, tc := range invalidUsernames {
		t.Run(tc.reason, func(t *testing.T) {
			assert.False(t, IsValidUsername(tc.username), "Expected %s to be invalid: %s", tc.username, tc.reason)
		})
	}
}

func TestIsValidUUID_ValidUUIDs(t *testing.T) {
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for _, uuid := range validUUIDs {
		t.Run(uuid, func(t *testing.T) {
			assert.True(t, IsValidUUID(uuid), "Expected %s to be valid UUID", uuid)
		})
	}
}

func TestIsValidUUID_InvalidUUIDs(t *testing.T) {
	invalidUUIDs := []struct {
		uuid   string
		reason string
	}{
		{"", "empty string"},
		{"not-a-uuid", "invalid format"},
		{"550e8400-e29b-41d4-a716", "too short"},
		{"550e8400-e29b-41d4-a716-446655440000-extra", "too long"},
		{"550e8400-e29b-41d4-a716-44665544000g", "invalid character"},
		{"ZZZZZZZZ-ZZZZ-ZZZZ-ZZZZ-ZZZZZZZZZZZZ", "invalid hex"},
	}

	for _, tc := range invalidUUIDs {
		t.Run(tc.reason, func(t *testing.T) {
			assert.False(t, IsValidUUID(tc.uuid), "Expected %s to be invalid UUID: %s", tc.uuid, tc.reason)
		})
	}
}

func TestIsValidImageType_ValidTypes(t *testing.T) {
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"image/gif",
	}

	for _, mimeType := range validTypes {
		t.Run(mimeType, func(t *testing.T) {
			assert.True(t, IsValidImageType(mimeType), "Expected %s to be valid", mimeType)
		})
	}
}

func TestIsValidImageType_InvalidTypes(t *testing.T) {
	invalidTypes := []string{
		"",
		"image/bmp",
		"image/svg+xml",
		"application/pdf",
		"text/html",
		"video/mp4",
	}

	for _, mimeType := range invalidTypes {
		t.Run(mimeType, func(t *testing.T) {
			assert.False(t, IsValidImageType(mimeType), "Expected %s to be invalid", mimeType)
		})
	}
}

func TestIsValidImageExtension_ValidExtensions(t *testing.T) {
	validFilenames := []string{
		"image.jpg",
		"photo.jpeg",
		"picture.png",
		"graphic.webp",
		"animation.gif",
		"IMAGE.JPG",  // Should handle case insensitive
		"photo.JPEG",
	}

	for _, filename := range validFilenames {
		t.Run(filename, func(t *testing.T) {
			assert.True(t, IsValidImageExtension(filename), "Expected %s to have valid extension", filename)
		})
	}
}

func TestIsValidImageExtension_InvalidExtensions(t *testing.T) {
	invalidFilenames := []string{
		"document.pdf",
		"file.txt",
		"video.mp4",
		"image.bmp",
		"image.svg",
		"noextension",
		"",
	}

	for _, filename := range invalidFilenames {
		t.Run(filename, func(t *testing.T) {
			assert.False(t, IsValidImageExtension(filename), "Expected %s to have invalid extension", filename)
		})
	}
}

func TestCalculateReadingTime_ShortContent(t *testing.T) {
	// Very short content should return minimum 1 minute
	content := "Short article."
	result := CalculateReadingTime(content)
	assert.Equal(t, 1, result)
}

func TestCalculateReadingTime_MediumContent(t *testing.T) {
	// ~200 words = 1 minute
	words := make([]string, 200)
	for i := range words {
		words[i] = "word"
	}
	content := ""
	for _, w := range words {
		content += w + " "
	}

	result := CalculateReadingTime(content)
	assert.GreaterOrEqual(t, result, 1)
}

func TestCalculateReadingTime_LongContent(t *testing.T) {
	// ~1000 words = 5 minutes
	words := make([]string, 1000)
	for i := range words {
		words[i] = "word"
	}
	content := ""
	for _, w := range words {
		content += w + " "
	}

	result := CalculateReadingTime(content)
	assert.GreaterOrEqual(t, result, 4)
	assert.LessOrEqual(t, result, 6)
}

func TestCalculateReadingTime_EmptyContent(t *testing.T) {
	result := CalculateReadingTime("")
	assert.Equal(t, 1, result) // Minimum is 1
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"FirstName", "first_name"},
		{"lastName", "last_name"},
		{"ID", "i_d"},
		{"userID", "user_i_d"},
		{"simple", "simple"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
