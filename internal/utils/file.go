// utils/file.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FileInfo represents file information
type FileInfo struct {
	OriginalName string
	FileName     string
	FilePath     string
	FileSize     int64
	MimeType     string
	Extension    string
}

// FileUploadResult represents the result of a file upload
type FileUploadResult struct {
	FileInfo
	URL          string
	ThumbnailURL string
	Success      bool
	Error        string
}

// CreateDirectory creates a directory if it doesn't exist
func CreateDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// GenerateFileName generates a unique filename with timestamp and random string
func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomString := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%d_%s%s", timestamp, randomString, ext)
}

// GenerateFileNameWithPrefix generates a filename with a prefix
func GenerateFileNameWithPrefix(prefix, originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 6)
	rand.Read(randomBytes)
	randomString := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("%s_%d_%s%s", prefix, timestamp, randomString, ext)
}

// GetFileExtension returns the file extension in lowercase
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// GetMimeType returns the MIME type based on file extension
func GetMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		// Images
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".tiff": "image/tiff",
		".svg":  "image/svg+xml",

		// Videos
		".mp4":  "video/mp4",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",
		".mkv":  "video/x-matroska",
		".webm": "video/webm",
		".flv":  "video/x-flv",
		".wmv":  "video/x-ms-wmv",
		".m4v":  "video/x-m4v",

		// Audio
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".aac":  "audio/aac",
		".flac": "audio/flac",
		".m4a":  "audio/mp4",
		".wma":  "audio/x-ms-wma",

		// Documents
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".rtf":  "application/rtf",
		".odt":  "application/vnd.oasis.opendocument.text",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
		".7z":   "application/x-7z-compressed",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// GetFileType returns the general file type category
func GetFileType(filename string) string {
	mimeType := GetMimeType(filename)

	if strings.HasPrefix(mimeType, "image/") {
		return "image"
	}
	if strings.HasPrefix(mimeType, "video/") {
		return "video"
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return "audio"
	}
	if strings.HasPrefix(mimeType, "application/") || strings.HasPrefix(mimeType, "text/") {
		return "document"
	}

	return "unknown"
}

// IsValidFileType checks if the file type is allowed
func IsValidFileType(filename string, allowedTypes []string) bool {
	mimeType := GetMimeType(filename)

	for _, allowedType := range allowedTypes {
		if mimeType == allowedType {
			return true
		}
	}

	return false
}

// IsImageFile checks if the file is an image
func IsImageFile(filename string) bool {
	return IsValidFileType(filename, SupportedImageTypes)
}

// IsVideoFile checks if the file is a video
func IsVideoFile(filename string) bool {
	return IsValidFileType(filename, SupportedVideoTypes)
}

// IsAudioFile checks if the file is an audio file
func IsAudioFile(filename string) bool {
	return IsValidFileType(filename, SupportedAudioTypes)
}

// IsDocumentFile checks if the file is a document
func IsDocumentFile(filename string) bool {
	return IsValidFileType(filename, SupportedDocumentTypes)
}

// ValidateFileSize checks if file size is within limits
func ValidateFileSize(size int64, fileType string) error {
	maxSize := int64(0)

	switch fileType {
	case "image":
		maxSize = MaxImageSizeMB * 1024 * 1024
	case "video":
		maxSize = MaxVideoSizeMB * 1024 * 1024
	case "audio":
		maxSize = MaxAudioSizeMB * 1024 * 1024
	case "document":
		maxSize = MaxDocumentSizeMB * 1024 * 1024
	default:
		maxSize = MaxDocumentSizeMB * 1024 * 1024
	}

	if size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %dMB", maxSize/(1024*1024))
	}

	return nil
}

// SaveUploadedFile saves an uploaded file to the specified directory
func SaveUploadedFile(file *multipart.FileHeader, uploadDir, category string) (*FileUploadResult, error) {
	// Validate file type
	fileType := GetFileType(file.Filename)
	if !isAllowedFileType(file.Filename) {
		return &FileUploadResult{
			Success: false,
			Error:   "Unsupported file type",
		}, fmt.Errorf("unsupported file type")
	}

	// Validate file size
	if err := ValidateFileSize(file.Size, fileType); err != nil {
		return &FileUploadResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// Generate unique filename
	fileName := GenerateFileNameWithPrefix(category, file.Filename)

	// Create upload directory
	fullUploadDir := filepath.Join(uploadDir, category, fileType)
	if err := CreateDirectory(fullUploadDir); err != nil {
		return &FileUploadResult{
			Success: false,
			Error:   "Failed to create upload directory",
		}, err
	}

	// Full file path
	filePath := filepath.Join(fullUploadDir, fileName)

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return &FileUploadResult{
			Success: false,
			Error:   "Failed to open uploaded file",
		}, err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return &FileUploadResult{
			Success: false,
			Error:   "Failed to create destination file",
		}, err
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		return &FileUploadResult{
			Success: false,
			Error:   "Failed to save file",
		}, err
	}

	// Generate URL (relative path)
	relativePath := filepath.Join(category, fileType, fileName)
	fileURL := "/" + strings.ReplaceAll(relativePath, "\\", "/")

	return &FileUploadResult{
		FileInfo: FileInfo{
			OriginalName: file.Filename,
			FileName:     fileName,
			FilePath:     filePath,
			FileSize:     file.Size,
			MimeType:     GetMimeType(file.Filename),
			Extension:    GetFileExtension(file.Filename),
		},
		URL:     fileURL,
		Success: true,
	}, nil
}

