// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/cursor.js
package executor

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CursorExecutor implements the Executor interface for Cursor API.
// ref: open-sse/executors/cursor.js:108
type CursorExecutor struct {
	BaseExecutor
}

// NewCursorExecutor creates a new Cursor executor.
// ref: open-sse/executors/cursor.js:109-111
func NewCursorExecutor() *CursorExecutor {
	return &CursorExecutor{
		BaseExecutor: NewBaseExecutor("cursor"),
	}
}

// CursorCredentials holds Cursor-specific authentication data.
type CursorCredentials struct {
	AccessToken         string
	MachineID           string
	GhostMode           bool
	ProviderSpecificData map[string]interface{}
}

// compressFlag represents ConnectRPC compression flags.
// ref: open-sse/executors/cursor.js:32-37
type compressFlag byte

const (
	compressNone       compressFlag = 0x00
	compressGzip       compressFlag = 0x01
	compressTrailer    compressFlag = 0x02
	compressGzipTrailer compressFlag = 0x03
)

// wireType represents protobuf wire types.
// ref: open-sse/utils/cursorProtobuf.js:17
type wireType int

const (
	wireVarint wireType = 0
	wireFixed64 wireType = 1
	wireLen     wireType = 2
	wireFixed32 wireType = 5
)

// Protobuf field constants for Cursor protocol.
// ref: open-sse/utils/cursorProtobuf.js:27-174
const (
	// StreamUnifiedChatRequestWithTools (top level)
	fieldRequest = 1

	// StreamUnifiedChatRequest
	fieldMessages       = 1
	fieldUnknown2       = 2
	fieldInstruction    = 3
	fieldUnknown4       = 4
	fieldModel          = 5
	fieldWebTool        = 8
	fieldUnknown13      = 13
	fieldCursorSetting  = 15
	fieldUnknown19      = 19
	fieldConversationID = 23
	fieldMetadata       = 26
	fieldIsAgentic      = 27
	fieldSupportedTools = 29
	fieldMessageIDs     = 30
	fieldMCPTools       = 34
	fieldLargeContext   = 35
	fieldUnknown38      = 38
	fieldUnifiedMode    = 46
	fieldUnknown47      = 47
	fieldDisableTools   = 48
	fieldThinkingLevel  = 49
	fieldUnknown51      = 51
	fieldUnknown53      = 53
	fieldUnifiedModeName = 54

	// ConversationMessage
	fieldMsgContent       = 1
	fieldMsgRole          = 2
	fieldMsgID            = 13
	fieldMsgToolResults   = 18
	fieldMsgIsAgentic     = 29
	fieldMsgUnifiedMode   = 47
	fieldMsgSupportedTools = 51

	// ConversationMessage.ToolResult
	fieldToolResultCallID      = 1
	fieldToolResultName        = 2
	fieldToolResultIndex       = 3
	fieldToolResultRawArgs     = 5
	fieldToolResultResult      = 8
	fieldToolResultToolCall    = 11
	fieldToolResultModelCallID = 12

	// Model
	fieldModelName  = 1
	fieldModelEmpty = 4

	// Instruction
	fieldInstructionText = 1

	// Metadata
	fieldMetaPlatform  = 1
	fieldMetaArch      = 2
	fieldMetaVersion   = 3
	fieldMetaCwd       = 4
	fieldMetaTimestamp = 5

	// MessageId
	fieldMsgIDID       = 1
	fieldMsgIDSummary  = 2
	fieldMsgIDRole     = 3

	// MCPTool
	fieldMCPToolName   = 1
	fieldMCPToolDesc   = 2
	fieldMCPToolParams = 3
	fieldMCPToolServer = 4

	// Response fields
	fieldToolCall   = 1
	fieldResponse   = 2
	fieldResponseText = 1
	fieldThinking   = 25
)

// Role constants for messages.
// ref: open-sse/utils/cursorProtobuf.js:19
const (
	roleUser      = 1
	roleAssistant = 2
)

