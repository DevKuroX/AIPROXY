package utils

import (
	"testing"
)

func TestBase64URLEncodeDecode(t *testing.T) {
	original := []byte("hello world")
	encoded := Base64URLEncode(original)
	decoded, err := Base64URLDecode(encoded)
	if err != nil {
		t.Fatalf("Base64URLDecode failed: %v", err)
	}
	if string(decoded) != string(original) {
		t.Fatalf("expected '%s', got '%s'", original, decoded)
	}
}

func TestGenerateSessionID(t *testing.T) {
	id := GenerateSessionID("test-auth-token")
	if id == "" {
		t.Fatal("GenerateSessionID returned empty")
	}
	t.Logf("Session ID: %s", id)
}

func TestGenerateHashed64Hex(t *testing.T) {
	hash := GenerateHashed64Hex("test-input", "test-salt")
	if hash == "" {
		t.Fatal("GenerateHashed64Hex returned empty")
	}
}

func TestDetectClientTool(t *testing.T) {
	headers := map[string]string{
		"User-Agent": "Cursor/v1.2",
	}
	tool := DetectClientTool(headers, nil)
	t.Logf("Detected tool: %s", tool)
}

func TestIsNativePassthrough(t *testing.T) {
	if IsNativePassthrough("cursor", "cursor") {
		t.Log("cursor->cursor is native passthrough")
	}
}
