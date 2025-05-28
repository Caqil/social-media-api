// s3.go
package storage

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Provider implements StorageProvider for Amazon S3
type S3Provider struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
	region     string
	cdnDomain  string
	baseURL    string
}

// NewS3Provider creates a new S3 storage provider
func NewS3Provider(config StorageConfig) (*S3Provider, error) {
	// Create AWS session
	sess, err := createS3Session(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 session: %w", err)
	}

	// Create S3 client
	client := s3.New(sess)

	// Test connection by checking if bucket exists
	_, err = client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(config.Bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to access S3 bucket %s: %w", config.Bucket, err)
	}

	// Determine base URL
	baseURL := config.CDNDomain
	if baseURL == "" {
		if config.Endpoint != "" {
			baseURL = config.Endpoint
		} else {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", config.Bucket, config.Region)
		}
	}

	return &S3Provider{
		client:     client,
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
		bucket:     config.Bucket,
		region:     config.Region,
		cdnDomain:  config.CDNDomain,
		baseURL:    baseURL,
	}, nil
}

// createS3Session creates an AWS session with the provided configuration
func createS3Session(config StorageConfig) (*session.Session, error) {
	awsConfig := &aws.Config{
		Region: aws.String(config.Region),
	}

	// Set credentials if provided
	if config.AccessKey != "" && config.SecretKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(
			config.AccessKey,
			config.SecretKey,
			"", // token
		)
	}

	// Set custom endpoint if provided (for S3-compatible services)
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	return session.NewSession(awsConfig)
}

// Upload uploads a file to S3
func (s *S3Provider) Upload(key string, data io.Reader, contentType string, size int64) (*UploadResult, error) {
	// Read data into buffer to allow multiple reads
	var buf bytes.Buffer
	actualSize, err := io.Copy(&buf, data)
	if err != nil {
		return nil, NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to read data: %v", err), key)
	}

	// Validate size if provided
	if size > 0 && actualSize != size {
		return nil, NewStorageErrorWithKey(ErrCodeInvalidInput,
			fmt.Sprintf("size mismatch: expected %d, got %d", size, actualSize), key)
	}

	// Prepare upload input
	uploadInput := &s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	}

	// Set ACL based on file type (make media files public by default)
	if isPublicContentType(contentType) {
		uploadInput.ACL = aws.String("public-read")
	}

	// Set cache control for media files
	if isMediaContentType(contentType) {
		uploadInput.CacheControl = aws.String("public, max-age=31536000") // 1 year
	}

	// Upload file
	result, err := s.uploader.Upload(uploadInput)
	if err != nil {
		return nil, s.handleS3Error(err, key)
	}

	// Generate URLs
	publicURL := s.generateURL(key)
	cdnURL := publicURL
	if s.cdnDomain != "" {
		cdnURL = fmt.Sprintf("https://%s/%s", strings.TrimPrefix(s.cdnDomain, "https://"), key)
	}

	return &UploadResult{
		Key:         key,
		URL:         publicURL,
		CDNUrl:      cdnURL,
		Size:        actualSize,
		ContentType: contentType,
		ETag:        strings.Trim(*result.ETag, "\""),
	}, nil
}

// Download downloads a file from S3
func (s *S3Provider) Download(key string) (io.ReadCloser, error) {
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(getInput)
	if err != nil {
		return nil, s.handleS3Error(err, key)
	}

	return result.Body, nil
}

// Delete removes a file from S3
func (s *S3Provider) Delete(key string) error {
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(deleteInput)
	if err != nil {
		return s.handleS3Error(err, key)
	}

	return nil
}

// Exists checks if a file exists in S3
func (s *S3Provider) Exists(key string) (bool, error) {
	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(headInput)
	if err != nil {
		if s.isNotFoundError(err) {
			return false, nil
		}
		return false, s.handleS3Error(err, key)
	}

	return true, nil
}

// GetURL generates a public URL for the file
func (s *S3Provider) GetURL(key string) (string, error) {
	return s.generateURL(key), nil
}

// GetSignedURL generates a temporary signed URL
func (s *S3Provider) GetSignedURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	signedURL, err := req.Presign(expiration)
	if err != nil {
		return "", NewStorageErrorWithKey(ErrCodeInternal,
			fmt.Sprintf("failed to generate signed URL: %v", err), key)
	}

	return signedURL, nil
}

// GetMetadata retrieves file metadata from S3
func (s *S3Provider) GetMetadata(key string) (*FileMetadata, error) {
	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(headInput)
	if err != nil {
		return nil, s.handleS3Error(err, key)
	}

	metadata := make(map[string]string)
	for k, v := range result.Metadata {
		if v != nil {
			metadata[k] = *v
		}
	}

	return &FileMetadata{
		Key:          key,
		Size:         *result.ContentLength,
		ContentType:  aws.StringValue(result.ContentType),
		LastModified: *result.LastModified,
		ETag:         strings.Trim(*result.ETag, "\""),
		Metadata:     metadata,
	}, nil
}

