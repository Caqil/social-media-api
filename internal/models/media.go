// models/media.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Media represents a media file (image, video, audio, document) in the system
type Media struct {
	BaseModel `bson:",inline"`

	// File Information
	OriginalName  string `json:"original_name" bson:"original_name" validate:"required"`
	FileName      string `json:"file_name" bson:"file_name" validate:"required"`
	FilePath      string `json:"file_path" bson:"file_path" validate:"required"`
	FileSize      int64  `json:"file_size" bson:"file_size" validate:"required"`
	MimeType      string `json:"mime_type" bson:"mime_type" validate:"required"`
	FileExtension string `json:"file_extension" bson:"file_extension" validate:"required"`

	// Media Type and Category
	Type     string `json:"type" bson:"type" validate:"required"`         // image, video, audio, document
	Category string `json:"category,omitempty" bson:"category,omitempty"` // profile_pic, cover_pic, post_media, story_media, etc.

	// Upload Information
	UploadedBy primitive.ObjectID `json:"uploaded_by" bson:"uploaded_by" validate:"required"`
	Uploader   UserResponse       `json:"uploader,omitempty" bson:"-"`              // Populated when querying
	Source     string             `json:"source,omitempty" bson:"source,omitempty"` // web, mobile, api

	// Media Metadata
	Width      int                    `json:"width,omitempty" bson:"width,omitempty"`
	Height     int                    `json:"height,omitempty" bson:"height,omitempty"`
	Duration   int                    `json:"duration,omitempty" bson:"duration,omitempty"` // For video/audio in seconds
	Bitrate    int                    `json:"bitrate,omitempty" bson:"bitrate,omitempty"`
	Framerate  float64                `json:"framerate,omitempty" bson:"framerate,omitempty"`
	ColorSpace string                 `json:"color_space,omitempty" bson:"color_space,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// Processing Status
	IsProcessed      bool       `json:"is_processed" bson:"is_processed"`
	ProcessingStatus string     `json:"processing_status" bson:"processing_status"` // pending, processing, completed, failed
	ProcessedAt      *time.Time `json:"processed_at,omitempty" bson:"processed_at,omitempty"`
	ProcessingError  string     `json:"processing_error,omitempty" bson:"processing_error,omitempty"`

	// Variants/Thumbnails
	Thumbnails []MediaVariant `json:"thumbnails,omitempty" bson:"thumbnails,omitempty"`
	Variants   []MediaVariant `json:"variants,omitempty" bson:"variants,omitempty"` // Different sizes/qualities

	// URLs
	URL          string `json:"url" bson:"url" validate:"required"`
	ThumbnailURL string `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	CDNUrl       string `json:"cdn_url,omitempty" bson:"cdn_url,omitempty"`

	// Access and Usage
	IsPublic      bool                 `json:"is_public" bson:"is_public"`
	AccessPolicy  string               `json:"access_policy,omitempty" bson:"access_policy,omitempty"` // public, private, restricted
	AllowedUsers  []primitive.ObjectID `json:"allowed_users,omitempty" bson:"allowed_users,omitempty"`
	DownloadCount int64                `json:"download_count" bson:"download_count"`
	ViewCount     int64                `json:"view_count" bson:"view_count"`

	// Content Moderation
	IsModerated      bool   `json:"is_moderated" bson:"is_moderated"`
	ModerationStatus string `json:"moderation_status" bson:"moderation_status"` // pending, approved, rejected
	ModerationNote   string `json:"moderation_note,omitempty" bson:"moderation_note,omitempty"`
	IsReported       bool   `json:"is_reported" bson:"is_reported"`
	ReportsCount     int64  `json:"reports_count" bson:"reports_count"`

	// Storage Information
	StorageProvider string                 `json:"storage_provider" bson:"storage_provider"` // local, s3, gcs, azure
	StorageKey      string                 `json:"storage_key" bson:"storage_key"`
	StorageMetadata map[string]interface{} `json:"storage_metadata,omitempty" bson:"storage_metadata,omitempty"`

	// Associated Content
	RelatedTo string              `json:"related_to,omitempty" bson:"related_to,omitempty"` // post, story, profile, group, event
	RelatedID *primitive.ObjectID `json:"related_id,omitempty" bson:"related_id,omitempty"`

	// Alt Text for Accessibility
	AltText     string `json:"alt_text,omitempty" bson:"alt_text,omitempty" validate:"max=200"`
	Description string `json:"description,omitempty" bson:"description,omitempty" validate:"max=1000"`

	// Expiration (for temporary media like stories)
	ExpiresAt *time.Time `json:"expires_at,omitempty" bson:"expires_at,omitempty"`
	IsExpired bool       `json:"is_expired" bson:"is_expired"`
}

