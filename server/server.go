package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"sync"

	pg_query "github.com/pganalyze/pg_query_go/v5"
	"github.com/satetsu888/vsql/parser"
	"github.com/satetsu888/vsql/storage"
)

type Server struct {
	port      int
	dataStore *storage.DataStore
	metaStore *storage.MetaStore
	listener  net.Listener
	wg        sync.WaitGroup
}

func New(port int, dataStore *storage.DataStore, metaStore *storage.MetaStore) *Server {
	return &Server{
		port:      port,
		dataStore: dataStore,
		metaStore: metaStore,
	}
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			return err
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	defer writer.Flush()

	// Create extended protocol state for this connection
	extState := NewExtendedProtocolState()

	if err := s.handleStartup(reader, writer); err != nil {
		fmt.Printf("Startup error: %v\n", err)
		return
	}

	for {
		msg, err := ReadMessage(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			return
		}

		switch msg.Type {
		case Query:
			query := string(bytes.TrimSuffix(msg.Data, []byte{0}))
			if err := s.handleQuery(writer, query); err != nil {
				WriteErrorResponse(writer, err.Error())
			}
			WriteReadyForQuery(writer)
			writer.Flush()
		case Parse:
			if err := s.handleParse(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			} else {
				WriteParseComplete(writer)
			}
			writer.Flush()
		case Bind:
			if err := s.handleBind(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			} else {
				WriteBindComplete(writer)
			}
			writer.Flush()
		case Execute:
			if err := s.handleExecute(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			}
			writer.Flush()
		case Describe:
			if err := s.handleDescribe(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			}
			writer.Flush()
		case Close:
			if err := s.handleClose(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			}
			writer.Flush()
		case Sync:
			// Sync completes the current extended query protocol sequence
			WriteReadyForQuery(writer)
			writer.Flush()
		case Flush:
			// Flush forces any pending output to be sent
			writer.Flush()
		case 'X': // Terminate message
			// Client is closing connection
			return
		default:
			// Send error response for unsupported message types
			WriteErrorResponse(writer, fmt.Sprintf("unsupported message type: %c", msg.Type))
			WriteReadyForQuery(writer)
			writer.Flush()
		}
	}
}

func (s *Server) handleStartup(reader io.Reader, writer *bufio.Writer) error {
	startupMsg := make([]byte, 8)
	if _, err := io.ReadFull(reader, startupMsg); err != nil {
		return err
	}

	length := binary.BigEndian.Uint32(startupMsg[:4])
	version := binary.BigEndian.Uint32(startupMsg[4:])

	params := make([]byte, length-8)
	if _, err := io.ReadFull(reader, params); err != nil {
		return err
	}

	if version == 80877103 {
		writer.WriteByte('N')
		writer.Flush()
		return s.handleStartup(reader, writer)
	}

	if err := WriteAuthenticationOk(writer); err != nil {
		return err
	}

	WriteParameterStatus(writer, "server_version", "12.0")
	WriteParameterStatus(writer, "server_encoding", "UTF8")
	WriteParameterStatus(writer, "client_encoding", "UTF8")
	WriteParameterStatus(writer, "DateStyle", "ISO, MDY")
	WriteBackendKeyData(writer, 12345, 67890)

	if err := WriteReadyForQuery(writer); err != nil {
		return err
	}

	return writer.Flush()
}

func (s *Server) handleQuery(w *bufio.Writer, query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	columns, rows, tag, err := parser.ExecutePgQuery(query, s.dataStore, s.metaStore)
	if err != nil {
		return err
	}

	if columns != nil {
		if err := WriteRowDescription(w, columns); err != nil {
			return err
		}

		for _, row := range rows {
			if err := WriteDataRow(w, row); err != nil {
				return err
			}
		}
	}

	return WriteCommandComplete(w, tag)
}

