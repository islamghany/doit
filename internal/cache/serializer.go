package cache

import (
	"encoding/json"
	"fmt"
)

// Serializer defines methods for serializing and deserializing values
type Serializer interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

// JSONSerializer implements Serializer using JSON encoding
type JSONSerializer struct{}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

func (s *JSONSerializer) Marshal(v any) ([]byte, error) {
	// Handle already-serialized types
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	case int, int8, int16, int32, int64:
		return []byte(fmt.Sprintf("%d", val)), nil
	case uint, uint8, uint16, uint32, uint64:
		return []byte(fmt.Sprintf("%d", val)), nil
	case float32, float64:
		return []byte(fmt.Sprintf("%f", val)), nil
	case bool:
		return []byte(fmt.Sprintf("%t", val)), nil
	}

	// Marshal complex types to JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSerialization, err)
	}
	return data, nil
}

func (s *JSONSerializer) Unmarshal(data []byte, v any) error {
	// If target is []byte or string, return as-is
	switch target := v.(type) {
	case *[]byte:
		*target = data
		return nil
	case *string:
		*target = string(data)
		return nil
	}

	// Unmarshal JSON for complex types
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("%w: %v", ErrSerialization, err)
	}
	return nil
}

// NoOpSerializer implements Serializer without any transformation
// Use this when you want to store []byte directly
type NoOpSerializer struct{}

func NewNoOpSerializer() *NoOpSerializer {
	return &NoOpSerializer{}
}

func (s *NoOpSerializer) Marshal(v any) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		return nil, fmt.Errorf("%w: NoOpSerializer only accepts []byte or string", ErrSerialization)
	}
}

func (s *NoOpSerializer) Unmarshal(data []byte, v any) error {
	switch target := v.(type) {
	case *[]byte:
		*target = data
		return nil
	case *string:
		*target = string(data)
		return nil
	default:
		return fmt.Errorf("%w: NoOpSerializer only supports []byte or string", ErrSerialization)
	}
}