// DeleteFile deletes a file from the filesystem
func DeleteFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, consider it deleted
	}

	return os.Remove(filePath)
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// FormatFileSize formats file size in human readable format
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// CleanupTempFiles removes temporary files older than specified duration
func CleanupTempFiles(tempDir string, maxAge time.Duration) error {
	return filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && time.Since(info.ModTime()) > maxAge {
			return os.Remove(path)
		}

		return nil
	})
}

// MoveFile moves a file from source to destination
func MoveFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := CreateDirectory(dstDir); err != nil {
		return err
	}

	// Try to rename first (works if on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails, copy and delete
	if err := copyFile(src, dst); err != nil {
		return err
	}

	return os.Remove(src)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// isAllowedFileType checks if file type is in the allowed list
func isAllowedFileType(filename string) bool {
	allAllowedTypes := append(append(append(SupportedImageTypes, SupportedVideoTypes...), SupportedAudioTypes...), SupportedDocumentTypes...)
	return IsValidFileType(filename, allAllowedTypes)
}

// GetUploadPath returns the upload path for a specific category
func GetUploadPath(category string) string {
	basePath := "uploads"

	switch category {
	case "profile":
		return filepath.Join(basePath, "profiles")
	case "post":
		return filepath.Join(basePath, "posts")
	case "story":
		return filepath.Join(basePath, "stories")
	case "group":
		return filepath.Join(basePath, "groups")
	case "event":
		return filepath.Join(basePath, "events")
	case "message":
		return filepath.Join(basePath, "messages")
	default:
		return filepath.Join(basePath, "misc")
	}
}

// ValidateFileName checks if filename is valid
func ValidateFileName(filename string) error {
	if len(filename) == 0 {
		return fmt.Errorf("filename cannot be empty")
	}

	if len(filename) > 255 {
		return fmt.Errorf("filename too long")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("filename contains invalid character: %s", char)
		}
	}

	// Check for reserved names (Windows)
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
	for _, reserved := range reservedNames {
		if strings.EqualFold(nameWithoutExt, reserved) {
			return fmt.Errorf("filename cannot be a reserved name: %s", reserved)
		}
	}

	return nil
}

// GetFileHash generates a hash for a file (for duplicate detection)
func GetFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := make([]byte, 32)
	if _, err := rand.Read(hasher); err != nil {
		return "", err
	}

	// Simple hash implementation - in production use crypto/sha256
	return hex.EncodeToString(hasher), nil
}
func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// InferMediaTypeFromExtension infers media type category from file extension
func InferMediaTypeFromExtension(ext string) string {
	ext = strings.ToLower(ext)

	// Remove leading dot if present
	if strings.HasPrefix(ext, ".") {
		ext = strings.TrimPrefix(ext, ".")
	}

	// Image extensions
	imageExts := []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "tiff", "svg"}
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return "image"
		}
	}

	// Video extensions
	videoExts := []string{"mp4", "mov", "avi", "mkv", "webm", "flv", "wmv", "m4v"}
	for _, vidExt := range videoExts {
		if ext == vidExt {
			return "video"
		}
	}

	// Audio extensions
	audioExts := []string{"mp3", "wav", "ogg", "aac", "flac", "m4a", "wma"}
	for _, audExt := range audioExts {
		if ext == audExt {
			return "audio"
		}
	}

	// Document extensions
	docExts := []string{"pdf", "doc", "docx", "txt", "rtf", "odt", "xls", "xlsx", "ppt", "pptx"}
	for _, docExt := range docExts {
		if ext == docExt {
			return "document"
		}
	}

	return "unknown"
}

// IsValidMediaType checks if the media type is supported
func IsValidMediaType(mediaType string) bool {
	validTypes := []string{"image", "video", "audio", "document"}

	for _, validType := range validTypes {
		if mediaType == validType {
			return true
		}
	}

	return false
}

// GetAllowedExtensionsForType returns allowed file extensions for a media type
func GetAllowedExtensionsForType(mediaType string) []string {
	switch mediaType {
	case "image":
		return []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "tiff"}
	case "video":
		return []string{"mp4", "mov", "avi", "mkv", "webm", "flv", "wmv", "m4v"}
	case "audio":
		return []string{"mp3", "wav", "ogg", "aac", "flac", "m4a", "wma"}
	case "document":
		return []string{"pdf", "doc", "docx", "txt", "rtf", "odt", "xls", "xlsx", "ppt", "pptx"}
	default:
		return []string{}
	}
}

// IsAllowedExtensionForType checks if file extension is allowed for media type
func IsAllowedExtensionForType(filename, mediaType string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	allowedExts := GetAllowedExtensionsForType(mediaType)

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return true
		}
	}

	return false
}

// ValidateMediaFile performs comprehensive media file validation
func ValidateMediaFile(header *multipart.FileHeader, mediaType string) error {
	if header == nil {
		return fmt.Errorf("no file provided")
	}

	// Validate filename
	if err := ValidateFileName(header.Filename); err != nil {
		return fmt.Errorf("invalid filename: %v", err)
	}

	// Validate media type
	if !IsValidMediaType(mediaType) {
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	// Validate file extension for media type
	if !IsAllowedExtensionForType(header.Filename, mediaType) {
		ext := filepath.Ext(header.Filename)
		return fmt.Errorf("file extension %s not allowed for media type %s", ext, mediaType)
	}

	// Validate file size
	if err := ValidateFileSize(header.Size, mediaType); err != nil {
		return err
	}

	return nil
}
