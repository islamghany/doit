// Package validator provides input validation and sanitization functions.
package validator

import (
	"fmt"
	"unicode"
)

// PasswordStrength defines the rules for password validation
type PasswordStrength struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

// DefaultPasswordStrength provides secure default password requirements
// Based on NIST SP 800-63B guidelines
var DefaultPasswordStrength = PasswordStrength{
	MinLength:        12, // NIST recommends minimum 8, but 12 is better
	RequireUppercase: true,
	RequireLowercase: true,
	RequireNumber:    true,
	RequireSpecial:   true,
}

// ValidatePasswordStrength checks if a password meets the specified strength requirements
// Implements OWASP A02:2021 (Cryptographic Failures) and A07:2021 (Authentication Failures)
func ValidatePasswordStrength(password string, rules PasswordStrength) error {
	// Check minimum length
	if len(password) < rules.MinLength {
		return fmt.Errorf("password must be at least %d characters long", rules.MinLength)
	}

	// Check maximum length (prevent DoS via extremely long passwords)
	if len(password) > 128 {
		return fmt.Errorf("password must be at most 128 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	// Check character requirements
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if rules.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if rules.RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if rules.RequireNumber && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if rules.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character (!@#$%%^&*)")
	}

	// Check against common passwords
	if IsCommonPassword(password) {
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	return nil
}

// commonPasswords is a list of commonly used passwords that should be rejected
// In production, use a more comprehensive list like the one from:
// https://github.com/danielmiessler/SecLists/tree/master/Passwords
var commonPasswords = map[string]bool{
	"password":       true,
	"Password":       true,
	"Password1":      true,
	"Password12":     true,
	"Password123":    true,
	"Password123!":   true,
	"Admin123":       true,
	"Admin123!":      true,
	"Welcome123":     true,
	"Welcome123!":    true,
	"Qwerty123":      true,
	"Qwerty123!":     true,
	"P@ssw0rd":       true,
	"P@ssword":       true,
	"P@ssword1":      true,
	"P@ssword123":    true,
	"Password1!":     true,
	"Password12!":    true,
	"MyPassword123!": true,
	"Test123!":       true,
	"Testing123!":    true,
	"User123!":       true,
	"User1234!":      true,
	"Change123!":     true,
	"Changeme123!":   true,
	"Letmein123!":    true,
	"Welcome1":       true,
	"Welcome1!":      true,
}

// IsCommonPassword checks if the password is in the list of common passwords
func IsCommonPassword(password string) bool {
	return commonPasswords[password]
}

// ValidatePasswordWithDefaults validates password using default strength requirements
func ValidatePasswordWithDefaults(password string) error {
	return ValidatePasswordStrength(password, DefaultPasswordStrength)
}