// handleParse handles the Parse message (P)
func (s *Server) handleParse(data []byte, extState *ExtendedProtocolState, w *bufio.Writer) error {
	buf := bytes.NewReader(data)
	
	// Read statement name
	statementName, err := readCString(buf)
	if err != nil {
		return fmt.Errorf("failed to read statement name: %v", err)
	}
	
	// Read query string
	query, err := readCString(buf)
	if err != nil {
		return fmt.Errorf("failed to read query: %v", err)
	}
	
	// Read number of parameter data types
	var numParams int16
	if err := binary.Read(buf, binary.BigEndian, &numParams); err != nil {
		return fmt.Errorf("failed to read parameter count: %v", err)
	}
	
	// Read parameter data types
	paramTypes := make([]int32, numParams)
	for i := int16(0); i < numParams; i++ {
		if err := binary.Read(buf, binary.BigEndian, &paramTypes[i]); err != nil {
			return fmt.Errorf("failed to read parameter type: %v", err)
		}
	}
	
	// Parse the query
	parsedQuery, err := parser.ParsePostgreSQL(query)
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}
	
	// If client didn't specify parameter types, analyze the query to determine them
	actualParamTypes := paramTypes
	if len(paramTypes) == 0 {
		// Count parameters in the query
		paramCount := countParameters(query)
		if paramCount > 0 {
			actualParamTypes = make([]int32, paramCount)
			// Default to unknown type (0) - PostgreSQL will infer types later
			for i := range actualParamTypes {
				actualParamTypes[i] = 0
			}
			// Try to infer types from query context
			actualParamTypes = inferParameterTypes(query, parsedQuery)
		}
	}
	
	// Create and store the prepared statement
	stmt := &PreparedStatement{
		Name:        statementName,
		Query:       query,
		ParsedQuery: parsedQuery,
		ParamTypes:  actualParamTypes,
	}
	
	extState.StorePreparedStatement(stmt)
	
	return nil
}

// handleBind handles the Bind message (B)
func (s *Server) handleBind(data []byte, extState *ExtendedProtocolState, w *bufio.Writer) error {
	buf := bytes.NewReader(data)
	
	// Read portal name
	portalName, err := readCString(buf)
	if err != nil {
		return fmt.Errorf("failed to read portal name: %v", err)
	}
	
	// Read statement name
	statementName, err := readCString(buf)
	if err != nil {
		return fmt.Errorf("failed to read statement name: %v", err)
	}
	
	// Get the prepared statement
	stmt, err := extState.GetPreparedStatement(statementName)
	if err != nil {
		return err
	}
	
	// Read parameter format codes
	var numParamFormats int16
	if err := binary.Read(buf, binary.BigEndian, &numParamFormats); err != nil {
		return fmt.Errorf("failed to read parameter format count: %v", err)
	}
	
	paramFormats := make([]int16, numParamFormats)
	for i := int16(0); i < numParamFormats; i++ {
		if err := binary.Read(buf, binary.BigEndian, &paramFormats[i]); err != nil {
			return fmt.Errorf("failed to read parameter format: %v", err)
		}
	}
	
	// Read parameter values
	var numParams int16
	if err := binary.Read(buf, binary.BigEndian, &numParams); err != nil {
		return fmt.Errorf("failed to read parameter value count: %v", err)
	}
	
	paramValues := make([][]byte, numParams)
	for i := int16(0); i < numParams; i++ {
		var length int32
		if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
			return fmt.Errorf("failed to read parameter length: %v", err)
		}
		
		if length == -1 {
			// NULL value
			paramValues[i] = nil
		} else {
			paramValues[i] = make([]byte, length)
			if _, err := io.ReadFull(buf, paramValues[i]); err != nil {
				return fmt.Errorf("failed to read parameter value: %v", err)
			}
		}
	}
	
	// Read result format codes
	var numResultFormats int16
	if err := binary.Read(buf, binary.BigEndian, &numResultFormats); err != nil {
		return fmt.Errorf("failed to read result format count: %v", err)
	}
	
	resultFormats := make([]int16, numResultFormats)
	for i := int16(0); i < numResultFormats; i++ {
		if err := binary.Read(buf, binary.BigEndian, &resultFormats[i]); err != nil {
			return fmt.Errorf("failed to read result format: %v", err)
		}
	}
	
	// Create and store the portal
	portal := &Portal{
		Name:              portalName,
		Statement:         stmt,
		ParameterValues:   paramValues,
		ParameterFormats:  paramFormats,
		ResultFormats:     resultFormats,
	}
	
	extState.StorePortal(portal)
	
	return nil
}

