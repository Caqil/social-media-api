// utils/image.go
package utils

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// ImageInfo represents image information
type ImageInfo struct {
	Width       int
	Height      int
	Format      string
	Size        int64
	AspectRatio float64
}

// ThumbnailSize represents thumbnail dimensions
type ThumbnailSize struct {
	Width  int
	Height int
	Name   string
}

// Predefined thumbnail sizes
var (
	ThumbnailSizes = []ThumbnailSize{
		{Width: 150, Height: 150, Name: "small"},
		{Width: 300, Height: 300, Name: "medium"},
		{Width: 600, Height: 600, Name: "large"},
	}

	ProfilePicSizes = []ThumbnailSize{
		{Width: 50, Height: 50, Name: "xs"},
		{Width: 100, Height: 100, Name: "sm"},
		{Width: 200, Height: 200, Name: "md"},
		{Width: 400, Height: 400, Name: "lg"},
	}

	CoverPicSizes = []ThumbnailSize{
		{Width: 400, Height: 150, Name: "sm"},
		{Width: 800, Height: 300, Name: "md"},
		{Width: 1200, Height: 450, Name: "lg"},
	}
)

// GetImageInfo extracts information from an image file
func GetImageInfo(imagePath string) (*ImageInfo, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Decode image to get dimensions
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	aspectRatio := float64(width) / float64(height)

	return &ImageInfo{
		Width:       width,
		Height:      height,
		Format:      format,
		Size:        fileInfo.Size(),
		AspectRatio: aspectRatio,
	}, nil
}

// ValidateImageDimensions checks if image dimensions are within limits
func ValidateImageDimensions(imagePath string) error {
	info, err := GetImageInfo(imagePath)
	if err != nil {
		return err
	}

	if info.Width > MaxImageWidth || info.Height > MaxImageHeight {
		return fmt.Errorf("image dimensions exceed maximum allowed size of %dx%d", MaxImageWidth, MaxImageHeight)
	}

	return nil
}

// IsValidImageFormat checks if the image format is supported
func IsValidImageFormat(imagePath string) bool {
	file, err := os.Open(imagePath)
	if err != nil {
		return false
	}
	defer file.Close()

	_, format, err := image.DecodeConfig(file)
	if err != nil {
		return false
	}

	supportedFormats := []string{"jpeg", "png", "gif", "webp"}
	for _, supported := range supportedFormats {
		if format == supported {
			return true
		}
	}

	return false
}

