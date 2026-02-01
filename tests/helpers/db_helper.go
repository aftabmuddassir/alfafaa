package helpers

import (
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Pure Go SQLite driver (no CGO required)
	"github.com/glebarez/sqlite"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create tables with SQLite-compatible schema
	err = createTables(db)
	if err != nil {
		log.Fatalf("Failed to create test database tables: %v", err)
	}

	return db
}

// createTables creates all required tables with SQLite-compatible schema
func createTables(db *gorm.DB) error {
	// Users table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			first_name TEXT,
			last_name TEXT,
			bio TEXT,
			profile_image_url TEXT,
			role TEXT DEFAULT 'reader',
			is_verified INTEGER DEFAULT 0,
			is_active INTEGER DEFAULT 1,
			last_login_at DATETIME,
			google_id TEXT UNIQUE,
			auth_provider TEXT DEFAULT 'local',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error; err != nil {
		return err
	}

	// Categories table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS categories (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			description TEXT,
			parent_id TEXT,
			display_order INTEGER DEFAULT 0,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (parent_id) REFERENCES categories(id)
		)
	`).Error; err != nil {
		return err
	}

	// Tags table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			description TEXT,
			usage_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error; err != nil {
		return err
	}

	// Articles table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			excerpt TEXT,
			content TEXT NOT NULL,
			featured_image_url TEXT,
			author_id TEXT NOT NULL,
			status TEXT DEFAULT 'draft',
			view_count INTEGER DEFAULT 0,
			reading_time_minutes INTEGER DEFAULT 1,
			is_staff_pick INTEGER DEFAULT 0,
			meta_title TEXT,
			meta_description TEXT,
			meta_keywords TEXT,
			published_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (author_id) REFERENCES users(id)
		)
	`).Error; err != nil {
		return err
	}

	// Article-Categories join table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS article_categories (
			article_id TEXT NOT NULL,
			category_id TEXT NOT NULL,
			PRIMARY KEY (article_id, category_id),
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	// Article-Tags join table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS article_tags (
			article_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			PRIMARY KEY (article_id, tag_id),
			FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	// Media table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS media (
			id TEXT PRIMARY KEY,
			filename TEXT NOT NULL,
			original_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size INTEGER NOT NULL,
			url TEXT NOT NULL,
			uploaded_by TEXT NOT NULL,
			alt_text TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (uploaded_by) REFERENCES users(id)
		)
	`).Error; err != nil {
		return err
	}

	// Comments table
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			article_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			parent_id TEXT,
			is_approved INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (article_id) REFERENCES articles(id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (parent_id) REFERENCES comments(id)
		)
	`).Error; err != nil {
		return err
	}

	// User follows table (social graph)
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_follows (
			follower_id TEXT NOT NULL,
			following_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (follower_id, following_id),
			FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	// User interests table (user to category relationship)
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS user_interests (
			user_id TEXT NOT NULL,
			category_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, category_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`).Error; err != nil {
		return err
	}

	return nil
}

// CleanupTestDB cleans all tables in the test database
func CleanupTestDB(db *gorm.DB) {
	// Disable foreign key checks for cleanup
	db.Exec("PRAGMA foreign_keys = OFF")

	tables := []string{
		"user_follows",
		"user_interests",
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
