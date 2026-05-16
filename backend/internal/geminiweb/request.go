package geminiweb

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// BuildChatPayload builds the form-encoded request payload for a Gemini chat request.
func (s *Session) BuildChatPayload(prompt string, modelName string) (string, string, error) {
	modelCfg, ok := FreeModels[modelName]
	if !ok {
		// Default to flash if model not found
		modelCfg = FreeModels["gemini-3-flash"]
	}

	s.reqCounter += 100000
	reqID := s.reqCounter

	messageContent := []interface{}{
		prompt,    // [0] text prompt
		0,         // [1] 
		nil,       // [2]
		nil,       // [3] file data
		nil,       // [4]
		nil,       // [5]
		0,         // [6]
	}

	// Build the 69-element inner request array
	innerReq := make([]interface{}, 69)
	innerReq[0] = messageContent
	innerReq[1] = []interface{}{s.Language}
	innerReq[2] = defaultMetadata
	innerReq[6] = []interface{}{1}
	innerReq[streamingFlagIndex] = 1
	innerReq[10] = 1
	innerReq[11] = 0
	innerReq[17] = [][]int{{0}}
	innerReq[18] = 0
	innerReq[27] = 1
	innerReq[30] = []int{4}
	innerReq[41] = []int{1}
	innerReq[53] = 0
	innerReq[61] = []interface{}{}
	innerReq[68] = 2

	// Generate UUID for the request
	uid := strings.ToUpper(uuid.New().String())
	innerReq[59] = uid

	innerJSON, err := json.Marshal(innerReq)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal inner req: %w", err)
	}
	freqPayload := []interface{}{
		nil,
		string(innerJSON),
	}
	freqJSON, err := json.Marshal(freqPayload)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal f.req: %w", err)
	}

	// Build form data
	formData := url.Values{}
	formData.Set("at", s.AccessToken)
	formData.Set("f.req", string(freqJSON))

	// Build model header
	modelHeader := fmt.Sprintf(`[1,null,null,null,"%s",null,null,0,[4],null,null,%d]`, modelCfg.ModelID, modelCfg.Capacity)

	// Build URL with query params
	queryParams := url.Values{}
	queryParams.Set("hl", s.Language)
	queryParams.Set("_reqid", fmt.Sprintf("%d", reqID))
	queryParams.Set("rt", "c")
	if s.BuildLabel != "" {
		queryParams.Set("bl", s.BuildLabel)
	}
	if s.SessionID != "" {
		queryParams.Set("f.sid", s.SessionID)
	}

	requestURL := endpointGenerate + "?" + queryParams.Encode()
	body := formData.Encode()

	// Build headers (returned as part of the URL for caller to use)
	_ = modelHeader // used by caller for headers

	return requestURL, body, nil
}

// BuildRequestHeaders returns the headers needed for the Gemini API request.
func BuildRequestHeaders(modelName string, uuidStr string) map[string]string {
	modelCfg, ok := FreeModels[modelName]
	if !ok {
		modelCfg = FreeModels["gemini-3-flash"]
	}

	modelHeader := fmt.Sprintf(`[1,null,null,null,"%s",null,null,0,[4],null,null,%d]`, modelCfg.ModelID, modelCfg.Capacity)

	return map[string]string{
		"Content-Type":                "application/x-www-form-urlencoded;charset=utf-8",
		"Origin":                      "https://gemini.google.com",
		"Referer":                     "https://gemini.google.com/",
		"x-goog-ext-525001261-jspb":   modelHeader,
		"x-goog-ext-73010989-jspb":    "[0]",
		"x-goog-ext-73010990-jspb":    "[0]",
		"x-goog-ext-525005358-jspb":   fmt.Sprintf(`["%s",1]`, uuidStr),
		"X-Same-Domain":               "1",
	}
}
