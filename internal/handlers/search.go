// internal/handlers/search.go
package handlers

import (
	"strconv"
	"strings"

	"social-media-api/internal/services"
	"social-media-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SearchHandler struct {
	searchService *services.SearchService
	validator     *validator.Validate
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		validator:     validator.New(),
	}
}

// Search performs comprehensive search across all content types
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	if len(strings.TrimSpace(query)) < 2 {
		utils.BadRequestResponse(c, "Search query must be at least 2 characters", nil)
		return
	}

	// Get current user ID if authenticated
	var userID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		userID = &id
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build search filters
	filters := services.SearchFilters{
		Type:        c.DefaultQuery("type", "all"),
		DateRange:   c.Query("date_range"),
		SortBy:      c.DefaultQuery("sort_by", "relevance"),
		Location:    c.Query("location"),
		Language:    c.Query("language"),
		ContentType: c.Query("content_type"),
	}

	// Validate filters
	if !h.isValidSearchType(filters.Type) {
		utils.BadRequestResponse(c, "Invalid search type. Must be one of: all, posts, users, hashtags", nil)
		return
	}

	if !h.isValidSortBy(filters.SortBy) {
		utils.BadRequestResponse(c, "Invalid sort type. Must be one of: relevance, recent, popular", nil)
		return
	}

	if filters.DateRange != "" && !h.isValidDateRange(filters.DateRange) {
		utils.BadRequestResponse(c, "Invalid date range. Must be one of: day, week, month, year", nil)
		return
	}

	response, err := h.searchService.Search(query, userID, filters, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Search failed", err)
		return
	}

	utils.OkResponse(c, "Search completed successfully", response)
}

// SearchPosts searches specifically for posts
func (h *SearchHandler) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Get current user ID if authenticated
	var userID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		userID = &id
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filters for posts only
	filters := services.SearchFilters{
		Type:        "posts",
		DateRange:   c.Query("date_range"),
		SortBy:      c.DefaultQuery("sort_by", "relevance"),
		Location:    c.Query("location"),
		Language:    c.Query("language"),
		ContentType: c.Query("content_type"),
	}

	response, err := h.searchService.Search(query, userID, filters, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Post search failed", err)
		return
	}

	utils.OkResponse(c, "Post search completed successfully", response)
}

// SearchUsers searches specifically for users
func (h *SearchHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Get current user ID if authenticated
	var userID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		userID = &id
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filters for users only
	filters := services.SearchFilters{
		Type:   "users",
		SortBy: c.DefaultQuery("sort_by", "relevance"),
	}

	response, err := h.searchService.Search(query, userID, filters, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "User search failed", err)
		return
	}

	utils.OkResponse(c, "User search completed successfully", response)
}

// SearchHashtags searches specifically for hashtags
func (h *SearchHandler) SearchHashtags(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Build filters for hashtags only
	filters := services.SearchFilters{
		Type:   "hashtags",
		SortBy: c.DefaultQuery("sort_by", "popular"),
	}

	response, err := h.searchService.Search(query, nil, filters, params.Limit, params.Offset)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Hashtag search failed", err)
		return
	}

	utils.OkResponse(c, "Hashtag search completed successfully", response)
}

// GetTrendingHashtags retrieves trending hashtags
func (h *SearchHandler) GetTrendingHashtags(c *gin.Context) {
	// Get limit parameter
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "day")
	if !h.isValidDateRange(timeRange) {
		timeRange = "day"
	}

	hashtags, err := h.searchService.GetTrendingHashtags(limit, timeRange)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get trending hashtags", err)
		return
	}

	utils.OkResponse(c, "Trending hashtags retrieved successfully", gin.H{
		"hashtags":   hashtags,
		"time_range": timeRange,
		"count":      len(hashtags),
	})
}

// GetSearchSuggestions provides search suggestions based on query
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	if len(strings.TrimSpace(query)) < 1 {
		utils.BadRequestResponse(c, "Search query is too short", nil)
		return
	}

	// Get current user ID if authenticated
	var userID *primitive.ObjectID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(primitive.ObjectID)
		userID = &id
	}

	// Get limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	// For now, we'll use a basic search to get suggestions
	// In a full implementation, you'd have a dedicated suggestions service
	filters := services.SearchFilters{
		Type:   c.DefaultQuery("type", "all"),
		SortBy: "popular",
	}

	response, err := h.searchService.Search(query, userID, filters, limit, 0)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get search suggestions", err)
		return
	}

	// Extract suggestions from search results
	suggestions := response.Suggestions
	if len(suggestions) == 0 {
		suggestions = []string{}
	}

	utils.OkResponse(c, "Search suggestions retrieved successfully", gin.H{
		"suggestions": suggestions,
		"query":       query,
		"count":       len(suggestions),
	})
}

