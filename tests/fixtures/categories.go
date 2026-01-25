package fixtures

import (
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
)

// ValidCategoryID is a valid UUID for testing
var ValidCategoryID = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

// TestCategories contains sample category data for testing
var TestCategories = struct {
	Deen        models.Category
	News        models.Category
	Purification models.Category // Subcategory of Deen
}{
	Deen: models.Category{
		ID:           uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"),
		Name:         "Deen",
		Slug:         "deen",
		Description:  "Islamic teachings and religious knowledge",
		DisplayOrder: 1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	},
	News: models.Category{
		ID:           uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"),
		Name:         "News",
		Slug:         "news",
		Description:  "Latest news and updates",
		DisplayOrder: 2,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	},
	Purification: models.Category{
		ID:           uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		Name:         "Purification",
		Slug:         "purification",
		Description:  "Spiritual purification and self-improvement",
		ParentID:     func() *uuid.UUID { id := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"); return &id }(),
		DisplayOrder: 1,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	},
}

// NewTestCategory creates a new category with default values
func NewTestCategory(opts ...func(*models.Category)) *models.Category {
	category := &models.Category{
		ID:           uuid.New(),
		Name:         "Test Category " + uuid.New().String()[:8],
		Slug:         "test-category-" + uuid.New().String()[:8],
		Description:  "Test category description",
		DisplayOrder: 0,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	for _, opt := range opts {
		opt(category)
	}

	return category
}

// WithCategoryName sets the category name
func WithCategoryName(name string) func(*models.Category) {
	return func(c *models.Category) {
		c.Name = name
	}
}

// WithCategorySlug sets the category slug
func WithCategorySlug(slug string) func(*models.Category) {
	return func(c *models.Category) {
		c.Slug = slug
	}
}

// WithParent sets the parent category ID
func WithParent(parentID uuid.UUID) func(*models.Category) {
	return func(c *models.Category) {
		c.ParentID = &parentID
	}
}

// WithCategoryActive sets the category active status
func WithCategoryActive(active bool) func(*models.Category) {
	return func(c *models.Category) {
		c.IsActive = active
	}
}

// TestTags contains sample tag data for testing
var TestTags = struct {
	FivePillars models.Tag
	Ramadan     models.Tag
	Hajj        models.Tag
}{
	FivePillars: models.Tag{
		ID:          uuid.MustParse("11111111-2222-3333-4444-555555555555"),
		Name:        "Five Pillars",
		Slug:        "five-pillars",
		Description: "Articles about the five pillars of Islam",
		UsageCount:  10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	Ramadan: models.Tag{
		ID:          uuid.MustParse("22222222-3333-4444-5555-666666666666"),
		Name:        "Ramadan",
		Slug:        "ramadan",
		Description: "Ramadan related content",
		UsageCount:  25,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
	Hajj: models.Tag{
		ID:          uuid.MustParse("33333333-4444-5555-6666-777777777777"),
		Name:        "Hajj",
		Slug:        "hajj",
		Description: "Hajj and pilgrimage articles",
		UsageCount:  15,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	},
}

// NewTestTag creates a new tag with default values
func NewTestTag(opts ...func(*models.Tag)) *models.Tag {
	tag := &models.Tag{
		ID:          uuid.New(),
		Name:        "Test Tag " + uuid.New().String()[:8],
		Slug:        "test-tag-" + uuid.New().String()[:8],
		Description: "Test tag description",
		UsageCount:  0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, opt := range opts {
		opt(tag)
	}

	return tag
}

// WithTagName sets the tag name
func WithTagName(name string) func(*models.Tag) {
	return func(t *models.Tag) {
		t.Name = name
	}
}

// WithTagSlug sets the tag slug
func WithTagSlug(slug string) func(*models.Tag) {
	return func(t *models.Tag) {
		t.Slug = slug
	}
}

// WithUsageCount sets the tag usage count
func WithUsageCount(count int) func(*models.Tag) {
	return func(t *models.Tag) {
		t.UsageCount = count
	}
}
