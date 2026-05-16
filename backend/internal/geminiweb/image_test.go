package geminiweb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGeminiImageUpload(t *testing.T) {
	// Read cookies
	cookieData, err := os.ReadFile("/tmp/gemini_cookies.json")
	if err != nil {
		t.Skip("cookies file /tmp/gemini_cookies.json not found")
	}
	var cookies map[string]string
	if err := json.Unmarshal(cookieData, &cookies); err != nil {
		t.Fatalf("parse cookies failed: %v", err)
	}

	s := NewSession(cookies["__Secure-1PSID"], cookies["__Secure-1PSIDTS"], "")
	if err := s.Init(); err != nil {
		t.Fatalf("session init failed: %v", err)
	}
	t.Logf("Access token: %s...", s.AccessToken[:20])
	t.Logf("Build label: %s", s.BuildLabel)
	t.Logf("Session ID: %s", s.SessionID)

	// Read test image
	imageData, err := os.ReadFile("/tmp/gemini_real_test.png")
	if err != nil {
		t.Fatalf("read test image failed: %v", err)
	}
	t.Logf("Image size: %d bytes", len(imageData))

	// Step 1: Upload image to Google
	t.Log("=== STEP 1: Uploading image ===")
	uploadedURL, err := uploadImage(s.client, imageData, "gemini_real_test.png")
	if err != nil {
		t.Fatalf("image upload failed: %v", err)
	}
	t.Logf("Uploaded URL: %s", uploadedURL)

	// Step 2: Build and send chat request with image
	t.Log("=== STEP 2: Sending chat with image ===")
	prompt := "Describe this image in one sentence"
	modelName := "gemini-3-flash"

	s.reqCounter += 100000
	reqID := s.reqCounter

	// Build message content WITH file data at index [3]
	messageContent := []interface{}{
		prompt, // [0] text prompt
		0,      // [1]
		nil,    // [2]
		[]interface{}{ // [3] file data (list of [[url], filename] entries)
			[]interface{}{
				[]interface{}{uploadedURL},
				"gemini_real_test.png",
			},
		},
		nil, // [4]
		nil, // [5]
		0,   // [6]
	}

	innerReq := make([]interface{}, 69)
	innerReq[0] = messageContent
	innerReq[1] = []interface{}{s.Language}
	innerReq[2] = defaultMetadata
	innerReq[6] = []interface{}{1}
	innerReq[streamingFlagIndex] = 1
	innerReq[10] = 1
	innerReq[11] = 0
	innerReq[17] = [][]int{{0}}
	innerReq[18] = 0
	innerReq[27] = 1
	innerReq[30] = []int{4}
	innerReq[41] = []int{1}
	innerReq[53] = 0
	innerReq[61] = []interface{}{}
	innerReq[68] = 2

	uid := strings.ToUpper(uuid.New().String())
	innerReq[59] = uid

	innerJSON, err := json.Marshal(innerReq)
	if err != nil {
		t.Fatalf("marshal inner req failed: %v", err)
	}
	freqPayload := []interface{}{
		nil,
		string(innerJSON),
	}
	freqJSON, err := json.Marshal(freqPayload)
	if err != nil {
		t.Fatalf("marshal f.req failed: %v", err)
	}

	// Build form data
	formData := url.Values{}
	formData.Set("at", s.AccessToken)
	formData.Set("f.req", string(freqJSON))
	bodyStr := formData.Encode()

	// Log the raw f.req payload
	fmt.Fprintf(os.Stderr, "\n========== RAW F.REQ (inner) ==========\n%s\n========================================\n", string(innerJSON))

	// Build URL
	queryParams := url.Values{}
	queryParams.Set("hl", s.Language)
	queryParams.Set("_reqid", fmt.Sprintf("%d", reqID))
	queryParams.Set("rt", "c")
	if s.BuildLabel != "" {
		queryParams.Set("bl", s.BuildLabel)
	}
	if s.SessionID != "" {
		queryParams.Set("f.sid", s.SessionID)
	}

	requestURL := endpointGenerate + "?" + queryParams.Encode()
	headers := BuildRequestHeaders(modelName, uid)

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(bodyStr))
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	// Log the request body
	fmt.Fprintf(os.Stderr, "\n========== RAW REQUEST BODY ==========\n%s\n=====================================\n", bodyStr)

	resp, err := s.client.do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("gemini returned %d: %s", resp.StatusCode, string(errBody))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	// Log the full raw response
	fmt.Fprintf(os.Stderr, "\n========== RAW GEMINI RESPONSE ==========\n%s\n==========================================\n", string(bodyBytes))

	// Parse the response
	result, err := ParseNonStreamingResponse(bodyBytes)
	if err != nil {
		t.Fatalf("parse response failed: %v", err)
	}
	t.Logf("Parsed response text: %q", result.Text)
	t.Logf("Done: %v", result.Done)
}

// uploadImage uploads an image to Google's content push service and returns the uploaded URL.
func uploadImage(client *httpClient, imageData []byte, filename string) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file failed: %w", err)
	}
	if _, err := fw.Write(imageData); err != nil {
		return "", fmt.Errorf("write image data failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("close multipart failed: %w", err)
	}

	req, err := http.NewRequest("POST", "https://content-push.googleapis.com/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("create upload request failed: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Origin", "https://gemini.google.com")
	req.Header.Set("Referer", "https://gemini.google.com/")
	req.Header.Set("X-Tenant-Id", "bard-storage")
	req.Header.Set("Push-ID", "feeds/mcudyrk2a4khkz")

	fmt.Fprintf(os.Stderr, "\n--- Uploading image (%d bytes) to content-push.googleapis.com ---\n", len(imageData))

	resp, err := client.do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload returned %d: %s", resp.StatusCode, string(errBody))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	uploadedURL := strings.TrimSpace(string(bodyBytes))

	fmt.Fprintf(os.Stderr, "Upload response (%d bytes): %s\n", len(bodyBytes), uploadedURL)

	return uploadedURL, nil
}