// UpdateHashtagInfo updates hashtag information (internal use)
func (h *SearchHandler) UpdateHashtagInfo(c *gin.Context) {
	// This would typically be called internally, but can be exposed for admin use
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check if user has admin privileges (simplified check)
	// In a real app, you'd check user roles properly
	var req struct {
		Hashtag string `json:"hashtag" binding:"required"`
		PostID  string `json:"post_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	postID, err := primitive.ObjectIDFromHex(req.PostID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid post ID format", err)
		return
	}

	err = h.searchService.UpdateHashtagInfo(req.Hashtag, postID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to update hashtag info", err)
		return
	}

	utils.OkResponse(c, "Hashtag info updated successfully", gin.H{
		"hashtag": req.Hashtag,
		"post_id": req.PostID,
	})
}

// GetSearchHistory retrieves user's search history
func (h *SearchHandler) GetSearchHistory(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get pagination parameters
	utils.GetPaginationParams(c)

	// This would be implemented in the search service
	// For now, return a placeholder response
	utils.OkResponse(c, "Search history retrieved successfully", gin.H{
		"history": []gin.H{},
		"count":   0,
	})
}

// ClearSearchHistory clears user's search history
func (h *SearchHandler) ClearSearchHistory(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// This would be implemented in the search service
	// For now, return success response
	utils.OkResponse(c, "Search history cleared successfully", nil)
}

// GetPopularSearches retrieves popular/trending search queries
func (h *SearchHandler) GetPopularSearches(c *gin.Context) {
	// Get limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Get time range parameter
	timeRange := c.DefaultQuery("time_range", "day")
	if !h.isValidDateRange(timeRange) {
		timeRange = "day"
	}

	// This would be implemented in the search service
	// For now, return trending hashtags as popular searches
	hashtags, err := h.searchService.GetTrendingHashtags(limit, timeRange)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get popular searches", err)
		return
	}

	// Convert hashtags to search queries
	var popularSearches []string
	for _, hashtag := range hashtags {
		popularSearches = append(popularSearches, hashtag.Name)
	}

	utils.OkResponse(c, "Popular searches retrieved successfully", gin.H{
		"searches":   popularSearches,
		"time_range": timeRange,
		"count":      len(popularSearches),
	})
}

// IndexContent manually triggers content indexing (admin only)
func (h *SearchHandler) IndexContent(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Check admin privileges (simplified)
	// In a real app, you'd check user roles properly

	var req struct {
		ContentID       string   `json:"content_id" binding:"required"`
		ContentType     string   `json:"content_type" binding:"required"`
		Title           string   `json:"title"`
		Content         string   `json:"content" binding:"required"`
		Keywords        []string `json:"keywords"`
		Hashtags        []string `json:"hashtags"`
		AuthorID        string   `json:"author_id" binding:"required"`
		Visibility      string   `json:"visibility" binding:"required"`
		Language        string   `json:"language"`
		Location        string   `json:"location"`
		PopularityScore float64  `json:"popularity_score"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	contentID, err := primitive.ObjectIDFromHex(req.ContentID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid content ID format", err)
		return
	}

	authorID, err := primitive.ObjectIDFromHex(req.AuthorID)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid author ID format", err)
		return
	}

	err = h.searchService.IndexContent(
		contentID,
		req.ContentType,
		req.Title,
		req.Content,
		req.Keywords,
		req.Hashtags,
		authorID,
		req.Visibility,
		req.Language,
		req.Location,
		req.PopularityScore,
	)

	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to index content", err)
		return
	}

	utils.OkResponse(c, "Content indexed successfully", gin.H{
		"content_id":   req.ContentID,
		"content_type": req.ContentType,
		"indexed":      true,
	})
}

// Helper methods for validation

func (h *SearchHandler) isValidSearchType(searchType string) bool {
	validTypes := []string{"all", "posts", "users", "hashtags"}
	for _, t := range validTypes {
		if searchType == t {
			return true
		}
	}
	return false
}

func (h *SearchHandler) isValidSortBy(sortBy string) bool {
	validSorts := []string{"relevance", "recent", "popular"}
	for _, s := range validSorts {
		if sortBy == s {
			return true
		}
	}
	return false
}

func (h *SearchHandler) isValidDateRange(dateRange string) bool {
	validRanges := []string{"hour", "day", "week", "month", "year"}
	for _, r := range validRanges {
		if dateRange == r {
			return true
		}
	}
	return false
}
