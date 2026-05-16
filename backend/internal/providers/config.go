package providers

// Kiro API endpoints from 9router
const (
	KIRO_REFRESH_URL = "https://prod.us-east-1.auth.desktop.kiro.dev/refreshToken"
	KIRO_USAGE_URL   = "https://codewhisperer.us-east-1.amazonaws.com/getUsageLimits"
)

// Default context window in tokens (128K)
const DefaultContextWindow = 128000

// ProviderConfig represents the configuration for an AI provider
type ProviderConfig struct {
	Name          string
	Type          string
	BaseURL       string
	AuthType      string
	Format        string
	Headers       map[string]string
	ContextWindow int // max context tokens, 0 = use DefaultContextWindow
}

// GetContextWindow returns the effective context window for this provider
func (p ProviderConfig) GetContextWindow() int {
	if p.ContextWindow > 0 {
		return p.ContextWindow
	}
	return DefaultContextWindow
}

const (
	TypeKiro     = "kiro"
	TypeOpenAI   = "openai"
	TypeClaude   = "claude"
	TypeGemini   = "gemini"
	TypeCodex    = "codex"
	TypeGitHub   = "github"
	TypeAntigrav = "antigravity"
	TypeOllama   = "ollama"
)

const (
	AuthTypeOAuth   = "oauth"
	AuthTypeBearer  = "bearer"
	AuthTypeNone    = "none"
	AuthTypeCookie  = "cookie"
	AuthTypeAPIKey  = "apikey"
)

const (
	FormatGeminiCLI      = "gemini-cli"
	FormatClaude         = "claude"
	FormatOpenAI         = "openai"
	FormatGemini         = "gemini"
	FormatKiro           = "kiro"
	FormatAntigravity    = "antigravity"
	FormatCursor         = "cursor"
	FormatOllama         = "ollama"
	FormatVertex         = "vertex"
	FormatGrokWeb        = "grok-web"
	FormatPerplexityWeb  = "perplexity-web"
	FormatCommandCode    = "commandcode"
	FormatCodexResponses = "openai-responses"
	FormatGeminiWeb      = "gemini-web"
)

var CLAUDE_API_HEADERS = map[string]string{
	"Anthropic-Version": "2023-06-01",
	"Anthropic-Beta":    "claude-code-20250219,interleaved-thinking-2025-05-14",
}

