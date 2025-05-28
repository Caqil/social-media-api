// Add this to internal/utils/datetime.go (create new file)
package utils

import (
	"fmt"
	"strings"
	"time"
)

// Common datetime formats
var DateTimeFormats = []string{
	time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
	time.RFC3339Nano,      // "2006-01-02T15:04:05.999999999Z07:00"
	"2006-01-02T15:04:05", // ISO format without timezone
	"2006-01-02 15:04:05", // MySQL datetime format
	"2006-01-02",          // Date only
	"15:04:05",            // Time only
	"2006/01/02 15:04:05", // Alternative format
	"2006/01/02",          // Alternative date format
	"01/02/2006",          // US date format
	"01/02/2006 15:04:05", // US datetime format
	"02/01/2006",          // European date format
	"02/01/2006 15:04:05", // European datetime format
}

// ParseDateTime parses various datetime string formats
func ParseDateTime(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Clean the input string
	dateStr = strings.TrimSpace(dateStr)

	// Try each format
	for _, format := range DateTimeFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	// Try parsing Unix timestamp
	if unixTime, err := parseUnixTimestamp(dateStr); err == nil {
		return unixTime, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", dateStr)
}

// parseUnixTimestamp attempts to parse Unix timestamp
func parseUnixTimestamp(s string) (time.Time, error) {
	// Try parsing as Unix timestamp (seconds)
	if len(s) == 10 {
		if timestamp, err := StringToInt64(s); err == nil {
			return time.Unix(timestamp, 0), nil
		}
	}

	// Try parsing as Unix timestamp (milliseconds)
	if len(s) == 13 {
		if timestamp, err := StringToInt64(s); err == nil {
			return time.Unix(0, timestamp*int64(time.Millisecond)), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid unix timestamp")
}

// FormatDateTime formats time to a standard format
func FormatDateTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatDateOnly formats time to date only
func FormatDateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatTimeOnly formats time to time only
func FormatTimeOnly(t time.Time) string {
	return t.Format("15:04:05")
}

// GetTimeAgo returns human-readable time difference
func GetTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// IsExpired checks if a time is in the past
func IsExpired(t *time.Time) bool {
	if t == nil {
		return false
	}
	return time.Now().After(*t)
}

// AddDuration adds duration to time with validation
func AddDuration(t time.Time, duration time.Duration) time.Time {
	return t.Add(duration)
}

// GetStartOfDay returns the start of the day for given time
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay returns the end of the day for given time
func GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// GetStartOfWeek returns the start of the week (Monday) for given time
func GetStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	days := weekday - 1
	return GetStartOfDay(t.AddDate(0, 0, -days))
}

// GetEndOfWeek returns the end of the week (Sunday) for given time
func GetEndOfWeek(t time.Time) time.Time {
	return GetEndOfDay(GetStartOfWeek(t).AddDate(0, 0, 6))
}

// GetStartOfMonth returns the start of the month for given time
func GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfMonth returns the end of the month for given time
func GetEndOfMonth(t time.Time) time.Time {
	return GetStartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// ValidateDateRange validates that start date is before end date
func ValidateDateRange(start, end time.Time) error {
	if start.After(end) {
		return fmt.Errorf("start date must be before end date")
	}
	return nil
}

// StringToInt64 converts string to int64
func StringToInt64(s string) (int64, error) {
	var result int64
	for _, digit := range s {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		result = result*10 + int64(digit-'0')
	}
	return result, nil
}
