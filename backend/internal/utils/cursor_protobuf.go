// ref: _ref/9router/open-sse/utils/cursorProtobuf.js
package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

const ProtobufSchemaVersion = "1.1.3"

type WireType int

const (
	WireTypeVarint WireType = 0
	WireTypeFixed64 WireType = 1
	WireTypeLen     WireType = 2
	WireTypeFixed32 WireType = 5
)

type Role int

const (
	RoleUser Role = 1
	RoleAssistant Role = 2
)

type UnifiedMode int

const (
	UnifiedModeChat   UnifiedMode = 1
	UnifiedModeAgent  UnifiedMode = 2
)

type ThinkingLevel int

const (
	ThinkingLevelUnspecified ThinkingLevel = 0
	ThinkingLevelMedium      ThinkingLevel = 1
	ThinkingLevelHigh        ThinkingLevel = 2
)

func EncodeVarint(value uint64) []byte {
	var bytes []byte
	for value >= 0x80 {
		bytes = append(bytes, byte((value&0x7F)|0x80))
		value >>= 7
	}
	bytes = append(bytes, byte(value&0x7F))
	return bytes
}

func DecodeVarint(data []byte, offset int) (value uint64, bytesRead int) {
	var result uint64
	var shift uint
	for i := offset; i < len(data); i++ {
		b := data[i]
		result |= uint64(b&0x7F) << shift
		shift += 7
		bytesRead++
		if b&0x80 == 0 {
			break
		}
	}
	return result, bytesRead
}

func EncodeField(fieldNum int, wireType WireType, value interface{}) []byte {
	tag := (fieldNum << 3) | int(wireType)
	tagBytes := EncodeVarint(uint64(tag))

	switch wireType {
	case WireTypeVarint:
		v, ok := value.(uint64)
		if !ok {
			if vi, ok := value.(int); ok {
				v = uint64(vi)
			} else {
				v = 0
			}
		}
		valueBytes := EncodeVarint(v)
		return append(tagBytes, valueBytes...)

	case WireTypeLen:
		var dataBytes []byte
		switch v := value.(type) {
		case string:
			dataBytes = []byte(v)
		case []byte:
			dataBytes = v
		default:
			dataBytes = []byte{}
		}
		lenBytes := EncodeVarint(uint64(len(dataBytes)))
		return append(append(tagBytes, lenBytes...), dataBytes...)

	default:
		return tagBytes
	}
}

func EncodeString(fieldNum int, value string) []byte {
	return EncodeField(fieldNum, WireTypeLen, value)
}

func EncodeBytes(fieldNum int, value []byte) []byte {
	return EncodeField(fieldNum, WireTypeLen, value)
}

func EncodeUint32(fieldNum int, value uint32) []byte {
	return EncodeField(fieldNum, WireTypeVarint, uint64(value))
}

func EncodeInt(fieldNum int, value int) []byte {
	return EncodeField(fieldNum, WireTypeVarint, uint64(value))
}

func ConcatArrays(arrays ...[]byte) []byte {
	var result []byte
	for _, arr := range arrays {
		result = append(result, arr...)
	}
	return result
}

func GzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(data)
	if err != nil {
		w.Close()
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
}

func GzipDecompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

type ConversationMessage struct {
	Content   []ContentBlock
	Role      Role
	ID        string
	ToolResults []ToolResult
}

type ContentBlock struct {
	Type string
	Text string
}

type ToolResult struct {
	CallID     string
	Name       string
	Index      int
	RawArgs    string
	Result     string
	ToolCall   *ProtobufToolCall
}

type ProtobufToolCall struct {
	ID         string
	Name       string
	RawArgs    string
	MCPParams  *MCPParams
}

type MCPParams struct {
	ToolsList []MCPTool
}

type MCPTool struct {
	Name        string
	Description string
	Params      json.RawMessage
	Server      string
}

type StreamRequest struct {
	Messages       []ConversationMessage
	Instruction    string
	Model          string
	ConversationID string
	UnifiedMode    UnifiedMode
	ThinkingLevel  ThinkingLevel
	MCPTools       []MCPTool
}

func NewConversationID() string {
	return uuid.New().String()
}

func ParseProtobufMessage(data []byte) (map[int]interface{}, error) {
	result := make(map[int]interface{})
	offset := 0

	for offset < len(data) {
		if offset >= len(data) {
			break
		}

		tag, bytesRead := DecodeVarint(data, offset)
		if bytesRead == 0 {
			break
		}
		offset += bytesRead

		fieldNum := int(tag >> 3)
		wireType := WireType(tag & 0x7)

		switch wireType {
		case WireTypeVarint:
			v, n := DecodeVarint(data, offset)
			result[fieldNum] = v
			offset += n

		case WireTypeLen:
			length, n := DecodeVarint(data, offset)
			offset += n
			if int(length) > len(data)-offset {
				return nil, fmt.Errorf("invalid length for field %d", fieldNum)
			}
			result[fieldNum] = data[offset : offset+int(length)]
			offset += int(length)

		case WireTypeFixed64:
			if offset+8 > len(data) {
				return nil, fmt.Errorf("insufficient data for fixed64 field %d", fieldNum)
			}
			result[fieldNum] = binary.LittleEndian.Uint64(data[offset : offset+8])
			offset += 8

		case WireTypeFixed32:
			if offset+4 > len(data) {
				return nil, fmt.Errorf("insufficient data for fixed32 field %d", fieldNum)
			}
			result[fieldNum] = binary.LittleEndian.Uint32(data[offset : offset+4])
			offset += 4

		default:
			return nil, fmt.Errorf("unknown wire type %d for field %d", wireType, fieldNum)
		}
	}

	return result, nil
}
