// storage_interface.go
package storage

import (
	"io"
	"time"
)

// StorageProvider defines the interface for different storage backends
type StorageProvider interface {
	// Upload uploads a file to storage and returns the file path/key and URL
	Upload(key string, data io.Reader, contentType string, size int64) (*UploadResult, error)

	// Download retrieves a file from storage
	Download(key string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(key string) error

	// Exists checks if a file exists in storage
	Exists(key string) (bool, error)

	// GetURL generates a public URL for the file
	GetURL(key string) (string, error)

	// GetSignedURL generates a temporary signed URL for private files
	GetSignedURL(key string, expiration time.Duration) (string, error)

	// GetMetadata retrieves file metadata
	GetMetadata(key string) (*FileMetadata, error)

	// Copy copies a file from one location to another
	Copy(sourceKey, destKey string) error

	// Move moves a file from one location to another
	Move(sourceKey, destKey string) error

	// ListFiles lists files with optional prefix
	ListFiles(prefix string, limit int) ([]FileInfo, error)

	// GetStorageInfo returns information about the storage provider
	GetStorageInfo() StorageInfo
}

// UploadResult contains the result of a file upload
type UploadResult struct {
	Key         string            `json:"key"`          // Storage key/path
	URL         string            `json:"url"`          // Public URL
	CDNUrl      string            `json:"cdn_url"`      // CDN URL if available
	Size        int64             `json:"size"`         // File size in bytes
	ContentType string            `json:"content_type"` // MIME type
	ETag        string            `json:"etag"`         // Entity tag for versioning
	Metadata    map[string]string `json:"metadata"`     // Additional metadata
}

// FileMetadata contains metadata about a stored file
type FileMetadata struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified time.Time         `json:"last_modified"`
	ETag         string            `json:"etag"`
	Metadata     map[string]string `json:"metadata"`
}

// FileInfo contains basic information about a file
type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	IsDir        bool      `json:"is_dir"`
}

// StorageInfo contains information about the storage provider
type StorageInfo struct {
	Provider string `json:"provider"` // "s3", "local", "gcs", etc.
	Region   string `json:"region"`   // For cloud providers
	Bucket   string `json:"bucket"`   // Bucket/container name
	Endpoint string `json:"endpoint"` // Custom endpoint if any
}

// UploadOptions contains options for file upload
type UploadOptions struct {
	ContentType     string            `json:"content_type"`
	CacheControl    string            `json:"cache_control"`
	ContentEncoding string            `json:"content_encoding"`
	Metadata        map[string]string `json:"metadata"`
	ACL             string            `json:"acl"` // Access control list
	StorageClass    string            `json:"storage_class"`
}

// StorageConfig contains configuration for storage providers
type StorageConfig struct {
	Provider     string            `json:"provider"`
	Region       string            `json:"region"`
	Bucket       string            `json:"bucket"`
	AccessKey    string            `json:"access_key"`
	SecretKey    string            `json:"secret_key"`
	Endpoint     string            `json:"endpoint"`
	CDNDomain    string            `json:"cdn_domain"`
	LocalPath    string            `json:"local_path"`
	MaxFileSize  int64             `json:"max_file_size"`
	AllowedTypes []string          `json:"allowed_types"`
	Options      map[string]string `json:"options"`
}

// Storage errors
type StorageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Key     string `json:"key,omitempty"`
}

func (e *StorageError) Error() string {
	if e.Key != "" {
		return e.Code + ": " + e.Message + " (key: " + e.Key + ")"
	}
	return e.Code + ": " + e.Message
}

// Common error codes
const (
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAccessDenied  = "ACCESS_DENIED"
	ErrCodeInvalidInput  = "INVALID_INPUT"
	ErrCodeQuotaExceeded = "QUOTA_EXCEEDED"
	ErrCodeNetworkError  = "NETWORK_ERROR"
	ErrCodeInternal      = "INTERNAL_ERROR"
)

// Helper functions for creating errors
func NewStorageError(code, message string) *StorageError {
	return &StorageError{Code: code, Message: message}
}

func NewStorageErrorWithKey(code, message, key string) *StorageError {
	return &StorageError{Code: code, Message: message, Key: key}
}

// Utility functions
func IsNotFoundError(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.Code == ErrCodeNotFound
	}
	return false
}

func IsAccessDeniedError(err error) bool {
	if storageErr, ok := err.(*StorageError); ok {
		return storageErr.Code == ErrCodeAccessDenied
	}
	return false
}
