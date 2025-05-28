// internal/services/media_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"
	"social-media-api/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MediaService struct {
	collection     *mongo.Collection
	userCollection *mongo.Collection
	db             *mongo.Database
	uploadPath     string
	baseURL        string
	maxFileSize    int64
	allowedTypes   map[string][]string
}

type UploadResult struct {
	Media    *models.Media `json:"media"`
	URL      string        `json:"url"`
	Filename string        `json:"filename"`
}

func NewMediaService(uploadPath, baseURL string) *MediaService {
	return &MediaService{
		collection:     config.DB.Collection("media"),
		userCollection: config.DB.Collection("users"),
		db:             config.DB,
		uploadPath:     uploadPath,
		baseURL:        baseURL,
		maxFileSize:    50 * 1024 * 1024, // 50MB default
		allowedTypes: map[string][]string{
			"image":    {"jpg", "jpeg", "png", "gif", "webp", "bmp"},
			"video":    {"mp4", "mov", "avi", "mkv", "webm"},
			"audio":    {"mp3", "wav", "ogg", "aac", "flac"},
			"document": {"pdf", "doc", "docx", "txt", "rtf"},
		},
	}
}

// UploadMedia handles file upload and creates media record
func (ms *MediaService) UploadMedia(userID primitive.ObjectID, file multipart.File, header *multipart.FileHeader, req models.CreateMediaRequest) (*UploadResult, error) {
	// Validate file
	if err := ms.validateFile(header, req.Type); err != nil {
		return nil, err
	}

	// Generate unique filename
	ext := strings.ToLower(filepath.Ext(header.Filename))
	filename := fmt.Sprintf("%s_%d%s", primitive.NewObjectID().Hex(), time.Now().Unix(), ext)

	// Create directory structure based on date and type
	dateFolder := time.Now().Format("2006/01/02")
	fullPath := filepath.Join(ms.uploadPath, req.Type, dateFolder)

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Save file
	filePath := filepath.Join(fullPath, filename)
	destFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	defer destFile.Close()

	size, err := io.Copy(destFile, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %v", err)
	}

	// Get file info
	mimeType := utils.GetMimeType(ext)
	width, height := 0, 0
	duration := 0

	// Extract metadata for images and videos
	if req.Type == "image" {
		width, height = ms.getImageDimensions(filePath)
	} else if req.Type == "video" {
		width, height, duration = ms.getVideoMetadata(filePath)
	} else if req.Type == "audio" {
		duration = ms.getAudioDuration(filePath)
	}

	// Convert related ID if provided
	var relatedID *primitive.ObjectID
	if req.RelatedID != "" {
		if rID, err := primitive.ObjectIDFromHex(req.RelatedID); err == nil {
			relatedID = &rID
		}
	}

	// Create media record
	media := &models.Media{
		OriginalName:    header.Filename,
		FileName:        filename,
		FilePath:        filePath,
		FileSize:        size,
		MimeType:        mimeType,
		FileExtension:   strings.TrimPrefix(ext, "."),
		Type:            req.Type,
		Category:        req.Category,
		UploadedBy:      userID,
		Width:           width,
		Height:          height,
		Duration:        duration,
		URL:             fmt.Sprintf("%s/media/%s/%s/%s", ms.baseURL, req.Type, dateFolder, filename),
		IsPublic:        req.IsPublic,
		AltText:         req.AltText,
		Description:     req.Description,
		RelatedTo:       req.RelatedTo,
		RelatedID:       relatedID,
		ExpiresAt:       req.ExpiresAt,
		StorageProvider: "local",
		StorageKey:      fmt.Sprintf("%s/%s/%s", req.Type, dateFolder, filename),
	}

	media.BeforeCreate()

	// Insert media record
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := ms.collection.InsertOne(ctx, media)
	if err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, err
	}

	media.ID = result.InsertedID.(primitive.ObjectID)

	// Generate thumbnails for images and videos
	go ms.generateThumbnails(media)

	// Process media asynchronously (resize, optimize, etc.)
	go ms.processMedia(media)

	return &UploadResult{
		Media:    media,
		URL:      media.URL,
		Filename: filename,
	}, nil
}

