// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/azure.js
package executor

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// AzureExecutor handles Azure OpenAI-specific request transformations.
// ref: open-sse/executors/azure.js:3
type AzureExecutor struct {
	BaseExecutor
}

// NewAzureExecutor creates a new Azure executor.
// ref: open-sse/executors/azure.js:4-6
func NewAzureExecutor() *AzureExecutor {
	return &AzureExecutor{
		BaseExecutor: NewBaseExecutor("azure"),
	}
}

// AzureCredentials holds Azure-specific credential data.
// ref: open-sse/executors/azure.js:9-20
type AzureCredentials struct {
	AzureEndpoint string
	APIVersion    string
	Deployment    string
	Organization  string
	APIKey        string
}

// PrepareRequest transforms the request for Azure OpenAI API.
// ref: open-sse/executors/azure.js:8-24
func (e *AzureExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Extract credentials from context if available
	creds := extractAzureCredentials(ctx)

	// Build Azure-specific URL
	// ref: open-sse/executors/azure.js:9-23
	azureEndpoint := creds.AzureEndpoint
	if azureEndpoint == "" {
		azureEndpoint = os.Getenv("AZURE_ENDPOINT")
	}
	if azureEndpoint == "" {
		azureEndpoint = "https://api.openai.com"
	}
	azureEndpoint = strings.TrimSuffix(azureEndpoint, "/")

	apiVersion := creds.APIVersion
	if apiVersion == "" {
		apiVersion = os.Getenv("AZURE_API_VERSION")
	}
	if apiVersion == "" {
		apiVersion = "2024-10-01-preview"
	}

	// Deployment name from credentials, or use model from request path
	deployment := creds.Deployment
	if deployment == "" {
		deployment = os.Getenv("AZURE_DEPLOYMENT")
	}
	if deployment == "" {
		deployment = "gpt-4"
	}

	// Construct Azure OpenAI URL
	// Format: {endpoint}/openai/deployments/{deployment}/chat/completions?api-version={version}
	// ref: open-sse/executors/azure.js:23
	azureURL := azureEndpoint + "/openai/deployments/" + deployment + "/chat/completions?api-version=" + url.QueryEscape(apiVersion)
	req.URL, _ = url.Parse(azureURL)

	// Set Azure-specific headers
	// ref: open-sse/executors/azure.js:26-51
	req.Header.Set("Content-Type", "application/json")

	// Azure uses "api-key" header instead of Authorization
	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey != "" {
		req.Header.Set("api-key", apiKey)
	}

	// Organization header if provided
	// ref: open-sse/executors/azure.js:40-45
	org := creds.Organization
	if org == "" {
		org = os.Getenv("AZURE_ORGANIZATION")
	}
	if org != "" {
		req.Header.Set("OpenAI-Organization", org)
	}

	return nil
}

// TransformResponse passes through the response unchanged.
// ref: open-sse/executors/azure.js:54-56
func (e *AzureExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	// Azure returns standard OpenAI format, no transformation needed
	return nil, nil
}

// HandleError processes Azure-specific errors.
func (e *AzureExecutor) HandleError(ctx context.Context, err error) error {
	return err
}

// extractAzureCredentials extracts Azure credentials from context.
func extractAzureCredentials(ctx context.Context) AzureCredentials {
	// TODO: Implement credential extraction from context when credential system is in place
	return AzureCredentials{}
}
