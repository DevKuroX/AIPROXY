package validation

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestChatCompletionResponseFormat(t *testing.T) {
	expectedFields := []string{
		"id",
		"object",
		"created",
		"model",
		"choices",
		"usage",
	}

	tc := APITestCase{
		Name:   "response_format_check",
		Method: "POST",
		Path:   "/v1/chat/completions",
		Body: map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("server returned %d, skipping format check", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("response missing required field: %s", field)
		}
	}

	t.Run("id_format", func(t *testing.T) {
		id, ok := result["id"].(string)
		if !ok {
			t.Error("id is not a string")
			return
		}
		if len(id) == 0 {
			t.Error("id is empty")
		}
	})

	t.Run("object_type", func(t *testing.T) {
		obj, ok := result["object"].(string)
		if !ok {
			t.Error("object is not a string")
			return
		}
		if obj != "chat.completion" {
			t.Errorf("object = %s, want chat.completion", obj)
		}
	})

	t.Run("choices_structure", func(t *testing.T) {
		choices, ok := result["choices"].([]interface{})
		if !ok {
			t.Fatal("choices is not an array")
		}
		if len(choices) == 0 {
			t.Fatal("choices array is empty")
		}

		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			t.Fatal("choice is not an object")
		}

		if _, ok := choice["index"]; !ok {
			t.Error("choice missing 'index'")
		}
		if _, ok := choice["message"]; !ok {
			t.Error("choice missing 'message'")
		}
		if _, ok := choice["finish_reason"]; !ok {
			t.Error("choice missing 'finish_reason'")
		}

		message, ok := choice["message"].(map[string]interface{})
		if !ok {
			t.Fatal("message is not an object")
		}

		if _, ok := message["role"]; !ok {
			t.Error("message missing 'role'")
		}
		if _, ok := message["content"]; !ok {
			t.Error("message missing 'content'")
		}
	})

	t.Run("usage_structure", func(t *testing.T) {
		usage, ok := result["usage"].(map[string]interface{})
		if !ok {
			t.Fatal("usage is not an object")
		}

		if _, ok := usage["prompt_tokens"]; !ok {
			t.Error("usage missing 'prompt_tokens'")
		}
		if _, ok := usage["completion_tokens"]; !ok {
			t.Error("usage missing 'completion_tokens'")
		}
		if _, ok := usage["total_tokens"]; !ok {
			t.Error("usage missing 'total_tokens'")
		}
	})
}

