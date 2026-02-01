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
	PasswordHash    string         `gorm:"type:varchar(255)" json:"-"` // Optional for OAuth users
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

	// OAuth fields
	GoogleID     *string `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	AuthProvider string  `gorm:"type:varchar(20);default:'local'" json:"auth_provider"` // local, google, etc.

	// Relationships
	Articles []Article `gorm:"foreignKey:AuthorID" json:"articles,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Media    []Media   `gorm:"foreignKey:UploadedBy" json:"media,omitempty"`

	// Social Graph - Interests (Categories the user wants to follow)
	Interests []Category `gorm:"many2many:user_interests;" json:"interests,omitempty"`

	// Social Graph - Followers/Following (User to User)
	Followers []*User `gorm:"many2many:user_follows;joinForeignKey:FollowingID;joinReferences:FollowerID" json:"followers,omitempty"`
	Following []*User `gorm:"many2many:user_follows;joinForeignKey:FollowerID;joinReferences:FollowingID" json:"following,omitempty"`
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

// IsOAuthUser checks if the user signed up via OAuth
func (u *User) IsOAuthUser() bool {
	return u.AuthProvider != "local" && u.AuthProvider != ""
}

// GetInterestIDs returns a slice of interest (category) IDs
func (u *User) GetInterestIDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(u.Interests))
	for i, cat := range u.Interests {
		ids[i] = cat.ID
	}
	return ids
}

// GetFollowerIDs returns a slice of follower user IDs
func (u *User) GetFollowerIDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(u.Followers))
	for i, user := range u.Followers {
		ids[i] = user.ID
	}
	return ids
}

// GetFollowingIDs returns a slice of following user IDs
func (u *User) GetFollowingIDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(u.Following))
	for i, user := range u.Following {
		ids[i] = user.ID
	}
	return ids
}

// UserFollow represents the follow relationship between users
// This is a custom join table for tracking follow timestamps
type UserFollow struct {
	FollowerID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"follower_id"`
	FollowingID uuid.UUID `gorm:"type:uuid;primaryKey" json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`

	Follower  *User `gorm:"foreignKey:FollowerID" json:"follower,omitempty"`
	Following *User `gorm:"foreignKey:FollowingID" json:"following,omitempty"`
}

// TableName returns the table name for the UserFollow model
func (UserFollow) TableName() string {
	return "user_follows"
}

// UserInterest represents the interest relationship between users and categories
// This is a custom join table for tracking when a user added an interest
type UserInterest struct {
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;primaryKey" json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`

	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName returns the table name for the UserInterest model
func (UserInterest) TableName() string {
	return "user_interests"
}