// Copy copies a file within S3
func (s *S3Provider) Copy(sourceKey, destKey string) error {
	copySource := fmt.Sprintf("%s/%s", s.bucket, sourceKey)

	copyInput := &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(destKey),
		CopySource: aws.String(copySource),
	}

	_, err := s.client.CopyObject(copyInput)
	if err != nil {
		return s.handleS3Error(err, destKey)
	}

	return nil
}

// Move moves a file within S3 (copy then delete)
func (s *S3Provider) Move(sourceKey, destKey string) error {
	// First copy the file
	if err := s.Copy(sourceKey, destKey); err != nil {
		return err
	}

	// Then delete the source
	return s.Delete(sourceKey)
}

// ListFiles lists files with optional prefix
func (s *S3Provider) ListFiles(prefix string, limit int) ([]FileInfo, error) {
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	if limit > 0 {
		listInput.MaxKeys = aws.Int64(int64(limit))
	}

	result, err := s.client.ListObjectsV2(listInput)
	if err != nil {
		return nil, NewStorageError(ErrCodeInternal,
			fmt.Sprintf("failed to list files: %v", err))
	}

	files := make([]FileInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		files = append(files, FileInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
			IsDir:        false,
		})
	}

	return files, nil
}

// GetStorageInfo returns information about this S3 provider
func (s *S3Provider) GetStorageInfo() StorageInfo {
	return StorageInfo{
		Provider: "s3",
		Region:   s.region,
		Bucket:   s.bucket,
		Endpoint: s.baseURL,
	}
}

// Helper methods

// generateURL generates a public URL for a file
func (s *S3Provider) generateURL(key string) string {
	if s.cdnDomain != "" {
		return fmt.Sprintf("https://%s/%s", strings.TrimPrefix(s.cdnDomain, "https://"), key)
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.baseURL, "/"), key)
}

// handleS3Error converts S3 errors to StorageError
func (s *S3Provider) handleS3Error(err error, key string) error {
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case s3.ErrCodeNoSuchKey, "NotFound":
			return NewStorageErrorWithKey(ErrCodeNotFound, "file not found", key)
		case "AccessDenied", "Forbidden":
			return NewStorageErrorWithKey(ErrCodeAccessDenied, "access denied", key)
		case "InvalidRequest", "InvalidArgument":
			return NewStorageErrorWithKey(ErrCodeInvalidInput, awsErr.Message(), key)
		default:
			return NewStorageErrorWithKey(ErrCodeInternal, awsErr.Message(), key)
		}
	}
	return NewStorageErrorWithKey(ErrCodeInternal, err.Error(), key)
}

// isNotFoundError checks if the error is a not found error
func (s *S3Provider) isNotFoundError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == s3.ErrCodeNoSuchKey || awsErr.Code() == "NotFound"
	}
	return false
}

// Utility functions for content type handling

// isPublicContentType determines if a content type should be publicly readable
func isPublicContentType(contentType string) bool {
	publicTypes := []string{
		"image/",
		"video/",
		"audio/",
		"text/css",
		"text/javascript",
		"application/javascript",
	}

	for _, publicType := range publicTypes {
		if strings.HasPrefix(contentType, publicType) {
			return true
		}
	}
	return false
}

// isMediaContentType determines if a content type is media
func isMediaContentType(contentType string) bool {
	mediaTypes := []string{
		"image/",
		"video/",
		"audio/",
	}

	for _, mediaType := range mediaTypes {
		if strings.HasPrefix(contentType, mediaType) {
			return true
		}
	}
	return false
}

// GenerateStorageKey generates a storage key based on file type and user ID
func GenerateStorageKey(userID, fileName, fileType string) string {
	// Clean the filename
	cleanName := filepath.Base(fileName)

	// Generate path based on file type
	switch fileType {
	case "profile_pic":
		return fmt.Sprintf("profiles/%s/avatar/%s", userID, cleanName)
	case "cover_pic":
		return fmt.Sprintf("profiles/%s/cover/%s", userID, cleanName)
	case "post_media":
		return fmt.Sprintf("posts/%s/%s", userID, cleanName)
	case "story_media":
		return fmt.Sprintf("stories/%s/%s", userID, cleanName)
	case "group_media":
		return fmt.Sprintf("groups/%s/%s", userID, cleanName)
	case "event_media":
		return fmt.Sprintf("events/%s/%s", userID, cleanName)
	case "message_media":
		return fmt.Sprintf("messages/%s/%s", userID, cleanName)
	default:
		return fmt.Sprintf("uploads/%s/%s", userID, cleanName)
	}
}

// CleanupExpiredFiles removes expired files (for stories, temp uploads, etc.)
func (s *S3Provider) CleanupExpiredFiles(prefix string, olderThan time.Time) error {
	// List files with prefix
	files, err := s.ListFiles(prefix, 0)
	if err != nil {
		return err
	}

	// Delete files older than specified time
	var deleteErrors []error
	for _, file := range files {
		if file.LastModified.Before(olderThan) {
			if err := s.Delete(file.Key); err != nil {
				deleteErrors = append(deleteErrors, err)
			}
		}
	}

	if len(deleteErrors) > 0 {
		return fmt.Errorf("failed to delete %d files", len(deleteErrors))
	}

	return nil
}