// handleExecute handles the Execute message (E)
func (s *Server) handleExecute(data []byte, extState *ExtendedProtocolState, w *bufio.Writer) error {
	buf := bytes.NewReader(data)
	
	// Read portal name
	portalName, err := readCString(buf)
	if err != nil {
		return fmt.Errorf("failed to read portal name: %v", err)
	}
	
	// Read maximum number of rows to return (0 = unlimited)
	var maxRows int32
	if err := binary.Read(buf, binary.BigEndian, &maxRows); err != nil {
		return fmt.Errorf("failed to read max rows: %v", err)
	}
	
	// Get the portal
	portal, err := extState.GetPortal(portalName)
	if err != nil {
		return err
	}
	
	// Execute the query with bound parameters
	columns, rows, tag, err := s.executePortal(portal)
	if err != nil {
		return err
	}
	
	// Send row description if this is a SELECT-like query
	if columns != nil {
		// Analyze the query to get proper column descriptions with types
		var colDescs []ColumnDescription
		if portal.Statement.ParsedQuery != nil && len(portal.Statement.ParsedQuery.Stmts) > 0 {
			if stmt := portal.Statement.ParsedQuery.Stmts[0].Stmt; stmt != nil {
				if selectStmt, ok := stmt.Node.(*pg_query.Node_SelectStmt); ok {
					colDescs, err = s.analyzeSelectColumns(selectStmt.SelectStmt)
					if err != nil {
						return err
					}
				}
			}
		}
		
		// If we couldn't analyze the query, fall back to simple column names
		if len(colDescs) == 0 {
			for _, col := range columns {
				colDescs = append(colDescs, ColumnDescription{
					Name:      col,
					TableOID:  0,
					ColumnNum: 0,
					TypeOID:   OIDText,
					TypeSize:  -1,
					TypeMod:   -1,
					Format:    0,
				})
			}
		}
		
		if err := WriteRowDescriptionExt(w, colDescs); err != nil {
			return err
		}
		
		// Send data rows (respecting maxRows if specified)
		rowCount := len(rows)
		if maxRows > 0 && int32(rowCount) > maxRows {
			rowCount = int(maxRows)
		}
		
		for i := 0; i < rowCount; i++ {
			if err := WriteDataRow(w, rows[i]); err != nil {
				return err
			}
		}
		
		// If we hit the row limit, send PortalSuspended instead of CommandComplete
		if maxRows > 0 && int32(len(rows)) > maxRows {
			return WritePortalSuspended(w)
		}
	}
	
	return WriteCommandComplete(w, tag)
}

// handleDescribe handles the Describe message (D)
func (s *Server) handleDescribe(data []byte, extState *ExtendedProtocolState, w *bufio.Writer) error {
	buf := bytes.NewReader(data)
	
	// Read type ('S' for statement, 'P' for portal)
	var describeType byte
	if err := binary.Read(buf, binary.BigEndian, &describeType); err != nil {
		return fmt.Errorf("failed to read describe type: %v", err)
	}
	
	// Read name
	name, err := readCString(buf)
	if err != nil {
		return fmt.Errorf("failed to read name: %v", err)
	}
	
	switch describeType {
	case 'S': // Describe statement
		stmt, err := extState.GetPreparedStatement(name)
		if err != nil {
			return err
		}
		
		// Send parameter description
		if err := WriteParameterDescription(w, stmt.ParamTypes); err != nil {
			return err
		}
		
		// Also need to send row description or no data
		if stmt.ParsedQuery != nil && len(stmt.ParsedQuery.Stmts) > 0 {
			if stmtNode := stmt.ParsedQuery.Stmts[0].Stmt; stmtNode != nil {
				switch node := stmtNode.Node.(type) {
				case *pg_query.Node_SelectStmt:
					// Analyze the SELECT query to get column descriptions
					colDescs, err := s.analyzeSelectColumns(node.SelectStmt)
					if err != nil {
						return err
					}
					return WriteRowDescriptionExt(w, colDescs)
				case *pg_query.Node_InsertStmt:
					if node.InsertStmt.ReturningList != nil {
						// INSERT ... RETURNING has result columns
						return WriteRowDescription(w, []string{"result"})
					}
					return WriteNoData(w)
				case *pg_query.Node_UpdateStmt:
					if node.UpdateStmt.ReturningList != nil {
						// UPDATE ... RETURNING has result columns
						return WriteRowDescription(w, []string{"result"})
					}
					return WriteNoData(w)
				case *pg_query.Node_DeleteStmt:
					if node.DeleteStmt.ReturningList != nil {
						// DELETE ... RETURNING has result columns
						return WriteRowDescription(w, []string{"result"})
					}
					return WriteNoData(w)
				default:
					// Other statement types have no data
					return WriteNoData(w)
				}
			}
		}
		
		return WriteNoData(w)
		
	case 'P': // Describe portal
		portal, err := extState.GetPortal(name)
		if err != nil {
			return err
		}
		
		// For portals, we need to describe the result columns
		// This is a simplified implementation - in a real system we'd analyze the query
		if portal.Statement.ParsedQuery != nil && len(portal.Statement.ParsedQuery.Stmts) > 0 {
			// Check if it's a SELECT-like query
			if stmt := portal.Statement.ParsedQuery.Stmts[0].Stmt; stmt != nil {
				switch stmt.Node.(type) {
				case *pg_query.Node_SelectStmt:
					// Analyze the SELECT query to get column descriptions
					if selectStmt, ok := stmt.Node.(*pg_query.Node_SelectStmt); ok {
						colDescs, err := s.analyzeSelectColumns(selectStmt.SelectStmt)
						if err != nil {
							return err
						}
						return WriteRowDescriptionExt(w, colDescs)
					}
					// Fallback to generic description
					return WriteRowDescription(w, []string{"result"})
				default:
					// Non-SELECT queries have no data
					return WriteNoData(w)
				}
			}
		}
		
		return WriteNoData(w)
		
	default:
		return fmt.Errorf("invalid describe type: %c", describeType)
	}
}

