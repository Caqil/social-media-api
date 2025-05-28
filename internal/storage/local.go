// local.go
package storage

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalProvider implements StorageProvider for local file system storage
type LocalProvider struct {
	basePath  string // Base directory for file storage
	baseURL   string // Base URL for serving files
	urlPrefix string // URL prefix for file access
}

// NewLocalProvider creates a new local storage provider
func NewLocalProvider(config StorageConfig) (*LocalProvider, error) {
	basePath := config.LocalPath
	if basePath == "" {
		basePath = "./uploads" // Default upload directory
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory %s: %w", basePath, err)
	}

	// Determine base URL for serving files
	baseURL := config.CDNDomain
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default base URL
	}

	urlPrefix := "/uploads" // Default URL prefix
	if prefix, exists := config.Options["url_prefix"]; exists {
		urlPrefix = prefix
	}

	return &LocalProvider{
		basePath:  basePath,
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		urlPrefix: strings.TrimSuffix(urlPrefix, "/"),
	}, nil
}

// Upload uploads a file to local storage
func (l *LocalProvider) Upload(key string, data io.Reader, contentType string, size int64) (*UploadResult, error) {
	// Generate full file path
	fullPath := filepath.Join(l.basePath, key)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to create directory: %v", err), key)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to create file: %v", err), key)
	}
	defer file.Close()

	// Copy data to file and calculate size and hash
	hash := md5.New()
	multiWriter := io.MultiWriter(file, hash)

	actualSize, err := io.Copy(multiWriter, data)
	if err != nil {
		// Clean up partial file
		os.Remove(fullPath)
		return nil, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to write file: %v", err), key)
	}

	// Validate size if provided
	if size > 0 && actualSize != size {
		os.Remove(fullPath)
		return nil, NewStorageErrorWithKey(ErrCodeInvalidInput,
			fmt.Sprintf("size mismatch: expected %d, got %d", size, actualSize), key)
	}

	// Set file permissions
	if err := file.Chmod(0644); err != nil {
		return nil, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to set file permissions: %v", err), key)
	}

	// Generate ETag from MD5 hash
	etag := fmt.Sprintf("%x", hash.Sum(nil))

	// Generate URL
	fileURL := l.generateURL(key)

	return &UploadResult{
		Key:         key,
		URL:         fileURL,
		CDNUrl:      fileURL, // Same as URL for local storage
		Size:        actualSize,
		ContentType: contentType,
		ETag:        etag,
	}, nil
}

// Download downloads a file from local storage
func (l *LocalProvider) Download(key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewStorageErrorWithKey(ErrCodeNotFound, "file not found", key)
		}
		return nil, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to open file: %v", err), key)
	}

	return file, nil
}

// Delete removes a file from local storage
func (l *LocalProvider) Delete(key string) error {
	fullPath := filepath.Join(l.basePath, key)

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, consider it deleted
		}
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to delete file: %v", err), key)
	}

	// Try to remove empty parent directories
	l.cleanupEmptyDirs(filepath.Dir(fullPath))

	return nil
}

// Exists checks if a file exists in local storage
func (l *LocalProvider) Exists(key string) (bool, error) {
	fullPath := filepath.Join(l.basePath, key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to check file existence: %v", err), key)
	}

	return true, nil
}

// GetURL generates a public URL for the file
func (l *LocalProvider) GetURL(key string) (string, error) {
	return l.generateURL(key), nil
}

// GetSignedURL generates a temporary signed URL (not applicable for local storage)
func (l *LocalProvider) GetSignedURL(key string, expiration time.Duration) (string, error) {
	// For local storage, we can't really create signed URLs
	// Return regular URL with a token or timestamp
	baseURL := l.generateURL(key)

	// Add expiration timestamp as query parameter
	expiryTime := time.Now().Add(expiration).Unix()
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to parse URL: %v", err), key)
	}

	query := parsedURL.Query()
	query.Set("expires", fmt.Sprintf("%d", expiryTime))
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// GetMetadata retrieves file metadata from local storage
func (l *LocalProvider) GetMetadata(key string) (*FileMetadata, error) {
	fullPath := filepath.Join(l.basePath, key)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewStorageErrorWithKey(ErrCodeNotFound, "file not found", key)
		}
		return nil, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to get file info: %v", err), key)
	}

	// Calculate MD5 hash for ETag
	etag, err := l.calculateFileHash(fullPath)
	if err != nil {
		etag = "" // Continue without ETag if calculation fails
	}

	// Determine content type from file extension
	contentType := getContentTypeFromExtension(filepath.Ext(fullPath))

	return &FileMetadata{
		Key:          key,
		Size:         info.Size(),
		ContentType:  contentType,
		LastModified: info.ModTime(),
		ETag:         etag,
		Metadata:     make(map[string]string), // Local storage doesn't support custom metadata
	}, nil
}