// GetMediaByID retrieves media by ID
func (ms *MediaService) GetMediaByID(mediaID primitive.ObjectID, currentUserID *primitive.ObjectID) (*models.Media, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var media models.Media
	err := ms.collection.FindOne(ctx, bson.M{
		"_id":        mediaID,
		"deleted_at": bson.M{"$exists": false},
	}).Decode(&media)

	if err != nil {
		return nil, err
	}

	// Check access permissions
	if !ms.canAccessMedia(&media, currentUserID) {
		return nil, errors.New("access denied")
	}

	// Increment view count
	if currentUserID != nil && *currentUserID != media.UploadedBy {
		go ms.incrementViewCount(mediaID)
	}

	return &media, nil
}

// GetUserMedia retrieves media uploaded by a user
func (ms *MediaService) GetUserMedia(userID primitive.ObjectID, currentUserID *primitive.ObjectID, mediaType string, limit, skip int) ([]models.Media, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"uploaded_by": userID,
		"deleted_at":  bson.M{"$exists": false},
	}

	if mediaType != "" {
		filter["type"] = mediaType
	}

	// Apply privacy filter if not viewing own media
	if currentUserID == nil || *currentUserID != userID {
		filter["is_public"] = true
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ms.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var media []models.Media
	if err := cursor.All(ctx, &media); err != nil {
		return nil, err
	}

	return media, nil
}

// UpdateMedia updates media information
func (ms *MediaService) UpdateMedia(mediaID, userID primitive.ObjectID, req models.UpdateMediaRequest) (*models.Media, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if media exists and user owns it
	var media models.Media
	err := ms.collection.FindOne(ctx, bson.M{
		"_id":         mediaID,
		"uploaded_by": userID,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&media)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("media not found or access denied")
		}
		return nil, err
	}

	// Build update document
	update := bson.M{"$set": bson.M{"updated_at": time.Now()}}

	if req.AltText != nil {
		update["$set"].(bson.M)["alt_text"] = *req.AltText
	}
	if req.Description != nil {
		update["$set"].(bson.M)["description"] = *req.Description
	}
	if req.IsPublic != nil {
		update["$set"].(bson.M)["is_public"] = *req.IsPublic
		// Update access policy
		if *req.IsPublic {
			update["$set"].(bson.M)["access_policy"] = "public"
		} else {
			update["$set"].(bson.M)["access_policy"] = "private"
		}
	}

	_, err = ms.collection.UpdateOne(ctx, bson.M{"_id": mediaID}, update)
	if err != nil {
		return nil, err
	}

	return ms.GetMediaByID(mediaID, &userID)
}

// DeleteMedia soft deletes media
func (ms *MediaService) DeleteMedia(mediaID, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if media exists and user owns it
	var media models.Media
	err := ms.collection.FindOne(ctx, bson.M{
		"_id":         mediaID,
		"uploaded_by": userID,
		"deleted_at":  bson.M{"$exists": false},
	}).Decode(&media)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("media not found or access denied")
		}
		return err
	}

	// Soft delete
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err = ms.collection.UpdateOne(ctx, bson.M{"_id": mediaID}, update)
	if err != nil {
		return err
	}

	// Schedule physical file deletion
	go ms.scheduleFileDeletion(media.FilePath, 24*time.Hour)

	return nil
}

// SearchMedia searches for media
func (ms *MediaService) SearchMedia(query string, mediaType string, currentUserID *primitive.ObjectID, limit, skip int) ([]models.Media, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	filter := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"original_name": bson.M{"$regex": query, "$options": "i"}},
					{"description": bson.M{"$regex": query, "$options": "i"}},
					{"alt_text": bson.M{"$regex": query, "$options": "i"}},
				},
			},
			{"deleted_at": bson.M{"$exists": false}},
			{"is_public": true}, // Only search public media
		},
	}

	if mediaType != "" {
		filter["$and"] = append(filter["$and"].([]bson.M), bson.M{"type": mediaType})
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(skip)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := ms.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var media []models.Media
	if err := cursor.All(ctx, &media); err != nil {
		return nil, err
	}

	return media, nil
}

