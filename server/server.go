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

	"vsql/parser"
	"vsql/storage"
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
		case 'X': // Terminate message
			// Client is closing connection
			return
		default:
			// Silently ignore unknown message types for better compatibility
			// fmt.Printf("Unhandled message type: %c\n", msg.Type)
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