// handleClose handles the Close message (C)
func (s *Server) handleClose(data []byte, extState *ExtendedProtocolState, w *bufio.Writer) error {
	buf := bytes.NewReader(data)
	
	// Read type ('S' for statement, 'P' for portal)
	var closeType byte
	if err := binary.Read(buf, binary.BigEndian, &closeType); err != nil {
		return fmt.Errorf("failed to read close type: %v", err)
	}
	
	// Read name
	name, err := readCString(buf)
	if err != nil {
		return fmt.Errorf("failed to read name: %v", err)
	}
	
	switch closeType {
	case 'S': // Close statement
		if err := extState.ClosePreparedStatement(name); err != nil {
			return err
		}
	case 'P': // Close portal
		if err := extState.ClosePortal(name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid close type: %c", closeType)
	}
	
	// Send CloseComplete
	return WriteMessage(w, CloseComplete, []byte{})
}

// executePortal executes a portal with bound parameters
func (s *Server) executePortal(portal *Portal) ([]string, [][]interface{}, string, error) {
	// Replace parameters in the query
	query := portal.Statement.Query
	
	// Create a map to track replacements to handle type casts
	replacements := make(map[string]string)
	
	// Simple parameter substitution for $1, $2, etc.
	// This is a basic implementation - a full implementation would handle this in the parser
	for i, paramValue := range portal.ParameterValues {
		placeholder := fmt.Sprintf("$%d", i+1)
		var value string
		if paramValue == nil {
			value = "NULL"
		} else {
			// Check the parameter type to determine formatting
			paramType := int32(0) // Default to unknown
			if i < len(portal.Statement.ParamTypes) {
				paramType = portal.Statement.ParamTypes[i]
			}
			
			// Check the parameter format (text vs binary)
			paramFormat := int16(0) // Default to text
			if i < len(portal.ParameterFormats) {
				paramFormat = portal.ParameterFormats[i]
			}
			
			// Format the value based on type
			switch paramType {
			case OIDInt2, OIDInt4, OIDInt8:
				// Numeric types - don't quote
				if paramFormat == 1 {
					// Binary format - decode as integer
					if len(paramValue) == 8 {
						// int64 in big-endian
						intVal := int64(0)
						for j := 0; j < 8; j++ {
							intVal = (intVal << 8) | int64(paramValue[j])
						}
						value = fmt.Sprintf("%d", intVal)
					} else if len(paramValue) == 4 {
						// int32 in big-endian
						intVal := int32(0)
						for j := 0; j < 4; j++ {
							intVal = (intVal << 8) | int32(paramValue[j])
						}
						value = fmt.Sprintf("%d", intVal)
					} else if len(paramValue) == 2 {
						// int16 in big-endian
						intVal := int16(0)
						for j := 0; j < 2; j++ {
							intVal = (intVal << 8) | int16(paramValue[j])
						}
						value = fmt.Sprintf("%d", intVal)
					} else {
						// Fallback to text
						value = string(paramValue)
					}
				} else {
					// Text format
					value = string(paramValue)
				}
			case OIDFloat4, OIDFloat8:
				// Float types - don't quote
				value = string(paramValue)
			case OIDBool:
				// Boolean type - don't quote
				value = string(paramValue)
			default:
				// Text and other types - quote them
				escaped := strings.ReplaceAll(string(paramValue), "'", "''")
				value = fmt.Sprintf("'%s'", escaped)
			}
		}
		replacements[placeholder] = value
	}
	
	// Replace parameters, handling type casts like $1::int
	for placeholder, value := range replacements {
		// Replace $N::type patterns
		query = strings.ReplaceAll(query, placeholder+"::int", value+"::int")
		query = strings.ReplaceAll(query, placeholder+"::text", value+"::text")
		query = strings.ReplaceAll(query, placeholder+"::float", value+"::float")
		query = strings.ReplaceAll(query, placeholder+"::boolean", value+"::boolean")
		// Replace plain $N
		query = strings.ReplaceAll(query, placeholder, value)
	}
	
	// Execute the substituted query
	return parser.ExecutePgQuery(query, s.dataStore, s.metaStore)
}

// readCString reads a null-terminated string from the buffer
func readCString(buf *bytes.Reader) (string, error) {
	var result []byte
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0 {
			break
		}
		result = append(result, b)
	}
	return string(result), nil
}

