package repositories

import (
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByIDWithRelations(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByGoogleID(googleID string) (*models.User, error)
	FindAll(filters UserFilters) ([]models.User, int64, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
	UpdateLastLogin(id uuid.UUID) error
	// Social graph methods
	FollowUser(followerID, followingID uuid.UUID) error
	UnfollowUser(followerID, followingID uuid.UUID) error
	IsFollowing(followerID, followingID uuid.UUID) (bool, error)
	GetFollowers(userID uuid.UUID, limit, offset int) ([]models.User, int64, error)
	GetFollowing(userID uuid.UUID, limit, offset int) ([]models.User, int64, error)
	GetFollowingIDs(userID uuid.UUID) ([]uuid.UUID, error)
	// Interest methods
	SetInterests(userID uuid.UUID, categoryIDs []uuid.UUID) error
	GetInterests(userID uuid.UUID) ([]models.Category, error)
	GetInterestIDs(userID uuid.UUID) ([]uuid.UUID, error)
}

// UserFilters contains filter options for querying users
type UserFilters struct {
	Role     string
	IsActive *bool
	Search   string
	Limit    int
	Offset   int
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds a user by username
func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAll finds all users with filters
func (r *userRepository) FindAll(filters UserFilters) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if filters.Role != "" {
		query = query.Where("role = ?", filters.Role)
	}
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
			search, search, search, search)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and order
	query = query.Order("created_at DESC")
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update updates a user
func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// ExistsByEmail checks if a user exists with the given email
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByUsername checks if a user exists with the given username
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// UpdateLastLogin updates the user's last login time
func (r *userRepository) UpdateLastLogin(id uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}

// FindByIDWithRelations finds a user by ID with interests and following/followers preloaded
func (r *userRepository) FindByIDWithRelations(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.
		Preload("Interests").
		Preload("Followers").
		Preload("Following").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGoogleID finds a user by Google ID
func (r *userRepository) FindByGoogleID(googleID string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "google_id = ?", googleID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FollowUser creates a follow relationship between two users
func (r *userRepository) FollowUser(followerID, followingID uuid.UUID) error {
	follow := models.UserFollow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}
	return r.db.Create(&follow).Error
}

// UnfollowUser removes a follow relationship between two users
func (r *userRepository) UnfollowUser(followerID, followingID uuid.UUID) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Delete(&models.UserFollow{}).Error
}

// IsFollowing checks if a user is following another user
func (r *userRepository) IsFollowing(followerID, followingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count).Error
	return count > 0, err
}

// GetFollowers returns users who follow the given user
func (r *userRepository) GetFollowers(userID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{}).
		Joins("JOIN user_follows ON user_follows.follower_id = users.id").
		Where("user_follows.following_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("user_follows.created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetFollowing returns users that the given user follows
func (r *userRepository) GetFollowing(userID uuid.UUID, limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{}).
		Joins("JOIN user_follows ON user_follows.following_id = users.id").
		Where("user_follows.follower_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("user_follows.created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetFollowingIDs returns the IDs of users that the given user follows
func (r *userRepository) GetFollowingIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&models.UserFollow{}).
		Where("follower_id = ?", userID).
		Pluck("following_id", &ids).Error
	return ids, err
}

// SetInterests replaces a user's interests with the given category IDs
func (r *userRepository) SetInterests(userID uuid.UUID, categoryIDs []uuid.UUID) error {
	// Start a transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing interests
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserInterest{}).Error; err != nil {
			return err
		}

		// Insert new interests
		for _, catID := range categoryIDs {
			interest := models.UserInterest{
				UserID:     userID,
				CategoryID: catID,
			}
			if err := tx.Create(&interest).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetInterests returns the categories a user is interested in
func (r *userRepository) GetInterests(userID uuid.UUID) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.Model(&models.Category{}).
		Joins("JOIN user_interests ON user_interests.category_id = categories.id").
		Where("user_interests.user_id = ?", userID).
		Find(&categories).Error
	return categories, err
}

// GetInterestIDs returns the category IDs a user is interested in
func (r *userRepository) GetInterestIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&models.UserInterest{}).
		Where("user_id = ?", userID).
		Pluck("category_id", &ids).Error
	return ids, err
}
