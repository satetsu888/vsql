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

func WriteRowDescription(w io.Writer, columns []string) error {
	var buf bytes.Buffer
	
	fieldCount := int16(len(columns))
	binary.Write(&buf, binary.BigEndian, fieldCount)
	
	for _, col := range columns {
		buf.Write([]byte(col))
		buf.WriteByte(0)
		
		binary.Write(&buf, binary.BigEndian, int32(0))
		binary.Write(&buf, binary.BigEndian, int16(0))
		binary.Write(&buf, binary.BigEndian, int32(25))
		binary.Write(&buf, binary.BigEndian, int16(-1))
		binary.Write(&buf, binary.BigEndian, int32(-1))
		binary.Write(&buf, binary.BigEndian, int16(0))
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