package request

// ref: https://platform.openai.com/docs/api-reference/images/create

// ImageGenerationRequest represents an OpenAI-compatible image generation request
type ImageGenerationRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model,omitempty"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Style          string `json:"style,omitempty"`
	User           string `json:"user,omitempty"`
}

// ImageGenerationResponse represents the response from an image generation API
type ImageGenerationResponse struct {
	Created int64          `json:"created"`
	Data    []ImageData    `json:"data"`
}

// ImageData represents a single generated image
type ImageData struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// DALLEImageRequest represents DALL-E specific request format
// ref: https://platform.openai.com/docs/api-reference/images/create
type DALLEImageRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Style          string `json:"style,omitempty"`
	User           string `json:"user,omitempty"`
}

// TranslateImageRequest translates an OpenAI-compatible image request to DALL-E format
func TranslateImageRequest(req *ImageGenerationRequest) *DALLEImageRequest {
	model := req.Model
	if model == "" {
		model = "dall-e-3"
	}

	size := req.Size
	if size == "" {
		size = "1024x1024"
	}

	n := req.N
	if n == 0 {
		n = 1
	}

	return &DALLEImageRequest{
		Prompt:         req.Prompt,
		Model:          model,
		N:              n,
		Size:           size,
		Quality:        req.Quality,
		ResponseFormat: req.ResponseFormat,
		Style:          req.Style,
		User:           req.User,
	}
}
