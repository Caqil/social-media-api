// internal/services/search_service.go
package services

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"social-media-api/internal/config"
	"social-media-api/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SearchService struct {
	postCollection          *mongo.Collection
	userCollection          *mongo.Collection
	hashtagCollection       *mongo.Collection
	searchHistoryCollection *mongo.Collection
	searchIndexCollection   *mongo.Collection
	db                      *mongo.Database
}

type SearchResult struct {
	Type        string      `json:"type"` // "post", "user", "hashtag", "location"
	Score       float64     `json:"score"`
	Data        interface{} `json:"data"`
	Highlighted string      `json:"highlighted,omitempty"`
	Context     string      `json:"context,omitempty"`
}

type SearchResponse struct {
	Query        string                    `json:"query"`
	Results      []SearchResult            `json:"results"`
	Suggestions  []string                  `json:"suggestions,omitempty"`
	Categories   map[string][]SearchResult `json:"categories"`
	TotalResults int                       `json:"total_results"`
	TimeTaken    time.Duration             `json:"time_taken"`
	Filters      SearchFilters             `json:"filters"`
}

type SearchFilters struct {
	Type        string `json:"type,omitempty"`       // "all", "posts", "users", "hashtags"
	DateRange   string `json:"date_range,omitempty"` // "day", "week", "month", "year"
	SortBy      string `json:"sort_by,omitempty"`    // "relevance", "recent", "popular"
	Location    string `json:"location,omitempty"`
	Language    string `json:"language,omitempty"`
	ContentType string `json:"content_type,omitempty"` // "text", "image", "video"
}

