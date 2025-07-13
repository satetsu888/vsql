package parser

import (
	"fmt"
	"sync"
	
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// SQLPreparedStatement represents a prepared statement created via SQL PREPARE command
type SQLPreparedStatement struct {
	Name       string
	QueryNode  *pg_query.Node // The parsed query node from PREPARE
	ParamTypes []string // Parameter type names from PREPARE statement
}

// Global store for SQL-level prepared statements
// This is separate from the Extended Protocol prepared statements
var (
	sqlPreparedStatements = make(map[string]*SQLPreparedStatement)
	sqlPreparedMutex      sync.RWMutex
)

// StoreSQLPreparedStatement stores a prepared statement from PREPARE command
func StoreSQLPreparedStatement(stmt *SQLPreparedStatement) {
	sqlPreparedMutex.Lock()
	defer sqlPreparedMutex.Unlock()
	sqlPreparedStatements[stmt.Name] = stmt
}

// GetSQLPreparedStatement retrieves a prepared statement by name
func GetSQLPreparedStatement(name string) (*SQLPreparedStatement, error) {
	sqlPreparedMutex.RLock()
	defer sqlPreparedMutex.RUnlock()
	
	stmt, exists := sqlPreparedStatements[name]
	if !exists {
		return nil, fmt.Errorf("prepared statement \"%s\" does not exist", name)
	}
	
	return stmt, nil
}

// DropSQLPreparedStatement removes a prepared statement
func DropSQLPreparedStatement(name string) {
	sqlPreparedMutex.Lock()
	defer sqlPreparedMutex.Unlock()
	delete(sqlPreparedStatements, name)
}

// ClearAllSQLPreparedStatements removes all prepared statements
func ClearAllSQLPreparedStatements() {
	sqlPreparedMutex.Lock()
	defer sqlPreparedMutex.Unlock()
	sqlPreparedStatements = make(map[string]*SQLPreparedStatement)
}