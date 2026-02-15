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
		// ── Root categories ────────────────────────────────────
		{Name: "Deen", Slug: "deen", Description: "Islamic teachings and religious knowledge", Order: 1},
		{Name: "Opinion", Slug: "opinion", Description: "Opinion pieces and perspectives", Order: 2},
		{Name: "Islam", Slug: "islam", Description: "Articles about Islamic topics", Order: 3},
		{Name: "News", Slug: "news", Description: "Latest news and updates", Order: 4},
		{Name: "People", Slug: "people", Description: "Stories about people and communities", Order: 5},
		{Name: "Science & Biology", Slug: "science-biology", Description: "Life sciences, physics, and the natural world", Order: 6},
		{Name: "Technology & Engineering", Slug: "technology-engineering", Description: "Software, AI, cybersecurity, and engineering disciplines", Order: 7},
		{Name: "Social & History", Slug: "social-history", Description: "Social sciences and historical studies", Order: 8},

		// ── Subcategories of "Deen" ────────────────────────────
		{Name: "Purification", Slug: "purification", Description: "Spiritual purification and self-improvement", ParentSlug: "deen", Order: 1},
		{Name: "Seerah", Slug: "seerah", Description: "The biography and life of Prophet Muhammad (PBUH)", ParentSlug: "deen", Order: 2},
		{Name: "Fiqh", Slug: "fiqh", Description: "Islamic Jurisprudence and rulings", ParentSlug: "deen", Order: 3},
		{Name: "Aqeedah", Slug: "aqeedah", Description: "Matters of faith, theology, and belief", ParentSlug: "deen", Order: 4},
		{Name: "Tafsir", Slug: "tafsir", Description: "Exegesis and deep explanation of the Quran", ParentSlug: "deen", Order: 5},
		{Name: "Tazkiyah", Slug: "tazkiyah", Description: "Spiritual purification and self-development", ParentSlug: "deen", Order: 6},
		{Name: "Contemporary Issues", Slug: "contemporary-issues", Description: "Modern challenges and Islamic solutions", ParentSlug: "deen", Order: 7},
		{Name: "Family & Parenting", Slug: "family-parenting", Description: "Building strong Muslim families", ParentSlug: "deen", Order: 8},

		// ── Subcategories of "Science & Biology" ───────────────
		{Name: "Biology", Slug: "biology", Description: "Study of life and living organisms", ParentSlug: "science-biology", Order: 1},
		{Name: "Physics", Slug: "physics", Description: "Matter, energy, and the universe", ParentSlug: "science-biology", Order: 2},
		{Name: "Environment", Slug: "environment", Description: "Ecology, climate change, and sustainability", ParentSlug: "science-biology", Order: 3},
		{Name: "Neuroscience", Slug: "neuroscience", Description: "Brain and nervous system", ParentSlug: "science-biology", Order: 4},
		{Name: "Astronomy", Slug: "astronomy", Description: "Celestial bodies and the universe (Islamic perspective on cosmos)", ParentSlug: "science-biology", Order: 5},

		// ── Subcategories of "Technology & Engineering" ─────────
		{Name: "Artificial Intelligence", Slug: "ai", Description: "Machine learning and robotics", ParentSlug: "technology-engineering", Order: 1},
		{Name: "Software Engineering", Slug: "software-engineering", Description: "Coding, development, and best practices", ParentSlug: "technology-engineering", Order: 2},
		{Name: "Civil Engineering", Slug: "civil-engineering", Description: "Infrastructure and urban planning", ParentSlug: "technology-engineering", Order: 3},
		{Name: "Cybersecurity", Slug: "cybersecurity", Description: "Digital safety and privacy", ParentSlug: "technology-engineering", Order: 4},
		{Name: "Ethical Tech", Slug: "ethical-tech", Description: "Privacy rights and ethics in technology", ParentSlug: "technology-engineering", Order: 5},

		// ── Subcategories of "Social & History" ────────────────
		{Name: "Sociology", Slug: "sociology", Description: "Study of social behavior and society", ParentSlug: "social-history", Order: 1},
		{Name: "Psychology", Slug: "psychology", Description: "Human mind and behavior", ParentSlug: "social-history", Order: 2},
		{Name: "Islamic History", Slug: "islamic-history", Description: "Specific historical events in the Muslim world", ParentSlug: "social-history", Order: 3},
		{Name: "Economics", Slug: "economics", Description: "Wealth, finance, and resource management", ParentSlug: "social-history", Order: 4},
		{Name: "Anthropology", Slug: "anthropology", Description: "Human societies and cultures", ParentSlug: "social-history", Order: 5},
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
		{Name: "Marriage", Slug: "marriage", Description: "Marriage in Islam"},
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
