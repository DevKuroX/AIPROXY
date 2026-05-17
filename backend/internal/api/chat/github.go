package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/api/middleware"
	"github.com/DevKuroX/AIPROXY/internal/auth/oauth"
	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

type GitHubHandler struct {
	db        *storage.DB
	flow      *oauth.GitHubFlow
}

func NewGitHubHandler(db *storage.DB, clientID, clientSecret string) *GitHubHandler {
	return &GitHubHandler{
		db:   db,
		flow: oauth.NewGitHubFlow(clientID, clientSecret, "repo,user"),
	}
}

type StartAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (h *GitHubHandler) StartAuth(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	deviceCode, err := h.flow.Start(ctx)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := StartAuthResponse{
		DeviceCode:      deviceCode.DeviceCode,
		UserCode:        deviceCode.UserCode,
		VerificationURI: deviceCode.VerificationURI,
		ExpiresIn:       deviceCode.ExpiresIn,
		Interval:        deviceCode.Interval,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type PollAuthRequest struct {
	DeviceCode string `json:"device_code"`
}

func (h *GitHubHandler) PollAuth(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req PollAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceCode == "" {
		errs.WriteJSONError(w, "device_code required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	token, err := h.flow.Poll(ctx, req.DeviceCode)
	if err != nil {
		errStr := err.Error()
		if errStr == "authorization_pending" || errStr == "slow_down" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]string{"status": "pending"})
			return
		}
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "complete",
		"access_token": token.AccessToken,
		"expires_in":   token.ExpiresIn,
	})
}

type GitHubAPIProxyRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Body    json.RawMessage   `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (h *GitHubHandler) ProxyAPI(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req GitHubAPIProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		errs.WriteJSONError(w, "path required", http.StatusBadRequest)
		return
	}

	method := req.Method
	if method == "" {
		method = "GET"
	}

	apiURL := "https://api.github.com" + req.Path

	var bodyReader io.Reader
	if req.Body != nil {
		bodyReader = strings.NewReader(string(req.Body))
	}

	apiReq, err := http.NewRequestWithContext(r.Context(), method, apiURL, bodyReader)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiReq.Header.Set("Accept", "application/vnd.github.v3+json")
	apiReq.Header.Set("User-Agent", "AIPROXY-Chat")
	if req.Body != nil {
		apiReq.Header.Set("Content-Type", "application/json")
	}
	for k, v := range req.Headers {
		apiReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	apiResp, err := client.Do(apiReq)
	if err != nil {
		errs.WriteJSONError(w, fmt.Sprintf("GitHub API error: %v", err), http.StatusBadGateway)
		return
	}
	defer apiResp.Body.Close()

	body, err := io.ReadAll(apiResp.Body)
	if err != nil {
		errs.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiResp.StatusCode)
	w.Write(body)
}