// PROVIDERS is the map of all available provider configurations
var PROVIDERS = map[string]ProviderConfig{
	// === API Key Providers (OpenAI format) ===
	"openai": {
		Name: "OpenAI", Type: TypeOpenAI,
		BaseURL: "https://api.openai.com/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"anthropic": {
		Name: "Anthropic", Type: TypeClaude, ContextWindow: 200000,
		BaseURL: "https://api.anthropic.com/v1/messages", AuthType: AuthTypeAPIKey, Format: FormatClaude,
		Headers: CLAUDE_API_HEADERS,
	},
	"openrouter": {
		Name: "OpenRouter", Type: TypeOpenAI,
		BaseURL: "https://openrouter.ai/api/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
		Headers: map[string]string{"HTTP-Referer": "https://endpoint-proxy.local", "X-Title": "Endpoint Proxy"},
	},
	"deepseek": {
		Name: "DeepSeek", Type: TypeOpenAI,
		BaseURL: "https://api.deepseek.com/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"groq": {
		Name: "Groq", Type: TypeOpenAI,
		BaseURL: "https://api.groq.com/openai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"xai": {
		Name: "xAI", Type: TypeOpenAI,
		BaseURL: "https://api.x.ai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"mistral": {
		Name: "Mistral", Type: TypeOpenAI,
		BaseURL: "https://api.mistral.ai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"perplexity": {
		Name: "Perplexity", Type: TypeOpenAI,
		BaseURL: "https://api.perplexity.ai/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"together": {
		Name: "Together AI", Type: TypeOpenAI,
		BaseURL: "https://api.together.xyz/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"fireworks": {
		Name: "Fireworks AI", Type: TypeOpenAI,
		BaseURL: "https://api.fireworks.ai/inference/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"cerebras": {
		Name: "Cerebras", Type: TypeOpenAI,
		BaseURL: "https://api.cerebras.ai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"cohere": {
		Name: "Cohere", Type: TypeOpenAI,
		BaseURL: "https://api.cohere.ai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"nebius": {
		Name: "Nebius", Type: TypeOpenAI,
		BaseURL: "https://api.studio.nebius.ai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"siliconflow": {
		Name: "SiliconFlow", Type: TypeOpenAI,
		BaseURL: "https://api.siliconflow.cn/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"hyperbolic": {
		Name: "Hyperbolic", Type: TypeOpenAI,
		BaseURL: "https://api.hyperbolic.xyz/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"nanobanana": {
		Name: "NanoBanana", Type: TypeOpenAI,
		BaseURL: "https://api.nanobananaapi.ai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"chutes": {
		Name: "Chutes", Type: TypeOpenAI,
		BaseURL: "https://llm.chutes.ai/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"nvidia": {
		Name: "NVIDIA", Type: TypeOpenAI,
		BaseURL: "https://integrate.api.nvidia.com/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"kilocode": {
		Name: "KiloCode", Type: TypeOpenAI,
		BaseURL: "https://api.kilo.ai/api/openrouter/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"azure": {
		Name: "Azure OpenAI", Type: TypeOpenAI,
		BaseURL: "https://{resource}.openai.azure.com/openai/deployments/{deployment}/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"alicode": {
		Name: "Aliyun Code", Type: TypeOpenAI,
		BaseURL: "https://coding.dashscope.aliyuncs.com/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"alicode-intl": {
		Name: "Aliyun Code Intl", Type: TypeOpenAI,
		BaseURL: "https://coding-intl.dashscope.aliyuncs.com/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"volcengine-ark": {
		Name: "Volcengine ARK", Type: TypeOpenAI,
		BaseURL: "https://ark.cn-beijing.volces.com/api/coding/v3/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"byteplus": {
		Name: "BytePlus", Type: TypeOpenAI,
		BaseURL: "https://ark.ap-southeast.bytepluses.com/api/coding/v3/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"gitlab": {
		Name: "GitLab Duo", Type: TypeOpenAI,
		BaseURL: "https://gitlab.com/api/v4/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"codebuddy": {
		Name: "CodeBuddy", Type: TypeOpenAI,
		BaseURL: "https://copilot.tencent.com/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"opencode-go": {
		Name: "OpenCode Go", Type: TypeOpenAI,
		BaseURL: "https://opencode.ai/zen/go/v1/chat/completions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"vertex-partner": {
		Name: "Vertex AI Partner", Type: TypeOpenAI,
		BaseURL: "https://aiplatform.googleapis.com", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},

	// === OAuth Providers (Claude format) ===
	"claude": {
		Name: "Anthropic Claude", Type: TypeClaude, ContextWindow: 200000,
		BaseURL: "https://api.anthropic.com/v1/messages", AuthType: AuthTypeOAuth, Format: FormatClaude,
		Headers: map[string]string{
			"Anthropic-Version": "2023-06-01",
			"Anthropic-Beta":    "claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14",
		},
	},
	"glm": {
		Name: "GLM", Type: TypeOpenAI,
		BaseURL: "https://api.z.ai/api/anthropic/v1/messages", AuthType: AuthTypeOAuth, Format: FormatClaude,
		Headers: CLAUDE_API_HEADERS,
	},
	"glm-cn": {
		Name: "GLM CN", Type: TypeOpenAI,
		BaseURL: "https://open.bigmodel.cn/api/coding/paas/v4/chat/completions", AuthType: AuthTypeOAuth, Format: FormatOpenAI,
	},
	"kimi": {
		Name: "Kimi", Type: TypeOpenAI,
		BaseURL: "https://api.kimi.com/coding/v1/messages", AuthType: AuthTypeOAuth, Format: FormatClaude,
		Headers: CLAUDE_API_HEADERS,
	},
	"minimax": {
		Name: "MiniMax", Type: TypeOpenAI,
		BaseURL: "https://api.minimax.io/anthropic/v1/messages", AuthType: AuthTypeOAuth, Format: FormatClaude,
		Headers: CLAUDE_API_HEADERS,
	},
	"minimax-cn": {
		Name: "MiniMax CN", Type: TypeOpenAI,
		BaseURL: "https://api.minimaxi.com/anthropic/v1/messages", AuthType: AuthTypeOAuth, Format: FormatClaude,
		Headers: CLAUDE_API_HEADERS,
	},
	// === OAuth Providers (OpenAI format) ===
	"qwen": {
		Name: "Qwen Code", Type: TypeOpenAI,
		BaseURL: "https://portal.qwen.ai/v1/chat/completions", AuthType: AuthTypeOAuth, Format: FormatOpenAI,
	},
	"iflow": {
		Name: "iFlow", Type: TypeOpenAI,
		BaseURL: "https://apis.iflow.cn/v1/chat/completions", AuthType: AuthTypeOAuth, Format: FormatOpenAI,
		Headers: map[string]string{"User-Agent": "iFlow-Cli"},
	},
	"qoder": {
		Name: "Qoder", Type: TypeOpenAI,
		BaseURL: "https://api.qoder.com/v1/chat/completions", AuthType: AuthTypeOAuth, Format: FormatOpenAI,
		Headers: map[string]string{"User-Agent": "Qoder-Cli"},
	},
	"github": {
		Name: "GitHub Copilot", Type: TypeGitHub,
		BaseURL: "https://api.githubcopilot.com/chat/completions", AuthType: AuthTypeOAuth, Format: FormatOpenAI,
		Headers: map[string]string{
			"copilot-integration-id":            "vscode-chat",
			"editor-version":                    "vscode/1.110.0",
			"editor-plugin-version":             "copilot-chat/0.38.0",
			"user-agent":                        "GitHubCopilotChat/0.38.0",
			"openai-intent":                     "conversation-panel",
			"x-github-api-version":              "2025-04-01",
		},
	},
	"cline": {
		Name: "Cline", Type: TypeOpenAI,
		BaseURL: "https://api.cline.bot/api/v1/chat/completions", AuthType: AuthTypeOAuth, Format: FormatOpenAI,
		Headers: map[string]string{"HTTP-Referer": "https://cline.bot", "X-Title": "Cline"},
	},
	"codex": {
		Name: "OpenAI Codex", Type: TypeCodex,
		BaseURL: "https://chatgpt.com/backend-api/codex/responses", AuthType: AuthTypeOAuth, Format: FormatCodexResponses,
		Headers: map[string]string{"originator": "codex-cli", "User-Agent": "codex-cli/1.0.18 (macOS; arm64)"},
	},

	// === Cloud OAuth Providers (Gemini format) ===
	"gemini": {
		Name: "Google Gemini", Type: TypeGemini, ContextWindow: 1000000,
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/models", AuthType: AuthTypeOAuth, Format: FormatGemini,
	},
	"gemini-cli": {
		Name: "Gemini CLI", Type: TypeGemini, ContextWindow: 1000000,
		BaseURL: "https://cloudcode-pa.googleapis.com/v1internal", AuthType: AuthTypeOAuth, Format: FormatGeminiCLI,
	},
	"antigravity": {
		Name: "Antigravity", Type: TypeAntigrav,
		BaseURL: "https://daily-cloudcode-pa.googleapis.com", AuthType: AuthTypeOAuth, Format: FormatAntigravity,
		Headers: map[string]string{"User-Agent": "antigravity/1.107.0"},
	},

	// === Special Executors ===
	"kiro": {
		Name: "AWS CodeWhisperer", Type: TypeKiro, ContextWindow: 200000,
		BaseURL: "https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse", AuthType: AuthTypeOAuth, Format: FormatKiro,
		Headers: map[string]string{
			"X-Amz-Target":     "AmazonCodeWhispererStreamingService.GenerateAssistantResponse",
			"Accept":           "application/vnd.amazon.eventstream",
			"User-Agent":       "AWS-SDK-JS/3.0.0 kiro-ide/1.0.0",
			"X-Amz-User-Agent": "aws-sdk-js/3.0.0 kiro-ide/1.0.0",
		},
	},
	"cursor": {
		Name: "Cursor", Type: TypeOpenAI,
		BaseURL: "https://api2.cursor.sh", AuthType: AuthTypeOAuth, Format: FormatCursor,
		Headers: map[string]string{
			"connect-accept-encoding": "gzip",
			"connect-protocol-version": "1",
			"Content-Type":            "application/connect+proto",
			"User-Agent":              "connect-es/1.6.1",
		},
	},
	"commandcode": {
		Name: "Command Code", Type: TypeOpenAI,
		BaseURL: "https://api.commandcode.ai/alpha/generate", AuthType: AuthTypeBearer, Format: FormatCommandCode,
		Headers: map[string]string{"x-command-code-version": "0.25.7", "x-cli-environment": "cli"},
	},
	"vertex": {
		Name: "Vertex AI", Type: TypeGemini,
		BaseURL: "https://aiplatform.googleapis.com", AuthType: AuthTypeOAuth, Format: FormatVertex,
	},

	// === Local / No-Auth Providers ===
	"ollama": {
		Name: "Ollama", Type: TypeOllama,
		BaseURL: "https://ollama.com/api/chat", AuthType: AuthTypeNone, Format: FormatOllama,
	},
	"ollama-local": {
		Name: "Ollama Local", Type: TypeOllama,
		BaseURL: "http://localhost:11434/api/chat", AuthType: AuthTypeNone, Format: FormatOllama,
	},
	"opencode": {
		Name: "OpenCode Free", Type: TypeOpenAI,
		BaseURL: "https://opencode.ai/zen/v1", AuthType: AuthTypeNone, Format: FormatOpenAI,
		Headers: map[string]string{"x-opencode-client": "desktop", "Authorization": "Bearer public"},
	},
	"oc": { // alias for opencode
		Name: "OpenCode Free", Type: TypeOpenAI,
		BaseURL: "https://opencode.ai/zen/v1", AuthType: AuthTypeNone, Format: FormatOpenAI,
		Headers: map[string]string{"x-opencode-client": "desktop", "Authorization": "Bearer public"},
	},

	// === Cookie-Based Providers ===
	"grok-web": {
		Name: "Grok Web", Type: TypeOpenAI,
		BaseURL: "https://grok.com/rest/app-chat/conversations/new", AuthType: AuthTypeCookie, Format: FormatGrokWeb,
	},
	"perplexity-web": {
		Name: "Perplexity Web", Type: TypeOpenAI,
		BaseURL: "https://www.perplexity.ai/rest/sse/perplexity_ask", AuthType: AuthTypeCookie, Format: FormatPerplexityWeb,
	},
	"gemini-web": {
		Name: "Gemini Web", Type: TypeGemini,
		BaseURL: "https://gemini.google.com", AuthType: AuthTypeCookie, Format: FormatGeminiWeb,
	},

	// === Audio Providers ===
	"deepgram": {
		Name: "Deepgram", Type: TypeOpenAI,
		BaseURL: "https://api.deepgram.com/v1/listen", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
	"assemblyai": {
		Name: "AssemblyAI", Type: TypeOpenAI,
		BaseURL: "https://api.assemblyai.com/v1/audio/transcriptions", AuthType: AuthTypeBearer, Format: FormatOpenAI,
	},
}

// GetProviderConfig retrieves a provider configuration by ID
func GetProviderConfig(providerID string) (ProviderConfig, bool) {
	config, exists := PROVIDERS[providerID]
	return config, exists
}
