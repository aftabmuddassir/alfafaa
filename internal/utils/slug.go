package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	// slugRegex matches non-alphanumeric characters
	slugRegex = regexp.MustCompile(`[^a-z0-9]+`)
	// multiDashRegex matches multiple consecutive dashes
	multiDashRegex = regexp.MustCompile(`-+`)
)

// GenerateSlug creates a URL-friendly slug from a string
func GenerateSlug(s string) string {
	// Normalize unicode characters
	s = norm.NFKD.String(s)

	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove non-ASCII characters
	s = removeNonASCII(s)

	// Replace non-alphanumeric characters with dashes
	s = slugRegex.ReplaceAllString(s, "-")

	// Replace multiple dashes with single dash
	s = multiDashRegex.ReplaceAllString(s, "-")

	// Trim leading and trailing dashes
	s = strings.Trim(s, "-")

	return s
}

// removeNonASCII removes non-ASCII characters from a string
func removeNonASCII(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r <= unicode.MaxASCII {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GenerateUniqueSlug generates a unique slug by appending a suffix if needed
func GenerateUniqueSlug(baseSlug string, suffix string) string {
	if suffix == "" {
		return baseSlug
	}
	return baseSlug + "-" + suffix
}

// TruncateSlug truncates a slug to a maximum length
func TruncateSlug(slug string, maxLength int) string {
	if len(slug) <= maxLength {
		return slug
	}

	// Truncate and remove trailing dash
	truncated := slug[:maxLength]
	truncated = strings.TrimRight(truncated, "-")

	return truncated
}

// IsValidSlug checks if a string is a valid slug
func IsValidSlug(s string) bool {
	if s == "" {
		return false
	}

	// Check for valid characters
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}

	// Check for leading/trailing dashes or consecutive dashes
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") || strings.Contains(s, "--") {
		return false
	}

	return true
}
