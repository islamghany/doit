package web

import (
	"doit/pkg/validator"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

const (
	MAX_BODY_SIZE = int64(1024 * 1024 * 2) // 2MB
)

// Decode decodes the request body into dst and validates it.
func Decode[T any](w http.ResponseWriter, r *http.Request, dst T) error {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_BODY_SIZE)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decoding request: %w", err)
	}
	// Validate the request body
	return validator.Check(dst)
}

func GetParamNumber(r *http.Request, key string) (int, error) {
	param := r.PathValue(key)
	if param == "" {
		return 0, fmt.Errorf("missing parameter %s", key)
	}
	return strconv.Atoi(param)
}

func GetParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := r.PathValue(key)
	if param == "" {
		return uuid.UUID{}, fmt.Errorf("missing parameter %s", key)
	}
	uid, err := uuid.Parse(param)
	if err != nil {
		return uuid.UUID{}, errors.New("url parameter id is not a valid uuid")
	}
	return uid, nil
}

func GetParamString(r *http.Request, key string) (string, error) {
	param := r.PathValue(key)
	if param == "" {
		return "", fmt.Errorf("missing parameter %s", key)
	}
	return param, nil
}

func GetQueryString(r *http.Request, key string, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetQueryInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return num
}

func GetHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

// Helper: Extract client IP address
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fallback to RemoteAddr
	return r.RemoteAddr
}

func GetUserAgent(r *http.Request) string {
	return r.UserAgent()
}

func GetDeviceName(r *http.Request) string {
	return r.Header.Get("X-Device-Name")
}

func GetLocation(r *http.Request) string {
	return r.Header.Get("X-Location")
}