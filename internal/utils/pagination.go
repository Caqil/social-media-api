// utils/pagination.go
package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	CurrentPage  int   `json:"current_page"`
	PerPage      int   `json:"per_page"`
	Total        int64 `json:"total"`
	TotalPages   int   `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
	From         int   `json:"from"`
	To           int   `json:"to"`
}

// PaginatedResult represents a paginated response
type PaginatedResult struct {
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// CursorPaginationParams represents cursor-based pagination parameters
type CursorPaginationParams struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit"`
}

// CursorPaginationMeta represents cursor pagination metadata
type CursorPaginationMeta struct {
	HasNext     bool   `json:"has_next"`
	HasPrevious bool   `json:"has_previous"`
	NextCursor  string `json:"next_cursor,omitempty"`
	PrevCursor  string `json:"prev_cursor,omitempty"`
	Count       int    `json:"count"`
}

// CursorPaginatedResult represents a cursor-based paginated response
type CursorPaginatedResult struct {
	Data       interface{}          `json:"data"`
	Pagination CursorPaginationMeta `json:"pagination"`
}

// GetPaginationParams extracts pagination parameters from request
func GetPaginationParams(c *gin.Context) PaginationParams {
	page := 1
	limit := DefaultPageSize

	// Get page parameter
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Get limit parameter
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > MaxPageSize {
				limit = MaxPageSize
			} else if l < MinPageSize {
				limit = MinPageSize
			} else {
				limit = l
			}
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

// GetCursorPaginationParams extracts cursor pagination parameters from request
func GetCursorPaginationParams(c *gin.Context) CursorPaginationParams {
	cursor := c.Query("cursor")
	limit := DefaultPageSize

	// Get limit parameter
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > MaxPageSize {
				limit = MaxPageSize
			} else if l < MinPageSize {
				limit = MinPageSize
			} else {
				limit = l
			}
		}
	}

	return CursorPaginationParams{
		Cursor: cursor,
		Limit:  limit,
	}
}

// CreatePaginationMeta creates pagination metadata
func CreatePaginationMeta(params PaginationParams, total int64) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	// Calculate from and to
	from := params.Offset + 1
	to := params.Offset + params.Limit
	if int64(to) > total {
		to = int(total)
	}
	if total == 0 {
		from = 0
		to = 0
	}

	meta := PaginationMeta{
		CurrentPage: params.Page,
		PerPage:     params.Limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     params.Page < totalPages,
		HasPrevious: params.Page > 1,
		From:        from,
		To:          to,
	}

	// Set next page
	if meta.HasNext {
		nextPage := params.Page + 1
		meta.NextPage = &nextPage
	}

	// Set previous page
	if meta.HasPrevious {
		prevPage := params.Page - 1
		meta.PreviousPage = &prevPage
	}

	return meta
}

// CreatePaginatedResult creates a paginated result
func CreatePaginatedResult(data interface{}, params PaginationParams, total int64) PaginatedResult {
	return PaginatedResult{
		Data:       data,
		Pagination: CreatePaginationMeta(params, total),
	}
}

// CreateCursorPaginatedResult creates a cursor-based paginated result
func CreateCursorPaginatedResult(data interface{}, meta CursorPaginationMeta) CursorPaginatedResult {
	return CursorPaginatedResult{
		Data:       data,
		Pagination: meta,
	}
}

// GetMongoFindOptions returns MongoDB find options for pagination
func GetMongoFindOptions(params PaginationParams) *options.FindOptions {
	return options.Find().
		SetSkip(int64(params.Offset)).
		SetLimit(int64(params.Limit))
}

// GetMongoFindOptionsWithSort returns MongoDB find options with sorting
func GetMongoFindOptionsWithSort(params PaginationParams, sortField string, sortOrder int) *options.FindOptions {
	return options.Find().
		SetSkip(int64(params.Offset)).
		SetLimit(int64(params.Limit)).
		SetSort(bson.D{{Key: sortField, Value: sortOrder}})
}

