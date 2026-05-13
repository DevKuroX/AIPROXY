// Package services provides shared services for the ai_proxy.
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Cloud Code API endpoints.
// ref: open-sse/config/appConstants.js:127-130
const (
	CloudCodeLoadCodeAssist = "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist"
	CloudCodeOnboardUser    = "https://cloudcode-pa.googleapis.com/v1internal:onboardUser"
)

// Cache and cleanup timing constants.
// ref: open-sse/services/projectId.js:17,24,28
const (
	CacheTTL        = 1 * time.Hour
	PendingTTL      = 2 * time.Minute
	CleanupInterval = 10 * time.Minute
)

// IDE and platform constants for Cloud Code API.
// ref: open-sse/config/appConstants.js:21-41,56-60
const (
	IDETypeAntigravity = 9
	PluginTypeGemini   = 2
)

// Platform constants for runtime detection.
// ref: open-sse/config/appConstants.js:28-35
const (
	PlatformUnspecified  = 0
	PlatformDarwinAMD64  = 1
	PlatformDarwinARM64  = 2
	PlatformLinuxAMD64   = 3
	PlatformLinuxARM64   = 4
	PlatformWindowsAMD64 = 5
)

// cacheEntry holds a cached project ID with its fetch timestamp.
// ref: open-sse/services/projectId.js:13-14
type cacheEntry struct {
	ProjectID string
	FetchedAt time.Time
}

// pendingFetch tracks an in-flight project ID fetch.
// ref: open-sse/services/projectId.js:20-21
type pendingFetch struct {
	Result    chan string
	Cancel    context.CancelFunc
	StartedAt time.Time
}

// ProjectIDService fetches and caches real Project IDs from Google Cloud Code API.
// ref: open-sse/services/projectId.js
type ProjectIDService struct {
	httpClient *http.Client
	logger     *slog.Logger

	mu           sync.RWMutex
	cache        map[string]*cacheEntry
	pending      map[string]*pendingFetch
	cleanupTimer *time.Timer
	stopCleanup  chan struct{}
}

// loadCodeAssistRequest is the request body for loadCodeAssist API.
// ref: open-sse/services/projectId.js:162
type loadCodeAssistRequest struct {
	Metadata metadataRequest `json:"metadata"`
}

// metadataRequest holds IDE metadata for Cloud Code API.
// ref: open-sse/config/appConstants.js:139-143
type metadataRequest struct {
	IDEType    int `json:"ideType"`
	Platform   int `json:"platform"`
	PluginType int `json:"pluginType"`
}

// loadCodeAssistResponse is the response from loadCodeAssist API.
// ref: open-sse/services/projectId.js:171-172
type loadCodeAssistResponse struct {
	CloudAICompanionProject interface{} `json:"cloudaicompanionProject,omitempty"`
	AllowedTiers            []tierInfo  `json:"allowedTiers,omitempty"`
}

// tierInfo represents a tier in the loadCodeAssist response.
// ref: open-sse/services/projectId.js:177-186
type tierInfo struct {
	ID        string `json:"id,omitempty"`
	IsDefault bool   `json:"isDefault,omitempty"`
}

// onboardUserRequest is the request body for onboardUser API.
// ref: open-sse/services/projectId.js:202
type onboardUserRequest struct {
	TierID   string          `json:"tierId"`
	Metadata metadataRequest `json:"metadata"`
}

// onboardUserResponse is the response from onboardUser API.
// ref: open-sse/services/projectId.js:230-233
type onboardUserResponse struct {
	Done     bool `json:"done,omitempty"`
	Response struct {
		CloudAICompanionProject interface{} `json:"cloudaicompanionProject,omitempty"`
	} `json:"response,omitempty"`
}

// NewProjectIDService creates a new ProjectIDService.
func NewProjectIDService(httpClient *http.Client, logger *slog.Logger) *ProjectIDService {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if logger == nil {
		logger = slog.Default()
	}

	s := &ProjectIDService{
		httpClient:  httpClient,
		logger:      logger,
		cache:       make(map[string]*cacheEntry),
		pending:     make(map[string]*pendingFetch),
		stopCleanup: make(chan struct{}),
	}

	return s
}

