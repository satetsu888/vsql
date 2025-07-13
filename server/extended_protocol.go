package server

import (
	"fmt"
	"sync"

	"github.com/satetsu888/vsql/storage"
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
	OIDUnknown   = 0
	OIDBool      = 16
	OIDInt8      = 20
	OIDInt2      = 21
	OIDInt4      = 23
	OIDText      = 25
	OIDFloat4    = 700
	OIDFloat8    = 701
	OIDVarchar   = 1043
	OIDTimestamp = 1114
	OIDDate      = 1082
	OIDTime      = 1083
)

// VSQLTypeToOID converts VSQL column types to PostgreSQL OIDs
func VSQLTypeToOID(colType storage.ColumnType) int32 {
	switch colType {
	case storage.TypeBoolean:
		return OIDBool
	case storage.TypeInteger:
		return OIDInt4
	case storage.TypeFloat:
		return OIDFloat8
	case storage.TypeString:
		return OIDText
	case storage.TypeTimestamp:
		return OIDTimestamp
	case storage.TypeUnknown:
		// Return text as a safe default for unknown types
		// This allows clients to work with the data even if type isn't determined yet
		return OIDText
	default:
		return OIDText
	}
}

// GetTypeSizeAndMod returns the size and type modifier for a PostgreSQL OID
func GetTypeSizeAndMod(oid int32) (size int16, mod int32) {
	switch oid {
	case OIDBool:
		return 1, -1
	case OIDInt2:
		return 2, -1
	case OIDInt4:
		return 4, -1
	case OIDInt8:
		return 8, -1
	case OIDFloat4:
		return 4, -1
	case OIDFloat8:
		return 8, -1
	case OIDText, OIDVarchar:
		return -1, -1  // Variable length
	case OIDTimestamp, OIDDate, OIDTime:
		return 8, -1
	default:
		return -1, -1
	}
}