type SearchHistory struct {
	models.BaseModel `bson:",inline"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	Query            string             `json:"query" bson:"query"`
	Type             string             `json:"type" bson:"type"`
	ResultsCount     int                `json:"results_count" bson:"results_count"`
	Clicked          bool               `json:"clicked" bson:"clicked"`
	ClickedResultID  string             `json:"clicked_result_id,omitempty" bson:"clicked_result_id,omitempty"`
}

type HashtagInfo struct {
	models.BaseModel `bson:",inline"`
	Name             string    `json:"name" bson:"name"`
	PostCount        int64     `json:"post_count" bson:"post_count"`
	TrendingScore    float64   `json:"trending_score" bson:"trending_score"`
	LastUsed         time.Time `json:"last_used" bson:"last_used"`
	IsBlocked        bool      `json:"is_blocked" bson:"is_blocked"`
	Category         string    `json:"category" bson:"category"`
}

type SearchIndex struct {
	models.BaseModel `bson:",inline"`
	ContentID        primitive.ObjectID `json:"content_id" bson:"content_id"`
	ContentType      string             `json:"content_type" bson:"content_type"` // "post", "user", "comment"
	Title            string             `json:"title" bson:"title"`
	Content          string             `json:"content" bson:"content"`
	Keywords         []string           `json:"keywords" bson:"keywords"`
	Hashtags         []string           `json:"hashtags" bson:"hashtags"`
	AuthorID         primitive.ObjectID `json:"author_id" bson:"author_id"`
	Visibility       string             `json:"visibility" bson:"visibility"`
	Language         string             `json:"language" bson:"language"`
	Location         string             `json:"location" bson:"location"`
	PopularityScore  float64            `json:"popularity_score" bson:"popularity_score"`
}

func NewSearchService() *SearchService {
	return &SearchService{
		postCollection:          config.DB.Collection("posts"),
		userCollection:          config.DB.Collection("users"),
		hashtagCollection:       config.DB.Collection("hashtags"),
		searchHistoryCollection: config.DB.Collection("search_history"),
		searchIndexCollection:   config.DB.Collection("search_index"),
		db:                      config.DB,
	}
}

// Search performs comprehensive search across all content types
func (ss *SearchService) Search(query string, userID *primitive.ObjectID, filters SearchFilters, limit, skip int) (*SearchResponse, error) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clean and prepare query
	cleanQuery := ss.cleanQuery(query)
	if cleanQuery == "" {
		return &SearchResponse{
			Query:        query,
			Results:      []SearchResult{},
			Categories:   make(map[string][]SearchResult),
			TotalResults: 0,
			TimeTaken:    time.Since(startTime),
			Filters:      filters,
		}, nil
	}

	var allResults []SearchResult
	categories := make(map[string][]SearchResult)

	// Search based on type filter
	switch filters.Type {
	case "posts":
		results, err := ss.searchPosts(ctx, cleanQuery, userID, filters, limit+skip)
		if err == nil {
			allResults = append(allResults, results...)
			categories["posts"] = results
		}
	case "users":
		results, err := ss.searchUsers(ctx, cleanQuery, userID, filters, limit+skip)
		if err == nil {
			allResults = append(allResults, results...)
			categories["users"] = results
		}
	case "hashtags":
		results, err := ss.searchHashtags(ctx, cleanQuery, filters, limit+skip)
		if err == nil {
			allResults = append(allResults, results...)
			categories["hashtags"] = results
		}
	default: // "all" or empty
		// Search all types
		postResults, _ := ss.searchPosts(ctx, cleanQuery, userID, filters, limit/2)
		userResults, _ := ss.searchUsers(ctx, cleanQuery, userID, filters, limit/4)
		hashtagResults, _ := ss.searchHashtags(ctx, cleanQuery, filters, limit/4)

		allResults = append(allResults, postResults...)
		allResults = append(allResults, userResults...)
		allResults = append(allResults, hashtagResults...)

		categories["posts"] = postResults
		categories["users"] = userResults
		categories["hashtags"] = hashtagResults
	}

	// Sort results by relevance score
	ss.sortResultsByScore(allResults, filters.SortBy)

	// Apply pagination
	totalResults := len(allResults)
	if skip >= len(allResults) {
		allResults = []SearchResult{}
	} else {
		end := skip + limit
		if end > len(allResults) {
			end = len(allResults)
		}
		allResults = allResults[skip:end]
	}

	// Get search suggestions
	suggestions := ss.getSearchSuggestions(ctx, cleanQuery, userID)

	// Record search history
	if userID != nil {
		go ss.recordSearchHistory(*userID, query, filters.Type, totalResults)
	}

	response := &SearchResponse{
		Query:        query,
		Results:      allResults,
		Suggestions:  suggestions,
		Categories:   categories,
		TotalResults: totalResults,
		TimeTaken:    time.Since(startTime),
		Filters:      filters,
	}

	return response, nil
}

// searchPosts searches for posts
func (ss *SearchService) searchPosts(ctx context.Context, query string, userID *primitive.ObjectID, filters SearchFilters, limit int) ([]SearchResult, error) {
	// Build search filter
	searchFilter := bson.M{
		"is_published": true,
		"deleted_at":   bson.M{"$exists": false},
	}

	// Add visibility filter
	if userID == nil {
		searchFilter["visibility"] = "public"
	} else {
		// Get user's following list for friends-only posts
		following, _ := ss.getUserFollowing(ctx, *userID)
		searchFilter["$or"] = []bson.M{
			{"visibility": "public"},
			{
				"$and": []bson.M{
					{"visibility": "friends"},
					{"user_id": bson.M{"$in": append(following, *userID)}},
				},
			},
		}
	}

	// Add text search
	searchTerms := ss.buildTextSearchQuery(query)
	if len(searchTerms) > 0 {
		searchFilter["$or"] = []bson.M{
			{"content": bson.M{"$regex": searchTerms, "$options": "i"}},
			{"hashtags": bson.M{"$in": ss.extractHashtags(query)}},
		}
	}

	// Add date filter
	if filters.DateRange != "" {
		dateFilter := ss.getDateFilter(filters.DateRange)
		searchFilter["created_at"] = bson.M{"$gte": dateFilter}
	}

	// Add content type filter
	if filters.ContentType != "" {
		searchFilter["content_type"] = filters.ContentType
	}

	// Add language filter
	if filters.Language != "" {
		searchFilter["language"] = filters.Language
	}

	// Build aggregation pipeline
	pipeline := []bson.M{
		{"$match": searchFilter},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "user_id",
				"foreignField": "_id",
				"as":           "author",
			},
		},
		{"$unwind": "$author"},
		{
			"$addFields": bson.M{
				"relevance_score": ss.buildRelevanceScore(query, "post"),
			},
		},
	}

	// Add sorting
	sortStage := ss.buildSortStage(filters.SortBy)
	pipeline = append(pipeline, sortStage)

	pipeline = append(pipeline, bson.M{"$limit": limit})

	cursor, err := ss.postCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var posts []struct {
		models.Post    `bson:",inline"`
		Author         models.User `bson:"author"`
		RelevanceScore float64     `bson:"relevance_score"`
	}

	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, post := range posts {
		post.Post.Author = post.Author.ToUserResponse()

		result := SearchResult{
			Type:        "post",
			Score:       post.RelevanceScore,
			Data:        post.Post.ToPostResponse(),
			Highlighted: ss.highlightText(post.Post.Content, query),
			Context:     "post",
		}
		results = append(results, result)
	}

	return results, nil
}

// searchUsers searches for users
func (ss *SearchService) searchUsers(ctx context.Context, query string, userID *primitive.ObjectID, filters SearchFilters, limit int) ([]SearchResult, error) {
	searchFilter := bson.M{
		"is_active":  true,
		"deleted_at": bson.M{"$exists": false},
	}

	// Build text search for users
	searchTerms := ss.buildTextSearchQuery(query)
	if len(searchTerms) > 0 {
		searchFilter["$or"] = []bson.M{
			{"username": bson.M{"$regex": searchTerms, "$options": "i"}},
			{"first_name": bson.M{"$regex": searchTerms, "$options": "i"}},
			{"last_name": bson.M{"$regex": searchTerms, "$options": "i"}},
			{"display_name": bson.M{"$regex": searchTerms, "$options": "i"}},
			{"bio": bson.M{"$regex": searchTerms, "$options": "i"}},
		}
	}

	pipeline := []bson.M{
		{"$match": searchFilter},
		{
			"$addFields": bson.M{
				"relevance_score": ss.buildRelevanceScore(query, "user"),
			},
		},
	}

	// Add sorting
	sortStage := ss.buildSortStage(filters.SortBy)
	if filters.SortBy == "popular" {
		sortStage = bson.M{"$sort": bson.M{"followers_count": -1, "relevance_score": -1}}
	}
	pipeline = append(pipeline, sortStage)

	pipeline = append(pipeline, bson.M{"$limit": limit})

	cursor, err := ss.userCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []struct {
		models.User    `bson:",inline"`
		RelevanceScore float64 `bson:"relevance_score"`
	}

	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, user := range users {
		// Check if current user is following this user
		isFollowing := false
		if userID != nil {
			isFollowing = ss.checkIfFollowing(ctx, *userID, user.User.ID)
		}

		userResponse := user.User.ToUserResponseWithContext(
			primitive.NilObjectID, // This would be properly set with relationship data
			isFollowing, false, false, false, 0,
		)

		result := SearchResult{
			Type:        "user",
			Score:       user.RelevanceScore + float64(user.User.FollowersCount)*0.1, // Boost popular users
			Data:        userResponse,
			Highlighted: ss.highlightUserText(user.User, query),
			Context:     "user",
		}
		results = append(results, result)
	}

	return results, nil
}

// searchHashtags searches for hashtags
func (ss *SearchService) searchHashtags(ctx context.Context, query string, filters SearchFilters, limit int) ([]SearchResult, error) {
	cleanQuery := strings.TrimPrefix(strings.ToLower(query), "#")

	searchFilter := bson.M{
		"name":       bson.M{"$regex": cleanQuery, "$options": "i"},
		"is_blocked": false,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.M{"trending_score": -1, "post_count": -1})

	cursor, err := ss.hashtagCollection.Find(ctx, searchFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var hashtags []HashtagInfo
	if err := cursor.All(ctx, &hashtags); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, hashtag := range hashtags {
		score := float64(hashtag.PostCount)*0.1 + hashtag.TrendingScore

		// Boost exact matches
		if strings.ToLower(hashtag.Name) == cleanQuery {
			score *= 2.0
		}

		result := SearchResult{
			Type:        "hashtag",
			Score:       score,
			Data:        hashtag,
			Highlighted: ss.highlightText(hashtag.Name, cleanQuery),
			Context:     fmt.Sprintf("%d posts", hashtag.PostCount),
		}
		results = append(results, result)
	}

	return results, nil
}

// GetTrendingHashtags returns trending hashtags
func (ss *SearchService) GetTrendingHashtags(limit int, timeRange string) ([]HashtagInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Calculate trending score based on recent usage
	dateFilter := ss.getDateFilter(timeRange)

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"last_used":  bson.M{"$gte": dateFilter},
				"is_blocked": false,
			},
		},
		{
			"$addFields": bson.M{
				"recency_bonus": bson.M{
					"$divide": []interface{}{
						1,
						bson.M{
							"$add": []interface{}{
								1,
								bson.M{
									"$divide": []interface{}{
										bson.M{"$subtract": []interface{}{time.Now(), "$last_used"}},
										1000 * 60 * 60 * 24, // Convert to days
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"final_trending_score": bson.M{
					"$add": []interface{}{
						"$trending_score",
						bson.M{"$multiply": []interface{}{"$recency_bonus", 10}},
					},
				},
			},
		},
		{
			"$sort": bson.M{
				"final_trending_score": -1,
				"post_count":           -1,
			},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := ss.hashtagCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var hashtags []HashtagInfo
	if err := cursor.All(ctx, &hashtags); err != nil {
		return nil, err
	}

	return hashtags, nil
}

// GetSearchSuggestions returns search suggestions based on query
func (ss *SearchService) getSearchSuggestions(ctx context.Context, query string, userID *primitive.ObjectID) []string {
	var suggestions []string

	// Get hashtag suggestions
	hashtagSuggestions := ss.getHashtagSuggestions(ctx, query, 3)
	suggestions = append(suggestions, hashtagSuggestions...)

	// Get user suggestions
	userSuggestions := ss.getUserSuggestions(ctx, query, 3)
	suggestions = append(suggestions, userSuggestions...)

	// Get trending queries
	trendingSuggestions := ss.getTrendingQueries(ctx, 2)
	suggestions = append(suggestions, trendingSuggestions...)

	// Remove duplicates and limit
	suggestions = ss.removeDuplicates(suggestions)
	if len(suggestions) > 8 {
		suggestions = suggestions[:8]
	}

	return suggestions
}

// UpdateHashtagInfo updates hashtag information when used in posts
func (ss *SearchService) UpdateHashtagInfo(hashtag string, postID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	// Calculate trending score boost (more recent usage = higher boost)
	trendingBoost := 1.0

	filter := bson.M{"name": strings.ToLower(hashtag)}
	update := bson.M{
		"$set": bson.M{
			"name":       strings.ToLower(hashtag),
			"last_used":  now,
			"updated_at": now,
		},
		"$inc": bson.M{
			"post_count":     1,
			"trending_score": trendingBoost,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := ss.hashtagCollection.UpdateOne(ctx, filter, update, opts)

	return err
}

// IndexContent adds content to search index
func (ss *SearchService) IndexContent(contentID primitive.ObjectID, contentType, title, content string, keywords, hashtags []string, authorID primitive.ObjectID, visibility, language, location string, popularityScore float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchIndex := &SearchIndex{
		ContentID:       contentID,
		ContentType:     contentType,
		Title:           title,
		Content:         content,
		Keywords:        keywords,
		Hashtags:        hashtags,
		AuthorID:        authorID,
		Visibility:      visibility,
		Language:        language,
		Location:        location,
		PopularityScore: popularityScore,
	}
	searchIndex.BeforeCreate()

	// Upsert search index
	filter := bson.M{
		"content_id":   contentID,
		"content_type": contentType,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := ss.searchIndexCollection.ReplaceOne(ctx, filter, searchIndex, opts)

	return err
}

// Helper methods

func (ss *SearchService) cleanQuery(query string) string {
	// Remove extra spaces and trim
	query = strings.TrimSpace(query)
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")

	// Minimum query length
	if len(query) < 2 {
		return ""
	}

	return query
}

func (ss *SearchService) buildTextSearchQuery(query string) string {
	// Escape special regex characters
	escaped := regexp.QuoteMeta(query)

	// Build regex pattern for partial matching
	words := strings.Fields(escaped)
	if len(words) == 0 {
		return escaped
	}

	// Create pattern that matches any of the words
	return strings.Join(words, "|")
}

func (ss *SearchService) buildRelevanceScore(query string, contentType string) bson.M {
	queryWords := strings.Fields(strings.ToLower(query))

	var conditions []bson.M
	baseScore := 1.0

	for _, word := range queryWords {
		// Exact match gets highest score
		conditions = append(conditions, bson.M{
			"$cond": []interface{}{
				bson.M{"$regexMatch": bson.M{
					"input": bson.M{"$toLower": "$content"},
					"regex": fmt.Sprintf("\\b%s\\b", regexp.QuoteMeta(word)),
				}},
				10.0, 0.0,
			},
		})

		// Partial match gets lower score
		conditions = append(conditions, bson.M{
			"$cond": []interface{}{
				bson.M{"$regexMatch": bson.M{
					"input": bson.M{"$toLower": "$content"},
					"regex": regexp.QuoteMeta(word),
				}},
				5.0, 0.0,
			},
		})
	}

	if len(conditions) == 0 {
		return bson.M{"$literal": baseScore}
	}

	return bson.M{"$add": conditions}
}

func (ss *SearchService) buildSortStage(sortBy string) bson.M {
	switch sortBy {
	case "recent":
		return bson.M{"$sort": bson.M{"created_at": -1}}
	case "popular":
		return bson.M{"$sort": bson.M{"likes_count": -1, "comments_count": -1}}
	default: // "relevance"
		return bson.M{"$sort": bson.M{"relevance_score": -1, "created_at": -1}}
	}
}

func (ss *SearchService) getDateFilter(dateRange string) time.Time {
	now := time.Now()
	switch dateRange {
	case "day":
		return now.Add(-24 * time.Hour)
	case "week":
		return now.Add(-7 * 24 * time.Hour)
	case "month":
		return now.Add(-30 * 24 * time.Hour)
	case "year":
		return now.Add(-365 * 24 * time.Hour)
	default:
		return time.Time{} // No filter
	}
}

func (ss *SearchService) extractHashtags(text string) []string {
	re := regexp.MustCompile(`#\w+`)
	matches := re.FindAllString(text, -1)

	var hashtags []string
	for _, match := range matches {
		hashtags = append(hashtags, strings.ToLower(strings.TrimPrefix(match, "#")))
	}

	return hashtags
}

