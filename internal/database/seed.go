package database

import (
	"log"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedDatabase seeds the database with initial data
func SeedDatabase(db *gorm.DB) error {
	log.Println("Seeding database...")

	// Seed categories
	if err := seedCategories(db); err != nil {
		return err
	}

	// Seed tags
	if err := seedTags(db); err != nil {
		return err
	}

	// Seed admin user
	if err := seedAdminUser(db); err != nil {
		return err
	}

	log.Println("Database seeding completed")
	return nil
}

// seedCategories seeds the initial categories
func seedCategories(db *gorm.DB) error {
	log.Println("Seeding categories...")

	categories := []struct {
		Name        string
		Slug        string
		Description string
		ParentSlug  string
		Order       int
	}{
		{Name: "Deen", Slug: "deen", Description: "Islamic teachings and religious knowledge", Order: 1},
		{Name: "Opinion", Slug: "opinion", Description: "Opinion pieces and perspectives", Order: 2},
		{Name: "Islam", Slug: "islam", Description: "Articles about Islamic topics", Order: 3},
		{Name: "News", Slug: "news", Description: "Latest news and updates", Order: 4},
		{Name: "People", Slug: "people", Description: "Stories about people and communities", Order: 5},
		{Name: "Science & Technology", Slug: "science-technology", Description: "Science and technology articles", Order: 6},
		{Name: "Purification", Slug: "purification", Description: "Spiritual purification and self-improvement", ParentSlug: "deen", Order: 1},
	}

	// Create a map to store category IDs by slug
	categoryIDs := make(map[string]uuid.UUID)

	// First pass: create root categories
	for _, cat := range categories {
		if cat.ParentSlug != "" {
			continue // Skip subcategories in first pass
		}

		// Check if category already exists
		var existing models.Category
		if err := db.Where("slug = ?", cat.Slug).First(&existing).Error; err == nil {
			categoryIDs[cat.Slug] = existing.ID
			log.Printf("Category '%s' already exists, skipping", cat.Name)
			continue
		}

		category := models.Category{
			Name:         cat.Name,
			Slug:         cat.Slug,
			Description:  cat.Description,
			DisplayOrder: cat.Order,
			IsActive:     true,
		}

		if err := db.Create(&category).Error; err != nil {
			log.Printf("Failed to create category '%s': %v", cat.Name, err)
			continue
		}

		categoryIDs[cat.Slug] = category.ID
		log.Printf("Created category: %s", cat.Name)
	}

	// Second pass: create subcategories
	for _, cat := range categories {
		if cat.ParentSlug == "" {
			continue // Skip root categories
		}

		// Check if category already exists
		var existing models.Category
		if err := db.Where("slug = ?", cat.Slug).First(&existing).Error; err == nil {
			log.Printf("Category '%s' already exists, skipping", cat.Name)
			continue
		}

		parentID, ok := categoryIDs[cat.ParentSlug]
		if !ok {
			log.Printf("Parent category '%s' not found for '%s'", cat.ParentSlug, cat.Name)
			continue
		}

		category := models.Category{
			Name:         cat.Name,
			Slug:         cat.Slug,
			Description:  cat.Description,
			ParentID:     &parentID,
			DisplayOrder: cat.Order,
			IsActive:     true,
		}

		if err := db.Create(&category).Error; err != nil {
			log.Printf("Failed to create category '%s': %v", cat.Name, err)
			continue
		}

		log.Printf("Created category: %s (subcategory of %s)", cat.Name, cat.ParentSlug)
	}

	return nil
}

// seedTags seeds the initial tags
func seedTags(db *gorm.DB) error {
	log.Println("Seeding tags...")

	tags := []struct {
		Name        string
		Slug        string
		Description string
	}{
		{Name: "Five Pillars", Slug: "five-pillars", Description: "Articles about the five pillars of Islam"},
		{Name: "Ramadan", Slug: "ramadan", Description: "Ramadan related content"},
		{Name: "Hajj", Slug: "hajj", Description: "Hajj and pilgrimage articles"},
		{Name: "Salah", Slug: "salah", Description: "Prayer related articles"},
		{Name: "Zakat", Slug: "zakat", Description: "Charity and Zakat articles"},
		{Name: "Fasting", Slug: "fasting", Description: "Fasting related content"},
		{Name: "Quran", Slug: "quran", Description: "Quran related articles"},
		{Name: "Hadith", Slug: "hadith", Description: "Hadith related articles"},
		{Name: "History", Slug: "history", Description: "Islamic history"},
		{Name: "Community", Slug: "community", Description: "Community related articles"},
	}

	for _, t := range tags {
		// Check if tag already exists
		var existing models.Tag
		if err := db.Where("slug = ?", t.Slug).First(&existing).Error; err == nil {
			log.Printf("Tag '%s' already exists, skipping", t.Name)
			continue
		}

		tag := models.Tag{
			Name:        t.Name,
			Slug:        t.Slug,
			Description: t.Description,
			UsageCount:  0,
		}

		if err := db.Create(&tag).Error; err != nil {
			log.Printf("Failed to create tag '%s': %v", t.Name, err)
			continue
		}

		log.Printf("Created tag: %s", t.Name)
	}

	return nil
}

// seedAdminUser creates the initial admin user
func seedAdminUser(db *gorm.DB) error {
	log.Println("Seeding admin user...")

	// Check if admin already exists
	var existing models.User
	if err := db.Where("email = ?", "admin@alfafaa.com").First(&existing).Error; err == nil {
		log.Println("Admin user already exists, skipping")
		return nil
	}

	// Hash the default password
	hashedPassword, err := utils.HashPassword("Admin@123")
	if err != nil {
		return err
	}

	admin := models.User{
		Username:     "admin",
		Email:        "admin@alfafaa.com",
		PasswordHash: hashedPassword,
		FirstName:    "Admin",
		LastName:     "User",
		Bio:          "System administrator for Alfafaa Blog",
		Role:         models.RoleAdmin,
		IsVerified:   true,
		IsActive:     true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Failed to create admin user: %v", err)
		return err
	}

	log.Println("Created admin user: admin@alfafaa.com (password: Admin@123)")
	log.Println("IMPORTANT: Please change the admin password after first login!")

	return nil
}
