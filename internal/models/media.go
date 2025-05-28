// Add this to internal/models/media.go (create new file)
package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Media represents a media file (image, video, audio, document)
type Media struct {
	BaseModel `bson:",inline"`

	// File information
	OriginalName  string `json:"original_name" bson:"original_name" validate:"required"`
	FileName      string `json:"file_name" bson:"file_name" validate:"required"`
	FilePath      string `json:"file_path" bson:"file_path" validate:"required"`
	FileSize      int64  `json:"file_size" bson:"file_size" validate:"required"`
	MimeType      string `json:"mime_type" bson:"mime_type" validate:"required"`
	FileExtension string `json:"file_extension" bson:"file_extension" validate:"required"`

	// Media properties
	Type     string `json:"type" bson:"type" validate:"required,oneof=image video audio document"`
	Category string `json:"category,omitempty" bson:"category,omitempty"`
	Width    int    `json:"width,omitempty" bson:"width,omitempty"`
	Height   int    `json:"height,omitempty" bson:"height,omitempty"`
	Duration int    `json:"duration,omitempty" bson:"duration,omitempty"` // in seconds

	// URLs
	URL string `json:"url" bson:"url" validate:"required"`

	// Content information
	AltText     string `json:"alt_text,omitempty" bson:"alt_text,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Caption     string `json:"caption,omitempty" bson:"caption,omitempty"`

	// Ownership and access
	UploadedBy   primitive.ObjectID   `json:"uploaded_by" bson:"uploaded_by" validate:"required"`
	IsPublic     bool                 `json:"is_public" bson:"is_public"`
	AccessPolicy string               `json:"access_policy" bson:"access_policy"` // public, private, restricted
	AllowedUsers []primitive.ObjectID `json:"allowed_users,omitempty" bson:"allowed_users,omitempty"`

	// Usage tracking
	ViewCount     int64 `json:"view_count" bson:"view_count"`
	DownloadCount int64 `json:"download_count" bson:"download_count"`

	// Related content
	RelatedTo string              `json:"related_to,omitempty" bson:"related_to,omitempty"` // post, story, message, profile
	RelatedID *primitive.ObjectID `json:"related_id,omitempty" bson:"related_id,omitempty"`

	// Processing status
	IsProcessed      bool       `json:"is_processed" bson:"is_processed"`
	ProcessingStatus string     `json:"processing_status" bson:"processing_status"` // pending, processing, completed, failed
	ProcessedAt      *time.Time `json:"processed_at,omitempty" bson:"processed_at,omitempty"`

	// Storage information
	StorageProvider string `json:"storage_provider" bson:"storage_provider"` // local, s3, cloudinary
	StorageKey      string `json:"storage_key" bson:"storage_key"`
	StorageBucket   string `json:"storage_bucket,omitempty" bson:"storage_bucket,omitempty"`

	// Thumbnails and variants
	Thumbnails []MediaVariant `json:"thumbnails,omitempty" bson:"thumbnails,omitempty"`
	Variants   []MediaVariant `json:"variants,omitempty" bson:"variants,omitempty"`

	// Expiration
	ExpiresAt *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	IsExpired bool       `json:"is_expired" bson:"is_expired"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Tags     []string               `json:"tags,omitempty" bson:"tags,omitempty"`

	// Moderation
	IsModerationRequired bool   `json:"is_moderation_required" bson:"is_moderation_required"`
	ModerationStatus     string `json:"moderation_status" bson:"moderation_status"` // pending, approved, rejected
	ModerationNotes      string `json:"moderation_notes,omitempty" bson:"moderation_notes,omitempty"`
}