// Unified mode constants.
// ref: open-sse/utils/cursorProtobuf.js:21
const (
	unifiedModeChat  = 1
	unifiedModeAgent = 2
)

// Thinking level constants.
// ref: open-sse/utils/cursorProtobuf.js:23
const (
	thinkingUnspecified = 0
	thinkingMedium      = 1
	thinkingHigh        = 2
)

// Client side tool V2 constant.
// ref: open-sse/utils/cursorProtobuf.js:24-25
const clientSideToolV2MCP = 19

// generateHashed64Hex generates SHA-256 hash as hex string.
// ref: open-sse/utils/cursorChecksum.js:17-19
func generateHashed64Hex(input, salt string) string {
	h := sha256.New()
	h.Write([]byte(input + salt))
	return hex.EncodeToString(h.Sum(nil))
}

// generateCursorSessionID generates a UUID v5 session ID for Cursor.
// ref: open-sse/utils/cursorChecksum.js:26-28
func generateCursorSessionID(authToken string) string {
	// UUID v5 with DNS namespace
	namespace := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8") // DNS namespace
	return uuid.NewSHA1(namespace, []byte(authToken)).String()
}

// generateCursorChecksum generates the x-cursor-checksum header value.
// ref: open-sse/utils/cursorChecksum.js:43-85
func generateCursorChecksum(machineID string) string {
	// Math.floor(Date.now() / 1e6)
	timestamp := time.Now().UnixMilli() / 1000000

	// Create byte array from timestamp (6 bytes, big-endian)
	byteArray := make([]byte, 6)
	byteArray[0] = byte((timestamp >> 40) & 0xFF)
	byteArray[1] = byte((timestamp >> 32) & 0xFF)
	byteArray[2] = byte((timestamp >> 24) & 0xFF)
	byteArray[3] = byte((timestamp >> 16) & 0xFF)
	byteArray[4] = byte((timestamp >> 8) & 0xFF)
	byteArray[5] = byte(timestamp & 0xFF)

	// Jyh cipher obfuscation
	t := byte(165)
	for i := 0; i < len(byteArray); i++ {
		byteArray[i] = (byteArray[i] ^ t) + byte(i%256)
		t = byteArray[i]
	}

	// URL-safe base64 encode (without padding)
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var encoded strings.Builder

	for i := 0; i < len(byteArray); i += 3 {
		a := byteArray[i]
		b := byte(0)
		c := byte(0)
		if i+1 < len(byteArray) {
			b = byteArray[i+1]
		}
		if i+2 < len(byteArray) {
			c = byteArray[i+2]
		}

		encoded.WriteByte(alphabet[a>>2])
		encoded.WriteByte(alphabet[((a&3)<<4)|(b>>4)])

		if i+1 < len(byteArray) {
			encoded.WriteByte(alphabet[((b&15)<<2)|(c>>6)])
		}
		if i+2 < len(byteArray) {
			encoded.WriteByte(alphabet[c&63])
		}
	}

	return encoded.String() + machineID
}