// MediaVariant represents different sizes/qualities of a media file
type MediaVariant struct {
	Name     string `json:"name" bson:"name"` // thumbnail, small, medium, large, original
	URL      string `json:"url" bson:"url"`
	Width    int    `json:"width,omitempty" bson:"width,omitempty"`
	Height   int    `json:"height,omitempty" bson:"height,omitempty"`
	FileSize int64  `json:"file_size" bson:"file_size"`
	Quality  string `json:"quality,omitempty" bson:"quality,omitempty"` // low, medium, high
	Format   string `json:"format" bson:"format"`                       // jpg, png, webp, mp4, etc.
}

// MediaResponse represents media data returned in API responses
type MediaResponse struct {
	ID               string         `json:"id"`
	OriginalName     string         `json:"original_name"`
	FileName         string         `json:"file_name"`
	FileSize         int64          `json:"file_size"`
	MimeType         string         `json:"mime_type"`
	FileExtension    string         `json:"file_extension"`
	Type             string         `json:"type"`
	Category         string         `json:"category,omitempty"`
	UploadedBy       string         `json:"uploaded_by"`
	Uploader         UserResponse   `json:"uploader,omitempty"`
	Width            int            `json:"width,omitempty"`
	Height           int            `json:"height,omitempty"`
	Duration         int            `json:"duration,omitempty"`
	IsProcessed      bool           `json:"is_processed"`
	ProcessingStatus string         `json:"processing_status"`
	Thumbnails       []MediaVariant `json:"thumbnails,omitempty"`
	Variants         []MediaVariant `json:"variants,omitempty"`
	URL              string         `json:"url"`
	ThumbnailURL     string         `json:"thumbnail_url,omitempty"`
	CDNUrl           string         `json:"cdn_url,omitempty"`
	IsPublic         bool           `json:"is_public"`
	DownloadCount    int64          `json:"download_count"`
	ViewCount        int64          `json:"view_count"`
	ModerationStatus string         `json:"moderation_status"`
	AltText          string         `json:"alt_text,omitempty"`
	Description      string         `json:"description,omitempty"`
	ExpiresAt        *time.Time     `json:"expires_at,omitempty"`
	IsExpired        bool           `json:"is_expired"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`

	// User-specific context
	CanDownload bool `json:"can_download,omitempty"`
	CanEdit     bool `json:"can_edit,omitempty"`
	CanDelete   bool `json:"can_delete,omitempty"`
}

// CreateMediaRequest represents media upload request
type CreateMediaRequest struct {
	OriginalName string     `json:"original_name" validate:"required"`
	Type         string     `json:"type" validate:"required,oneof=image video audio document"`
	Category     string     `json:"category,omitempty"`
	IsPublic     bool       `json:"is_public"`
	AltText      string     `json:"alt_text,omitempty" validate:"max=200"`
	Description  string     `json:"description,omitempty" validate:"max=1000"`
	RelatedTo    string     `json:"related_to,omitempty"`
	RelatedID    string     `json:"related_id,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// UpdateMediaRequest represents media update request
type UpdateMediaRequest struct {
	AltText     *string `json:"alt_text,omitempty" validate:"omitempty,max=200"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}


// Methods for Media model