// GetProjectID fetches and caches the project ID for a connection.
// Returns empty string on failure (callers should fall back to random generation).
// ref: open-sse/services/projectId.js:86-122
func (s *ProjectIDService) GetProjectID(ctx context.Context, connectionID, accessToken string) string {
	if connectionID == "" || accessToken == "" {
		return ""
	}

	// Check cache first (read lock)
	s.mu.RLock()
	cached, ok := s.cache[connectionID]
	s.mu.RUnlock()

	if ok && time.Since(cached.FetchedAt) < CacheTTL {
		return cached.ProjectID
	}

	// Check for pending fetch (read lock)
	s.mu.RLock()
	pending, hasPending := s.pending[connectionID]
	s.mu.RUnlock()

	if hasPending {
		// Wait for the pending fetch to complete
		select {
		case result := <-pending.Result:
			return result
		case <-ctx.Done():
			return ""
		}
	}

	// Create a new pending fetch
	resultChan := make(chan string, 1)
	fetchCtx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	// Double-check under write lock
	if p, exists := s.pending[connectionID]; exists {
		s.mu.Unlock()
		cancel() // clean up unused context
		select {
		case result := <-p.Result:
			return result
		case <-ctx.Done():
			return ""
		}
	}

	s.pending[connectionID] = &pendingFetch{
		Result:    resultChan,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	s.mu.Unlock()

	// Perform the fetch
	go func() {
		defer func() {
			cancel()
			s.mu.Lock()
			delete(s.pending, connectionID)
			s.mu.Unlock()
		}()

		projectID := s.fetchProjectID(fetchCtx, accessToken)

		if projectID != "" {
			s.mu.Lock()
			s.cache[connectionID] = &cacheEntry{
				ProjectID: projectID,
				FetchedAt: time.Now(),
			}
			s.mu.Unlock()
		} else {
			s.logger.Warn("could not fetch projectId for connection", "connectionID", connectionID[:min(len(connectionID), 8)])
		}

		select {
		case resultChan <- projectID:
		default:
		}
	}()

	// Wait for result
	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		cancel()
		return ""
	}
}

// InvalidateProjectID removes the cached project ID for a connection.
// Call this when a connection's credentials are fully revoked or refreshed.
// ref: open-sse/services/projectId.js:128-130
func (s *ProjectIDService) InvalidateProjectID(connectionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, connectionID)
}

// RemoveConnection aborts any in-flight fetch and deletes the cached project ID.
// Wire this into connection close / disconnect lifecycle events.
// ref: open-sse/services/projectId.js:138-146
func (s *ProjectIDService) RemoveConnection(connectionID string) {
	if connectionID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cache, connectionID)

	if pending, ok := s.pending[connectionID]; ok {
		pending.Cancel()
		delete(s.pending, connectionID)
	}
}

// CleanupNow runs one sweep immediately: evicts stale cache entries and aborts orphaned pending fetches.
// ref: open-sse/services/projectId.js:33-52
func (s *ProjectIDService) CleanupNow() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Evict stale cache entries
	for id, entry := range s.cache {
		if entry == nil || now.Sub(entry.FetchedAt) >= CacheTTL {
			delete(s.cache, id)
		}
	}

	// Abort orphaned pending fetches
	for id, pending := range s.pending {
		if pending == nil {
			delete(s.pending, id)
			continue
		}
		if now.Sub(pending.StartedAt) > PendingTTL {
			pending.Cancel()
			delete(s.pending, id)
		}
	}
}

// StartCacheCleanup starts the periodic background cleanup (idempotent).
// ref: open-sse/services/projectId.js:55-64
func (s *ProjectIDService) StartCacheCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cleanupTimer != nil {
		return
	}

	s.cleanupTimer = time.NewTimer(CleanupInterval)

	go func() {
		for {
			select {
			case <-s.cleanupTimer.C:
				s.CleanupNow()
				s.mu.Lock()
				if s.cleanupTimer != nil {
					s.cleanupTimer.Reset(CleanupInterval)
				}
				s.mu.Unlock()
			case <-s.stopCleanup:
				return
			}
		}
	}()
}

// StopCacheCleanup stops the periodic background cleanup.
// ref: open-sse/services/projectId.js:67-71
func (s *ProjectIDService) StopCacheCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cleanupTimer == nil {
		return
	}

	close(s.stopCleanup)
	s.cleanupTimer.Stop()
	s.cleanupTimer = nil
}

