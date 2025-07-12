package storage

import (
	"fmt"
	"time"
)

// ColumnType represents the data type of a column
type ColumnType int

const (
	TypeUnknown ColumnType = iota // Type not yet determined (only NULLs seen)
	TypeInteger                   // Integer type
	TypeFloat                     // Floating point type
	TypeString                    // String type
	TypeBoolean                   // Boolean type
)

// ColumnTypeInfo stores type information for a column
type ColumnTypeInfo struct {
	CurrentType    ColumnType
	IsConfirmed    bool      // Whether type has been confirmed by non-NULL value
	LastUpdateTime time.Time
}

// TypeMismatchError represents a type mismatch error
type TypeMismatchError struct {
	Table    string
	Column   string
	Expected ColumnType
	Actual   ColumnType
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch for column %s.%s: expected %s, got %s",
		e.Table, e.Column, TypeToString(e.Expected), TypeToString(e.Actual))
}

// TypeToString converts a ColumnType to its string representation
func TypeToString(t ColumnType) string {
	switch t {
	case TypeUnknown:
		return "unknown"
	case TypeInteger:
		return "integer"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeBoolean:
		return "boolean"
	default:
		return "invalid"
	}
}

// IsTypeCompatible checks if a type change is allowed
func IsTypeCompatible(currentType, newType ColumnType) bool {
	if currentType == TypeUnknown {
		return true // Unknown can become any type
	}
	if currentType == newType {
		return true // Same type is always compatible
	}
	if currentType == TypeInteger && newType == TypeFloat {
		return true // Integer can be promoted to Float
	}
	return false
}

// InferTypeFromValue infers the ColumnType from a value
func InferTypeFromValue(value interface{}) ColumnType {
	switch value.(type) {
	case nil:
		return TypeUnknown
	case bool:
		return TypeBoolean
	case int, int32, int64, uint, uint32, uint64:
		return TypeInteger
	case float32, float64:
		return TypeFloat
	case string:
		// Treat strings as strings (no automatic conversion to numbers)
		return TypeString
	default:
		// Default to string for unknown types
		return TypeString
	}
}