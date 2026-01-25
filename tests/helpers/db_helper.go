package helpers

import (
	"log"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate models
	err = db.AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.Category{},
		&models.Tag{},
		&models.Media{},
		&models.Comment{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans all tables in the test database
func CleanupTestDB(db *gorm.DB) {
	// Disable foreign key checks for cleanup
	db.Exec("PRAGMA foreign_keys = OFF")

	tables := []string{
		"article_categories",
		"article_tags",
		"comments",
		"media",
		"articles",
		"categories",
		"tags",
		"users",
	}

	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
	}

	db.Exec("PRAGMA foreign_keys = ON")
}