// Copy copies a file within local storage
func (l *LocalProvider) Copy(sourceKey, destKey string) error {
	sourcePath := filepath.Join(l.basePath, sourceKey)
	destPath := filepath.Join(l.basePath, destKey)

	// Check if source exists
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			return NewStorageErrorWithKey(ErrCodeNotFound, "source file not found", sourceKey)
		}
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to access source file: %v", err), sourceKey)
	}

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to create destination directory: %v", err), destKey)
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to open source file: %v", err), sourceKey)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to create destination file: %v", err), destKey)
	}
	defer destFile.Close()

	// Copy data
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		os.Remove(destPath) // Clean up partial file
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to copy file data: %v", err), destKey)
	}

	// Set permissions
	if err := destFile.Chmod(0644); err != nil {
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to set file permissions: %v", err), destKey)
	}

	return nil
}

// Move moves a file within local storage
func (l *LocalProvider) Move(sourceKey, destKey string) error {
	sourcePath := filepath.Join(l.basePath, sourceKey)
	destPath := filepath.Join(l.basePath, destKey)

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to create destination directory: %v", err), destKey)
	}

	// Move file
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewStorageErrorWithKey(ErrCodeNotFound, "source file not found", sourceKey)
		}
		// If rename fails, try copy and delete
		if copyErr := l.Copy(sourceKey, destKey); copyErr != nil {
			return copyErr
		}
		return l.Delete(sourceKey)
	}

	// Clean up empty source directory
	l.cleanupEmptyDirs(filepath.Dir(sourcePath))

	return nil
}

// ListFiles lists files with optional prefix
func (l *LocalProvider) ListFiles(prefix string, limit int) ([]FileInfo, error) {
	searchPath := filepath.Join(l.basePath, prefix)
	var files []FileInfo
	count := 0

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip inaccessible files/directories
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check limit
		if limit > 0 && count >= limit {
			return filepath.SkipDir
		}

		// Get relative path from base path
		relPath, err := filepath.Rel(l.basePath, path)
		if err != nil {
			return nil // Skip files that can't be processed
		}

		// Convert to forward slashes for consistency
		relPath = filepath.ToSlash(relPath)

		files = append(files, FileInfo{
			Key:          relPath,
			Size:         info.Size(),
			LastModified: info.ModTime(),
			IsDir:        false,
		})

		count++
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, NewStorageError(ErrCodeInternal,
			fmt.Sprintf("failed to list files: %v", err))
	}

	return files, nil
}

// GetStorageInfo returns information about this local provider
func (l *LocalProvider) GetStorageInfo() StorageInfo {
	return StorageInfo{
		Provider: "local",
		Region:   "local",
		Bucket:   l.basePath,
		Endpoint: l.baseURL,
	}
}

// Helper methods

// generateURL generates a public URL for a file
func (l *LocalProvider) generateURL(key string) string {
	// Ensure key starts with forward slash for URL consistency
	if !strings.HasPrefix(key, "/") {
		key = "/" + key
	}
	return fmt.Sprintf("%s%s%s", l.baseURL, l.urlPrefix, key)
}

// calculateFileHash calculates MD5 hash of a file for ETag
func (l *LocalProvider) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// cleanupEmptyDirs removes empty directories up the tree
func (l *LocalProvider) cleanupEmptyDirs(dirPath string) {
	// Don't remove the base path
	if dirPath == l.basePath || dirPath == "." {
		return
	}

	// Check if directory is empty
	entries, err := os.ReadDir(dirPath)
	if err != nil || len(entries) > 0 {
		return
	}

	// Remove empty directory
	if err := os.Remove(dirPath); err != nil {
		return
	}

	// Recursively check parent directory
	l.cleanupEmptyDirs(filepath.Dir(dirPath))
}

// getContentTypeFromExtension determines content type from file extension
func getContentTypeFromExtension(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	// Images
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".tiff", ".tif":
		return "image/tiff"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"

	// Videos
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"

	// Audio
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	case ".aac":
		return "audio/aac"
	case ".flac":
		return "audio/flac"
	case ".m4a":
		return "audio/mp4"

	// Documents
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".txt":
		return "text/plain"
	case ".rtf":
		return "application/rtf"
	case ".csv":
		return "text/csv"

	// Archives
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".7z":
		return "application/x-7z-compressed"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"

	// Default
	default:
		return "application/octet-stream"
	}
}

// CleanupExpiredFiles removes expired files from local storage
func (l *LocalProvider) CleanupExpiredFiles(prefix string, olderThan time.Time) error {
	searchPath := filepath.Join(l.basePath, prefix)
	var deleteErrors []error

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible files
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Delete files older than specified time
		if info.ModTime().Before(olderThan) {
			if err := os.Remove(path); err != nil {
				deleteErrors = append(deleteErrors, err)
			}
		}

		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to walk directory: %v", err)
	}

	if len(deleteErrors) > 0 {
		return fmt.Errorf("failed to delete %d files", len(deleteErrors))
	}

	return nil
}

// GetDiskUsage returns disk usage statistics for the storage directory
func (l *LocalProvider) GetDiskUsage() (int64, error) {
	var totalSize int64

	err := filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible files
		}

		if !info.IsDir() {
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate disk usage: %v", err)
	}

	return totalSize, nil
}
