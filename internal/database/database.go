package database

import (
	"fmt"
	"log"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection
var DB *gorm.DB

// Connect establishes a connection to the PostgreSQL database
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// Configure GORM logger based on environment
	gormLogger := logger.Default.LogMode(logger.Info)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	log.Println("Database connection established successfully")
	return db, nil
}

// AutoMigrate runs automatic migrations for all models
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Use raw SQL for all migrations to avoid GORM AutoMigrate
	// "insufficient arguments" error caused by self-referencing
	// many2many relationships (User.Followers/Following []*User)
	// combined with explicit join table structs (UserFollow).
	// GORM v1.31.1 + postgres driver v1.5.4 chokes during schema
	// scanning at SELECT * FROM "users" LIMIT 1.
	if err := migrateAllTablesRawSQL(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// migrateAllTablesRawSQL creates all tables using raw SQL to bypass
// GORM AutoMigrate issues with self-referencing many2many relationships
func migrateAllTablesRawSQL(db *gorm.DB) error {
	queries := []string{
		// ==================== USERS ====================
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(30) NOT NULL,
			email VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255),
			first_name VARCHAR(100) DEFAULT '',
			last_name VARCHAR(100) DEFAULT '',
			bio TEXT DEFAULT '',
			profile_image_url VARCHAR(500),
			role VARCHAR(20) NOT NULL DEFAULT 'reader',
			is_verified BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			last_login_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			google_id VARCHAR(255),
			auth_provider VARCHAR(20) DEFAULT 'local'
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`,
		`CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at)`,
		// Add columns that may not exist yet (idempotent)
		`DO $$ BEGIN
			ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255);
			ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(20) DEFAULT 'local';
		EXCEPTION WHEN others THEN NULL;
		END $$`,

		// ==================== CATEGORIES ====================
		// Note: Category model has NO deleted_at (no soft deletes)
		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(100) NOT NULL,
			slug VARCHAR(100) NOT NULL,
			description TEXT DEFAULT '',
			parent_id UUID,
			display_order INT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id) REFERENCES categories(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_name ON categories(name)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active)`,

		// ==================== TAGS ====================
		// Note: Tag model has NO deleted_at (no soft deletes)
		`CREATE TABLE IF NOT EXISTS tags (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL,
			slug VARCHAR(50) NOT NULL,
			description TEXT DEFAULT '',
			usage_count INT DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_slug ON tags(slug)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name ON tags(name)`,
		`CREATE INDEX IF NOT EXISTS idx_tags_usage_count ON tags(usage_count)`,

		// ==================== ARTICLES ====================
		`CREATE TABLE IF NOT EXISTS articles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			excerpt TEXT DEFAULT '',
			content TEXT NOT NULL,
			featured_image_url VARCHAR(500),
			author_id UUID NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			published_at TIMESTAMPTZ,
			view_count INT DEFAULT 0,
			reading_time_minutes INT DEFAULT 1,
			is_staff_pick BOOLEAN DEFAULT FALSE,
			meta_title VARCHAR(70) DEFAULT '',
			meta_description VARCHAR(160) DEFAULT '',
			meta_keywords VARCHAR(255) DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			CONSTRAINT fk_articles_author FOREIGN KEY (author_id) REFERENCES users(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_title ON articles(title)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_author_id ON articles(author_id)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_is_staff_pick ON articles(is_staff_pick)`,
		`CREATE INDEX IF NOT EXISTS idx_articles_deleted_at ON articles(deleted_at)`,
		// Add columns that may not exist yet
		`DO $$ BEGIN
			ALTER TABLE articles ADD COLUMN IF NOT EXISTS is_staff_pick BOOLEAN DEFAULT FALSE;
			ALTER TABLE articles ADD COLUMN IF NOT EXISTS reading_time_minutes INT DEFAULT 1;
			ALTER TABLE articles ADD COLUMN IF NOT EXISTS meta_title VARCHAR(70) DEFAULT '';
			ALTER TABLE articles ADD COLUMN IF NOT EXISTS meta_description VARCHAR(160) DEFAULT '';
			ALTER TABLE articles ADD COLUMN IF NOT EXISTS meta_keywords VARCHAR(255) DEFAULT '';
		EXCEPTION WHEN others THEN NULL;
		END $$`,

		// ==================== ARTICLE_CATEGORIES (join) ====================
		`CREATE TABLE IF NOT EXISTS article_categories (
			article_id UUID NOT NULL,
			category_id UUID NOT NULL,
			PRIMARY KEY (article_id, category_id),
			CONSTRAINT fk_ac_article FOREIGN KEY (article_id) REFERENCES articles(id),
			CONSTRAINT fk_ac_category FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,

		// ==================== ARTICLE_TAGS (join) ====================
		`CREATE TABLE IF NOT EXISTS article_tags (
			article_id UUID NOT NULL,
			tag_id UUID NOT NULL,
			PRIMARY KEY (article_id, tag_id),
			CONSTRAINT fk_at_article FOREIGN KEY (article_id) REFERENCES articles(id),
			CONSTRAINT fk_at_tag FOREIGN KEY (tag_id) REFERENCES tags(id)
		)`,

		// ==================== MEDIA ====================
		// Note: Media model has NO deleted_at or updated_at, has is_featured
		`CREATE TABLE IF NOT EXISTS media (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			filename VARCHAR(255) NOT NULL,
			original_filename VARCHAR(255) NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			file_size BIGINT NOT NULL DEFAULT 0,
			mime_type VARCHAR(100) NOT NULL,
			uploaded_by UUID NOT NULL,
			is_featured BOOLEAN DEFAULT FALSE,
			alt_text VARCHAR(255) DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT fk_media_uploader FOREIGN KEY (uploaded_by) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_media_uploaded_by ON media(uploaded_by)`,

		// ==================== COMMENTS ====================
		`CREATE TABLE IF NOT EXISTS comments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			article_id UUID NOT NULL,
			user_id UUID NOT NULL,
			parent_id UUID,
			content TEXT NOT NULL,
			is_approved BOOLEAN DEFAULT FALSE,
			likes_count INT DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			CONSTRAINT fk_comments_article FOREIGN KEY (article_id) REFERENCES articles(id),
			CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_comments_parent FOREIGN KEY (parent_id) REFERENCES comments(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_article_id ON comments(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_is_approved ON comments(is_approved)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments(deleted_at)`,
		`ALTER TABLE comments ADD COLUMN IF NOT EXISTS likes_count INT DEFAULT 0`,

		// ==================== USER_FOLLOWS (social graph) ====================
		`CREATE TABLE IF NOT EXISTS user_follows (
			follower_id UUID NOT NULL,
			following_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (follower_id, following_id),
			CONSTRAINT fk_uf_follower FOREIGN KEY (follower_id) REFERENCES users(id),
			CONSTRAINT fk_uf_following FOREIGN KEY (following_id) REFERENCES users(id)
		)`,

		// ==================== USER_INTERESTS (social graph) ====================
		`CREATE TABLE IF NOT EXISTS user_interests (
			user_id UUID NOT NULL,
			category_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (user_id, category_id),
			CONSTRAINT fk_ui_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_ui_category FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,

		// ==================== LIKES (engagement) ====================
		`CREATE TABLE IF NOT EXISTS likes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			article_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_likes_article FOREIGN KEY (article_id) REFERENCES articles(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_likes_article_id ON likes(article_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_likes_user_article ON likes(user_id, article_id)`,

		// ==================== BOOKMARKS (engagement) ====================
		`CREATE TABLE IF NOT EXISTS bookmarks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			article_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT fk_bookmarks_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_bookmarks_article FOREIGN KEY (article_id) REFERENCES articles(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_bookmarks_article_id ON bookmarks(article_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_bookmarks_user_article ON bookmarks(user_id, article_id)`,

		// ==================== NOTIFICATIONS (engagement) ====================
		`CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			actor_id UUID NOT NULL,
			type VARCHAR(20) NOT NULL,
			message TEXT NOT NULL,
			article_id UUID,
			read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id),
			CONSTRAINT fk_notifications_actor FOREIGN KEY (actor_id) REFERENCES users(id),
			CONSTRAINT fk_notifications_article FOREIGN KEY (article_id) REFERENCES articles(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_actor_id ON notifications(actor_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_article_id ON notifications(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications(read)`,
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			return fmt.Errorf("migration query failed [%s]: %w", truncateQuery(query), err)
		}
	}

	return nil
}

// truncateQuery returns the first 80 chars of a query for error messages
func truncateQuery(q string) string {
	if len(q) <= 80 {
		return q
	}
	return q[:80] + "..."
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