// GetSortOrder returns sort order from query parameter
func GetSortOrder(c *gin.Context, defaultField string, defaultOrder int) (string, int) {
	sortField := defaultField
	sortOrder := defaultOrder

	// Get sort field
	if sort := c.Query("sort"); sort != "" {
		sortField = sort
	}

	// Get sort order
	if order := c.Query("order"); order != "" {
		switch order {
		case "asc", "ascending", "1":
			sortOrder = 1
		case "desc", "descending", "-1":
			sortOrder = -1
		}
	}

	return sortField, sortOrder
}

// GetSortOrderFromString converts string sort order to int
func GetSortOrderFromString(order string) int {
	switch order {
	case "asc", "ascending", "1":
		return 1
	case "desc", "descending", "-1":
		return -1
	default:
		return -1 // Default to descending
	}
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(params PaginationParams) PaginationParams {
	// Validate page
	if params.Page < 1 {
		params.Page = 1
	}

	// Validate limit
	if params.Limit < MinPageSize {
		params.Limit = MinPageSize
	} else if params.Limit > MaxPageSize {
		params.Limit = MaxPageSize
	}

	// Recalculate offset
	params.Offset = (params.Page - 1) * params.Limit

	return params
}

// PaginationQuery represents a query builder for pagination
type PaginationQuery struct {
	Filter bson.M
	Sort   bson.D
	Params PaginationParams
}

// NewPaginationQuery creates a new pagination query builder
func NewPaginationQuery() *PaginationQuery {
	return &PaginationQuery{
		Filter: bson.M{},
		Sort:   bson.D{},
		Params: PaginationParams{
			Page:  1,
			Limit: DefaultPageSize,
		},
	}
}

// SetFilter sets MongoDB filter
func (pq *PaginationQuery) SetFilter(filter bson.M) *PaginationQuery {
	pq.Filter = filter
	return pq
}

// AddFilter adds to MongoDB filter
func (pq *PaginationQuery) AddFilter(key string, value interface{}) *PaginationQuery {
	pq.Filter[key] = value
	return pq
}

// SetSort sets MongoDB sort
func (pq *PaginationQuery) SetSort(sort bson.D) *PaginationQuery {
	pq.Sort = sort
	return pq
}

// AddSort adds to MongoDB sort
func (pq *PaginationQuery) AddSort(field string, order int) *PaginationQuery {
	pq.Sort = append(pq.Sort, bson.E{Key: field, Value: order})
	return pq
}

// SetParams sets pagination parameters
func (pq *PaginationQuery) SetParams(params PaginationParams) *PaginationQuery {
	pq.Params = ValidatePaginationParams(params)
	return pq
}

// GetFindOptions returns MongoDB find options
func (pq *PaginationQuery) GetFindOptions() *options.FindOptions {
	opts := options.Find().
		SetSkip(int64(pq.Params.Offset)).
		SetLimit(int64(pq.Params.Limit))

	if len(pq.Sort) > 0 {
		opts.SetSort(pq.Sort)
	}

	return opts
}

// SearchPaginationParams represents search with pagination
type SearchPaginationParams struct {
	PaginationParams
	Query    string `json:"query"`
	Category string `json:"category,omitempty"`
	Type     string `json:"type,omitempty"`
}

// GetSearchPaginationParams extracts search pagination parameters
func GetSearchPaginationParams(c *gin.Context) SearchPaginationParams {
	params := GetPaginationParams(c)

	return SearchPaginationParams{
		PaginationParams: params,
		Query:            c.Query("q"),
		Category:         c.Query("category"),
		Type:             c.Query("type"),
	}
}

// TimeBasedPaginationParams represents time-based pagination
type TimeBasedPaginationParams struct {
	Before string `json:"before,omitempty"` // Timestamp or ID
	After  string `json:"after,omitempty"`  // Timestamp or ID
	Limit  int    `json:"limit"`
}

// GetTimeBasedPaginationParams extracts time-based pagination parameters
func GetTimeBasedPaginationParams(c *gin.Context) TimeBasedPaginationParams {
	limit := DefaultPageSize

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > MaxPageSize {
				limit = MaxPageSize
			} else {
				limit = l
			}
		}
	}

	return TimeBasedPaginationParams{
		Before: c.Query("before"),
		After:  c.Query("after"),
		Limit:  limit,
	}
}

