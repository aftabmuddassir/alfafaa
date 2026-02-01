package repositories

import (
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ArticleRepository defines the interface for article data access
type ArticleRepository interface {
	Create(article *models.Article) error
	FindByID(id uuid.UUID) (*models.Article, error)
	FindBySlug(slug string) (*models.Article, error)
	FindAll(filters ArticleFilters) ([]models.Article, int64, error)
	FindByAuthor(authorID uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error)
	FindByCategory(categoryID uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error)
	FindByTag(tagID uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error)
	FindTrending(limit int) ([]models.Article, error)
	FindRecent(limit int) ([]models.Article, error)
	FindRelated(articleID uuid.UUID, categoryIDs, tagIDs []uuid.UUID, limit int) ([]models.Article, error)
	FindForUser(userID uuid.UUID, followingIDs, interestCategoryIDs []uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error)
	FindStaffPicks(filters ArticleFilters) ([]models.Article, int64, error)
	Update(article *models.Article) error
	Delete(id uuid.UUID) error
	ExistsBySlug(slug string) (bool, error)
	IncrementViewCount(id uuid.UUID) error
	UpdateCategories(article *models.Article, categories []models.Category) error
	UpdateTags(article *models.Article, tags []models.Tag) error
	Search(query string, filters ArticleFilters) ([]models.Article, int64, error)
	SetStaffPick(id uuid.UUID, isStaffPick bool) error
	// WithTx returns a new repository instance using the provided transaction
	WithTx(tx *gorm.DB) ArticleRepository
}

// ArticleFilters contains filter options for querying articles
type ArticleFilters struct {
	Status   string
	AuthorID *uuid.UUID
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
	Sort     string
}

type articleRepository struct {
	db *gorm.DB
}

// NewArticleRepository creates a new article repository
func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

// WithTx returns a new repository instance using the provided transaction
func (r *articleRepository) WithTx(tx *gorm.DB) ArticleRepository {
	return &articleRepository{db: tx}
}

// Create creates a new article
func (r *articleRepository) Create(article *models.Article) error {
	return r.db.Create(article).Error
}

// FindByID finds an article by ID with preloaded relationships
func (r *articleRepository) FindByID(id uuid.UUID) (*models.Article, error) {
	var article models.Article
	err := r.db.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		First(&article, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// FindBySlug finds an article by slug with preloaded relationships
func (r *articleRepository) FindBySlug(slug string) (*models.Article, error) {
	var article models.Article
	err := r.db.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		First(&article, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// FindAll finds all articles with filters
func (r *articleRepository) FindAll(filters ArticleFilters) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{})
	query = r.applyFilters(query, filters)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	query = r.applyPaginationAndSort(query, filters)

	// Preload relationships
	err := query.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, total, err
}

// FindByAuthor finds articles by author
func (r *articleRepository) FindByAuthor(authorID uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error) {
	filters.AuthorID = &authorID
	return r.FindAll(filters)
}

// FindByCategory finds articles in a category
func (r *articleRepository) FindByCategory(categoryID uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).
		Joins("JOIN article_categories ON article_categories.article_id = articles.id").
		Where("article_categories.category_id = ?", categoryID)

	query = r.applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPaginationAndSort(query, filters)

	err := query.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, total, err
}

// FindByTag finds articles with a tag
func (r *articleRepository) FindByTag(tagID uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).
		Joins("JOIN article_tags ON article_tags.article_id = articles.id").
		Where("article_tags.tag_id = ?", tagID)

	query = r.applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPaginationAndSort(query, filters)

	err := query.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, total, err
}

// FindTrending finds trending articles by view count
func (r *articleRepository) FindTrending(limit int) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.
		Where("status = ?", models.StatusPublished).
		Order("view_count DESC").
		Limit(limit).
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error
	return articles, err
}

// FindRecent finds recently published articles
func (r *articleRepository) FindRecent(limit int) ([]models.Article, error) {
	var articles []models.Article
	err := r.db.
		Where("status = ?", models.StatusPublished).
		Order("published_at DESC").
		Limit(limit).
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error
	return articles, err
}

// FindRelated finds related articles based on categories and tags
func (r *articleRepository) FindRelated(articleID uuid.UUID, categoryIDs, tagIDs []uuid.UUID, limit int) ([]models.Article, error) {
	var articles []models.Article

	query := r.db.Model(&models.Article{}).
		Where("id != ?", articleID).
		Where("status = ?", models.StatusPublished)

	// Find articles that share categories or tags
	if len(categoryIDs) > 0 || len(tagIDs) > 0 {
		subQuery := r.db.Model(&models.Article{}).Select("DISTINCT articles.id")

		if len(categoryIDs) > 0 {
			subQuery = subQuery.
				Joins("LEFT JOIN article_categories ON article_categories.article_id = articles.id").
				Where("article_categories.category_id IN ?", categoryIDs)
		}
		if len(tagIDs) > 0 {
			subQuery = subQuery.
				Joins("LEFT JOIN article_tags ON article_tags.article_id = articles.id").
				Or("article_tags.tag_id IN ?", tagIDs)
		}

		query = query.Where("id IN (?)", subQuery)
	}

	err := query.
		Order("published_at DESC").
		Limit(limit).
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, err
}

