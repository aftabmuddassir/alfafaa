package models

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Media represents an uploaded file/media
type Media struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Filename         string    `gorm:"type:varchar(255);not null" json:"filename"`
	OriginalFilename string    `gorm:"type:varchar(255);not null" json:"original_filename"`
	FilePath         string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize         int64     `gorm:"not null" json:"file_size"`
	MimeType         string    `gorm:"type:varchar(100);not null" json:"mime_type"`
	UploadedBy       uuid.UUID `gorm:"type:uuid;not null;index" json:"uploaded_by"`
	IsFeatured       bool      `gorm:"default:false" json:"is_featured"`
	AltText          string    `gorm:"type:varchar(255)" json:"alt_text"`
	CreatedAt        time.Time `json:"created_at"`

	// Relationships
	Uploader *User `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}

// TableName returns the table name for the Media model
func (Media) TableName() string {
	return "media"
}

// BeforeCreate is a GORM hook that runs before creating a media record
func (m *Media) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// GetExtension returns the file extension
func (m *Media) GetExtension() string {
	return strings.ToLower(filepath.Ext(m.Filename))
}

// IsImage checks if the media is an image
func (m *Media) IsImage() bool {
	imageTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	for _, t := range imageTypes {
		if m.MimeType == t {
			return true
		}
	}
	return false
}

// GetFileSizeFormatted returns the file size in a human-readable format
func (m *Media) GetFileSizeFormatted() string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case m.FileSize >= GB:
		return formatFloat(float64(m.FileSize)/float64(GB)) + " GB"
	case m.FileSize >= MB:
		return formatFloat(float64(m.FileSize)/float64(MB)) + " MB"
	case m.FileSize >= KB:
		return formatFloat(float64(m.FileSize)/float64(KB)) + " KB"
	default:
		return formatInt(m.FileSize) + " B"
	}
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(
		strings.Replace(
			string(append([]byte{}, []byte(strings.TrimRight(strings.TrimRight(
				string(append([]byte{}, []byte{byte('0' + int(f)/10), byte('0' + int(f)%10), '.', byte('0' + int(f*10)%10), byte('0' + int(f*100)%10)}...)),
				"0"), "."))...)),
			".", ",", 1),
		"0"), ",")
}

func formatInt(i int64) string {
	s := ""
	for i > 0 {
		s = string(byte('0'+i%10)) + s
		i /= 10
	}
	if s == "" {
		s = "0"
	}
	return s
}

// GetURL returns the public URL for the media
func (m *Media) GetURL(baseURL string) string {
	return baseURL + "/" + m.FilePath
}
