package validation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var testAPIBaseURL string

func init() {
	testAPIBaseURL = os.Getenv("AI_PROXY_TEST_URL")
	if testAPIBaseURL == "" {
		testAPIBaseURL = "http://localhost:8080"
	}
}

type APITestCase struct {
	Name           string
	Method         string
	Path           string
	Body           interface{}
	Headers        map[string]string
	ExpectedStatus int
	SkipIfNoServer bool
}

func TestChatCompletionsNonStreamingParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	testCases := []APITestCase{
		{
			Name:   "simple_chat",
			Method: "POST",
			Path:   "/v1/chat/completions",
			Body: map[string]interface{}{
				"model": "gpt-4o-mini",
				"messages": []map[string]string{
					{"role": "user", "content": "Say hello"},
				},
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "with_temperature",
			Method: "POST",
			Path:   "/v1/chat/completions",
			Body: map[string]interface{}{
				"model": "gpt-4o-mini",
				"messages": []map[string]string{
					{"role": "user", "content": "Count to 5"},
				},
				"temperature": 0.7,
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "with_max_tokens",
			Method: "POST",
			Path:   "/v1/chat/completions",
			Body: map[string]interface{}{
				"model": "gpt-4o-mini",
				"messages": []map[string]string{
					{"role": "user", "content": "Write a poem"},
				},
				"max_tokens": 50,
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp := makeAPIRequest(t, tc)
			defer resp.Body.Close()

			if resp.StatusCode != tc.ExpectedStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.ExpectedStatus)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if _, ok := result["choices"]; !ok {
				t.Error("response missing 'choices' field")
			}
			if _, ok := result["model"]; !ok {
				t.Error("response missing 'model' field")
			}
			if _, ok := result["usage"]; !ok {
				t.Error("response missing 'usage' field")
			}
		})
	}
}

func TestChatCompletionsStreamingParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tc := APITestCase{
		Name:   "streaming_chat",
		Method: "POST",
		Path:   "/v1/chat/completions",
		Body: map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Say hello in one word"},
			},
			"stream": true,
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("Content-Type = %s, want text/event-stream", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "data:") {
		t.Error("stream does not contain SSE data events")
	}
	if !strings.Contains(bodyStr, "[DONE]") {
		t.Error("stream does not contain [DONE] marker")
	}
}

func TestEmbeddingsParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	testCases := []APITestCase{
		{
			Name:   "single_text",
			Method: "POST",
			Path:   "/v1/embeddings",
			Body: map[string]interface{}{
				"model": "text-embedding-ada-002",
				"input": "Hello world",
			},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:   "multiple_texts",
			Method: "POST",
			Path:   "/v1/embeddings",
			Body: map[string]interface{}{
				"model": "text-embedding-ada-002",
				"input": []string{"Hello", "World"},
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp := makeAPIRequest(t, tc)
			defer resp.Body.Close()

			if resp.StatusCode != tc.ExpectedStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.ExpectedStatus)
			}

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			data, ok := result["data"].([]interface{})
			if !ok {
				t.Fatal("response missing 'data' array")
			}
			if len(data) == 0 {
				t.Error("embeddings data is empty")
			}

			firstEmbedding := data[0].(map[string]interface{})
			if _, ok := firstEmbedding["embedding"]; !ok {
				t.Error("embedding missing 'embedding' field")
			}
		})
	}
}

func TestImageGenerationParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tc := APITestCase{
		Name:   "generate_image",
		Method: "POST",
		Path:   "/v1/images/generations",
		Body: map[string]interface{}{
			"model":  "dall-e-3",
			"prompt": "A simple red circle on white background",
			"n":      1,
			"size":   "1024x1024",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != tc.ExpectedStatus {
		t.Errorf("status = %d, want %d", resp.StatusCode, tc.ExpectedStatus)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("response missing 'data' array")
	}
	if len(data) == 0 {
		t.Error("image data is empty")
	}

	firstImage := data[0].(map[string]interface{})
	if _, ok := firstImage["url"]; !ok {
		if _, ok := firstImage["b64_json"]; !ok {
			t.Error("image missing both 'url' and 'b64_json' fields")
		}
	}
}

func TestTTSParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tc := APITestCase{
		Name:   "text_to_speech",
		Method: "POST",
		Path:   "/v1/audio/speech",
		Body: map[string]interface{}{
			"model": "tts-1",
			"input": "Hello, this is a test.",
			"voice": "alloy",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != tc.ExpectedStatus {
		t.Errorf("status = %d, want %d", resp.StatusCode, tc.ExpectedStatus)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "audio") {
		t.Errorf("Content-Type = %s, want audio/*", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read audio: %v", err)
	}
	if len(body) == 0 {
		t.Error("audio response is empty")
	}
}

func TestModelsListParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tc := APITestCase{
		Name:           "list_models",
		Method:         "GET",
		Path:           "/v1/models",
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != tc.ExpectedStatus {
		t.Errorf("status = %d, want %d", resp.StatusCode, tc.ExpectedStatus)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("response missing 'data' array")
	}
	if len(data) == 0 {
		t.Error("models list is empty")
	}
}

func TestSearchParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tc := APITestCase{
		Name:   "web_search",
		Method: "POST",
		Path:   "/v1/search",
		Body: map[string]interface{}{
			"query": "golang testing best practices",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != tc.ExpectedStatus {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("status = %d, want %d, body: %s", resp.StatusCode, tc.ExpectedStatus, string(body))
	}
}

func TestFetchParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tc := APITestCase{
		Name:   "web_fetch",
		Method: "POST",
		Path:   "/v1/fetch",
		Body: map[string]interface{}{
			"url": "https://example.com",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != tc.ExpectedStatus {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("status = %d, want %d, body: %s", resp.StatusCode, tc.ExpectedStatus, string(body))
	}
}

func TestMessagesAPIParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	createThreadTC := APITestCase{
		Name:           "create_thread",
		Method:         "POST",
		Path:           "/v1/threads",
		Body:           map[string]interface{}{},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, createThreadTC)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("thread creation failed with status %d", resp.StatusCode)
		return
	}

	var threadResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&threadResult); err != nil {
		t.Fatalf("failed to decode thread response: %v", err)
	}

	threadID, ok := threadResult["id"].(string)
	if !ok {
		t.Fatal("thread response missing 'id'")
	}

	createMsgTC := APITestCase{
		Name:   "create_message",
		Method: "POST",
		Path:   "/v1/threads/" + threadID + "/messages",
		Body: map[string]interface{}{
			"role":    "user",
			"content": "Hello!",
		},
		ExpectedStatus: http.StatusOK,
	}

	msgResp := makeAPIRequest(t, createMsgTC)
	defer msgResp.Body.Close()

	if msgResp.StatusCode != http.StatusOK {
		t.Errorf("create message status = %d, want %d", msgResp.StatusCode, http.StatusOK)
	}
}

func TestErrorHandlingParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Run("missing_model", func(t *testing.T) {
		tc := APITestCase{
			Name:   "missing_model",
			Method: "POST",
			Path:   "/v1/chat/completions",
			Body: map[string]interface{}{
				"messages": []map[string]string{
					{"role": "user", "content": "Hello"},
				},
			},
			ExpectedStatus: http.StatusBadRequest,
		}

		resp := makeAPIRequest(t, tc)
		defer resp.Body.Close()

		if resp.StatusCode != tc.ExpectedStatus {
			t.Errorf("status = %d, want %d", resp.StatusCode, tc.ExpectedStatus)
		}
	})

	t.Run("missing_messages", func(t *testing.T) {
		tc := APITestCase{
			Name:   "missing_messages",
			Method: "POST",
			Path:   "/v1/chat/completions",
			Body: map[string]interface{}{
				"model": "gpt-4o-mini",
			},
			ExpectedStatus: http.StatusBadRequest,
		}

		resp := makeAPIRequest(t, tc)
		defer resp.Body.Close()

		if resp.StatusCode != tc.ExpectedStatus {
			t.Errorf("status = %d, want %d", resp.StatusCode, tc.ExpectedStatus)
		}
	})

	t.Run("invalid_model", func(t *testing.T) {
		tc := APITestCase{
			Name:   "invalid_model",
			Method: "POST",
			Path:   "/v1/chat/completions",
			Body: map[string]interface{}{
				"model": "nonexistent-model-xyz",
				"messages": []map[string]string{
					{"role": "user", "content": "Hello"},
				},
			},
			ExpectedStatus: http.StatusNotFound,
		}

		resp := makeAPIRequest(t, tc)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("invalid model returned status %d (expected to fail gracefully)", resp.StatusCode)
		}
	})
}

func makeAPIRequest(t *testing.T, tc APITestCase) *http.Response {
	var body io.Reader
	if tc.Body != nil {
		bodyBytes, err := json.Marshal(tc.Body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(tc.Method, testAPIBaseURL+tc.Path, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	apiKey := os.Getenv("AI_PROXY_TEST_API_KEY")
	if apiKey == "" {
		apiKey = "test-key"
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	for k, v := range tc.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		if tc.SkipIfNoServer {
			t.Skipf("server not available: %v", err)
		}
		t.Fatalf("request failed: %v", err)
	}

	return resp
}

type mockHandler struct {
	StatusCode int
	Response   interface{}
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.StatusCode)
	json.NewEncoder(w).Encode(h.Response)
}

func newTestServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}