// analyzeSelectColumns analyzes a SELECT statement and returns column descriptions
func (s *Server) analyzeSelectColumns(stmt *pg_query.SelectStmt) ([]ColumnDescription, error) {
	var colDescs []ColumnDescription
	
	// Extract table name from FROM clause
	var tableName string
	if len(stmt.FromClause) > 0 {
		tableName = extractTableNameFromNode(stmt.FromClause[0])
		// Handle quoted names
		tableName = strings.Trim(tableName, `"`)
	}
	
	// Process target list
	colNum := int16(1)
	for _, target := range stmt.TargetList {
		if resTarget, ok := target.Node.(*pg_query.Node_ResTarget); ok {
			// Get column name
			var colName string
			if resTarget.ResTarget.Name != "" {
				// Alias is provided
				colName = resTarget.ResTarget.Name
			} else if colRef, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
				// Column reference
				colName = extractColumnName(colRef.ColumnRef)
			} else if funcCall, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_FuncCall); ok {
				// Function call
				colName = extractFunctionName(funcCall.FuncCall)
			} else if _, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_AConst); ok {
				// Constant
				colName = "?column?"
			} else {
				// Default name
				colName = "?column?"
			}
			
			// Get column type
			var typeOID int32 = OIDText  // Default to text
			var typeSize int16 = -1
			var typeMod int32 = -1
			
			// Try to determine type from metadata if it's a column reference
			if tableName != "" && resTarget.ResTarget.Val != nil {
				if colRef, ok := resTarget.ResTarget.Val.Node.(*pg_query.Node_ColumnRef); ok {
					actualColName := extractColumnName(colRef.ColumnRef)
					colType := s.metaStore.GetColumnType(tableName, actualColName)
					// Always use the actual type, even if unknown
					typeOID = VSQLTypeToOID(colType)
					typeSize, typeMod = GetTypeSizeAndMod(typeOID)
				}
			}
			
			colDesc := ColumnDescription{
				Name:      colName,
				TableOID:  0,  // We don't track table OIDs yet
				ColumnNum: colNum,
				TypeOID:   typeOID,
				TypeSize:  typeSize,
				TypeMod:   typeMod,
				Format:    0,  // Text format
			}
			colDescs = append(colDescs, colDesc)
			colNum++
		}
	}
	
	// Handle SELECT * case
	if len(colDescs) == 0 && tableName != "" {
		// Get all columns from the table
		columns := s.metaStore.GetTableColumns(tableName)
		if len(columns) == 0 {
			// If no columns defined in metastore, try to get from first row
			// This supports truly schema-less operation
			table, exists := s.dataStore.GetTable(tableName)
			if exists {
				rows := table.GetRows()
				if len(rows) > 0 {
					// Get columns from first row
					for colName := range rows[0] {
						columns = append(columns, colName)
					}
					// Sort for consistent ordering
					sort.Strings(columns)
				}
			}
		}
		
		colNum := int16(1)
		for _, col := range columns {
			colType := s.metaStore.GetColumnType(tableName, col)
			typeOID := VSQLTypeToOID(colType)
			typeSize, typeMod := GetTypeSizeAndMod(typeOID)
			
			colDesc := ColumnDescription{
				Name:      col,
				TableOID:  0,
				ColumnNum: colNum,
				TypeOID:   typeOID,
				TypeSize:  typeSize,
				TypeMod:   typeMod,
				Format:    0,
			}
			colDescs = append(colDescs, colDesc)
			colNum++
		}
	}
	
	// If still no columns, return a default
	if len(colDescs) == 0 {
		colDescs = append(colDescs, ColumnDescription{
			Name:      "result",
			TableOID:  0,
			ColumnNum: 0,
			TypeOID:   OIDText,
			TypeSize:  -1,
			TypeMod:   -1,
			Format:    0,
		})
	}
	
	return colDescs, nil
}

