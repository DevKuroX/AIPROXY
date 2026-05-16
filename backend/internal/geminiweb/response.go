package geminiweb

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf16"
)

// ParseStream reads the Gemini streaming response and returns parsed chunks.
func ParseStream(reader io.Reader, chunkChan chan<- GeminiResponse) error {
	defer close(chunkChan)
	buf := make([]byte, 32*1024)
	var sb strings.Builder
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
			content := sb.String()
			if strings.HasPrefix(content, ")]}'") {
				content = strings.TrimLeft(content[4:], " \t\r\n")
			}
			for {
				frame, remaining, ok := parseNextFrame(content)
				if !ok {
					break
				}
				content = remaining
				chunk, chunkErr := extractResponseFromFrame(frame)
				if chunkErr == nil && chunk != nil {
					chunkChan <- *chunk
					if chunk.Done {
						// Gemini stream complete - break to close channel and return
						// Remaining data in buffer is end marker only, not needed
						return nil
					}
				}
			}
			sb.Reset()
			sb.WriteString(content)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// parseNextFrame extracts the next length-prefixed JSON frame from the buffer.
// Gemini framing: [digits]\n[json_payload] where digits = UTF-16 code unit count of \n + json_payload + trailing newline.
func parseNextFrame(buffer string) (string, string, bool) {
	// Only trim leading whitespace (like Python's isspace() skip).
	// Do NOT trim trailing - the trailing \n is part of the frame length count.
	buffer = strings.TrimLeft(buffer, " \t\r\n")
	if buffer == "" {
		return "", buffer, false
	}

	nl := strings.IndexByte(buffer, '\n')
	if nl < 0 {
		return "", buffer, false
	}

	lengthStr := strings.TrimSpace(buffer[:nl])
	length, err := strconv.Atoi(lengthStr)
	if err != nil || length <= 0 {
		return "", buffer, false
	}

	frameStart := nl
	if frameStart >= len(buffer) {
		return "", buffer, false
	}

	// Measure UTF-16 code units starting from \n (which is counted in the length)
	payload := buffer[frameStart:]
	codeUnits := 0
	bytesConsumed := 0
	for i, r := range payload {
		codeUnits++
		if r >= 0x10000 {
			codeUnits++
		}
		if codeUnits >= length {
			bytesConsumed = i + len(string(r))
			break
		}
	}

	if codeUnits < length {
		return "", buffer, false
	}

	frame := payload[:bytesConsumed]
	if len(frame) > 0 && frame[0] == '\n' {
		frame = frame[1:]
	}
	remaining := payload[bytesConsumed:]
	return frame, remaining, true
}

// extractResponseFromText parses the complete response body (non-streaming).
func extractResponseFromText(body []byte) (*GeminiResponse, error) {
	raw := string(body)
	if strings.HasPrefix(raw, ")]}'") {
		raw = strings.TrimLeft(raw[4:], " \t\r\n")
	}

	// 1. Try frame protocol
	remaining := raw
	var lastResult *GeminiResponse
	for {
		frame, rem, ok := parseNextFrame(remaining)
		if !ok {
			break
		}
		remaining = rem
		chunk, err := extractResponseFromFrame(frame)
		if err == nil && chunk != nil && chunk.Text != "" {
			lastResult = chunk
		}
	}
	if lastResult != nil {
		return lastResult, nil
	}

	// 2. Try full JSON
	var parsed []interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err == nil {
		result := &GeminiResponse{}
		for _, item := range parsed {
			if tuple, ok := item.([]interface{}); ok && len(tuple) >= 1 {
				code, _ := tuple[0].(string)
				switch code {
				case "wra", "wrb.fr":
					if len(tuple) >= 3 {
						if payloadStr, ok := tuple[2].(string); ok && payloadStr != "" {
							parseInnerPayload(payloadStr, result)
						}
					}
				case "di":
					if len(tuple) > 1 {
						if v, ok := tuple[1].(float64); ok && v == 2 {
							result.Done = true
						}
					}
				}
			}
		}
		if result.Text != "" {
			return result, nil
		}
	}

	// 3. Try NDJSON
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var parsedItem interface{}
		if err := json.Unmarshal([]byte(line), &parsedItem); err == nil {
			if tuple, ok := parsedItem.([]interface{}); ok && len(tuple) >= 1 {
				code, _ := tuple[0].(string)
				if code == "wra" || code == "wrb.fr" {
					if len(tuple) >= 3 {
						if payloadStr, ok := tuple[2].(string); ok && payloadStr != "" {
							result := &GeminiResponse{}
							parseInnerPayload(payloadStr, result)
							if result.Text != "" {
								return result, nil
							}
						}
					}
				}
			}
		}
	}

	sample := raw
	if len(sample) > 200 {
		sample = sample[:200]
	}
	return nil, fmt.Errorf("no valid response found (body: %s)", sample)
}

// extractResponseFromFrame parses a single frame (JSON array of response tuples).
func extractResponseFromFrame(frame string) (*GeminiResponse, error) {
	frame = strings.TrimSpace(frame)
	if frame == "" {
		return nil, fmt.Errorf("empty frame")
	}

	var tuples []interface{}
	if err := json.Unmarshal([]byte(frame), &tuples); err != nil {
		return nil, fmt.Errorf("json parse error: %w", err)
	}

	result := &GeminiResponse{}

	for _, t := range tuples {
		tuple, ok := t.([]interface{})
		if !ok || len(tuple) < 1 {
			continue
		}

		code, _ := tuple[0].(string)

		switch code {
		case "wra", "wrb.fr":
			// wrapper: [code, null, "json_string"]
			payloadStr, ok := tuple[2].(string)
			if !ok || payloadStr == "" {
				continue
			}
			parseInnerPayload(payloadStr, result)

		case "di":
			// done indicator: [code, value]
			if len(tuple) > 1 {
				if v, ok := tuple[1].(float64); ok && v >= 2 {
					result.Done = true
				}
			}

		case "e":
			// End of stream frame: ["e",9,null,null,length]
			result.Done = true
		}
	}

	return result, nil
}

// parseInnerPayload extracts text and images from the inner JSON payload.
func parseInnerPayload(payloadStr string, result *GeminiResponse) {
	var payload []interface{}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return
	}
	if len(payload) < 5 {
		return
	}

	candidates, ok := payload[4].([]interface{})
	if !ok {
		return
	}
	for _, cand := range candidates {
		candData, ok := cand.([]interface{})
		if !ok || len(candData) < 2 {
			continue
		}
		if content, ok := candData[1].([]interface{}); ok && len(content) > 0 {
			if text, ok := content[0].(string); ok && text != "" {
				result.Text = text
			}
		}
	}
}

// ParseNonStreamingResponse is the public entry point for non-streaming responses.
func ParseNonStreamingResponse(body []byte) (*GeminiResponse, error) {
	return extractResponseFromText(body)
}

// Helper to estimate UTF-16 string length (JavaScript String.length)
func utf16Len(s string) int {
	count := 0
	for _, r := range s {
		count++
		if r >= 0x10000 {
			count++
		}
	}
	return count
}

// unused alias preserved for compatibility
var _ = utf16.Encode
