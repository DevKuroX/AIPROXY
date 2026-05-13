package validation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

var benchmarkAPIURL string
var benchmarkAPIKey string

func init() {
	benchmarkAPIURL = os.Getenv("AI_PROXY_TEST_URL")
	if benchmarkAPIURL == "" {
		benchmarkAPIURL = "http://localhost:8080"
	}
	benchmarkAPIKey = os.Getenv("AI_PROXY_TEST_API_KEY")
	if benchmarkAPIKey == "" {
		benchmarkAPIKey = "test-key"
	}
}

func BenchmarkChatCompletionLatency(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Say hello"},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", benchmarkAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			b.Errorf("request failed: %v", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkStreamingLatency(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Count from 1 to 10"},
			},
			"stream": true,
		}
		bodyBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", benchmarkAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			b.Errorf("request failed: %v", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkEmbeddingLatency(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := map[string]interface{}{
			"model": "text-embedding-ada-002",
			"input": "Hello world benchmark test",
		}
		bodyBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", benchmarkAPIURL+"/v1/embeddings", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			b.Errorf("request failed: %v", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkModelsListLatency(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", benchmarkAPIURL+"/v1/models", nil)
		req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			b.Errorf("request failed: %v", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkConcurrentRequests(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			body := map[string]interface{}{
				"model": "gpt-4o-mini",
				"messages": []map[string]string{
					{"role": "user", "content": "Hello"},
				},
			}
			bodyBytes, _ := json.Marshal(body)

			req, _ := http.NewRequest("POST", benchmarkAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				b.Errorf("request failed: %v", err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}

func BenchmarkThroughput(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	concurrency := 10
	if b.N < concurrency {
		concurrency = b.N
	}

	var wg sync.WaitGroup
	requestsPerWorker := b.N / concurrency
	if requestsPerWorker == 0 {
		requestsPerWorker = 1
	}

	b.ResetTimer()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				body := map[string]interface{}{
					"model": "gpt-4o-mini",
					"messages": []map[string]string{
						{"role": "user", "content": "Say hello"},
					},
				}
				bodyBytes, _ := json.Marshal(body)

				req, _ := http.NewRequest("POST", benchmarkAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

				client := &http.Client{Timeout: 30 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
}

func BenchmarkMemoryUsage(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}

	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	beforeAlloc := m.Alloc

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := map[string]interface{}{
			"model": "gpt-4o-mini",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello"},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", benchmarkAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, _ := client.Do(req)
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m)
	afterAlloc := m.Alloc

	if afterAlloc > beforeAlloc {
		allocPerRequest := (afterAlloc - beforeAlloc) / uint64(b.N)
		b.ReportMetric(float64(allocPerRequest), "bytes/request")
	}
}

func BenchmarkRequestSerialization(b *testing.B) {
	body := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": "This is a test message for serialization benchmark"},
		},
		"temperature": 0.7,
		"max_tokens":  100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(body)
	}
}

func BenchmarkResponseDeserialization(b *testing.B) {
	responseJSON := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1677652288,
		"model": "gpt-4o-mini",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Hello! How can I help you today?"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 9,
			"completion_tokens": 12,
			"total_tokens": 21
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(responseJSON), &result)
	}
}

func BenchmarkLargeRequest(b *testing.B) {
	messages := make([]map[string]string, 100)
	for i := range messages {
		messages[i] = map[string]string{
			"role":    "user",
			"content": "This is a longer test message that simulates a more realistic request with additional context and information.",
		}
	}

	body := map[string]interface{}{
		"model":    "gpt-4o-mini",
		"messages": messages,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bodyBytes, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", benchmarkAPIURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+benchmarkAPIKey)

		client := &http.Client{Timeout: 60 * time.Second}
		resp, _ := client.Do(req)
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}
