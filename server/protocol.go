package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type MessageType byte

const (
	AuthenticationOk    MessageType = 'R'
	Query               MessageType = 'Q'
	ParseComplete       MessageType = '1'
	BindComplete        MessageType = '2'
	DataRow             MessageType = 'D'
	CommandComplete     MessageType = 'C'
	ReadyForQuery       MessageType = 'Z'
	ErrorResponse       MessageType = 'E'
	NoticeResponse      MessageType = 'N'
	ParameterStatus     MessageType = 'S'
	BackendKeyData      MessageType = 'K'
	RowDescription      MessageType = 'T'
	NoData              MessageType = 'n'
	EmptyQueryResponse  MessageType = 'I'
	
	// Extended Query Protocol
	Parse               MessageType = 'P'
	Bind                MessageType = 'B'
	Execute             MessageType = 'E'
	Describe            MessageType = 'D'
	Close               MessageType = 'C'
	Sync                MessageType = 'S'
	Flush               MessageType = 'H'
	ParameterDescription MessageType = 't'
	PortalSuspended     MessageType = 's'
	CloseComplete       MessageType = '3'
)

type Message struct {
	Type MessageType
	Data []byte
}

func ReadMessage(r io.Reader) (*Message, error) {
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(r, msgType); err != nil {
		return nil, err
	}

	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBytes) - 4
	
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return &Message{
		Type: MessageType(msgType[0]),
		Data: data,
	}, nil
}

func WriteMessage(w io.Writer, msgType MessageType, data []byte) error {
	var buf bytes.Buffer
	
	buf.WriteByte(byte(msgType))
	
	length := uint32(len(data) + 4)
	if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
		return err
	}
	
	buf.Write(data)
	
	_, err := w.Write(buf.Bytes())
	return err
}

func WriteAuthenticationOk(w io.Writer) error {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, int32(0))
	return WriteMessage(w, AuthenticationOk, buf.Bytes())
}

func WriteReadyForQuery(w io.Writer) error {
	return WriteMessage(w, ReadyForQuery, []byte{'I'})
}

func WriteCommandComplete(w io.Writer, tag string) error {
	data := append([]byte(tag), 0)
	return WriteMessage(w, CommandComplete, data)
}

func WriteErrorResponse(w io.Writer, msg string) error {
	var buf bytes.Buffer
	
	buf.WriteByte('S')
	buf.Write([]byte("ERROR"))
	buf.WriteByte(0)
	
	buf.WriteByte('M')
	buf.Write([]byte(msg))
	buf.WriteByte(0)
	
	buf.WriteByte(0)
	
	return WriteMessage(w, ErrorResponse, buf.Bytes())
}

// ColumnDescription represents metadata for a result column
type ColumnDescription struct {
	Name        string
	TableOID    int32  // OID of table (0 if not from a table)
	ColumnNum   int16  // Column number in table (0 if not from a table)
	TypeOID     int32  // PostgreSQL type OID
	TypeSize    int16  // Type size (-1 for variable length)
	TypeMod     int32  // Type modifier (-1 if not applicable)
	Format      int16  // Format code (0 = text, 1 = binary)
}

func WriteRowDescription(w io.Writer, columns []string) error {
	// Convert simple column names to ColumnDescription with default type (text)
	colDescs := make([]ColumnDescription, len(columns))
	for i, name := range columns {
		colDescs[i] = ColumnDescription{
			Name:     name,
			TableOID: 0,
			ColumnNum: 0,
			TypeOID:  25,  // text type
			TypeSize: -1,
			TypeMod:  -1,
			Format:   0,
		}
	}
	return WriteRowDescriptionExt(w, colDescs)
}

// WriteRowDescriptionExt writes a RowDescription message with full column metadata
func WriteRowDescriptionExt(w io.Writer, columns []ColumnDescription) error {
	var buf bytes.Buffer
	
	fieldCount := int16(len(columns))
	binary.Write(&buf, binary.BigEndian, fieldCount)
	
	for _, col := range columns {
		buf.Write([]byte(col.Name))
		buf.WriteByte(0)
		
		binary.Write(&buf, binary.BigEndian, col.TableOID)
		binary.Write(&buf, binary.BigEndian, col.ColumnNum)
		binary.Write(&buf, binary.BigEndian, col.TypeOID)
		binary.Write(&buf, binary.BigEndian, col.TypeSize)
		binary.Write(&buf, binary.BigEndian, col.TypeMod)
		binary.Write(&buf, binary.BigEndian, col.Format)
	}
	
	return WriteMessage(w, RowDescription, buf.Bytes())
}

func WriteDataRow(w io.Writer, values []interface{}) error {
	var buf bytes.Buffer
	
	fieldCount := int16(len(values))
	binary.Write(&buf, binary.BigEndian, fieldCount)
	
	for _, val := range values {
		if val == nil {
			binary.Write(&buf, binary.BigEndian, int32(-1))
		} else {
			str := fmt.Sprintf("%v", val)
			binary.Write(&buf, binary.BigEndian, int32(len(str)))
			buf.Write([]byte(str))
		}
	}
	
	return WriteMessage(w, DataRow, buf.Bytes())
}

func WriteParameterStatus(w io.Writer, name, value string) error {
	var buf bytes.Buffer
	buf.Write([]byte(name))
	buf.WriteByte(0)
	buf.Write([]byte(value))
	buf.WriteByte(0)
	return WriteMessage(w, ParameterStatus, buf.Bytes())
}

func WriteBackendKeyData(w io.Writer, processID, secretKey int32) error {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, processID)
	binary.Write(&buf, binary.BigEndian, secretKey)
	return WriteMessage(w, BackendKeyData, buf.Bytes())
}

func WriteParseComplete(w io.Writer) error {
	return WriteMessage(w, ParseComplete, []byte{})
}

func WriteBindComplete(w io.Writer) error {
	return WriteMessage(w, BindComplete, []byte{})
}

func WriteNoData(w io.Writer) error {
	return WriteMessage(w, NoData, []byte{})
}

func WriteEmptyQueryResponse(w io.Writer) error {
	return WriteMessage(w, EmptyQueryResponse, []byte{})
}

func WriteParameterDescription(w io.Writer, paramTypes []int32) error {
	var buf bytes.Buffer
	
	// Number of parameters
	binary.Write(&buf, binary.BigEndian, int16(len(paramTypes)))
	
	// OID of each parameter type
	for _, oid := range paramTypes {
		binary.Write(&buf, binary.BigEndian, oid)
	}
	
	return WriteMessage(w, ParameterDescription, buf.Bytes())
}

func WritePortalSuspended(w io.Writer) error {
	return WriteMessage(w, PortalSuspended, []byte{})
}