func TestEmbeddingResponseFormat(t *testing.T) {
	tc := APITestCase{
		Name:   "embedding_format_check",
		Method: "POST",
		Path:   "/v1/embeddings",
		Body: map[string]interface{}{
			"model": "text-embedding-ada-002",
			"input": "test",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("server returned %d, skipping format check", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Run("required_fields", func(t *testing.T) {
		requiredFields := []string{"object", "data", "model", "usage"}
		for _, field := range requiredFields {
			if _, ok := result[field]; !ok {
				t.Errorf("response missing required field: %s", field)
			}
		}
	})

	t.Run("object_type", func(t *testing.T) {
		obj, ok := result["object"].(string)
		if !ok {
			t.Error("object is not a string")
			return
		}
		if obj != "list" {
			t.Errorf("object = %s, want list", obj)
		}
	})

	t.Run("data_structure", func(t *testing.T) {
		data, ok := result["data"].([]interface{})
		if !ok {
			t.Fatal("data is not an array")
		}
		if len(data) == 0 {
			t.Fatal("data array is empty")
		}

		embedding, ok := data[0].(map[string]interface{})
		if !ok {
			t.Fatal("embedding is not an object")
		}

		if _, ok := embedding["object"]; !ok {
			t.Error("embedding missing 'object'")
		}
		if _, ok := embedding["index"]; !ok {
			t.Error("embedding missing 'index'")
		}
		if _, ok := embedding["embedding"]; !ok {
			t.Error("embedding missing 'embedding'")
		}

		embeddingVector, ok := embedding["embedding"].([]interface{})
		if !ok {
			t.Fatal("embedding vector is not an array")
		}
		if len(embeddingVector) == 0 {
			t.Error("embedding vector is empty")
		}
	})
}

func TestImageGenerationResponseFormat(t *testing.T) {
	tc := APITestCase{
		Name:   "image_format_check",
		Method: "POST",
		Path:   "/v1/images/generations",
		Body: map[string]interface{}{
			"model":  "dall-e-3",
			"prompt": "A red circle",
			"n":      1,
			"size":   "1024x1024",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("server returned %d, skipping format check", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Run("required_fields", func(t *testing.T) {
		requiredFields := []string{"created", "data"}
		for _, field := range requiredFields {
			if _, ok := result[field]; !ok {
				t.Errorf("response missing required field: %s", field)
			}
		}
	})

	t.Run("data_structure", func(t *testing.T) {
		data, ok := result["data"].([]interface{})
		if !ok {
			t.Fatal("data is not an array")
		}
		if len(data) == 0 {
			t.Fatal("data array is empty")
		}

		image, ok := data[0].(map[string]interface{})
		if !ok {
			t.Fatal("image is not an object")
		}

		hasURL := false
		hasB64 := false
		if _, ok := image["url"]; ok {
			hasURL = true
		}
		if _, ok := image["b64_json"]; ok {
			hasB64 = true
		}
		if !hasURL && !hasB64 {
			t.Error("image missing both 'url' and 'b64_json'")
		}
	})
}

func TestModelsListResponseFormat(t *testing.T) {
	tc := APITestCase{
		Name:           "models_format_check",
		Method:         "GET",
		Path:           "/v1/models",
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("server returned %d, skipping format check", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Run("required_fields", func(t *testing.T) {
		requiredFields := []string{"object", "data"}
		for _, field := range requiredFields {
			if _, ok := result[field]; !ok {
				t.Errorf("response missing required field: %s", field)
			}
		}
	})

	t.Run("object_type", func(t *testing.T) {
		obj, ok := result["object"].(string)
		if !ok {
			t.Error("object is not a string")
			return
		}
		if obj != "list" {
			t.Errorf("object = %s, want list", obj)
		}
	})

	t.Run("data_structure", func(t *testing.T) {
		data, ok := result["data"].([]interface{})
		if !ok {
			t.Fatal("data is not an array")
		}

		if len(data) > 0 {
			model, ok := data[0].(map[string]interface{})
			if !ok {
				t.Fatal("model is not an object")
			}

			modelFields := []string{"id", "object", "created", "owned_by"}
			for _, field := range modelFields {
				if _, ok := model[field]; !ok {
					t.Errorf("model missing required field: %s", field)
				}
			}
		}
	})
}

func TestErrorResponseFormat(t *testing.T) {
	tc := APITestCase{
		Name:   "error_format_check",
		Method: "POST",
		Path:   "/v1/chat/completions",
		Body: map[string]interface{}{
			"model": "",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		},
		ExpectedStatus: http.StatusBadRequest,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Skip("server accepted invalid request, cannot test error format")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Logf("Response body: %s", string(body))
		t.Skipf("response is not JSON, skipping error format check")
		return
	}

	t.Run("error_field_exists", func(t *testing.T) {
		if _, ok := result["error"]; !ok {
			if _, ok := result["message"]; ok {
				return
			}
			t.Error("error response missing 'error' or 'message' field")
		}
	})
}

func TestStreamingResponseFormat(t *testing.T) {
	tc := APITestCase{
		Name:   "streaming_format_check",
		Method: "POST",
		Path:   "/v1/chat/completions",
		Body: map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Say hello"},
			},
			"stream": true,
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("server returned %d, skipping format check", resp.StatusCode)
		return
	}

	t.Run("content_type", func(t *testing.T) {
		contentType := resp.Header.Get("Content-Type")
		if contentType != "text/event-stream" {
			t.Errorf("Content-Type = %s, want text/event-stream", contentType)
		}
	})

	t.Run("transfer_encoding", func(t *testing.T) {
		transferEncoding := resp.Header.Get("Transfer-Encoding")
		if transferEncoding != "chunked" {
			t.Logf("Transfer-Encoding = %s (chunked preferred for streaming)", transferEncoding)
		}
	})
}

func TestUsageTrackingFormat(t *testing.T) {
	tc := APITestCase{
		Name:   "usage_format_check",
		Method: "POST",
		Path:   "/v1/chat/completions",
		Body: map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Skipf("server returned %d, skipping format check", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	usage, ok := result["usage"].(map[string]interface{})
	if !ok {
		t.Fatal("usage is not an object")
	}

	t.Run("token_counts", func(t *testing.T) {
		promptTokens, ok := usage["prompt_tokens"].(float64)
		if !ok {
			t.Error("prompt_tokens is not a number")
			return
		}
		if promptTokens <= 0 {
			t.Errorf("prompt_tokens = %v, want > 0", promptTokens)
		}

		completionTokens, ok := usage["completion_tokens"].(float64)
		if !ok {
			t.Error("completion_tokens is not a number")
			return
		}
		if completionTokens < 0 {
			t.Errorf("completion_tokens = %v, want >= 0", completionTokens)
		}

		totalTokens, ok := usage["total_tokens"].(float64)
		if !ok {
			t.Error("total_tokens is not a number")
			return
		}

		expectedTotal := promptTokens + completionTokens
		if totalTokens != expectedTotal {
			t.Errorf("total_tokens = %v, want %v", totalTokens, expectedTotal)
		}
	})
}