// buildCursorHeaders builds all Cursor API headers.
// ref: open-sse/utils/cursorChecksum.js:95-142
func buildCursorHeaders(accessToken, machineID string, ghostMode bool) http.Header {
	// Clean token if it has prefix
	cleanToken := accessToken
	if strings.Contains(accessToken, "::") {
		parts := strings.SplitN(accessToken, "::", 2)
		if len(parts) > 1 {
			cleanToken = parts[1]
		}
	}

	// Generate machine ID if not provided
	effectiveMachineID := machineID
	if effectiveMachineID == "" {
		effectiveMachineID = generateHashed64Hex(cleanToken, "machineId")
	}

	// Generate derived values
	sessionID := generateCursorSessionID(cleanToken)
	clientKey := generateHashed64Hex(cleanToken, "")
	checksum := generateCursorChecksum(effectiveMachineID)

	// Detect OS
	os := "linux"
	switch runtime.GOOS {
	case "windows":
		os = "windows"
	case "darwin":
		os = "macos"
	}

	// Detect architecture
	arch := "x64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	}

	headers := make(http.Header)
	headers.Set("authorization", "Bearer "+cleanToken)
	headers.Set("connect-accept-encoding", "gzip")
	headers.Set("connect-protocol-version", "1")
	headers.Set("content-type", "application/connect+proto")
	headers.Set("user-agent", "connect-es/1.6.1")
	headers.Set("x-amzn-trace-id", "Root="+uuid.New().String())
	headers.Set("x-client-key", clientKey)
	headers.Set("x-cursor-checksum", checksum)
	headers.Set("x-cursor-client-version", "3.1.0")
	headers.Set("x-cursor-client-type", "ide")
	headers.Set("x-cursor-client-os", os)
	headers.Set("x-cursor-client-arch", arch)
	headers.Set("x-cursor-client-device-type", "desktop")
	headers.Set("x-cursor-config-version", uuid.New().String())
	headers.Set("x-cursor-timezone", "UTC")
	headers.Set("x-ghost-mode", fmt.Sprintf("%v", ghostMode))
	headers.Set("x-request-id", uuid.New().String())
	headers.Set("x-session-id", sessionID)

	return headers
}

// encodeVarint encodes a varint value.
// ref: open-sse/utils/cursorProtobuf.js:191-199
func encodeVarint(value uint64) []byte {
	var bytes []byte
	for value >= 0x80 {
		bytes = append(bytes, byte((value&0x7F)|0x80))
		value >>= 7
	}
	bytes = append(bytes, byte(value&0x7F))
	return bytes
}

// encodeField encodes a protobuf field.
// ref: open-sse/utils/cursorProtobuf.js:201-222
func encodeField(fieldNum int, wt wireType, value interface{}) []byte {
	tag := uint64((fieldNum << 3) | int(wt))
	tagBytes := encodeVarint(tag)

	switch wt {
	case wireVarint:
		var v uint64
		switch val := value.(type) {
		case int:
			v = uint64(val)
		case uint64:
			v = val
		case int64:
			v = uint64(val)
		default:
			return nil
		}
		valueBytes := encodeVarint(v)
		return append(tagBytes, valueBytes...)
	case wireLen:
		var dataBytes []byte
		switch v := value.(type) {
		case string:
			dataBytes = []byte(v)
		case []byte:
			dataBytes = v
		default:
			dataBytes = []byte{}
		}
		lengthBytes := encodeVarint(uint64(len(dataBytes)))
		return append(append(tagBytes, lengthBytes...), dataBytes...)
	}
	return nil
}

// encodeModel encodes the model field.
// ref: open-sse/utils/cursorProtobuf.js (encodeModel function)
func encodeModel(modelName string) []byte {
	return append(
		encodeField(fieldModelName, wireLen, modelName),
		encodeField(fieldModelEmpty, wireVarint, 0)...,
	)
}

// encodeInstruction encodes the instruction field.
// ref: open-sse/utils/cursorProtobuf.js (encodeInstruction function)
func encodeInstruction(text string) []byte {
	return encodeField(fieldInstructionText, wireLen, text)
}

// encodeCursorSetting encodes the cursor setting field.
// ref: open-sse/utils/cursorProtobuf.js (encodeCursorSetting function)
func encodeCursorSetting() []byte {
	// Minimal cursor setting with required fields
	return encodeField(1, wireLen, "") // SETTING_PATH
}

// encodeMetadata encodes the metadata field.
// ref: open-sse/utils/cursorProtobuf.js (encodeMetadata function)
func encodeMetadata() []byte {
	return append(
		append(
			append(
				append(
					encodeField(fieldMetaPlatform, wireLen, runtime.GOOS),
					encodeField(fieldMetaArch, wireLen, runtime.GOARCH)...,
				),
				encodeField(fieldMetaVersion, wireLen, "3.1.0")...,
			),
			encodeField(fieldMetaCwd, wireLen, "/")...,
		),
		encodeField(fieldMetaTimestamp, wireVarint, time.Now().Unix())...,
	)
}

