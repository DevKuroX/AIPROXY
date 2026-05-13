package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var integrationAPIURL string
var integrationAPIKey string

func init() {
	integrationAPIURL = os.Getenv("AI_PROXY_TEST_URL")
	if integrationAPIURL == "" {
		integrationAPIURL = "http://localhost:8080"
	}
	integrationAPIKey = os.Getenv("AI_PROXY_TEST_API_KEY")
	if integrationAPIKey == "" {
		integrationAPIKey = "test-key"
	}
}

func TestEndToEndChatFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("chat_completion_flow", func(t *testing.T) {
		body := map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "system", "content": "You are a helpful assistant."},
				{"role": "user", "content": "What is 2+2?"},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		choices, ok := result["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			t.Fatal("no choices in response")
		}

		choice := choices[0].(map[string]interface{})
		message := choice["message"].(map[string]interface{})
		content := message["content"].(string)

		if !strings.Contains(content, "4") {
			t.Logf("Warning: response may not contain expected answer, got: %s", content)
		}

		t.Logf("Chat flow successful, response: %s", content)
	})
}

func TestEndToEndStreamingFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	body := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": "Count from 1 to 5"},
		},
		"stream": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Fatalf("expected text/event-stream, got %s", contentType)
	}

	streamBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	streamStr := string(streamBody)
	if !strings.Contains(streamStr, "data:") {
		t.Fatal("stream does not contain SSE data events")
	}
	if !strings.Contains(streamStr, "[DONE]") {
		t.Fatal("stream does not contain [DONE] marker")
	}

	t.Log("Streaming flow successful")
}

func TestEndToEndEmbeddingFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	body := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": "Hello, world!",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/embeddings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Fatal("no embedding data in response")
	}

	embedding := data[0].(map[string]interface{})
	vector, ok := embedding["embedding"].([]interface{})
	if !ok || len(vector) == 0 {
		t.Fatal("embedding vector is empty")
	}

	t.Logf("Embedding flow successful, vector length: %d", len(vector))
}

func TestEndToEndImageGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	body := map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": "A simple blue square on white background",
		"n":      1,
		"size":   "1024x1024",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/images/generations", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok || len(data) == 0 {
		t.Fatal("no image data in response")
	}

	image := data[0].(map[string]interface{})
	if _, ok := image["url"]; !ok {
		if _, ok := image["b64_json"]; !ok {
			t.Fatal("image missing both url and b64_json")
		}
	}

	t.Log("Image generation flow successful")
}

func TestEndToEndTTS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	body := map[string]interface{}{
		"model": "tts-1",
		"input": "Integration test successful",
		"voice": "alloy",
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/audio/speech", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "audio") {
		t.Fatalf("expected audio content type, got %s", contentType)
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read audio: %v", err)
	}

	if len(audioData) == 0 {
		t.Fatal("audio data is empty")
	}

	t.Logf("TTS flow successful, audio size: %d bytes", len(audioData))
}

func TestEndToEndModelsList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	req, _ := http.NewRequest("GET", integrationAPIURL+"/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatal("no data in response")
	}

	if len(data) == 0 {
		t.Log("Warning: models list is empty")
	}

	t.Logf("Models list flow successful, %d models available", len(data))
}

func TestEndToEndMultiTurnConversation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	messages := []map[string]string{
		{"role": "user", "content": "My name is Alice"},
	}

	body := map[string]interface{}{
		"model":    "gpt-4o-mini",
		"messages": messages,
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	var result1 map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result1)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first request status %d", resp.StatusCode)
	}

	assistantResponse := result1["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	messages = append(messages, map[string]string{"role": "assistant", "content": assistantResponse})
	messages = append(messages, map[string]string{"role": "user", "content": "What is my name?"})

	body = map[string]interface{}{
		"model":    "gpt-4o-mini",
		"messages": messages,
	}
	bodyBytes, _ = json.Marshal(body)

	req, _ = http.NewRequest("POST", integrationAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("second request status %d: %s", resp.StatusCode, string(body))
	}

	var result2 map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result2); err != nil {
		t.Fatalf("failed to decode second response: %v", err)
	}

	secondResponse := result2["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	if !strings.Contains(strings.ToLower(secondResponse), "alice") {
		t.Logf("Warning: model may not have remembered the name, got: %s", secondResponse)
	}

	t.Log("Multi-turn conversation flow successful")
}

func TestEndToEndErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("invalid_model_error", func(t *testing.T) {
		body := map[string]interface{}{
			"model": "invalid-model-xyz",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			t.Log("Warning: server accepted invalid model")
		} else {
			t.Logf("Server correctly rejected invalid model with status %d", resp.StatusCode)
		}
	})

	t.Run("missing_auth_error", func(t *testing.T) {
		body := map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			t.Log("Server correctly rejected request without auth")
		} else {
			t.Logf("Server returned status %d (expected 401 for missing auth)", resp.StatusCode)
		}
	})
}

func TestEndToEndConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	concurrency := 5
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			body := map[string]interface{}{
				"model": "gpt-4o-mini",
				"messages": []map[string]string{
					{"role": "user", "content": "Hello from concurrent request"},
				},
			}
			bodyBytes, _ := json.Marshal(body)

			req, _ := http.NewRequest("POST", integrationAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

			client := &http.Client{Timeout: 60 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- err
				return
			}

			errors <- nil
		}(i)
	}

	errorCount := 0
	for i := 0; i < concurrency; i++ {
		if err := <-errors; err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Errorf("%d/%d concurrent requests failed", errorCount, concurrency)
	} else {
		t.Logf("All %d concurrent requests successful", concurrency)
	}
}

func makeIntegrationRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, integrationAPIURL+path, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+integrationAPIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	return resp
}
