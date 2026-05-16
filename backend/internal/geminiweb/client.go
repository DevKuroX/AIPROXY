package geminiweb

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SendChat sends a chat request to Gemini and returns the parsed response.
// For streaming, use SendChatStream instead.
func (s *Session) SendChat(prompt string, modelName string) (*GeminiResponse, error) {
	// Ensure authenticated
	if !s.IsAuthenticated() {
		if err := s.Init(); err != nil {
			return nil, fmt.Errorf("auth failed: %w", err)
		}
	}

	if s.IsTokenExpired() {
		if err := s.RefreshAccessToken(); err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
	}

	url, body, err := s.BuildChatPayload(prompt, modelName)
	if err != nil {
		return nil, fmt.Errorf("build payload failed: %w", err)
	}

	uid := strings.ToUpper(uuid.New().String())
	headers := BuildRequestHeaders(modelName, uid)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := s.client.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini returned %d: %s", resp.StatusCode, string(errBody))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	return ParseNonStreamingResponse(bodyBytes)
}

// SendChatStream sends a streaming chat request to Gemini.
// Returns parsed chunks via the callback function.
func (s *Session) SendChatStream(prompt string, modelName string, onChunk func(GeminiResponse)) error {
	if !s.IsAuthenticated() {
		if err := s.Init(); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	if s.IsTokenExpired() {
		if err := s.RefreshAccessToken(); err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
	}

	url, body, err := s.BuildChatPayload(prompt, modelName)
	if err != nil {
		return fmt.Errorf("build payload failed: %w", err)
	}

	uid := strings.ToUpper(uuid.New().String())
	headers := BuildRequestHeaders(modelName, uid)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	// Use longer timeout for streaming
	s.client.client.Timeout = 180 * time.Second

	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("stream request failed: %w", err)
	}

	if resp.StatusCode != 200 {
		errBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("gemini returned %d: %s", resp.StatusCode, string(errBody))
	}

	chunkChan := make(chan GeminiResponse, 100)
	errChan := make(chan error, 1)

	go func() {
		errChan <- ParseStream(resp.Body, chunkChan)
	}()

	for chunk := range chunkChan {
		onChunk(chunk)
	}

	err = <-errChan
	resp.Body.Close()
	return err
}