// encodeMessageID encodes a message ID.
// ref: open-sse/utils/cursorProtobuf.js (encodeMessageId function)
func encodeMessageID(messageID string, role int) []byte {
	return append(
		encodeField(fieldMsgIDID, wireLen, messageID),
		encodeField(fieldMsgIDRole, wireVarint, role)...,
	)
}

// encodeMCPTool encodes an MCP tool definition.
// ref: open-sse/utils/cursorProtobuf.js (encodeMcpTool function)
func encodeMCPTool(tool map[string]interface{}) []byte {
	name, _ := tool["name"].(string)
	desc, _ := tool["description"].(string)

	result := encodeField(fieldMCPToolName, wireLen, name)
	result = append(result, encodeField(fieldMCPToolDesc, wireLen, desc)...)

	// Encode parameters if present
	if params, ok := tool["parameters"].(map[string]interface{}); ok {
		if paramsJSON, err := json.Marshal(params); err == nil {
			result = append(result, encodeField(fieldMCPToolParams, wireLen, paramsJSON)...)
		}
	}

	return result
}

// encodeMessage encodes a conversation message.
// ref: open-sse/utils/cursorProtobuf.js (encodeMessage function)
func encodeMessage(content string, role int, messageID string, toolResults []map[string]interface{}, isLast, hasTools bool) []byte {
	result := encodeField(fieldMsgContent, wireLen, content)
	result = append(result, encodeField(fieldMsgRole, wireVarint, role)...)
	result = append(result, encodeField(fieldMsgID, wireLen, messageID)...)

	if len(toolResults) > 0 {
		for _, tr := range toolResults {
			encodedTR := encodeToolResult(tr)
			result = append(result, encodeField(fieldMsgToolResults, wireLen, encodedTR)...)
		}
	}

	if hasTools && isLast {
		result = append(result, encodeField(fieldMsgIsAgentic, wireVarint, 1)...)
		result = append(result, encodeField(fieldMsgUnifiedMode, wireVarint, unifiedModeAgent)...)
		result = append(result, encodeField(fieldMsgSupportedTools, wireVarint, 1)...)
	}

	return result
}

// encodeToolResult encodes a tool result.
// ref: open-sse/utils/cursorProtobuf.js:348-396
func encodeToolResult(toolResult map[string]interface{}) []byte {
	toolCallID, _ := toolResult["tool_call_id"].(string)
	toolName, _ := toolResult["name"].(string)
	if toolName == "" {
		toolName, _ = toolResult["tool_name"].(string)
	}
	toolIndex, _ := toolResult["tool_index"].(int)
	if toolIndex == 0 {
		toolIndex = 1
	}
	rawArgs, _ := toolResult["raw_args"].([]byte)
	resultContent, _ := toolResult["result"].([]byte)

	encoded := encodeField(fieldToolResultCallID, wireLen, toolCallID)
	encoded = append(encoded, encodeField(fieldToolResultName, wireLen, toolName)...)
	encoded = append(encoded, encodeField(fieldToolResultIndex, wireVarint, toolIndex)...)

	if len(rawArgs) > 0 {
		encoded = append(encoded, encodeField(fieldToolResultRawArgs, wireLen, rawArgs)...)
	}
	if len(resultContent) > 0 {
		encoded = append(encoded, encodeField(fieldToolResultResult, wireLen, resultContent)...)
	}

	return encoded
}

