// Package geminiweb implements a client for the Google Gemini web app API.
// This is a reverse-engineered protocol used by gemini.google.com.
package geminiweb

import "time"

// Session represents an authenticated Gemini web session.
type Session struct {
	// Cookies for authentication
	Secure1PSID  string
	Secure1PSIDTS string

	// Access token (SNlM0e) extracted from page
	AccessToken string

	// Session metadata from page
	BuildLabel string
	SessionID  string
	Language   string

	// HTTP proxy if needed
	Proxy string

	// Internal state
	client     *httpClient
	reqCounter int
	lastRefresh time.Time
}

// ModelConfig holds the model-specific header configuration.
type ModelConfig struct {
	ModelName string
	ModelID   string
	Capacity  int
}

// Known models for the free tier
var FreeModels = map[string]ModelConfig{
	"gemini-3-flash": {
		ModelName: "gemini-3-flash",
		ModelID:   "fbb127bbb056c959",
		Capacity:  1,
	},
	"gemini-3-flash-thinking": {
		ModelName: "gemini-3-flash-thinking",
		ModelID:   "5bf011840784117a",
		Capacity:  1,
	},
	"gemini-3-pro": {
		ModelName: "gemini-3-pro",
		ModelID:   "9d8ca3786ebdfbea",
		Capacity:  1,
	},
}

// GeminiResponse represents a parsed response from the Gemini API.
type GeminiResponse struct {
	Text     string
	Metadata []string // [cid, rid]
	RCID     string
	Images   []ImageInfo
	Done     bool
}

// ImageInfo holds info about a generated image.
type ImageInfo struct {
	URL  string
	Alt  string
	Title string
}

const (
	endpointGoogle     = "https://www.google.com"
	endpointInit       = "https://gemini.google.com/app"
	endpointGenerate   = "https://gemini.google.com/_/BardChatUi/data/assistant.lamda.BardFrontendService/StreamGenerate"
	endpointRotate     = "https://accounts.google.com/RotateCookies"

	streamingFlagIndex = 7
	gemFlagIndex      = 19
)

var defaultMetadata = []interface{}{"", "", "", nil, nil, nil, nil, nil, nil, ""}
