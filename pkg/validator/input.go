// Package validator provides input validation and sanitization functions.
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// SafeTextRegex allows alphanumeric characters and common punctuation
	SafeTextRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-\_\.\,\!\?\'\"\:\;]+$`)

	// SearchQueryRegex is stricter - for search inputs
	SearchQueryRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-\_\.]+$`)

	// UsernameRegex allows alphanumeric, underscore, hyphen
	UsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9\_\-]+$`)

	// EmailRegex is a basic email validation pattern
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// SQL injection patterns to explicitly reject
var sqlInjectionPatterns = []string{
	"--",
	"/*",
	"*/",
	"xp_",
	"sp_",
	"';",
	"\";",
	"OR 1=1",
	"OR '1'='1",
	"UNION SELECT",
	"DROP TABLE",
	"DROP DATABASE",
	"INSERT INTO",
	"UPDATE ",
	"DELETE FROM",
	"EXEC(",
	"EXECUTE(",
	"SCRIPT",
	"JAVASCRIPT",
	"ONERROR",
	"ONLOAD",
	"<script>",
	"</script>",
}

// ValidateSearchQuery validates and sanitizes search query input
// Implements OWASP A03:2021 (Injection) prevention
func ValidateSearchQuery(query string) error {
	if len(query) == 0 {
		return fmt.Errorf("search query cannot be empty")
	}

	if len(query) > 100 {
		return fmt.Errorf("search query too long (max 100 characters)")
	}

	// Check for SQL injection patterns
	upperQuery := strings.ToUpper(query)
	for _, pattern := range sqlInjectionPatterns {
		if strings.Contains(upperQuery, strings.ToUpper(pattern)) {
			return fmt.Errorf("search query contains invalid characters or patterns")
		}
	}

	// Check against regex pattern
	if !SearchQueryRegex.MatchString(query) {
		return fmt.Errorf("search query contains invalid characters")
	}

	return nil
}

// ValidateUsername validates username format
func ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 30 {
		return fmt.Errorf("username must be at most 30 characters long")
	}

	if !UsernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscores, and hyphens")
	}

	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if len(email) == 0 {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 254 {
		return fmt.Errorf("email too long (max 254 characters)")
	}

	if !EmailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// SanitizeString removes dangerous characters and limits length
func SanitizeString(input string, maxLength int) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove other control characters
	input = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1 // Drop control characters
		}
		return r
	}, input)

	// Limit length
	if len(input) > maxLength {
		input = input[:maxLength]
	}

	// Trim whitespace
	return strings.TrimSpace(input)
}

// ValidateID validates UUID format (basic check)
func ValidateID(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("invalid ID format")
	}

	// Basic UUID format check
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(id) {
		return fmt.Errorf("invalid ID format")
	}

	return nil
}

// ContainsSQLInjection checks if a string contains potential SQL injection patterns
func ContainsSQLInjection(input string) bool {
	upperInput := strings.ToUpper(input)
	for _, pattern := range sqlInjectionPatterns {
		if strings.Contains(upperInput, strings.ToUpper(pattern)) {
			return true
		}
	}
	return false
}

// ContainsXSS checks for basic XSS patterns
func ContainsXSS(input string) bool {
	xssPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"<iframe",
		"<object",
		"<embed",
		"<img",
		"vbscript:",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}