// generateCursorBody generates the Cursor request body in protobuf format.
// ref: open-sse/utils/cursorProtobuf.js:538-584
func generateCursorBody(messages []map[string]interface{}, model string, tools []map[string]interface{}, reasoningEffort string, forceAgentMode bool) []byte {
	hasTools := len(tools) > 0
	isAgentic := hasTools || forceAgentMode

	// Normalize and format messages
	type formattedMsg struct {
		content     string
		role        int
		messageID   string
		isLast      bool
		hasTools    bool
		toolResults []map[string]interface{}
	}
	var formattedMessages []formattedMsg
	var messageIDs []struct {
		messageID string
		role      int
	}

	for i, msg := range messages {
		content, _ := msg["content"].(string)
		role := roleUser
		if r, ok := msg["role"].(string); ok && r == "assistant" {
			role = roleAssistant
		}
		msgID := uuid.New().String()
		isLast := i == len(messages)-1

		var toolResults []map[string]interface{}
		if tr, ok := msg["tool_results"].([]map[string]interface{}); ok {
			toolResults = tr
		}

		formattedMessages = append(formattedMessages, formattedMsg{
			content:     content,
			role:        role,
			messageID:   msgID,
			isLast:      isLast,
			hasTools:    hasTools,
			toolResults: toolResults,
		})
		messageIDs = append(messageIDs, struct {
			messageID string
			role      int
		}{messageID: msgID, role: role})
	}

	// Map reasoning effort to thinking level
	thinkingLevel := thinkingUnspecified
	switch reasoningEffort {
	case "medium":
		thinkingLevel = thinkingMedium
	case "high":
		thinkingLevel = thinkingHigh
	}

	// Build request
	var request []byte

	// Messages
	for _, fm := range formattedMessages {
		msgBytes := encodeMessage(fm.content, fm.role, fm.messageID, fm.toolResults, fm.isLast, fm.hasTools)
		request = append(request, encodeField(fieldMessages, wireLen, msgBytes)...)
	}

	// Static fields
	request = append(request, encodeField(fieldUnknown2, wireVarint, 1)...)
	request = append(request, encodeField(fieldInstruction, wireLen, encodeInstruction(""))...)
	request = append(request, encodeField(fieldUnknown4, wireVarint, 1)...)
	request = append(request, encodeField(fieldModel, wireLen, encodeModel(model))...)
	request = append(request, encodeField(fieldWebTool, wireLen, "")...)
	request = append(request, encodeField(fieldUnknown13, wireVarint, 1)...)
	request = append(request, encodeField(fieldCursorSetting, wireLen, encodeCursorSetting())...)
	request = append(request, encodeField(fieldUnknown19, wireVarint, 1)...)
	request = append(request, encodeField(fieldConversationID, wireLen, uuid.New().String())...)
	request = append(request, encodeField(fieldMetadata, wireLen, encodeMetadata())...)

	// Tool-related fields
	request = append(request, encodeField(fieldIsAgentic, wireVarint, boolToInt(isAgentic))...)
	if isAgentic {
		request = append(request, encodeField(fieldSupportedTools, wireLen, encodeVarint(1))...)
	}

	// Message IDs
	for _, mid := range messageIDs {
		request = append(request, encodeField(fieldMessageIDs, wireLen, encodeMessageID(mid.messageID, mid.role))...)
	}

	// MCP Tools
	for _, tool := range tools {
		request = append(request, encodeField(fieldMCPTools, wireLen, encodeMCPTool(tool))...)
	}

	// Mode fields
	request = append(request, encodeField(fieldLargeContext, wireVarint, 0)...)
	request = append(request, encodeField(fieldUnknown38, wireVarint, 0)...)
	if isAgentic {
		request = append(request, encodeField(fieldUnifiedMode, wireVarint, unifiedModeAgent)...)
	} else {
		request = append(request, encodeField(fieldUnifiedMode, wireVarint, unifiedModeChat)...)
	}
	request = append(request, encodeField(fieldUnknown47, wireLen, "")...)
	if isAgentic {
		request = append(request, encodeField(fieldDisableTools, wireVarint, 0)...)
	} else {
		request = append(request, encodeField(fieldDisableTools, wireVarint, 1)...)
	}
	request = append(request, encodeField(fieldThinkingLevel, wireVarint, thinkingLevel)...)
	request = append(request, encodeField(fieldUnknown51, wireVarint, 0)...)
	request = append(request, encodeField(fieldUnknown53, wireVarint, 1)...)
	if isAgentic {
		request = append(request, encodeField(fieldUnifiedModeName, wireLen, "Agent")...)
	} else {
		request = append(request, encodeField(fieldUnifiedModeName, wireLen, "Ask")...)
	}

	// Wrap in request field
	return encodeField(fieldRequest, wireLen, request)
}

