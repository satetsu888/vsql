package server

import (
	"fmt"
	"sync"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// PreparedStatement represents a parsed SQL statement with parameter placeholders
type PreparedStatement struct {
	Name         string
	Query        string
	ParsedQuery  *pg_query.ParseResult
	ParamTypes   []int32 // PostgreSQL OIDs for parameter types
	ColumnNames  []string
	ColumnTypes  []int32 // PostgreSQL OIDs for column types
}

// Portal represents a bound prepared statement ready for execution
type Portal struct {
	Name              string
	Statement         *PreparedStatement
	ParameterValues   [][]byte
	ParameterFormats  []int16 // 0 = text, 1 = binary
	ResultFormats     []int16 // 0 = text, 1 = binary
}

// ExtendedProtocolState manages prepared statements and portals for a connection
type ExtendedProtocolState struct {
	mu                sync.RWMutex
	preparedStatements map[string]*PreparedStatement
	portals           map[string]*Portal
}

// NewExtendedProtocolState creates a new state manager
func NewExtendedProtocolState() *ExtendedProtocolState {
	return &ExtendedProtocolState{
		preparedStatements: make(map[string]*PreparedStatement),
		portals:           make(map[string]*Portal),
	}
}

// StorePreparedStatement stores a prepared statement
func (eps *ExtendedProtocolState) StorePreparedStatement(stmt *PreparedStatement) {
	eps.mu.Lock()
	defer eps.mu.Unlock()
	
	// Empty name means unnamed statement
	if stmt.Name == "" {
		// Clear previous unnamed statement
		delete(eps.preparedStatements, "")
	}
	
	eps.preparedStatements[stmt.Name] = stmt
}

// GetPreparedStatement retrieves a prepared statement by name
func (eps *ExtendedProtocolState) GetPreparedStatement(name string) (*PreparedStatement, error) {
	eps.mu.RLock()
	defer eps.mu.RUnlock()
	
	stmt, exists := eps.preparedStatements[name]
	if !exists {
		return nil, fmt.Errorf("prepared statement %q does not exist", name)
	}
	
	return stmt, nil
}

// StorePortal stores a portal
func (eps *ExtendedProtocolState) StorePortal(portal *Portal) {
	eps.mu.Lock()
	defer eps.mu.Unlock()
	
	// Empty name means unnamed portal
	if portal.Name == "" {
		// Clear previous unnamed portal
		delete(eps.portals, "")
	}
	
	eps.portals[portal.Name] = portal
}

// GetPortal retrieves a portal by name
func (eps *ExtendedProtocolState) GetPortal(name string) (*Portal, error) {
	eps.mu.RLock()
	defer eps.mu.RUnlock()
	
	portal, exists := eps.portals[name]
	if !exists {
		return nil, fmt.Errorf("portal %q does not exist", name)
	}
	
	return portal, nil
}

// ClosePreparedStatement removes a prepared statement
func (eps *ExtendedProtocolState) ClosePreparedStatement(name string) error {
	eps.mu.Lock()
	defer eps.mu.Unlock()
	
	if _, exists := eps.preparedStatements[name]; !exists {
		return fmt.Errorf("prepared statement %q does not exist", name)
	}
	
	delete(eps.preparedStatements, name)
	return nil
}

// ClosePortal removes a portal
func (eps *ExtendedProtocolState) ClosePortal(name string) error {
	eps.mu.Lock()
	defer eps.mu.Unlock()
	
	if _, exists := eps.portals[name]; !exists {
		return fmt.Errorf("portal %q does not exist", name)
	}
	
	delete(eps.portals, name)
	return nil
}

// Clear removes all prepared statements and portals
func (eps *ExtendedProtocolState) Clear() {
	eps.mu.Lock()
	defer eps.mu.Unlock()
	
	eps.preparedStatements = make(map[string]*PreparedStatement)
	eps.portals = make(map[string]*Portal)
}

// PostgreSQL type OIDs
const (
	OIDUnknown = 0
	OIDBool    = 16
	OIDInt8    = 20
	OIDInt2    = 21
	OIDInt4    = 23
	OIDText    = 25
	OIDFloat4  = 700
	OIDFloat8  = 701
	OIDVarchar = 1043
)