// fetchProjectID fetches project ID via loadCodeAssist endpoint.
// Falls back to onboardUser when loadCodeAssist returns no project.
// ref: open-sse/services/projectId.js:158-189
func (s *ProjectIDService) fetchProjectID(ctx context.Context, accessToken string) string {
	reqBody := loadCodeAssistRequest{
		Metadata: metadataRequest{
			IDEType:    IDETypeAntigravity,
			Platform:   getPlatformEnum(),
			PluginType: PluginTypeGemini,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.Warn("failed to marshal loadCodeAssist request", "error", err)
		return ""
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, CloudCodeLoadCodeAssist, bytes.NewReader(body))
	if err != nil {
		s.logger.Warn("failed to create loadCodeAssist request", "error", err)
		return ""
	}

	s.setLoadCodeAssistHeaders(req, accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Warn("loadCodeAssist request failed", "error", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("loadCodeAssist failed", "status", resp.StatusCode)
		return ""
	}

	var data loadCodeAssistResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		s.logger.Warn("failed to decode loadCodeAssist response", "error", err)
		return ""
	}

	projectID := extractProjectID(data.CloudAICompanionProject)
	if projectID != "" {
		return projectID
	}

	// Determine the tier to use for onboarding
	// ref: open-sse/services/projectId.js:176-188
	tierID := "legacy-tier"
	for _, tier := range data.AllowedTiers {
		if tier.IsDefault && tier.ID != "" {
			tierID = tier.ID
			break
		}
	}

	return s.onboardUser(ctx, accessToken, tierID)
}

// onboardUser fetches project ID via onboardUser endpoint (polls until done).
// ref: open-sse/services/projectId.js:199-266
func (s *ProjectIDService) onboardUser(ctx context.Context, accessToken, tierID string) string {
	s.logger.Debug("onboarding user with tier", "tierID", tierID)

	reqBody := onboardUserRequest{
		TierID: tierID,
		Metadata: metadataRequest{
			IDEType:    IDETypeAntigravity,
			Platform:   getPlatformEnum(),
			PluginType: PluginTypeGemini,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.Warn("failed to marshal onboardUser request", "error", err)
		return ""
	}

	const maxAttempts = 5

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ""
		default:
		}

		// Per-attempt timeout
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)

		req, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, CloudCodeOnboardUser, bytes.NewReader(body))
		if err != nil {
			cancel()
			s.logger.Warn("failed to create onboardUser request", "error", err)
			continue
		}

		s.setLoadCodeAssistHeaders(req, accessToken)

		resp, err := s.httpClient.Do(req)
		cancel()

		if err != nil {
			s.logger.Warn("onboardUser request failed", "attempt", attempt, "error", err)
			if attempt < maxAttempts {
				time.Sleep(2 * time.Second)
			}
			continue
		}

		var data onboardUserResponse
		err = json.NewDecoder(resp.Body).Decode(&resp.Body)
		resp.Body.Close()

		if err != nil {
			s.logger.Warn("failed to decode onboardUser response", "error", err)
			if attempt < maxAttempts {
				time.Sleep(2 * time.Second)
			}
			continue
		}

		if data.Done {
			projectID := extractProjectID(data.Response.CloudAICompanionProject)
			if projectID != "" {
				s.logger.Debug("successfully onboarded", "projectID", projectID)
				return projectID
			}
			s.logger.Warn("onboardUser done but no project_id in response")
			return ""
		}

		// Server not done yet – wait and retry
		s.logger.Debug("onboard attempt not done, waiting", "attempt", attempt, "maxAttempts", maxAttempts)
		if attempt < maxAttempts {
			time.Sleep(2 * time.Second)
		}
	}

	s.logger.Warn("onboardUser failed after max attempts", "attempts", maxAttempts)
	return ""
}

// setLoadCodeAssistHeaders sets the required headers for Cloud Code API.
// ref: open-sse/config/appConstants.js:132-137
func (s *ProjectIDService) setLoadCodeAssistHeaders(req *http.Request, accessToken string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "google-api-nodejs-client/9.15.1")
	req.Header.Set("X-Goog-Api-Client", "google-cloud-sdk vscode_cloudshelleditor/0.1")
	req.Header.Set("Client-Metadata", fmt.Sprintf(`{"ideType":%d,"platform":%d,"pluginType":%d}`, IDETypeAntigravity, getPlatformEnum(), PluginTypeGemini))
	req.Header.Set("Authorization", "Bearer "+accessToken)
}

// extractProjectID extracts project ID from loadCodeAssist response.
// ref: open-sse/services/projectId.js:271-285
func extractProjectID(project interface{}) string {
	if project == nil {
		return ""
	}

	// Try string type
	if str, ok := project.(string); ok {
		if str != "" {
			return str
		}
		return ""
	}

	// Try object with id field
	if m, ok := project.(map[string]interface{}); ok {
		if id, ok := m["id"].(string); ok && id != "" {
			return id
		}
	}

	return ""
}

// getPlatformEnum returns the platform enum for the current runtime.
// ref: open-sse/config/appConstants.js:43-50
func getPlatformEnum() int {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return PlatformDarwinARM64
		}
		return PlatformDarwinAMD64
	case "linux":
		if runtime.GOARCH == "arm64" {
			return PlatformLinuxARM64
		}
		return PlatformLinuxAMD64
	case "windows":
		return PlatformWindowsAMD64
	default:
		return PlatformUnspecified
	}
}