// GetMediaStats retrieves media statistics
func (ms *MediaService) GetMediaStats(userID *primitive.ObjectID) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	matchStage := bson.M{
		"deleted_at": bson.M{"$exists": false},
	}

	if userID != nil {
		matchStage["uploaded_by"] = *userID
	}

	pipeline := []bson.M{
		{
			"$match": matchStage,
		},
		{
			"$group": bson.M{
				"_id":             "$type",
				"count":           bson.M{"$sum": 1},
				"total_size":      bson.M{"$sum": "$file_size"},
				"avg_size":        bson.M{"$avg": "$file_size"},
				"total_views":     bson.M{"$sum": "$view_count"},
				"total_downloads": bson.M{"$sum": "$download_count"},
			},
		},
	}

	cursor, err := ms.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID             string  `bson:"_id"`
		Count          int64   `bson:"count"`
		TotalSize      int64   `bson:"total_size"`
		AvgSize        float64 `bson:"avg_size"`
		TotalViews     int64   `bson:"total_views"`
		TotalDownloads int64   `bson:"total_downloads"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	totalCount := int64(0)
	totalSize := int64(0)
	totalViews := int64(0)
	totalDownloads := int64(0)

	for _, result := range results {
		stats[result.ID] = map[string]interface{}{
			"count":           result.Count,
			"total_size":      result.TotalSize,
			"avg_size":        result.AvgSize,
			"total_views":     result.TotalViews,
			"total_downloads": result.TotalDownloads,
		}
		totalCount += result.Count
		totalSize += result.TotalSize
		totalViews += result.TotalViews
		totalDownloads += result.TotalDownloads
	}

	stats["total"] = map[string]interface{}{
		"count":           totalCount,
		"total_size":      totalSize,
		"total_views":     totalViews,
		"total_downloads": totalDownloads,
	}

	return stats, nil
}

// IncrementDownloadCount increments download count for media
func (ms *MediaService) IncrementDownloadCount(mediaID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$inc": bson.M{"download_count": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}

	_, err := ms.collection.UpdateOne(ctx, bson.M{"_id": mediaID}, update)
	return err
}

// CleanupExpiredMedia removes expired media
func (ms *MediaService) CleanupExpiredMedia() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
		"is_expired": false,
	}

	// Mark as expired first
	update := bson.M{
		"$set": bson.M{
			"is_expired": true,
			"updated_at": time.Now(),
		},
	}

	result, err := ms.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	// Get expired media for file cleanup
	cursor, err := ms.collection.Find(ctx, bson.M{"is_expired": true})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var expiredMedia []models.Media
	if err := cursor.All(ctx, &expiredMedia); err != nil {
		return err
	}

	// Schedule file deletion
	for _, media := range expiredMedia {
		go ms.scheduleFileDeletion(media.FilePath, 0)
	}

	fmt.Printf("Marked %d media items as expired\n", result.ModifiedCount)
	return nil
}

// Private helper methods

func (ms *MediaService) validateFile(header *multipart.FileHeader, mediaType string) error {
	// Check file size
	if header.Size > ms.maxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", ms.maxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	allowedExts, exists := ms.allowedTypes[mediaType]
	if !exists {
		return fmt.Errorf("unsupported media type: %s", mediaType)
	}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported file extension: %s", ext)
}

func (ms *MediaService) canAccessMedia(media *models.Media, userID *primitive.ObjectID) bool {
	// Owner can always access
	if userID != nil && media.UploadedBy == *userID {
		return true
	}

	// Check if media is public
	if media.IsPublic && media.AccessPolicy == "public" {
		return true
	}

	// Check if user is in allowed users list (for restricted access)
	if userID != nil && media.AccessPolicy == "restricted" {
		for _, allowedUserID := range media.AllowedUsers {
			if allowedUserID == *userID {
				return true
			}
		}
	}

	return false
}

func (ms *MediaService) incrementViewCount(mediaID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$inc": bson.M{"view_count": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}

	ms.collection.UpdateOne(ctx, bson.M{"_id": mediaID}, update)
}

func (ms *MediaService) generateThumbnails(media *models.Media) {
	// Generate thumbnails for images and videos
	// This would integrate with image processing libraries like imaging or ffmpeg for videos

	if media.Type == "image" {
		// Generate different sized thumbnails
		thumbnails := []models.MediaVariant{
			{Name: "thumbnail", Width: 150, Height: 150},
			{Name: "small", Width: 300, Height: 300},
			{Name: "medium", Width: 600, Height: 600},
		}

		for _, thumb := range thumbnails {
			// Generate thumbnail and save
			thumbPath := ms.generateImageThumbnail(media.FilePath, thumb.Width, thumb.Height)
			if thumbPath != "" {
				thumb.URL = fmt.Sprintf("%s/media/thumbnails/%s", ms.baseURL, filepath.Base(thumbPath))
				thumb.FileSize = ms.getFileSize(thumbPath)
				thumb.Format = media.FileExtension
				media.Thumbnails = append(media.Thumbnails, thumb)
			}
		}

		// Update media record with thumbnails
		ms.updateMediaThumbnails(media.ID, media.Thumbnails)
	}
}

func (ms *MediaService) processMedia(media *models.Media) {
	// Process media (optimize, resize, etc.)
	// This would integrate with image/video processing tools

	// Mark as processed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"is_processed":      true,
			"processing_status": "completed",
			"processed_at":      time.Now(),
			"updated_at":        time.Now(),
		},
	}

	ms.collection.UpdateOne(ctx, bson.M{"_id": media.ID}, update)
}

func (ms *MediaService) getImageDimensions(filePath string) (int, int) {
	// Use image processing library to get dimensions
	// Placeholder implementation
	return 800, 600
}

func (ms *MediaService) getVideoMetadata(filePath string) (int, int, int) {
	// Use ffmpeg or similar to get video metadata
	// Placeholder implementation
	return 1920, 1080, 120 // width, height, duration in seconds
}

func (ms *MediaService) getAudioDuration(filePath string) int {
	// Use audio processing library to get duration
	// Placeholder implementation
	return 180 // duration in seconds
}

func (ms *MediaService) generateImageThumbnail(originalPath string, width, height int) string {
	// Generate thumbnail using image processing library
	// Placeholder implementation
	return ""
}

func (ms *MediaService) getFileSize(filePath string) int64 {
	if info, err := os.Stat(filePath); err == nil {
		return info.Size()
	}
	return 0
}

func (ms *MediaService) updateMediaThumbnails(mediaID primitive.ObjectID, thumbnails []models.MediaVariant) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"thumbnails": thumbnails,
			"updated_at": time.Now(),
		},
	}

	ms.collection.UpdateOne(ctx, bson.M{"_id": mediaID}, update)
}

func (ms *MediaService) scheduleFileDeletion(filePath string, delay time.Duration) {
	if delay > 0 {
		time.Sleep(delay)
	}

	if err := os.Remove(filePath); err != nil {
		fmt.Printf("Failed to delete file %s: %v\n", filePath, err)
	}
}

// GetMediaURL returns the public URL for accessing media
func (ms *MediaService) GetMediaURL(media *models.Media, variant string) string {
	if variant == "" || variant == "original" {
		return media.URL
	}

	// Find variant URL
	for _, thumb := range media.Thumbnails {
		if thumb.Name == variant {
			return thumb.URL
		}
	}

	for _, variantMedia := range media.Variants {
		if variantMedia.Name == variant {
			return variantMedia.URL
		}
	}

	return media.URL // Return original if variant not found
}
