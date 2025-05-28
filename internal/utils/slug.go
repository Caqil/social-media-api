// utils/slug.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// SlugOptions represents options for slug generation
type SlugOptions struct {
	MaxLength     int
	Separator     string
	Lowercase     bool
	AllowUnicode  bool
	ReplaceSpaces bool
	Prefix        string
	Suffix        string
}

// DefaultSlugOptions returns default slug generation options
func DefaultSlugOptions() SlugOptions {
	return SlugOptions{
		MaxLength:     100,
		Separator:     "-",
		Lowercase:     true,
		AllowUnicode:  false,
		ReplaceSpaces: true,
	}
}

// GenerateSlug generates a URL-friendly slug from text
func GenerateSlug(text string) string {
	return GenerateSlugWithOptions(text, DefaultSlugOptions())
}

// GenerateSlugWithOptions generates a slug with custom options
func GenerateSlugWithOptions(text string, options SlugOptions) string {
	if text == "" {
		return ""
	}

	slug := text

	// Convert to lowercase if specified
	if options.Lowercase {
		slug = strings.ToLower(slug)
	}

	// Remove or replace unicode characters if not allowed
	if !options.AllowUnicode {
		slug = removeUnicodeCharacters(slug)
	}

	// Replace spaces with separator if specified
	if options.ReplaceSpaces {
		slug = strings.ReplaceAll(slug, " ", options.Separator)
	}

	// Replace multiple spaces/separators with single separator
	multipleSpaces := regexp.MustCompile(`\s+`)
	slug = multipleSpaces.ReplaceAllString(slug, options.Separator)

	// Remove special characters (keep alphanumeric, hyphens, underscores)
	if options.AllowUnicode {
		specialChars := regexp.MustCompile(`[^\p{L}\p{N}_-]+`)
		slug = specialChars.ReplaceAllString(slug, options.Separator)
	} else {
		specialChars := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
		slug = specialChars.ReplaceAllString(slug, options.Separator)
	}

	// Replace multiple separators with single separator
	multipleSeparators := regexp.MustCompile(regexp.QuoteMeta(options.Separator) + `+`)
	slug = multipleSeparators.ReplaceAllString(slug, options.Separator)

	// Trim separators from beginning and end
	slug = strings.Trim(slug, options.Separator)

	// Add prefix if specified
	if options.Prefix != "" {
		slug = options.Prefix + options.Separator + slug
	}

	// Add suffix if specified
	if options.Suffix != "" {
		slug = slug + options.Separator + options.Suffix
	}

	// Truncate if too long
	if options.MaxLength > 0 && len(slug) > options.MaxLength {
		slug = slug[:options.MaxLength]
		// Ensure we don't end with a separator
		slug = strings.TrimRight(slug, options.Separator)
	}

	// Ensure slug is not empty
	if slug == "" {
		return generateRandomSlug(8)
	}

	return slug
}

// GenerateUniqueSlug generates a unique slug by appending a number if needed
func GenerateUniqueSlug(text string, existingSlugs []string) string {
	baseSlug := GenerateSlug(text)

	if !contains(existingSlugs, baseSlug) {
		return baseSlug
	}

	// Try appending numbers
	for i := 2; i <= 1000; i++ {
		slug := fmt.Sprintf("%s-%d", baseSlug, i)
		if !contains(existingSlugs, slug) {
			return slug
		}
	}

	// If still not unique, append random string
	randomSuffix := generateRandomString(6)
	return fmt.Sprintf("%s-%s", baseSlug, randomSuffix)
}

// GenerateSlugWithTimestamp generates a slug with timestamp
func GenerateSlugWithTimestamp(text string) string {
	baseSlug := GenerateSlug(text)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", baseSlug, timestamp)
}

// GenerateSlugWithRandomSuffix generates a slug with random suffix
func GenerateSlugWithRandomSuffix(text string, suffixLength int) string {
	baseSlug := GenerateSlug(text)
	suffix := generateRandomString(suffixLength)
	return fmt.Sprintf("%s-%s", baseSlug, suffix)
}

// GenerateUsernameSlug generates a slug suitable for usernames
func GenerateUsernameSlug(name string) string {
	options := SlugOptions{
		MaxLength:     30,
		Separator:     "",
		Lowercase:     true,
		AllowUnicode:  false,
		ReplaceSpaces: false,
	}

	slug := GenerateSlugWithOptions(name, options)

	// Remove separators for username
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "_", "")

	// Ensure it starts with a letter
	if len(slug) > 0 && !unicode.IsLetter(rune(slug[0])) {
		slug = "user" + slug
	}

	// Ensure minimum length
	if len(slug) < 3 {
		slug = slug + generateRandomString(3-len(slug))
	}

	return slug
}

// GenerateGroupSlug generates a slug suitable for groups
func GenerateGroupSlug(name string) string {
	options := SlugOptions{
		MaxLength:     50,
		Separator:     "-",
		Lowercase:     true,
		AllowUnicode:  false,
		ReplaceSpaces: true,
	}

	return GenerateSlugWithOptions(name, options)
}

// GenerateEventSlug generates a slug suitable for events
func GenerateEventSlug(title string) string {
	options := SlugOptions{
		MaxLength:     60,
		Separator:     "-",
		Lowercase:     true,
		AllowUnicode:  false,
		ReplaceSpaces: true,
	}

	return GenerateSlugWithOptions(title, options)
}

