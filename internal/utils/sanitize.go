package utils

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var (
	// strictPolicy strips ALL HTML tags - use for titles, usernames, etc.
	strictPolicy = bluemonday.StrictPolicy()

	// ugcPolicy allows safe user-generated content HTML (links, basic formatting)
	ugcPolicy = bluemonday.UGCPolicy()

	// Regex patterns for additional sanitization
	multipleSpaces = regexp.MustCompile(`\s+`)
	sqlKeywords    = regexp.MustCompile(`(?i)(DROP|DELETE|INSERT|UPDATE|EXEC|EXECUTE|UNION|SELECT|ALTER|CREATE|TRUNCATE)\s`)
)

// SanitizeStrict removes all HTML tags and trims whitespace
// Use for: titles, usernames, email, slugs, short text fields
func SanitizeStrict(input string) string {
	// Remove all HTML tags
	sanitized := strictPolicy.Sanitize(input)
	// Normalize whitespace
	sanitized = multipleSpaces.ReplaceAllString(sanitized, " ")
	// Trim
	return strings.TrimSpace(sanitized)
}

// SanitizeHTML allows safe HTML tags (links, bold, italic, etc.)
// Use for: article content, comments, rich text fields
func SanitizeHTML(input string) string {
	// Allow safe HTML tags
	sanitized := ugcPolicy.Sanitize(input)
	return strings.TrimSpace(sanitized)
}

// SanitizeSlug sanitizes a string for use as a URL slug
func SanitizeSlug(input string) string {
	// First strip all HTML
	sanitized := strictPolicy.Sanitize(input)
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	// Replace spaces with hyphens
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	// Remove any character that isn't alphanumeric or hyphen
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	sanitized = reg.ReplaceAllString(sanitized, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	sanitized = reg.ReplaceAllString(sanitized, "-")
	// Trim hyphens from start and end
	return strings.Trim(sanitized, "-")
}

// SanitizeEmail sanitizes an email address
func SanitizeEmail(input string) string {
	// Strip HTML
	sanitized := strictPolicy.Sanitize(input)
	// Lowercase
	sanitized = strings.ToLower(sanitized)
	// Trim
	return strings.TrimSpace(sanitized)
}

// SanitizeUsername sanitizes a username
func SanitizeUsername(input string) string {
	// Strip HTML
	sanitized := strictPolicy.Sanitize(input)
	// Remove any character that isn't alphanumeric, underscore, or hyphen
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	sanitized = reg.ReplaceAllString(sanitized, "")
	return strings.TrimSpace(sanitized)
}

// SanitizeSearchQuery sanitizes a search query to prevent injection
func SanitizeSearchQuery(input string) string {
	// Strip HTML
	sanitized := strictPolicy.Sanitize(input)
	// Remove potential SQL injection patterns
	sanitized = sqlKeywords.ReplaceAllString(sanitized, "")
	// Normalize whitespace
	sanitized = multipleSpaces.ReplaceAllString(sanitized, " ")
	// Limit length
	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
	}
	return strings.TrimSpace(sanitized)
}

// SanitizeFilename sanitizes a filename
func SanitizeFilename(input string) string {
	// Strip HTML
	sanitized := strictPolicy.Sanitize(input)
	// Remove path separators
	sanitized = strings.ReplaceAll(sanitized, "/", "")
	sanitized = strings.ReplaceAll(sanitized, "\\", "")
	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	// Keep only safe characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-\.]`)
	sanitized = reg.ReplaceAllString(sanitized, "_")
	return strings.TrimSpace(sanitized)
}

// ContainsDangerousPatterns checks if input contains potentially dangerous patterns
func ContainsDangerousPatterns(input string) bool {
	dangerous := []string{
		"<script", "</script>",
		"javascript:", "vbscript:",
		"onclick", "onerror", "onload",
		"eval(", "expression(",
	}

	lower := strings.ToLower(input)
	for _, pattern := range dangerous {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}

// EscapeHTML escapes HTML special characters without stripping tags
// Useful when you want to display HTML as text
func EscapeHTML(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}

// TruncateString truncates a string to a maximum length
// Adds "..." if truncated
func TruncateString(input string, maxLen int) string {
	if len(input) <= maxLen {
		return input
	}
	if maxLen <= 3 {
		return input[:maxLen]
	}
	return input[:maxLen-3] + "..."
}
