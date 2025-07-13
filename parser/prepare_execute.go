package parser

import (
	"fmt"
	"strings"
	"sync"
)

// SimplePreparedStatement represents a prepared statement for SQL PREPARE/EXECUTE
type SimplePreparedStatement struct {
	Name       string
	Query      string   // The original query with $1, $2, etc.
	ParamTypes []string // Parameter type names from PREPARE statement
}

// Global store for SQL PREPARE statements
var (
	simplePreparedStatements = make(map[string]*SimplePreparedStatement)
	simplePreparedMutex      sync.RWMutex
)

// StoreSimplePreparedStatement stores a prepared statement from PREPARE command
func StoreSimplePreparedStatement(name, query string, paramTypes []string) {
	simplePreparedMutex.Lock()
	defer simplePreparedMutex.Unlock()
	
	simplePreparedStatements[name] = &SimplePreparedStatement{
		Name:       name,
		Query:      query,
		ParamTypes: paramTypes,
	}
}

// GetSimplePreparedStatement retrieves a prepared statement by name
func GetSimplePreparedStatement(name string) (*SimplePreparedStatement, error) {
	simplePreparedMutex.RLock()
	defer simplePreparedMutex.RUnlock()
	
	stmt, exists := simplePreparedStatements[name]
	if !exists {
		return nil, fmt.Errorf("prepared statement \"%s\" does not exist", name)
	}
	
	return stmt, nil
}

// DropSimplePreparedStatement removes a prepared statement
func DropSimplePreparedStatement(name string) {
	simplePreparedMutex.Lock()
	defer simplePreparedMutex.Unlock()
	delete(simplePreparedStatements, name)
}

// ClearAllSimplePreparedStatements removes all prepared statements
func ClearAllSimplePreparedStatements() {
	simplePreparedMutex.Lock()
	defer simplePreparedMutex.Unlock()
	simplePreparedStatements = make(map[string]*SimplePreparedStatement)
}

// SubstituteParameters replaces $1, $2, etc. with actual parameter values
func SubstituteParameters(query string, paramValues []interface{}) string {
	result := query
	
	for i, value := range paramValues {
		placeholder := fmt.Sprintf("$%d", i+1)
		var replacement string
		
		if value == nil {
			replacement = "NULL"
		} else {
			switch v := value.(type) {
			case int, int32, int64, float32, float64:
				replacement = fmt.Sprintf("%v", v)
			case bool:
				replacement = fmt.Sprintf("%v", v)
			default:
				// String and other types - escape quotes
				strVal := fmt.Sprintf("%v", v)
				escaped := strings.ReplaceAll(strVal, "'", "''")
				replacement = fmt.Sprintf("'%s'", escaped)
			}
		}
		
		// Replace all occurrences of the placeholder
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	
	return result
}