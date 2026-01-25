package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleReader UserRole = "reader"
	RoleAuthor UserRole = "author"
	RoleEditor UserRole = "editor"
	RoleAdmin  UserRole = "admin"
)

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleReader, RoleAuthor, RoleEditor, RoleAdmin:
		return true
	}
	return false
}

// HasPermission checks if a role has a minimum permission level
func (r UserRole) HasPermission(required UserRole) bool {
	roleHierarchy := map[UserRole]int{
		RoleReader: 1,
		RoleAuthor: 2,
		RoleEditor: 3,
		RoleAdmin:  4,
	}
	return roleHierarchy[r] >= roleHierarchy[required]
}

// User represents a user in the system
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username        string         `gorm:"type:varchar(30);uniqueIndex;not null" json:"username"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash    string         `gorm:"type:varchar(255);not null" json:"-"`
	FirstName       string         `gorm:"type:varchar(100)" json:"first_name"`
	LastName        string         `gorm:"type:varchar(100)" json:"last_name"`
	Bio             string         `gorm:"type:text" json:"bio"`
	ProfileImageURL *string        `gorm:"type:varchar(500)" json:"profile_image_url"`
	Role            UserRole       `gorm:"type:varchar(20);not null;default:'reader';index" json:"role"`
	IsVerified      bool           `gorm:"default:false" json:"is_verified"`
	IsActive        bool           `gorm:"default:true;index" json:"is_active"`
	LastLoginAt     *time.Time     `json:"last_login_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Articles []Article `gorm:"foreignKey:AuthorID" json:"articles,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Media    []Media   `gorm:"foreignKey:UploadedBy" json:"media,omitempty"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	if u.FirstName == "" {
		return u.LastName
	}
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

// CanCreateArticle checks if the user can create articles
func (u *User) CanCreateArticle() bool {
	return u.Role.HasPermission(RoleAuthor) && u.IsActive
}

// CanEditArticle checks if the user can edit a specific article
func (u *User) CanEditArticle(article *Article) bool {
	if !u.IsActive {
		return false
	}
	// Authors can edit their own articles
	if u.ID == article.AuthorID {
		return true
	}
	// Editors and admins can edit any article
	return u.Role.HasPermission(RoleEditor)
}

// CanDeleteArticle checks if the user can delete a specific article
func (u *User) CanDeleteArticle(article *Article) bool {
	return u.CanEditArticle(article)
}

// CanPublishArticle checks if the user can publish/unpublish articles
func (u *User) CanPublishArticle() bool {
	return u.Role.HasPermission(RoleEditor) && u.IsActive
}

// CanManageCategories checks if the user can manage categories
func (u *User) CanManageCategories() bool {
	return u.Role.HasPermission(RoleEditor) && u.IsActive
}

// CanManageTags checks if the user can manage tags
func (u *User) CanManageTags() bool {
	return u.Role.HasPermission(RoleEditor) && u.IsActive
}

// CanManageUsers checks if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin && u.IsActive
}

// CanViewAllUsers checks if the user can view the user list
func (u *User) CanViewAllUsers() bool {
	return u.Role.HasPermission(RoleEditor) && u.IsActive
}
