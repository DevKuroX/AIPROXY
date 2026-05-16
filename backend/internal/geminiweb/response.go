package geminiweb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// ParseStream reads the Gemini streaming response and returns parsed chunks.
// The response format is newline-delimited JSON arrays.
func ParseStream(reader io.Reader, chunkChan chan<- GeminiResponse) error {
	defer close(chunkChan)

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	buffer := ""
	for scanner.Scan() {
		line := scanner.Text()
		buffer += line

		// Each chunk is a complete JSON line
		if !isCompleteJSON(line) {
			continue
		}

		chunk, err := parseResponseLine(line)
		if err != nil {
			continue // Skip unparseable lines
		}
		chunkChan <- *chunk
	}

	return scanner.Err()
}

func isCompleteJSON(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// Lines from Gemini are typically arrays starting with [
	if !strings.HasPrefix(s, "[") {
		return false
	}
	// Check if it ends with ]
	return strings.HasSuffix(s, "]")
}

// parseResponseLine parses a single line from the Gemini streaming response.
// Format: [,"<code>",null,"<json_string>"]
// The json_string at index [3] contains the actual response data.
func parseResponseLine(line string) (*GeminiResponse, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	// Try to parse as JSON array
	var raw []interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, fmt.Errorf("json parse error: %w", err)
	}

	// The payload is at index 3
	if len(raw) < 4 {
		return nil, fmt.Errorf("unexpected array length: %d", len(raw))
	}

	payloadStr, ok := raw[3].(string)
	if !ok || payloadStr == "" {
		return nil, fmt.Errorf("no payload string at index 3")
	}

	// Parse the inner JSON
	var payload []interface{}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return nil, fmt.Errorf("inner json parse error: %w", err)
	}

	// Extract response data from payload
	// payload[1] contains the main response data
	if len(payload) < 2 {
		return nil, fmt.Errorf("payload too short")
	}

	data, ok := payload[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("payload[1] is not an array")
	}

	// data[0] = text content
	// data[1][0] = first candidate
	// data[4] = candidates list
	result := &GeminiResponse{}

	// Extract text from data[0] (simple case)
	if len(data) > 0 {
		if text, ok := data[0].(string); ok {
			result.Text = text
		}
	}

	// Check for candidates at data[4]
	if len(data) > 4 {
		if candidates, ok := data[4].([]interface{}); ok && len(candidates) > 0 {
			parseCandidates(candidates, result)
		}
	}

	// Check for error codes at payload[5][2][0][1][0]
	if errCode := extractErrorCode(payload); errCode > 0 {
		return nil, fmt.Errorf("gemini error code: %d", errCode)
	}

	return result, nil
}

func parseCandidates(candidates []interface{}, result *GeminiResponse) {
	for _, cand := range candidates {
		candData, ok := cand.([]interface{})
		if !ok || len(candData) < 3 {
			continue
		}

		// candData[0] = rcid
		if rcid, ok := candData[0].(string); ok {
			result.RCID = rcid
		}

		// candData[1] = content array
		if content, ok := candData[1].([]interface{}); ok && len(content) > 0 {
			if text, ok := content[0].(string); ok && text != "" {
				result.Text = text
			}
		}

		// candData[4] = images
		if len(candData) > 4 {
			if images, ok := candData[4].([]interface{}); ok {
				for _, img := range images {
					if imgData, ok := img.([]interface{}); ok && len(imgData) >= 3 {
						imgInfo := ImageInfo{}
						if len(imgData) > 0 {
							if u, ok := imgData[0].(string); ok {
								imgInfo.URL = u
							}
						}
						if len(imgData) > 1 {
							if a, ok := imgData[1].(string); ok {
								imgInfo.Alt = a
							}
						}
						if len(imgData) > 2 {
							if t, ok := imgData[2].(string); ok {
								imgInfo.Title = t
							}
						}
						result.Images = append(result.Images, imgInfo)
					}
				}
			}
		}

		// Check completion indicator at candData[8][0]
		if len(candData) > 8 {
			if indicator, ok := candData[8].([]interface{}); ok && len(indicator) > 0 {
				if val, ok := indicator[0].(float64); ok && val == 1 {
					result.Done = true
				}
			}
		}
	}
}

func extractErrorCode(payload []interface{}) int {
	if len(payload) < 6 {
		return 0
	}
	arr5, ok := payload[5].([]interface{})
	if !ok || len(arr5) < 3 {
		return 0
	}
	arr52, ok := arr5[2].([]interface{})
	if !ok || len(arr52) < 1 {
		return 0
	}
	arr520, ok := arr52[0].([]interface{})
	if !ok || len(arr520) < 2 {
		return 0
	}
	arr5201, ok := arr520[1].([]interface{})
	if !ok || len(arr5201) < 1 {
		return 0
	}
	if code, ok := arr5201[0].(float64); ok {
		return int(code)
	}
	return 0
}

// ParseNonStreamingResponse parses a complete (non-streaming) response body.
func ParseNonStreamingResponse(body []byte) (*GeminiResponse, error) {
	lines := strings.Split(string(body), "\n")
	var lastResult *GeminiResponse

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result, err := parseResponseLine(line)
		if err != nil {
			continue
		}
		if result.Text != "" {
			lastResult = result
		}
	}

	if lastResult == nil {
		return nil, fmt.Errorf("no valid response found")
	}
	return lastResult, nil
}