// boolToInt converts bool to int (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// decompressPayload decompresses a payload based on compression flags.
// ref: open-sse/executors/cursor.js:44-86
func decompressPayload(payload []byte, flags compressFlag) []byte {
	// Check if payload is JSON error (starts with {"error")
	if len(payload) > 10 && payload[0] == 0x7b && payload[1] == 0x22 {
		if bytes.HasPrefix(payload, []byte(`{"error"`)) {
			return payload
		}
	}

	switch flags {
	case compressGzip, compressTrailer, compressGzipTrailer:
		// Try gzip decompression
		reader, err := gzip.NewReader(bytes.NewReader(payload))
		if err == nil {
			decompressed, err := io.ReadAll(reader)
			reader.Close()
			if err == nil {
				return decompressed
			}
		}

		// Fallback: try raw zlib deflate
		decompressed, err := io.ReadAll(flate.NewReader(bytes.NewReader(payload)))
		if err == nil {
			return decompressed
		}
	}

	return payload
}

// parseConnectRPCFrame parses a ConnectRPC frame.
// ref: open-sse/utils/cursorProtobuf.js:7-8 (parseConnectRPCFrame)
func parseConnectRPCFrame(data []byte, offset int) (flags compressFlag, length uint32, payload []byte, newOffset int, err error) {
	if offset+5 > len(data) {
		err = fmt.Errorf("insufficient data for frame header")
		return
	}

	flags = compressFlag(data[offset])
	length = binary.BigEndian.Uint32(data[offset+1 : offset+5])

	if offset+5+int(length) > len(data) {
		err = fmt.Errorf("incomplete frame")
		return
	}

	payload = data[offset+5 : offset+5+int(length)]
	newOffset = offset + 5 + int(length)
	return
}

// extractTextFromResponse extracts text from a protobuf response.
// ref: open-sse/utils/cursorProtobuf.js:8 (extractTextFromResponse)
func extractTextFromResponse(payload []byte) (text string, toolCalls []map[string]interface{}, errMsg string) {
	// Simple protobuf parsing for response text
	// Field 1 (RESPONSE_TEXT) contains the text
	// Field 25 (THINKING) contains thinking content
	offset := 0

	for offset < len(payload) {
		if offset >= len(payload) {
			break
		}

		// Read tag
		tag := payload[offset]
		offset++

		if offset >= len(payload) {
			break
		}

		fieldNum := int(tag >> 3)
		wt := int(tag & 0x07)

		switch wt {
		case int(wireVarint):
			// Skip varint
			for offset < len(payload) && payload[offset]&0x80 != 0 {
				offset++
			}
			if offset < len(payload) {
				offset++
			}

		case int(wireLen):
			if offset >= len(payload) {
				break
			}
			// Read length
			length := int(payload[offset])
			offset++
			if length&0x80 != 0 {
				// Multi-byte length
				length = int(payload[offset-1] & 0x7f)
				shift := 7
				for offset < len(payload) && payload[offset-1]&0x80 != 0 {
					length |= int(payload[offset]&0x7f) << shift
					offset++
					shift += 7
				}
			}

			if offset+length > len(payload) {
				break
			}

			data := payload[offset : offset+length]
			offset += length

			// Field 1 is response text, field 25 is thinking
			if fieldNum == 1 || fieldNum == 25 {
				text += string(data)
			}
		}
	}

	return text, toolCalls, ""
}

