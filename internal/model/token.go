package model

import (
	"time"

	"github.com/google/uuid"
)

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// DeviceInfo represents client device information
type DeviceInfo struct {
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	DeviceName string `json:"device_name,omitempty"`
	Location  string `json:"location,omitempty"`
}

// Session represents an active user session
type Session struct {
	ID         uuid.UUID              `json:"id"`
	CreatedAt  time.Time              `json:"created_at"`
	LastUsedAt time.Time              `json:"last_used_at"`
	DeviceInfo map[string]interface{} `json:"device_info"`
	IsCurrent  bool                   `json:"is_current"`
}