// FeedPaginationParams represents feed pagination parameters
type FeedPaginationParams struct {
	LastID    string `json:"last_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Limit     int    `json:"limit"`
}

// GetFeedPaginationParams extracts feed pagination parameters
func GetFeedPaginationParams(c *gin.Context) FeedPaginationParams {
	limit := DefaultPageSize

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > MaxPageSize {
				limit = MaxPageSize
			} else {
				limit = l
			}
		}
	}

	return FeedPaginationParams{
		LastID:    c.Query("last_id"),
		Timestamp: c.Query("timestamp"),
		Limit:     limit,
	}
}

// InfiniteScrollParams represents infinite scroll pagination
type InfiniteScrollParams struct {
	Page    int    `json:"page"`
	Size    int    `json:"size"`
	LastID  string `json:"last_id,omitempty"`
	HasMore bool   `json:"has_more"`
}

// GetInfiniteScrollParams extracts infinite scroll parameters
func GetInfiniteScrollParams(c *gin.Context) InfiniteScrollParams {
	page := 1
	size := DefaultPageSize

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			if s > MaxPageSize {
				size = MaxPageSize
			} else {
				size = s
			}
		}
	}

	return InfiniteScrollParams{
		Page:   page,
		Size:   size,
		LastID: c.Query("last_id"),
	}
}

// CalculateInfiniteScrollMeta calculates infinite scroll metadata
func CalculateInfiniteScrollMeta(params InfiniteScrollParams, returnedCount int) InfiniteScrollParams {
	params.HasMore = returnedCount >= params.Size
	return params
}

// PaginationLinks represents pagination links for APIs
type PaginationLinks struct {
	First    string `json:"first,omitempty"`
	Previous string `json:"previous,omitempty"`
	Next     string `json:"next,omitempty"`
	Last     string `json:"last,omitempty"`
	Self     string `json:"self"`
}

// GeneratePaginationLinks generates pagination links
func GeneratePaginationLinks(baseURL string, params PaginationParams, meta PaginationMeta) PaginationLinks {
	links := PaginationLinks{
		Self: fmt.Sprintf("%s?page=%d&limit=%d", baseURL, params.Page, params.Limit),
	}

	// First page link
	if meta.TotalPages > 0 {
		links.First = fmt.Sprintf("%s?page=1&limit=%d", baseURL, params.Limit)
	}

	// Previous page link
	if meta.HasPrevious {
		links.Previous = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, params.Page-1, params.Limit)
	}

	// Next page link
	if meta.HasNext {
		links.Next = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, params.Page+1, params.Limit)
	}

	// Last page link
	if meta.TotalPages > 0 {
		links.Last = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, meta.TotalPages, params.Limit)
	}

	return links
}

// PaginatedResultWithLinks represents paginated result with navigation links
type PaginatedResultWithLinks struct {
	Data       interface{}     `json:"data"`
	Pagination PaginationMeta  `json:"pagination"`
	Links      PaginationLinks `json:"links"`
}

// CreatePaginatedResultWithLinks creates paginated result with links
func CreatePaginatedResultWithLinks(data interface{}, params PaginationParams, total int64, baseURL string) PaginatedResultWithLinks {
	meta := CreatePaginationMeta(params, total)
	links := GeneratePaginationLinks(baseURL, params, meta)

	return PaginatedResultWithLinks{
		Data:       data,
		Pagination: meta,
		Links:      links,
	}
}