// PrepareRequest modifies the outgoing request for Cursor API.
// ref: open-sse/executors/cursor.js:129-139
func (e *CursorExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Extract credentials from context or request headers
	accessToken := req.Header.Get("authorization")
	if strings.HasPrefix(accessToken, "Bearer ") {
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	}

	// Get machine ID and ghost mode from headers or use defaults
	machineID := req.Header.Get("x-cursor-machine-id")
	ghostMode := req.Header.Get("x-ghost-mode") != "false"

	// Build Cursor-specific headers
	cursorHeaders := buildCursorHeaders(accessToken, machineID, ghostMode)

	// Merge headers
	for key, values := range cursorHeaders {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	// Parse the incoming request body and transform to Cursor protobuf format
	var requestBody map[string]interface{}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	model, _ := requestBody["model"].(string)
	messagesRaw, _ := requestBody["messages"].([]interface{})
	toolsRaw, _ := requestBody["tools"].([]interface{})
	reasoningEffort, _ := requestBody["reasoning_effort"].(string)

	// Convert messages
	messages := make([]map[string]interface{}, len(messagesRaw))
	for i, m := range messagesRaw {
		if msg, ok := m.(map[string]interface{}); ok {
			messages[i] = msg
		}
	}

	// Convert tools
	tools := make([]map[string]interface{}, len(toolsRaw))
	for i, t := range toolsRaw {
		if tool, ok := t.(map[string]interface{}); ok {
			tools[i] = tool
		}
	}

	// Generate Cursor protobuf body
	transformedBody := generateCursorBody(messages, model, tools, reasoningEffort, false)

	// Set the transformed body
	req.Body = io.NopCloser(bytes.NewReader(transformedBody))
	req.ContentLength = int64(len(transformedBody))

	return nil
}

// TransformResponse transforms the response from Cursor API.
// ref: open-sse/executors/cursor.js:241-244
func (e *CursorExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return json.Marshal(map[string]interface{}{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("[%d]: %s", resp.StatusCode, string(body)),
				"type":    "invalid_request_error",
				"code":    "",
			},
		})
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse ConnectRPC frames and extract content
	var totalContent string
	var allToolCalls []map[string]interface{}
	offset := 0

	for offset < len(body) {
		flags, _, payload, newOffset, err := parseConnectRPCFrame(body, offset)
		if err != nil {
			break
		}
		offset = newOffset

		// Decompress if needed
		payload = decompressPayload(payload, flags)

		// Check for JSON error
		if len(payload) > 0 && payload[0] == 0x7b {
			if bytes.Contains(payload, []byte(`"error"`)) {
				// Return error if we haven't accumulated content
				if totalContent == "" && len(allToolCalls) == 0 {
					return payload, nil
				}
				break
			}
		}

		// Extract text from protobuf response
		text, toolCalls, errMsg := extractTextFromResponse(payload)
		if errMsg != "" && totalContent == "" {
			return json.Marshal(map[string]interface{}{
				"error": map[string]interface{}{
					"message": errMsg,
					"type":    "rate_limit_error",
					"code":    "rate_limited",
				},
			})
		}

		totalContent += text
		allToolCalls = append(allToolCalls, toolCalls...)
	}

	// Build OpenAI-compatible response
	response := map[string]interface{}{
		"id":      "chatcmpl-cursor-" + uuid.New().String()[:8],
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   "cursor",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": totalContent,
				},
				"finish_reason": "stop",
			},
		},
	}

	if len(allToolCalls) > 0 {
		response["choices"].([]map[string]interface{})[0]["message"].(map[string]interface{})["tool_calls"] = allToolCalls
	}

	return json.Marshal(response)
}

// HandleError processes errors from Cursor API.
// ref: open-sse/executors/cursor.js:246-258
func (e *CursorExecutor) HandleError(ctx context.Context, err error) error {
	return fmt.Errorf("cursor API error: %w", err)
}

// init registers the Cursor executor.
