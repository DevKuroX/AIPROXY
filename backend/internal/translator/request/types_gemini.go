package request

// ref: open-sse/translator/request/openai-to-gemini.js

// GeminiRequest represents a Gemini API request.
type GeminiRequest struct {
	Model            string              `json:"model,omitempty"`
	Contents         []GeminiContent     `json:"contents"`
	GenerationConfig *GeminiGenConfig    `json:"generationConfig,omitempty"`
	SafetySettings   []GeminiSafety      `json:"safetySettings,omitempty"`
	Tools            []GeminiTool        `json:"tools,omitempty"`
	SystemInstruction *GeminiContent     `json:"systemInstruction,omitempty"`
}

// GeminiContent represents a content block in Gemini format.
type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part in Gemini content.
type GeminiPart struct {
	Text             string                 `json:"text,omitempty"`
	InlineData       *GeminiInlineData      `json:"inlineData,omitempty"`
	FunctionCall     *GeminiFunctionCall    `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
	Thought          bool                   `json:"thought,omitempty"`
	ThoughtSignature string                 `json:"thoughtSignature,omitempty"`
}

// GeminiInlineData represents inline data (images, etc.).
type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// GeminiFunctionCall represents a function call in Gemini.
type GeminiFunctionCall struct {
	ID   string                 `json:"id,omitempty"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// GeminiFunctionResponse represents a function response in Gemini.
type GeminiFunctionResponse struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
}

// GeminiGenConfig represents generation config.
type GeminiGenConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
}

// GeminiSafety represents safety settings.
type GeminiSafety struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GeminiTool represents a tool in Gemini.
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDecl `json:"functionDeclarations,omitempty"`
}

// GeminiFunctionDecl represents a function declaration.
type GeminiFunctionDecl struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}
