package services

import (
	"context"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/models"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

// ref: open-sse/services/model.js:1-89
// Provider alias to ID mapping
var aliasToProviderID = map[string]string{
	// Short aliases
	"cc":  "claude",
	"cx":  "codex",
	"gc":  "gemini-cli",
	"qw":  "qwen",
	"if":  "iflow",
	"ag":  "antigravity",
	"gh":  "github",
	"kr":  "kiro",
	"cu":  "cursor",
	"kc":  "kilocode",
	"kmc": "kimi-coding",
	"cl":  "cline",
	"oc":  "opencode",
	"ocg": "opencode-go",
	// TTS providers
	"el": "elevenlabs",
	// API Key providers
	"openai":     "openai",
	"anthr0pic":  "anthr0pic",
	"gemini":     "gemini",
	"openrouter": "openrouter",
	"glm":        "glm",
	"kimi":       "kimi",
	"minimax":    "minimax",
	"minimax-cn": "minimax-cn",
	"ds":         "deepseek",
	"deepseek":   "deepseek",
	"cmc":        "commandcode",
	"commandcode": "commandcode",
	"groq":       "groq",
	"xai":        "xai",
	"mistral":    "mistral",
	"pplx":       "perplexity",
	"perplexity": "perplexity",
	"together":   "together",
	"fireworks":  "fireworks",
	"cerebras":   "cerebras",
	"cohere":     "cohere",
	"nvidia":     "nvidia",
	"nebius":     "nebius",
	"siliconflow": "siliconflow",
	"hyp":        "hyperbolic",
	"hyperbolic": "hyperbolic",
	"dg":         "deepgram",
	"deepgram":   "deepgram",
	"aai":        "assemblyai",
	"assemblyai": "assemblyai",
	"nb":         "nanobanana",
	"nanobanana": "nanobanana",
	"ch":         "chutes",
	"chutes":     "chutes",
	"ark":        "volcengine-ark",
	"volcengine-ark": "volcengine-ark",
	"byteplus":   "byteplus",
	"bpm":        "byteplus",
	"cursor":     "cursor",
	"vx":         "vertex",
	"vertex":     "vertex",
	"vxp":        "vertex-partner",
	"vertex-partner": "vertex-partner",
	// Web cookie providers
	"gw":          "grok-web",
	"grok-web":    "grok-web",
	"pw":          "perplexity-web",
	"perplexity-web": "perplexity-web",
	"mimo":        "xiaomi-mimo",
	"xiaomi-mimo": "xiaomi-mimo",
	"cf":          "cloudflare-ai",
	"cloudflare-ai": "cloudflare-ai",
	// Image/video providers
	"fal":            "fal-ai",
	"fal-ai":         "fal-ai",
	"stability":      "stability-ai",
	"stability-ai":   "stability-ai",
	"bfl":            "black-forest-labs",
	"black-forest-labs": "black-forest-labs",
	"recraft":        "recraft",
	"topaz":          "topaz",
	"runway":         "runwayml",
	"runwayml":       "runwayml",
	// Embedding/rerank
	"jina":    "jina-ai",
	"jina-ai": "jina-ai",
	// TTS
	"polly":     "aws-polly",
	"aws-polly": "aws-polly",
}

// ModelInfo contains parsed model information
// ref: open-sse/services/model.js:103
type ModelInfo struct {
	Provider     string // Resolved provider ID
	Model        string // Model name
	IsAlias      bool   // True if the model string was an alias (no provider prefix)
	ProviderAlias string // Original provider alias if any
}

// ResolvedModel contains fully resolved model information
type ResolvedModel struct {
	Provider string
	Model    string
}

// ModelService handles model resolution and parsing
type ModelService struct {
	db     *storage.DB
	// aliases could be cached here if needed
}

// NewModelService creates a new model service
func NewModelService(db *storage.DB) *ModelService {
	return &ModelService{db: db}
}

// ResolveProviderAlias resolves a provider alias to its canonical ID
// ref: open-sse/services/model.js:94-96
func (s *ModelService) ResolveProviderAlias(aliasOrID string) string {
	if id, ok := aliasToProviderID[aliasOrID]; ok {
		return id
	}
	return aliasOrID
}

// ParseModel parses a model string in "provider/model" or "alias/model" format
// ref: open-sse/services/model.js:101-122
func (s *ModelService) ParseModel(modelStr string) ModelInfo {
	if modelStr == "" {
		return ModelInfo{
			Provider: "",
			Model:    "",
			IsAlias:  false,
			ProviderAlias: "",
		}
	}

	// Check if standard format: provider/model or alias/model
	if strings.Contains(modelStr, "/") {
		firstSlash := strings.Index(modelStr, "/")
		providerOrAlias := modelStr[:firstSlash]
		model := modelStr[firstSlash+1:]
		provider := s.ResolveProviderAlias(providerOrAlias)
		return ModelInfo{
			Provider:      provider,
			Model:         model,
			IsAlias:       false,
			ProviderAlias: providerOrAlias,
		}
	}

	// Alias format (model alias, not provider alias)
	return ModelInfo{
		Provider:      "",
		Model:         modelStr,
		IsAlias:       true,
		ProviderAlias: "",
	}
}

// ResolveModelAlias resolves a model alias from the stored aliases
// ref: open-sse/services/model.js:128-154
func (s *ModelService) ResolveModelAlias(ctx context.Context, alias string) (*ResolvedModel, error) {
	modelAlias, err := s.db.GetModelAliasByAlias(ctx, alias)
	if err != nil {
		return nil, err
	}

	// TargetModel is in "provider/model" format
	targetModel := modelAlias.TargetModel
	if strings.Contains(targetModel, "/") {
		firstSlash := strings.Index(targetModel, "/")
		providerOrAlias := targetModel[:firstSlash]
		return &ResolvedModel{
			Provider: s.ResolveProviderAlias(providerOrAlias),
			Model:    targetModel[firstSlash+1:],
		}, nil
	}

	// Fallback: target model without provider
	return &ResolvedModel{
		Provider: "",
		Model:    targetModel,
	}, nil
}

// ResolveModelAliasFromMap resolves a model alias from a provided map
// ref: open-sse/services/model.js:128-154
func (s *ModelService) ResolveModelAliasFromMap(alias string, aliases map[string]string) *ResolvedModel {
	if aliases == nil {
		return nil
	}

	resolved, ok := aliases[alias]
	if !ok {
		return nil
	}

	// Resolved value is "provider/model" format
	if strings.Contains(resolved, "/") {
		firstSlash := strings.Index(resolved, "/")
		providerOrAlias := resolved[:firstSlash]
		return &ResolvedModel{
			Provider: s.ResolveProviderAlias(providerOrAlias),
			Model:    resolved[firstSlash+1:],
		}
	}

	return nil
}

// inferProviderFromModelName infers provider from model name prefix
// ref: open-sse/services/model.js:194-205
func inferProviderFromModelName(modelName string) string {
	if modelName == "" {
		return "openai"
	}

	m := strings.ToLower(modelName)
	switch {
	case strings.HasPrefix(m, "claude-"):
		return "anthr0pic"
	case strings.HasPrefix(m, "gemini-"):
		return "gemini"
	case strings.HasPrefix(m, "gpt-"):
		return "openai"
	case strings.HasPrefix(m, "o1"), strings.HasPrefix(m, "o3"), strings.HasPrefix(m, "o4"):
		return "openai"
	case strings.HasPrefix(m, "deepseek-"):
		return "openrouter"
	default:
		return "openai"
	}
}

// GetModelInfo gets full model info by parsing and resolving aliases
// ref: open-sse/services/model.js:161-188
func (s *ModelService) GetModelInfo(ctx context.Context, modelStr string) (*ResolvedModel, error) {
	parsed := s.ParseModel(modelStr)

	if !parsed.IsAlias {
		return &ResolvedModel{
			Provider: parsed.Provider,
			Model:    parsed.Model,
		}, nil
	}

	// Try to resolve alias from database
	resolved, err := s.ResolveModelAlias(ctx, parsed.Model)
	if err == nil && resolved != nil {
		if resolved.Provider == "" {
			// Infer provider from model name if not specified
			resolved.Provider = inferProviderFromModelName(resolved.Model)
		}
		return resolved, nil
	}

	// Fallback: infer provider from model name prefix
	return &ResolvedModel{
		Provider: inferProviderFromModelName(parsed.Model),
		Model:    parsed.Model,
	}, nil
}

// GetModelInfoWithAliases gets full model info using provided aliases map
// ref: open-sse/services/model.js:161-188
func (s *ModelService) GetModelInfoWithAliases(modelStr string, aliases map[string]string) *ResolvedModel {
	parsed := s.ParseModel(modelStr)

	if !parsed.IsAlias {
		return &ResolvedModel{
			Provider: parsed.Provider,
			Model:    parsed.Model,
		}
	}

	// Resolve alias from map
	resolved := s.ResolveModelAliasFromMap(parsed.Model, aliases)
	if resolved != nil {
		return resolved
	}

	// Fallback: infer provider from model name prefix
	return &ResolvedModel{
		Provider: inferProviderFromModelName(parsed.Model),
		Model:    parsed.Model,
	}
}

// ListModelAliases retrieves all model aliases from storage
func (s *ModelService) ListModelAliases(ctx context.Context) ([]models.ModelAlias, error) {
	return s.db.ListModelAliases(ctx)
}

// CreateModelAlias creates a new model alias
func (s *ModelService) CreateModelAlias(ctx context.Context, alias *models.ModelAlias) error {
	return s.db.CreateModelAlias(ctx, alias)
}

// DeleteModelAlias deletes a model alias by ID
func (s *ModelService) DeleteModelAlias(ctx context.Context, id string) error {
	return s.db.DeleteModelAlias(ctx, id)
}
