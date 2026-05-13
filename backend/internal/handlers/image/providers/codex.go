package providers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/handlers/image"
	"github.com/google/uuid"
)

const (
	codexResponsesURL = "https://chatgpt.com/backend-api/codex/responses"
	codexUserAgent    = "codex-imagen/0.2.6"
	codexVersion      = "0.122.0"
	codexOriginator   = "codex_cli_rs"
	codexModelSuffix  = "-image"
	codexRefDetail    = "high"
)

type codexAdapter struct {
	image.BaseAdapter
}

func newCodexAdapter() *codexAdapter {
	return &codexAdapter{
		BaseAdapter: image.BaseAdapter{
			Async:     false,
			Streaming: true,
			SkipAuth:  false,
		},
	}
}

func decodeAccountID(idToken string) string {
	if idToken == "" {
		return ""
	}
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return ""
	}
	payload := parts[1]
	payload = strings.ReplaceAll(payload, "-", "+")
	payload = strings.ReplaceAll(payload, "_", "/")
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := io.ReadAll(base64DecodedReader(payload))
	if err != nil {
		return ""
	}

	var data struct {
		ChatGPTAccountID string `json:"https://api.openai.com/auth"`
	}
	if err := json.Unmarshal(decoded, &data); err != nil {
		return ""
	}
	return data.ChatGPTAccountID
}

func base64DecodedReader(payload string) io.Reader {
	r := strings.NewReader(payload)
	return io.Reader(r)
}

func stripImageSuffix(model string) string {
	if strings.HasSuffix(model, codexModelSuffix) {
		return model[:len(model)-len(codexModelSuffix)]
	}
	return model
}

func toDataUrl(input string) string {
	if input == "" {
		return ""
	}
	if strings.HasPrefix(input, "data:image/") || strings.HasPrefix(input, "https://") || strings.HasPrefix(input, "http://") {
		return input
	}
	return "data:image/png;base64," + input
}

func buildContent(prompt string, refs []string, detail string) []interface{} {
	var content []interface{}
	for i, url := range refs {
		content = append(content, map[string]interface{}{
			"type": "input_text",
			"text": fmt.Sprintf("<image name=image%d>", i+1),
		})
		content = append(content, map[string]interface{}{
			"type":      "input_image",
			"image_url": url,
			"detail":    detail,
		})
		content = append(content, map[string]interface{}{
			"type": "input_text",
			"text": "</image>",
		})
	}
	content = append(content, map[string]interface{}{
		"type": "input_text",
		"text": prompt,
	})
	return content
}

func (a *codexAdapter) BuildURL(model string, creds *image.ImageCredentials) (string, error) {
	return codexResponsesURL, nil
}

func (a *codexAdapter) BuildHeaders(creds *image.ImageCredentials, requestBody interface{}, model string, body *image.ImageRequest) (map[string]string, error) {
	accountID := ""
	if creds.ProviderSpecificData != nil {
		if id, ok := creds.ProviderSpecificData["chatgptAccountId"].(string); ok {
			accountID = id
		}
	}
	if accountID == "" {
		accountID = decodeAccountID(creds.IDToken)
	}

	return map[string]string{
		"accept":               "text/event-stream, application/json",
		"authorization":        "Bearer " + creds.AccessToken,
		"chatgpt-account-id":   accountID,
		"content-type":         "application/json",
		"originator":           codexOriginator,
		"session_id":           uuid.New().String(),
		"user-agent":           codexUserAgent,
		"version":              codexVersion,
		"x-client-request-id":  uuid.New().String(),
	}, nil
}

func (a *codexAdapter) BuildBody(model string, body *image.ImageRequest) (interface{}, error) {
	var refs []string
	for _, img := range body.Images {
		if u := toDataUrl(img); u != "" {
			refs = append(refs, u)
		}
	}
	if body.Image != "" {
		if u := toDataUrl(body.Image); u != "" {
			refs = append(refs, u)
		}
	}

	detail := body.ImageDetail
	if detail == "" {
		detail = codexRefDetail
	}

	outputFormat := body.OutputFormat
	if outputFormat == "" {
		outputFormat = "png"
	}

	imgTool := map[string]interface{}{
		"type":          "image_generation",
		"output_format": strings.ToLower(outputFormat),
	}
	if body.Size != "" {
		imgTool["size"] = body.Size
	}
	if body.Quality != "" {
		imgTool["quality"] = body.Quality
	}
	if body.Background != "" {
		imgTool["background"] = body.Background
	}

	return map[string]interface{}{
		"model":      stripImageSuffix(model),
		"instructions": "",
		"input": []interface{}{
			map[string]interface{}{
				"type":    "message",
				"role":    "user",
				"content": buildContent(body.Prompt, refs, detail),
			},
		},
		"tools":              []interface{}{imgTool},
		"tool_choice":        "auto",
		"parallel_tool_calls": false,
		"prompt_cache_key":   uuid.New().String(),
		"stream":             true,
		"store":              false,
		"reasoning":          nil,
	}, nil
}

func (a *codexAdapter) ParseResponse(ctx context.Context, response *http.Response, opts *image.ParseResponseOptions) (*image.ImageResponse, error) {
	imageB64, err := a.parseStream(response, opts)
	if err != nil {
		return nil, err
	}
	if imageB64 == "" {
		return nil, fmt.Errorf("Codex did not return an image. Account may not be entitled (Plus/Pro required)")
	}
	return &image.ImageResponse{
		Created: image.NowSec(),
		Data:    []image.ImageData{{B64JSON: imageB64}},
	}, nil
}

func (a *codexAdapter) parseStream(response *http.Response, opts *image.ParseResponseOptions) (string, error) {
	reader := bufio.NewReader(response.Body)
	var imageB64 string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "event:") {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataStr := strings.TrimPrefix(line, "data:")
			dataStr = strings.TrimSpace(dataStr)
			if dataStr == "" || dataStr == "[DONE]" {
				continue
			}

			var data struct {
				Item struct {
					Type   string `json:"type"`
					Result string `json:"result"`
				} `json:"item"`
			}
			if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
				continue
			}

			if data.Item.Type == "image_generation_call" && data.Item.Result != "" {
				imageB64 = data.Item.Result
			}
		}
	}

	return imageB64, nil
}

func (a *codexAdapter) Normalize(responseBody interface{}, prompt string) (*image.ImageResponse, error) {
	switch v := responseBody.(type) {
	case *image.ImageResponse:
		return v, nil
	}
	return &image.ImageResponse{Created: image.NowSec(), Data: []image.ImageData{}}, nil
}

func init() {
	image.RegisterAdapter("codex", newCodexAdapter())
}
