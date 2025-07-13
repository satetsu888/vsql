package server

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
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
		case Bind:
			if err := s.handleBind(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			} else {
				WriteBindComplete(writer)
			}
		case Execute:
			if err := s.handleExecute(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			}
		case Describe:
			if err := s.handleDescribe(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			}
		case Close:
			if err := s.handleClose(msg.Data, extState, writer); err != nil {
				WriteErrorResponse(writer, err.Error())
			}
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
	
	// Create and store the prepared statement
	stmt := &PreparedStatement{
		Name:        statementName,
		Query:       query,
		ParsedQuery: parsedQuery,
		ParamTypes:  paramTypes,
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
		if err := WriteRowDescription(w, columns); err != nil {
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
		return WriteParameterDescription(w, stmt.ParamTypes)
		
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
					// For now, we'll send a generic row description
					// A full implementation would analyze the query to determine actual columns
					columns := []string{"result"}
					return WriteRowDescription(w, columns)
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
	return WriteMessage(w, '3', []byte{})
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
			// For now, treat all parameters as text
			// A full implementation would handle different formats based on ParameterFormats
			value = fmt.Sprintf("'%s'", string(paramValue))
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