// GenerateThumbnails creates thumbnails for an image
func GenerateThumbnails(originalPath string, sizes []ThumbnailSize) ([]string, error) {
	if !IsValidImageFormat(originalPath) {
		return nil, fmt.Errorf("unsupported image format")
	}

	// Open original image
	file, err := os.Open(originalPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode original image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	var thumbnailPaths []string

	// Create thumbnails for each size
	for _, size := range sizes {
		thumbnailPath, err := createThumbnail(img, originalPath, size, format)
		if err != nil {
			continue // Skip this size if creation fails
		}
		thumbnailPaths = append(thumbnailPaths, thumbnailPath)
	}

	return thumbnailPaths, nil
}

// createThumbnail creates a single thumbnail
func createThumbnail(img image.Image, originalPath string, size ThumbnailSize, format string) (string, error) {
	// Calculate thumbnail dimensions maintaining aspect ratio
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	var newWidth, newHeight int

	// Calculate dimensions to fit within the specified size
	if originalWidth > originalHeight {
		newWidth = size.Width
		newHeight = int(float64(originalHeight) * float64(size.Width) / float64(originalWidth))
	} else {
		newHeight = size.Height
		newWidth = int(float64(originalWidth) * float64(size.Height) / float64(originalHeight))
	}

	// Ensure minimum dimensions
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	// Create resized image (simplified - in production use image processing library)
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest neighbor scaling (for demo purposes)
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * originalWidth / newWidth
			srcY := y * originalHeight / newHeight
			if srcX >= originalWidth {
				srcX = originalWidth - 1
			}
			if srcY >= originalHeight {
				srcY = originalHeight - 1
			}
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}

	// Generate thumbnail filename
	dir := filepath.Dir(originalPath)
	filename := filepath.Base(originalPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	thumbnailPath := filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, size.Name, ext))

	// Create thumbnail file
	thumbnailFile, err := os.Create(thumbnailPath)
	if err != nil {
		return "", err
	}
	defer thumbnailFile.Close()

	// Encode thumbnail based on format
	switch format {
	case "jpeg":
		err = jpeg.Encode(thumbnailFile, resized, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(thumbnailFile, resized)
	case "gif":
		err = gif.Encode(thumbnailFile, resized, &gif.Options{})
	default:
		err = jpeg.Encode(thumbnailFile, resized, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		return "", err
	}

	return thumbnailPath, nil
}

// GenerateProfilePicThumbnails creates profile picture thumbnails
func GenerateProfilePicThumbnails(originalPath string) ([]string, error) {
	return GenerateThumbnails(originalPath, ProfilePicSizes)
}

// GenerateCoverPicThumbnails creates cover picture thumbnails
func GenerateCoverPicThumbnails(originalPath string) ([]string, error) {
	return GenerateThumbnails(originalPath, CoverPicSizes)
}

// CompressImage compresses an image to reduce file size
func CompressImage(inputPath, outputPath string, quality int) error {
	if quality < 1 || quality > 100 {
		quality = 85 // Default quality
	}

	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	// Decode image
	img, format, err := image.Decode(inputFile)
	if err != nil {
		return err
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Encode with compression
	switch format {
	case "jpeg":
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(outputFile, img)
	case "gif":
		err = gif.Encode(outputFile, img, &gif.Options{})
	default:
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	}

	return err
}

// IsSquareImage checks if an image is square (1:1 aspect ratio)
func IsSquareImage(imagePath string) (bool, error) {
	info, err := GetImageInfo(imagePath)
	if err != nil {
		return false, err
	}

	// Consider square if aspect ratio is close to 1 (within 5% tolerance)
	tolerance := 0.05
	return info.AspectRatio >= (1.0-tolerance) && info.AspectRatio <= (1.0+tolerance), nil
}

// IsLandscapeImage checks if an image is landscape orientation
func IsLandscapeImage(imagePath string) (bool, error) {
	info, err := GetImageInfo(imagePath)
	if err != nil {
		return false, err
	}

	return info.AspectRatio > 1.0, nil
}

// IsPortraitImage checks if an image is portrait orientation
func IsPortraitImage(imagePath string) (bool, error) {
	info, err := GetImageInfo(imagePath)
	if err != nil {
		return false, err
	}

	return info.AspectRatio < 1.0, nil
}

// CalculateOptimalSize calculates optimal dimensions for image display
func CalculateOptimalSize(originalWidth, originalHeight, maxWidth, maxHeight int) (int, int) {
	if originalWidth <= maxWidth && originalHeight <= maxHeight {
		return originalWidth, originalHeight
	}

	widthRatio := float64(maxWidth) / float64(originalWidth)
	heightRatio := float64(maxHeight) / float64(originalHeight)

	// Use the smaller ratio to ensure image fits within bounds
	ratio := widthRatio
	if heightRatio < widthRatio {
		ratio = heightRatio
	}

	newWidth := int(float64(originalWidth) * ratio)
	newHeight := int(float64(originalHeight) * ratio)

	// Ensure minimum dimensions
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	return newWidth, newHeight
}

// GetImageOrientation returns the orientation of an image
func GetImageOrientation(imagePath string) (string, error) {
	isSquare, err := IsSquareImage(imagePath)
	if err != nil {
		return "", err
	}

	if isSquare {
		return "square", nil
	}

	isLandscape, err := IsLandscapeImage(imagePath)
	if err != nil {
		return "", err
	}

	if isLandscape {
		return "landscape", nil
	}

	return "portrait", nil
}

// GenerateImageURL generates a URL for an image with optional size parameter
func GenerateImageURL(basePath, imagePath, size string) string {
	if size == "" {
		return basePath + imagePath
	}

	// Extract filename and extension
	dir := filepath.Dir(imagePath)
	filename := filepath.Base(imagePath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Generate thumbnail filename
	thumbnailName := fmt.Sprintf("%s_%s%s", nameWithoutExt, size, ext)
	thumbnailPath := filepath.Join(dir, thumbnailName)

	return basePath + thumbnailPath
}

// CleanupImageThumbnails removes all thumbnails for an image
func CleanupImageThumbnails(imagePath string) error {
	dir := filepath.Dir(imagePath)
	filename := filepath.Base(imagePath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Remove thumbnails for all sizes
	allSizes := append(append(ThumbnailSizes, ProfilePicSizes...), CoverPicSizes...)

	for _, size := range allSizes {
		thumbnailPath := filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, size.Name, ext))
		if FileExists(thumbnailPath) {
			os.Remove(thumbnailPath)
		}
	}

	return nil
}

// ValidateImageFile performs comprehensive image validation
func ValidateImageFile(imagePath string) error {
	// Check if file exists
	if !FileExists(imagePath) {
		return fmt.Errorf("image file does not exist")
	}

	// Check if it's a valid image format
	if !IsValidImageFormat(imagePath) {
		return fmt.Errorf("invalid or unsupported image format")
	}

	// Check dimensions
	if err := ValidateImageDimensions(imagePath); err != nil {
		return err
	}

	// Check file size
	info, err := GetImageInfo(imagePath)
	if err != nil {
		return err
	}

	if err := ValidateFileSize(info.Size, "image"); err != nil {
		return err
	}

	return nil
}

// GetImageThumbnailPath returns the path to a thumbnail for a given size
func GetImageThumbnailPath(originalPath, sizeName string) string {
	dir := filepath.Dir(originalPath)
	filename := filepath.Base(originalPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, sizeName, ext))
}

// HasImageThumbnail checks if a thumbnail exists for a given size
func HasImageThumbnail(originalPath, sizeName string) bool {
	thumbnailPath := GetImageThumbnailPath(originalPath, sizeName)
	return FileExists(thumbnailPath)
}

// GetAvailableThumbnailSizes returns available thumbnail sizes for an image
func GetAvailableThumbnailSizes(imagePath string) []string {
	var availableSizes []string

	allSizes := append(append(ThumbnailSizes, ProfilePicSizes...), CoverPicSizes...)

	for _, size := range allSizes {
		if HasImageThumbnail(imagePath, size.Name) {
			availableSizes = append(availableSizes, size.Name)
		}
	}

	return availableSizes
}