// GenerateHashtagSlug generates a slug suitable for hashtags
func GenerateHashtagSlug(tag string) string {
	// Remove # if present
	tag = strings.TrimPrefix(tag, "#")

	options := SlugOptions{
		MaxLength:     50,
		Separator:     "",
		Lowercase:     true,
		AllowUnicode:  false,
		ReplaceSpaces: false,
	}

	slug := GenerateSlugWithOptions(tag, options)

	// Remove all separators for hashtags
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "_", "")

	return slug
}

// ValidateSlug validates if a string is a valid slug
func ValidateSlug(slug string) bool {
	if slug == "" {
		return false
	}

	// Check length
	if len(slug) > 100 {
		return false
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	validSlug := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validSlug.MatchString(slug) {
		return false
	}

	// Ensure it doesn't start or end with separator
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}
	if strings.HasPrefix(slug, "_") || strings.HasSuffix(slug, "_") {
		return false
	}

	// Ensure no consecutive separators
	if strings.Contains(slug, "--") || strings.Contains(slug, "__") {
		return false
	}

	return true
}

// IsReservedSlug checks if a slug is reserved
func IsReservedSlug(slug string) bool {
	reservedSlugs := []string{
		"admin", "api", "www", "mail", "ftp", "localhost", "app", "application",
		"user", "users", "profile", "profiles", "account", "accounts",
		"setting", "settings", "config", "configuration", "help", "support",
		"about", "contact", "privacy", "terms", "legal", "policy",
		"blog", "news", "home", "index", "root", "public", "static",
		"assets", "images", "css", "js", "javascript", "style", "stylesheet",
		"login", "logout", "signin", "signup", "register", "auth", "oauth",
		"dashboard", "panel", "control", "manage", "management", "system",
		"post", "posts", "comment", "comments", "like", "likes", "share",
		"group", "groups", "event", "events", "message", "messages",
		"notification", "notifications", "search", "explore", "trending",
		"feed", "timeline", "activity", "activities", "follow", "followers",
		"following", "friend", "friends", "story", "stories", "live",
	}

	for _, reserved := range reservedSlugs {
		if strings.EqualFold(slug, reserved) {
			return true
		}
	}

	return false
}

// SuggestSlugAlternatives suggests alternative slugs
func SuggestSlugAlternatives(text string, count int) []string {
	if count <= 0 {
		count = 5
	}

	baseSlug := GenerateSlug(text)
	suggestions := make([]string, 0, count)

	// Add base slug
	suggestions = append(suggestions, baseSlug)

	// Add numbered variations
	for i := 2; len(suggestions) < count && i <= count+1; i++ {
		suggestions = append(suggestions, fmt.Sprintf("%s-%d", baseSlug, i))
	}

	// Add random variations if needed
	for len(suggestions) < count {
		randomSuffix := generateRandomString(4)
		suggestions = append(suggestions, fmt.Sprintf("%s-%s", baseSlug, randomSuffix))
	}

	return suggestions
}

// ConvertToSlug converts various inputs to slug format
func ConvertToSlug(input interface{}) string {
	var text string

	switch v := input.(type) {
	case string:
		text = v
	case fmt.Stringer:
		text = v.String()
	default:
		text = fmt.Sprintf("%v", v)
	}

	return GenerateSlug(text)
}

// SlugToTitle converts a slug back to a readable title
func SlugToTitle(slug string) string {
	if slug == "" {
		return ""
	}

	// Replace separators with spaces
	title := strings.ReplaceAll(slug, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")

	// Capitalize first letter of each word
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// Helper functions

// removeUnicodeCharacters removes non-ASCII characters
func removeUnicodeCharacters(text string) string {
	// Simple ASCII-only implementation
	var result strings.Builder
	for _, r := range text {
		if r <= 127 {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// generateRandomSlug generates a random slug
func generateRandomSlug(length int) string {
	return "slug-" + generateRandomString(length)
}

// generateRandomString generates a random string
func generateRandomString(length int) string {
	bytes := make([]byte, length/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CleanSlug cleans an existing slug
func CleanSlug(slug string) string {
	return GenerateSlug(slug)
}

// ExtractSlugFromURL extracts slug from URL path
func ExtractSlugFromURL(urlPath string) string {
	// Remove leading/trailing slashes
	path := strings.Trim(urlPath, "/")

	// Split by slash and get last segment
	segments := strings.Split(path, "/")
	if len(segments) > 0 {
		return segments[len(segments)-1]
	}

	return ""
}

// IsValidUsernameSlug validates username slug format
func IsValidUsernameSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 30 {
		return false
	}

	// Must start with letter
	if !unicode.IsLetter(rune(slug[0])) {
		return false
	}

	// Only alphanumeric and underscores
	validUsername := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	return validUsername.MatchString(slug)
}

// GenerateShortSlug generates a short slug (useful for URLs)
func GenerateShortSlug(text string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 20
	}

	options := DefaultSlugOptions()
	options.MaxLength = maxLength

	slug := GenerateSlugWithOptions(text, options)

	// If still too long, truncate at word boundary
	if len(slug) > maxLength {
		words := strings.Split(slug, "-")
		result := ""

		for _, word := range words {
			if len(result)+len(word)+1 <= maxLength {
				if result != "" {
					result += "-"
				}
				result += word
			} else {
				break
			}
		}

		if result != "" {
			slug = result
		} else {
			// If no words fit, just truncate
			slug = slug[:maxLength]
			slug = strings.TrimRight(slug, "-")
		}
	}

	return slug
}