// BeforeCreate sets default values before creating media
func (m *Media) BeforeCreate() {
	m.BaseModel.BeforeCreate()

	// Set default values
	m.IsProcessed = false
	m.ProcessingStatus = "pending"
	m.IsPublic = true
	m.DownloadCount = 0
	m.ViewCount = 0
	m.IsModerated = false
	m.ModerationStatus = "pending"
	m.IsReported = false
	m.ReportsCount = 0
	m.IsExpired = false
	m.StorageProvider = "local" // Default to local storage

	// Set default access policy
	if m.AccessPolicy == "" {
		if m.IsPublic {
			m.AccessPolicy = "public"
		} else {
			m.AccessPolicy = "private"
		}
	}
}

// ToMediaResponse converts Media to MediaResponse
func (m *Media) ToMediaResponse() MediaResponse {
	return MediaResponse{
		ID:               m.ID.Hex(),
		OriginalName:     m.OriginalName,
		FileName:         m.FileName,
		FileSize:         m.FileSize,
		MimeType:         m.MimeType,
		FileExtension:    m.FileExtension,
		Type:             m.Type,
		Category:         m.Category,
		UploadedBy:       m.UploadedBy.Hex(),
		Width:            m.Width,
		Height:           m.Height,
		Duration:         m.Duration,
		IsProcessed:      m.IsProcessed,
		ProcessingStatus: m.ProcessingStatus,
		Thumbnails:       m.Thumbnails,
		Variants:         m.Variants,
		URL:              m.URL,
		ThumbnailURL:     m.ThumbnailURL,
		CDNUrl:           m.CDNUrl,
		IsPublic:         m.IsPublic,
		DownloadCount:    m.DownloadCount,
		ViewCount:        m.ViewCount,
		ModerationStatus: m.ModerationStatus,
		AltText:          m.AltText,
		Description:      m.Description,
		ExpiresAt:        m.ExpiresAt,
		IsExpired:        m.IsExpired,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

// MarkAsProcessed marks the media as processed
func (m *Media) MarkAsProcessed() {
	m.IsProcessed = true
	m.ProcessingStatus = "completed"
	now := time.Now()
	m.ProcessedAt = &now
	m.BeforeUpdate()
}

// MarkProcessingFailed marks the media processing as failed
func (m *Media) MarkProcessingFailed(error string) {
	m.ProcessingStatus = "failed"
	m.ProcessingError = error
	m.BeforeUpdate()
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

// CanAccessMedia checks if a user can access this media
func (m *Media) CanAccessMedia(currentUserID primitive.ObjectID) bool {
	// Owner can always access
	if m.UploadedBy == currentUserID {
		return true
	}

	// Check if public
	if m.IsPublic && m.AccessPolicy == "public" {
		return true
	}

	// Check if user is in allowed users list
	if m.AccessPolicy == "restricted" {
		for _, allowedUserID := range m.AllowedUsers {
			if allowedUserID == currentUserID {
				return true
			}
		}
	}

	return false
}

// CheckExpiration checks and updates expiration status
func (m *Media) CheckExpiration() {
	if !m.IsExpired && m.ExpiresAt != nil && time.Now().After(*m.ExpiresAt) {
		m.IsExpired = true
		m.BeforeUpdate()
	}
}


// GetMediaTypes returns all supported media types
func GetMediaTypes() []string {
	return []string{"image", "video", "audio", "document"}
}

// IsValidMediaType checks if a media type is valid
func IsValidMediaType(mediaType string) bool {
	validTypes := GetMediaTypes()
	for _, validType := range validTypes {
		if mediaType == validType {
			return true
		}
	}
	return false
}

// GetSupportedImageFormats returns supported image formats
func GetSupportedImageFormats() []string {
	return []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "tiff"}
}

// GetSupportedVideoFormats returns supported video formats
func GetSupportedVideoFormats() []string {
	return []string{"mp4", "mov", "avi", "mkv", "webm", "flv", "wmv"}
}

// GetSupportedAudioFormats returns supported audio formats
func GetSupportedAudioFormats() []string {
	return []string{"mp3", "wav", "ogg", "aac", "flac", "m4a"}
}

// GetSupportedDocumentFormats returns supported document formats
func GetSupportedDocumentFormats() []string {
	return []string{"pdf", "doc", "docx", "txt", "rtf", "odt"}
}
