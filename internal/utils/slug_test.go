package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug_BasicString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "hello world",
			expected: "hello-world",
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "with numbers",
			input:    "Article 123",
			expected: "article-123",
		},
		{
			name:     "single word",
			input:    "test",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSlug_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with punctuation",
			input:    "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "with ampersand",
			input:    "Science & Technology",
			expected: "science-technology",
		},
		{
			name:     "with parentheses",
			input:    "Test (Article)",
			expected: "test-article",
		},
		{
			name:     "with apostrophe",
			input:    "It's a Test",
			expected: "it-s-a-test",
		},
		{
			name:     "with quotes",
			input:    `"Quoted Title"`,
			expected: "quoted-title",
		},
		{
			name:     "with colons",
			input:    "Part 1: Introduction",
			expected: "part-1-introduction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSlug_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with accents",
			input:    "café résumé",
			expected: "cafe-resume",
		},
		{
			name:     "with german umlaut",
			input:    "über alles",
			expected: "uber-alles",
		},
		{
			name:     "arabic text removed",
			input:    "Hello مرحبا World",
			expected: "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSlug_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple spaces",
			input:    "hello    world",
			expected: "hello-world",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  hello world  ",
			expected: "hello-world",
		},
		{
			name:     "multiple dashes should become one",
			input:    "hello---world",
			expected: "hello-world",
		},
		{
			name:     "leading dashes removed",
			input:    "---hello",
			expected: "hello",
		},
		{
			name:     "trailing dashes removed",
			input:    "hello---",
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSlug_EmptyString(t *testing.T) {
	result := GenerateSlug("")
	assert.Equal(t, "", result)
}

func TestGenerateSlug_OnlySpecialChars(t *testing.T) {
	result := GenerateSlug("!@#$%^&*()")
	assert.Equal(t, "", result)
}

func TestGenerateUniqueSlug_NoSuffix(t *testing.T) {
	result := GenerateUniqueSlug("test-slug", "")
	assert.Equal(t, "test-slug", result)
}

func TestGenerateUniqueSlug_WithSuffix(t *testing.T) {
	result := GenerateUniqueSlug("test-slug", "2")
	assert.Equal(t, "test-slug-2", result)
}

func TestGenerateUniqueSlug_WithLongSuffix(t *testing.T) {
	result := GenerateUniqueSlug("test-slug", "abc123")
	assert.Equal(t, "test-slug-abc123", result)
}

func TestTruncateSlug_NoTruncationNeeded(t *testing.T) {
	slug := "short-slug"
	result := TruncateSlug(slug, 50)
	assert.Equal(t, "short-slug", result)
}

func TestTruncateSlug_ExactLength(t *testing.T) {
	slug := "exact-len"
	result := TruncateSlug(slug, 9)
	assert.Equal(t, "exact-len", result)
}

func TestTruncateSlug_TruncationNeeded(t *testing.T) {
	slug := "this-is-a-very-long-slug"
	result := TruncateSlug(slug, 10)
	assert.Equal(t, "this-is-a", result) // Trailing dash removed
	assert.LessOrEqual(t, len(result), 10)
}

func TestTruncateSlug_RemovesTrailingDash(t *testing.T) {
	slug := "hello-world-test"
	result := TruncateSlug(slug, 12)
	assert.Equal(t, "hello-world", result)
	assert.False(t, result[len(result)-1] == '-')
}

func TestIsValidSlug_ValidSlugs(t *testing.T) {
	validSlugs := []string{
		"hello-world",
		"test",
		"article-123",
		"a",
		"abc-def-ghi",
		"test123",
	}

	for _, slug := range validSlugs {
		t.Run(slug, func(t *testing.T) {
			assert.True(t, IsValidSlug(slug), "Expected %s to be valid", slug)
		})
	}
}

func TestIsValidSlug_InvalidSlugs(t *testing.T) {
	invalidSlugs := []struct {
		slug   string
		reason string
	}{
		{"", "empty string"},
		{"-hello", "leading dash"},
		{"hello-", "trailing dash"},
		{"hello--world", "consecutive dashes"},
		{"Hello-World", "uppercase letters"},
		{"hello world", "contains space"},
		{"hello_world", "contains underscore"},
		{"hello.world", "contains period"},
		{"hello@world", "contains special char"},
	}

	for _, tc := range invalidSlugs {
		t.Run(tc.reason, func(t *testing.T) {
			assert.False(t, IsValidSlug(tc.slug), "Expected %s to be invalid: %s", tc.slug, tc.reason)
		})
	}
}

func TestGenerateSlug_RealWorldExamples(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Understanding the Five Pillars of Islam",
			expected: "understanding-the-five-pillars-of-islam",
		},
		{
			input:    "Ramadan 2024: A Complete Guide",
			expected: "ramadan-2024-a-complete-guide",
		},
		{
			input:    "Q&A: Common Questions About Prayer",
			expected: "q-a-common-questions-about-prayer",
		},
		{
			input:    "The Prophet's (PBUH) Life Story",
			expected: "the-prophet-s-pbuh-life-story",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
