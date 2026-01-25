package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category represents a blog category with hierarchical support
type Category struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Slug         string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Description  string     `gorm:"type:text" json:"description"`
	ParentID     *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	DisplayOrder int        `gorm:"default:0" json:"display_order"`
	IsActive     bool       `gorm:"default:true;index" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Relationships
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Articles []Article  `gorm:"many2many:article_categories;" json:"articles,omitempty"`
}

// TableName returns the table name for the Category model
func (Category) TableName() string {
	return "categories"
}

// BeforeCreate is a GORM hook that runs before creating a category
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsRoot checks if the category is a root category (no parent)
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren checks if the category has any children
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

// GetPath returns the full path of category names from root to this category
func (c *Category) GetPath() []string {
	path := []string{c.Name}
	if c.Parent != nil {
		parentPath := c.Parent.GetPath()
		path = append(parentPath, path...)
	}
	return path
}

// GetDepth returns the depth of the category in the hierarchy
func (c *Category) GetDepth() int {
	if c.Parent == nil {
		return 0
	}
	return c.Parent.GetDepth() + 1
}