func (ss *SearchService) highlightText(text, query string) string {
	if query == "" {
		return text
	}

	words := strings.Fields(query)
	highlighted := text

	for _, word := range words {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
		highlighted = re.ReplaceAllStringFunc(highlighted, func(match string) string {
			return fmt.Sprintf("<mark>%s</mark>", match)
		})
	}

	return highlighted
}

func (ss *SearchService) highlightUserText(user models.User, query string) string {
	// Highlight in username, display name, or bio
	if strings.Contains(strings.ToLower(user.Username), strings.ToLower(query)) {
		return ss.highlightText(user.Username, query)
	}
	if strings.Contains(strings.ToLower(user.DisplayName), strings.ToLower(query)) {
		return ss.highlightText(user.DisplayName, query)
	}
	if strings.Contains(strings.ToLower(user.Bio), strings.ToLower(query)) {
		return ss.highlightText(user.Bio, query)
	}
	return user.Username
}

func (ss *SearchService) sortResultsByScore(results []SearchResult, sortBy string) {
	switch sortBy {
	case "recent":
		// Sort by creation time (would need to extract from data)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score // Fallback to score
		})
	case "popular":
		// Sort by popularity (likes, followers, etc.)
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	default: // "relevance"
		sort.Slice(results, func(i, j int) bool {
			return results[i].Score > results[j].Score
		})
	}
}

