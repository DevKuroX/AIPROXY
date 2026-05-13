package auth

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	secret := "test-secret-key"

	key, hash, err := GenerateAPIKey(secret)
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if !strings.HasPrefix(key, KeyPrefix+"_") {
		t.Errorf("Key should start with %s_, got: %s", KeyPrefix, key)
	}

	parts := strings.Split(key, "_")
	if len(parts) != 3 {
		t.Errorf("Key should have 3 parts separated by _, got %d parts: %s", len(parts), key)
	}

	if parts[0] != KeyPrefix {
		t.Errorf("First part should be %s, got: %s", KeyPrefix, parts[0])
	}

	if len(parts[1]) != RandomPartLength {
		t.Errorf("Random part should be %d chars, got %d: %s", RandomPartLength, len(parts[1]), parts[1])
	}

	if len(parts[2]) != 64 {
		t.Errorf("Signature should be 64 chars (SHA256 hex), got %d: %s", len(parts[2]), parts[2])
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == key {
		t.Error("Hash should not equal the key (we store hash, not key)")
	}
}

func TestVerifyAPIKey(t *testing.T) {
	secret := "test-secret-key"

	key, hash, err := GenerateAPIKey(secret)
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if !VerifyAPIKey(key, hash, secret) {
		t.Error("VerifyAPIKey should return true for valid key")
	}
}

func TestVerifyAPIKey_InvalidKey(t *testing.T) {
	secret := "test-secret-key"

	_, hash, err := GenerateAPIKey(secret)
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if VerifyAPIKey("invalid_key", hash, secret) {
		t.Error("VerifyAPIKey should return false for invalid key format")
	}
}

func TestVerifyAPIKey_WrongSecret(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret"

	key, hash, err := GenerateAPIKey(secret)
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if VerifyAPIKey(key, hash, wrongSecret) {
		t.Error("VerifyAPIKey should return false when secret doesn't match")
	}
}

func TestVerifyAPIKey_WrongHash(t *testing.T) {
	secret := "test-secret-key"

	key, _, err := GenerateAPIKey(secret)
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	wrongHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	if VerifyAPIKey(key, wrongHash, secret) {
		t.Error("VerifyAPIKey should return false when hash doesn't match")
	}
}

func TestVerifyAPIKey_DifferentKeySameSecret(t *testing.T) {
	secret := "test-secret-key"

	key1, _, err := GenerateAPIKey(secret)
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	_, hash2, err := GenerateAPIKey(secret)
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	if VerifyAPIKey(key1, hash2, secret) {
		t.Error("VerifyAPIKey should return false for different key with same secret")
	}
}