// Update updates an article
func (r *articleRepository) Update(article *models.Article) error {
	return r.db.Save(article).Error
}

// Delete soft deletes an article
func (r *articleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Article{}, "id = ?", id).Error
}

// ExistsBySlug checks if an article exists with the given slug
func (r *articleRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Article{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

// IncrementViewCount increments the view count of an article
func (r *articleRepository) IncrementViewCount(id uuid.UUID) error {
	return r.db.Model(&models.Article{}).Where("id = ?", id).Update("view_count", gorm.Expr("view_count + 1")).Error
}

// UpdateCategories updates the categories of an article
func (r *articleRepository) UpdateCategories(article *models.Article, categories []models.Category) error {
	return r.db.Model(article).Association("Categories").Replace(categories)
}

// UpdateTags updates the tags of an article
func (r *articleRepository) UpdateTags(article *models.Article, tags []models.Tag) error {
	return r.db.Model(article).Association("Tags").Replace(tags)
}

// Search searches articles by title, content, and excerpt
func (r *articleRepository) Search(query string, filters ArticleFilters) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	search := "%" + query + "%"
	dbQuery := r.db.Model(&models.Article{}).
		Where("title ILIKE ? OR excerpt ILIKE ? OR content ILIKE ?", search, search, search)

	dbQuery = r.applyFilters(dbQuery, filters)

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dbQuery = r.applyPaginationAndSort(dbQuery, filters)

	err := dbQuery.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, total, err
}

// applyFilters applies common filters to a query
func (r *articleRepository) applyFilters(query *gorm.DB, filters ArticleFilters) *gorm.DB {
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.AuthorID != nil {
		query = query.Where("author_id = ?", *filters.AuthorID)
	}
	if filters.FromDate != nil {
		query = query.Where("published_at >= ?", *filters.FromDate)
	}
	if filters.ToDate != nil {
		query = query.Where("published_at <= ?", *filters.ToDate)
	}
	return query
}

// applyPaginationAndSort applies pagination and sorting to a query
func (r *articleRepository) applyPaginationAndSort(query *gorm.DB, filters ArticleFilters) *gorm.DB {
	// Apply sorting
	switch filters.Sort {
	case "oldest":
		query = query.Order("created_at ASC")
	case "popular":
		query = query.Order("view_count DESC")
	case "alphabetical":
		query = query.Order("title ASC")
	default: // newest
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	return query
}

// FindForUser finds articles personalized for a user based on followed authors and interests
func (r *articleRepository) FindForUser(userID uuid.UUID, followingIDs, interestCategoryIDs []uuid.UUID, filters ArticleFilters) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).
		Where("status = ?", models.StatusPublished)

	// Build the personalized feed query
	if len(followingIDs) > 0 || len(interestCategoryIDs) > 0 {
		subQuery := r.db

		if len(followingIDs) > 0 && len(interestCategoryIDs) > 0 {
			// Articles from followed authors OR in interested categories
			subQuery = r.db.Model(&models.Article{}).
				Select("DISTINCT articles.id").
				Joins("LEFT JOIN article_categories ON article_categories.article_id = articles.id").
				Where("articles.author_id IN ? OR article_categories.category_id IN ?", followingIDs, interestCategoryIDs)
			query = query.Where("id IN (?)", subQuery)
		} else if len(followingIDs) > 0 {
			// Only followed authors
			query = query.Where("author_id IN ?", followingIDs)
		} else {
			// Only interested categories
			subQuery = r.db.Model(&models.Article{}).
				Select("DISTINCT articles.id").
				Joins("JOIN article_categories ON article_categories.article_id = articles.id").
				Where("article_categories.category_id IN ?", interestCategoryIDs)
			query = query.Where("id IN (?)", subQuery)
		}
	}

	// Apply additional filters
	query = r.applyFilters(query, filters)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	query = r.applyPaginationAndSort(query, filters)

	// Preload relationships
	err := query.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, total, err
}

// FindStaffPicks finds articles marked as staff picks
func (r *articleRepository) FindStaffPicks(filters ArticleFilters) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).
		Where("is_staff_pick = ?", true).
		Where("status = ?", models.StatusPublished)

	query = r.applyFilters(query, filters)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = r.applyPaginationAndSort(query, filters)

	err := query.
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Find(&articles).Error

	return articles, total, err
}

// SetStaffPick sets the staff pick status of an article
func (r *articleRepository) SetStaffPick(id uuid.UUID, isStaffPick bool) error {
	return r.db.Model(&models.Article{}).Where("id = ?", id).Update("is_staff_pick", isStaffPick).Error
}
