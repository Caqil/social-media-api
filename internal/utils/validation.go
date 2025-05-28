package utils

import (
	"regexp"
	"strings"
)

// Custom validation functions for social media platform

// IsValidUsername checks if username meets social media standards
func IsValidUsername(username string) bool {
	// Length check (3-50 characters)
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// Only alphanumeric and underscores allowed
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return false
	}

	// Must start with a letter or number (not underscore)
	if username[0] == '_' {
		return false
	}

	// Cannot end with underscore
	if username[len(username)-1] == '_' {
		return false
	}

	// No consecutive underscores
	if strings.Contains(username, "__") {
		return false
	}

	return true
}