func (ss *SearchService) getUserFollowing(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	// This would typically query the follows collection
	// For now, return empty slice
	return []primitive.ObjectID{}, nil
}

func (ss *SearchService) checkIfFollowing(ctx context.Context, userID, targetUserID primitive.ObjectID) bool {
	// Check if userID follows targetUserID
	// Implementation would query follows collection
	return false
}

func (ss *SearchService) getHashtagSuggestions(ctx context.Context, query string, limit int) []string {
	filter := bson.M{
		"name":       bson.M{"$regex": query, "$options": "i"},
		"is_blocked": false,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.M{"post_count": -1})

	cursor, err := ss.hashtagCollection.Find(ctx, filter, opts)
	if err != nil {
		return []string{}
	}
	defer cursor.Close(ctx)

	var hashtags []HashtagInfo
	cursor.All(ctx, &hashtags)

	var suggestions []string
	for _, hashtag := range hashtags {
		suggestions = append(suggestions, "#"+hashtag.Name)
	}

	return suggestions
}

func (ss *SearchService) getUserSuggestions(ctx context.Context, query string, limit int) []string {
	filter := bson.M{
		"username":  bson.M{"$regex": query, "$options": "i"},
		"is_active": true,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.M{"followers_count": -1})

	cursor, err := ss.userCollection.Find(ctx, filter, opts)
	if err != nil {
		return []string{}
	}
	defer cursor.Close(ctx)

	var users []models.User
	cursor.All(ctx, &users)

	var suggestions []string
	for _, user := range users {
		suggestions = append(suggestions, "@"+user.Username)
	}

	return suggestions
}