// extractTableNameFromNode extracts table name from a FROM clause node
func extractTableNameFromNode(node *pg_query.Node) string {
	if rangeVar, ok := node.Node.(*pg_query.Node_RangeVar); ok && rangeVar.RangeVar != nil {
		// Return the table name, removing quotes if present
		return strings.Trim(rangeVar.RangeVar.Relname, `"`)
	}
	return ""
}

// extractColumnName extracts column name from a ColumnRef
func extractColumnName(colRef *pg_query.ColumnRef) string {
	if len(colRef.Fields) > 0 {
		// Handle qualified names (table.column)
		var parts []string
		for _, field := range colRef.Fields {
			if str, ok := field.Node.(*pg_query.Node_String_); ok {
				parts = append(parts, str.String_.Sval)
			}
		}
		if len(parts) > 0 {
			// Return the last part (column name), removing quotes
			colName := parts[len(parts)-1]
			return strings.Trim(colName, `"`)
		}
	}
	return "?column?"
}

// extractFunctionName extracts function name from a FuncCall
func extractFunctionName(funcCall *pg_query.FuncCall) string {
	if len(funcCall.Funcname) > 0 {
		if str, ok := funcCall.Funcname[0].Node.(*pg_query.Node_String_); ok {
			return str.String_.Sval
		}
	}
	return "?column?"
}

// countParameters counts the number of parameters ($1, $2, etc.) in a query
func countParameters(query string) int {
	maxParam := 0
	// Simple regex to find $N patterns
	for i := 0; i < len(query); i++ {
		if query[i] == '$' && i+1 < len(query) {
			// Parse the number after $
			j := i + 1
			for j < len(query) && query[j] >= '0' && query[j] <= '9' {
				j++
			}
			if j > i+1 {
				paramNum := 0
				for k := i+1; k < j; k++ {
					paramNum = paramNum*10 + int(query[k]-'0')
				}
				if paramNum > maxParam {
					maxParam = paramNum
				}
			}
		}
	}
	return maxParam
}

// inferParameterTypes tries to infer parameter types from query context
func inferParameterTypes(query string, parsedQuery *pg_query.ParseResult) []int32 {
	// For now, return int8 (OID 20) for OFFSET parameters
	// This is a simple implementation - a full implementation would analyze the AST
	paramCount := countParameters(query)
	if paramCount == 0 {
		return nil
	}
	
	paramTypes := make([]int32, paramCount)
	
	// Check if the query contains OFFSET $N
	if strings.Contains(query, "OFFSET") && strings.Contains(query, "$") {
		// For OFFSET parameters, use int8
		for i := range paramTypes {
			paramTypes[i] = OIDInt8  // 20
		}
	} else {
		// Default to unknown
		for i := range paramTypes {
			paramTypes[i] = 0
		}
	}
	
	return paramTypes
}

