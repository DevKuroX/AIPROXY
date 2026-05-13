// ref: _ref/9router/open-sse/config/providers.js
package config

type OAuthEndpoint struct {
	ProviderID    string `json:"provider_id"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	AuthorizeURL  string `json:"authorize_url"`
	TokenURL      string `json:"token_url"`
	RedirectURL   string `json:"redirect_url"`
	Scope         string `json:"scope"`
}

var OAuthEndpoints = map[string]OAuthEndpoint{
	"claude": {
		ProviderID:   "claude",
		ClientID:     "9d1c250a-e61b-44d9-88ed-5944d1962f5e",
		TokenURL:     "https://api.anthr0pic.com/v1/oauth/token",
	},
	"gemini": {
		ProviderID:    "gemini",
		ClientID:      "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com",
		ClientSecret:  "",
		AuthorizeURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:      "https://oauth2.googleapis.com/token",
		RedirectURL:   "http://localhost:20128/oauth/callback/gemini",
		Scope:         "https://www.googleapis.com/auth/cloud-platform",
	},
	"gemini-cli": {
		ProviderID:    "gemini-cli",
		ClientID:      "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com",
		ClientSecret:  "",
		AuthorizeURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:      "https://oauth2.googleapis.com/token",
		RedirectURL:   "http://localhost:20128/oauth/callback/gemini-cli",
		Scope:         "https://www.googleapis.com/auth/cloud-platform",
	},
	"github-copilot": {
		ProviderID:    "github-copilot",
		ClientID:      "Iv1.b507a08c87ecfe98",
		AuthorizeURL:  "https://github.com/login/oauth/authorize",
		TokenURL:      "https://github.com/login/oauth/access_token",
		Scope:         "user",
	},
	"cursor": {
		ProviderID:    "cursor",
		AuthorizeURL:  "https://cursor.sh/login",
		TokenURL:      "https://api2.cursor.sh/auth/token",
	},
}

func GetOAuthEndpoint(providerID string) *OAuthEndpoint {
	if ep, ok := OAuthEndpoints[providerID]; ok {
		return &ep
	}
	return nil
}

func GetOAuthTokenURL(providerID string) string {
	if ep := GetOAuthEndpoint(providerID); ep != nil {
		return ep.TokenURL
	}
	return ""
}

func GetOAuthClientID(providerID string) string {
	if ep := GetOAuthEndpoint(providerID); ep != nil {
		return ep.ClientID
	}
	return ""
}

func GetOAuthClientSecret(providerID string) string {
	if ep := GetOAuthEndpoint(providerID); ep != nil {
		return ep.ClientSecret
	}
	return ""
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token,omitempty"`
}

func BuildTokenURL(providerID string) string {
	switch providerID {
	case "claude":
		return "https://api.anthr0pic.com/v1/oauth/token"
	case "gemini", "gemini-cli":
		return "https://oauth2.googleapis.com/token"
	case "github-copilot":
		return "https://github.com/login/oauth/access_token"
	default:
		return ""
	}
}

func BuildAuthorizeURL(providerID, redirectURI, state string) string {
	switch providerID {
	case "gemini", "gemini-cli":
		return "https://accounts.google.com/o/oauth2/v2/auth?client_id=" + GetOAuthClientID(providerID) + "&redirect_uri=" + redirectURI + "&response_type=code&scope=https://www.googleapis.com/auth/cloud-platform&state=" + state
	case "github-copilot":
		return "https://github.com/login/oauth/authorize?client_id=" + GetOAuthClientID(providerID) + "&redirect_uri=" + redirectURI + "&state=" + state
	default:
		return ""
	}
}
