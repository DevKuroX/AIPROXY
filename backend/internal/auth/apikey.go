package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	// KeyPrefix is the prefix for all API keys
	KeyPrefix = "aiproxy"
	// RandomPartLength is the length of the random portion of the key
	RandomPartLength = 32
)

var (
	ErrInvalidKeyFormat = errors.New("invalid API key format")
	ErrInvalidSignature = errors.New("invalid API key signature")
)

// GenerateAPIKey creates a new API key with HMAC-SHA256 signature.
// Returns: key (the full API key to give to user), hash (to store in database), error
// The key format is: aiproxy_<random_32_chars>_<hmac_signature>
func GenerateAPIKey(secret string) (key, hash string, err error) {
	// Generate random bytes for the key
	randomBytes := make([]byte, RandomPartLength/2) // hex encoding doubles length
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomPart := hex.EncodeToString(randomBytes)

	// Create the key without signature first
	keyWithoutSig := fmt.Sprintf("%s_%s", KeyPrefix, randomPart)

	// Generate HMAC signature
	signature := computeHMAC(keyWithoutSig, secret)

	// Full key with signature
	key = fmt.Sprintf("%s_%s", keyWithoutSig, signature)

	// Store hash of the full key (not the key itself)
	hash = computeHash(key)

	return key, hash, nil
}

// VerifyAPIKey verifies an API key against the stored hash.
// key: the API key provided by the user
// storedHash: the hash stored in the database
// secret: the HMAC secret used for signature verification
func VerifyAPIKey(key, storedHash, secret string) bool {
	// Verify the key format and signature
	if !isValidKeyFormat(key) {
		return false
	}

	// Extract the key parts
	parts := splitKey(key)
	if len(parts) != 3 {
		return false
	}

	// Reconstruct key without signature
	keyWithoutSig := fmt.Sprintf("%s_%s", parts[0], parts[1])
	providedSig := parts[2]

	// Verify HMAC signature
	expectedSig := computeHMAC(keyWithoutSig, secret)
	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return false
	}

	// Verify against stored hash
	computedHash := computeHash(key)
	return hmac.Equal([]byte(computedHash), []byte(storedHash))
}

// computeHMAC generates an HMAC-SHA256 signature
func computeHMAC(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// computeHash generates a SHA256 hash of the key for storage
func computeHash(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// isValidKeyFormat checks if the key has the expected format
func isValidKeyFormat(key string) bool {
	parts := splitKey(key)
	if len(parts) != 3 {
		return false
	}
	return parts[0] == KeyPrefix && len(parts[1]) == RandomPartLength && len(parts[2]) == 64 // SHA256 hex length
}

// splitKey splits the key by underscore, expecting 3 parts
func splitKey(key string) []string {
	// We need to split into exactly 3 parts: prefix, random, signature
	// Format: aiproxy_<random>_<signature>
	// Since signature contains no underscores, we can use SplitN
	parts := make([]string, 0, 3)
	underscoreCount := 0
	lastIndex := 0

	for i, c := range key {
		if c == '_' {
			underscoreCount++
			if underscoreCount <= 2 {
				parts = append(parts, key[lastIndex:i])
				lastIndex = i + 1
			}
		}
	}

	// Add the remaining part (signature)
	if lastIndex < len(key) {
		parts = append(parts, key[lastIndex:])
	}

	return parts
}