func (ss *SearchService) getTrendingQueries(ctx context.Context, limit int) []string {
	// Get most searched queries from search history
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"created_at": bson.M{"$gte": time.Now().Add(-24 * time.Hour)},
			},
		},
		{
			"$group": bson.M{
				"_id":   "$query",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := ss.searchHistoryCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return []string{}
	}
	defer cursor.Close(ctx)

	var results []struct {
		Query string `bson:"_id"`
		Count int    `bson:"count"`
	}
	cursor.All(ctx, &results)

	var suggestions []string
	for _, result := range results {
		suggestions = append(suggestions, result.Query)
	}

	return suggestions
}

func (ss *SearchService) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (ss *SearchService) recordSearchHistory(userID primitive.ObjectID, query, searchType string, resultsCount int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	history := &SearchHistory{
		UserID:       userID,
		Query:        query,
		Type:         searchType,
		ResultsCount: resultsCount,
		Clicked:      false,
	}
	history.BeforeCreate()

	ss.searchHistoryCollection.InsertOne(ctx, history)
}

// CreateIndexes creates necessary indexes for search functionality
func (ss *SearchService) CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Text indexes for full-text search
	postTextIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "content", Value: "text"},
			{Key: "hashtags", Value: "text"},
		},
		Options: options.Index().SetBackground(true),
	}

	userTextIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "username", Value: "text"},
			{Key: "first_name", Value: "text"},
			{Key: "last_name", Value: "text"},
			{Key: "display_name", Value: "text"},
			{Key: "bio", Value: "text"},
		},
		Options: options.Index().SetBackground(true),
	}

	// Create indexes
	_, err := ss.postCollection.Indexes().CreateOne(ctx, postTextIndex)
	if err != nil {
		return err
	}

	_, err = ss.userCollection.Indexes().CreateOne(ctx, userTextIndex)
	if err != nil {
		return err
	}

	// Hashtag indexes
	hashtagIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "trending_score", Value: -1}},
			Options: options.Index().SetBackground(true),
		},
		{
			Keys:    bson.D{{Key: "post_count", Value: -1}},
			Options: options.Index().SetBackground(true),
		},
	}

	_, err = ss.hashtagCollection.Indexes().CreateMany(ctx, hashtagIndexes)
	return err
}
