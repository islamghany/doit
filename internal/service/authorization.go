// Package service contains business logic and service layer implementations.
package service

import (
	"fmt"

	"doit/internal/model"

	"github.com/google/uuid"
)

// Authorization errors
var (
	ErrUnauthorized      = fmt.Errorf("unauthorized: resource does not belong to user")
	ErrForbidden         = fmt.Errorf("forbidden: insufficient permissions")
	ErrInvalidResourceID = fmt.Errorf("invalid resource ID")
)

// VerifyOwnership checks if a resource belongs to the requesting user
// Implements OWASP A01:2021 (Broken Access Control) prevention
func VerifyOwnership(resourceUserID, requestUserID uuid.UUID, resourceType string) error {
	if resourceUserID != requestUserID {
		return fmt.Errorf("unauthorized: %s does not belong to user", resourceType)
	}
	return nil
}

// VerifyOwnershipOrRole checks ownership or verifies user has required role
// Allows admins to access any resource
func VerifyOwnershipOrRole(resourceUserID, requestUserID uuid.UUID, userRole model.UserRole, resourceType string, allowedRoles ...model.UserRole) error {
	// Check if user has one of the allowed roles
	for _, role := range allowedRoles {
		if userRole == role {
			return nil // User has required role, allow access
		}
	}

	// User doesn't have special role, check ownership
	return VerifyOwnership(resourceUserID, requestUserID, resourceType)
}

// CanAccessResource checks if user can access a resource
// Returns true if user is owner or has admin role
func CanAccessResource(resourceUserID, requestUserID uuid.UUID, userRole model.UserRole) bool {
	// Admins can access anything
	if userRole == model.UserRoleAdmin {
		return true
	}

	// Otherwise, must be owner
	return resourceUserID == requestUserID
}

// CanModifyResource checks if user can modify a resource
// Returns true if user is owner or has admin/moderator role
func CanModifyResource(resourceUserID, requestUserID uuid.UUID, userRole model.UserRole) bool {
	// Admins and moderators can modify anything
	if userRole == model.UserRoleAdmin || userRole == model.UserRoleModerator {
		return true
	}

	// Otherwise, must be owner
	return resourceUserID == requestUserID
}

// CanDeleteResource checks if user can delete a resource
// Returns true if user is owner or has admin role
func CanDeleteResource(resourceUserID, requestUserID uuid.UUID, userRole model.UserRole) bool {
	// Only admins can delete anything
	if userRole == model.UserRoleAdmin {
		return true
	}

	// Otherwise, must be owner
	return resourceUserID == requestUserID
}

// RequireAdmin returns error if user is not an admin
func RequireAdmin(userRole model.UserRole) error {
	if userRole != model.UserRoleAdmin {
		return ErrForbidden
	}
	return nil
}

// RequireRole returns error if user doesn't have one of the required roles
func RequireRole(userRole model.UserRole, allowedRoles ...model.UserRole) error {
	for _, role := range allowedRoles {
		if userRole == role {
			return nil
		}
	}
	return ErrForbidden
}