// MediaVariant represents different sizes/formats of media
type MediaVariant struct {
	Name      string    `json:"name" bson:"name"` // thumbnail, small, medium, large
	URL       string    `json:"url" bson:"url"`
	Width     int       `json:"width,omitempty" bson:"width,omitempty"`
	Height    int       `json:"height,omitempty" bson:"height,omitempty"`
	FileSize  int64     `json:"file_size" bson:"file_size"`
	Format    string    `json:"format" bson:"format"` // jpg, png, webp, etc.
	Quality   int       `json:"quality,omitempty" bson:"quality,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// MediaResponse represents media data returned in API responses
type MediaResponse struct {
	ID               string                 `json:"id"`
	OriginalName     string                 `json:"original_name"`
	FileName         string                 `json:"file_name"`
	FileSize         int64                  `json:"file_size"`
	MimeType         string                 `json:"mime_type"`
	FileExtension    string                 `json:"file_extension"`
	Type             string                 `json:"type"`
	Category         string                 `json:"category,omitempty"`
	Width            int                    `json:"width,omitempty"`
	Height           int                    `json:"height,omitempty"`
	Duration         int                    `json:"duration,omitempty"`
	URL              string                 `json:"url"`
	AltText          string                 `json:"alt_text,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Caption          string                 `json:"caption,omitempty"`
	UploadedBy       string                 `json:"uploaded_by"`
	IsPublic         bool                   `json:"is_public"`
	ViewCount        int64                  `json:"view_count"`
	DownloadCount    int64                  `json:"download_count"`
	RelatedTo        string                 `json:"related_to,omitempty"`
	RelatedID        string                 `json:"related_id,omitempty"`
	IsProcessed      bool                   `json:"is_processed"`
	ProcessingStatus string                 `json:"processing_status"`
	StorageProvider  string                 `json:"storage_provider"`
	Thumbnails       []MediaVariant         `json:"thumbnails,omitempty"`
	Variants         []MediaVariant         `json:"variants,omitempty"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	IsExpired        bool                   `json:"is_expired"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// CreateMediaRequest represents the request to upload media
type CreateMediaRequest struct {
	Type        string                 `json:"type" validate:"required,oneof=image video audio document"`
	Category    string                 `json:"category,omitempty"`
	AltText     string                 `json:"alt_text,omitempty" validate:"max=250"`
	Description string                 `json:"description,omitempty" validate:"max=1000"`
	Caption     string                 `json:"caption,omitempty" validate:"max=500"`
	RelatedTo   string                 `json:"related_to,omitempty"`
	RelatedID   string                 `json:"related_id,omitempty"`
	IsPublic    bool                   `json:"is_public"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateMediaRequest represents the request to update media
type UpdateMediaRequest struct {
	AltText     *string  `json:"alt_text,omitempty" validate:"omitempty,max=250"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	Caption     *string  `json:"caption,omitempty" validate:"omitempty,max=500"`
	IsPublic    *bool    `json:"is_public,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// MediaSearchRequest represents media search parameters
type MediaSearchRequest struct {
	Query    string   `json:"query" validate:"required,min=2"`
	Type     string   `json:"type,omitempty" validate:"omitempty,oneof=image video audio document"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	UserID   string   `json:"user_id,omitempty"`
	IsPublic *bool    `json:"is_public,omitempty"`
	Page     int      `json:"page,omitempty" validate:"min=1"`
	Limit    int      `json:"limit,omitempty" validate:"min=1,max=50"`
}

// Methods for Media model

// BeforeCreate sets default values before creating media
func (m *Media) BeforeCreate() {
	m.BaseModel.BeforeCreate()

	// Set default values
	m.ViewCount = 0
	m.DownloadCount = 0
	m.IsProcessed = false
	m.ProcessingStatus = "pending"
	m.IsExpired = false
	m.IsModerationRequired = false
	m.ModerationStatus = "pending"

	// Set default access policy
	if m.AccessPolicy == "" {
		if m.IsPublic {
			m.AccessPolicy = "public"
		} else {
			m.AccessPolicy = "private"
		}
	}

	// Set default storage provider
	if m.StorageProvider == "" {
		m.StorageProvider = "local"
	}
}

// ToMediaResponse converts Media model to MediaResponse
func (m *Media) ToMediaResponse() MediaResponse {
	response := MediaResponse{
		ID:               m.ID.Hex(),
		OriginalName:     m.OriginalName,
		FileName:         m.FileName,
		FileSize:         m.FileSize,
		MimeType:         m.MimeType,
		FileExtension:    m.FileExtension,
		Type:             m.Type,
		Category:         m.Category,
		Width:            m.Width,
		Height:           m.Height,
		Duration:         m.Duration,
		URL:              m.URL,
		AltText:          m.AltText,
		Description:      m.Description,
		Caption:          m.Caption,
		UploadedBy:       m.UploadedBy.Hex(),
		IsPublic:         m.IsPublic,
		ViewCount:        m.ViewCount,
		DownloadCount:    m.DownloadCount,
		RelatedTo:        m.RelatedTo,
		IsProcessed:      m.IsProcessed,
		ProcessingStatus: m.ProcessingStatus,
		StorageProvider:  m.StorageProvider,
		Thumbnails:       m.Thumbnails,
		Variants:         m.Variants,
		ExpiresAt:        m.ExpiresAt,
		IsExpired:        m.IsExpired,
		Metadata:         m.Metadata,
		Tags:             m.Tags,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}

	if m.RelatedID != nil {
		response.RelatedID = m.RelatedID.Hex()
	}

	return response
}

// IncrementViewCount increments the view count
func (m *Media) IncrementViewCount() {
	m.ViewCount++
	m.BeforeUpdate()
}

// IncrementDownloadCount increments the download count
func (m *Media) IncrementDownloadCount() {
	m.DownloadCount++
	m.BeforeUpdate()
}

// MarkAsProcessed marks the media as processed
func (m *Media) MarkAsProcessed() {
	m.IsProcessed = true
	m.ProcessingStatus = "completed"
	now := time.Now()
	m.ProcessedAt = &now
	m.BeforeUpdate()
}

// MarkProcessingFailed marks processing as failed
func (m *Media) MarkProcessingFailed() {
	m.IsProcessed = false
	m.ProcessingStatus = "failed"
	m.BeforeUpdate()
}

// CheckExpiration checks and updates expiration status
func (m *Media) CheckExpiration() {
	if !m.IsExpired && m.ExpiresAt != nil && time.Now().After(*m.ExpiresAt) {
		m.IsExpired = true
		m.BeforeUpdate()
	}
}

// AddThumbnail adds a thumbnail variant
func (m *Media) AddThumbnail(variant MediaVariant) {
	variant.CreatedAt = time.Now()
	m.Thumbnails = append(m.Thumbnails, variant)
	m.BeforeUpdate()
}

// AddVariant adds a media variant
func (m *Media) AddVariant(variant MediaVariant) {
	variant.CreatedAt = time.Now()
	m.Variants = append(m.Variants, variant)
	m.BeforeUpdate()
}

// GetVariantURL returns URL for a specific variant
func (m *Media) GetVariantURL(variantName string) string {
	// Check thumbnails first
	for _, thumb := range m.Thumbnails {
		if thumb.Name == variantName {
			return thumb.URL
		}
	}

	// Check variants
	for _, variant := range m.Variants {
		if variant.Name == variantName {
			return variant.URL
		}
	}

	// Return original URL if variant not found
	return m.URL
}

// HasThumbnail checks if media has a specific thumbnail
func (m *Media) HasThumbnail(name string) bool {
	for _, thumb := range m.Thumbnails {
		if thumb.Name == name {
			return true
		}
	}
	return false
}

// GetFileExtensionFromMime returns file extension based on MIME type
func (m *Media) GetFileExtensionFromMime() string {
	switch m.MimeType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "audio/mpeg":
		return "mp3"
	case "audio/wav":
		return "wav"
	case "application/pdf":
		return "pdf"
	default:
		return m.FileExtension
	}
}

// IsImage checks if media is an image
func (m *Media) IsImage() bool {
	return m.Type == "image"
}

// IsVideo checks if media is a video
func (m *Media) IsVideo() bool {
	return m.Type == "video"
}

// IsAudio checks if media is audio
func (m *Media) IsAudio() bool {
	return m.Type == "audio"
}

// IsDocument checks if media is a document
func (m *Media) IsDocument() bool {
	return m.Type == "document"
}

// GetAspectRatio returns aspect ratio for images/videos
func (m *Media) GetAspectRatio() float64 {
	if m.Height == 0 {
		return 0
	}
	return float64(m.Width) / float64(m.Height)
}

// GetFormattedFileSize returns human-readable file size
func (m *Media) GetFormattedFileSize() string {
	const unit = 1024
	if m.FileSize < unit {
		return fmt.Sprintf("%d B", m.FileSize)
	}

	div, exp := int64(unit), 0
	for n := m.FileSize / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(m.FileSize)/float64(div), units[exp])
}
