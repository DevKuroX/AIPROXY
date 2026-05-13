// Package crypto provides AES-256-GCM encryption utilities for secure token storage.
//
// Security considerations:
//   - AES-256-GCM provides authenticated encryption (confidentiality + integrity)
//   - Each encryption uses a unique random 12-byte nonce (never reused)
//   - Nonce is prepended to ciphertext: nonce || ciphertext || tag
//   - Key must be exactly 32 bytes (AES-256)
//   - Never log plaintext data or encryption keys
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

const (
	// KeySize is the required key length for AES-256 (32 bytes)
	KeySize = 32
	// NonceSize is the nonce length for GCM (12 bytes recommended by NIST)
	NonceSize = 12
)

var (
	ErrInvalidKeySize    = errors.New("key must be exactly 32 bytes for AES-256")
	ErrInvalidCiphertext = errors.New("ciphertext is too short to contain nonce")
	ErrDecryptFailed     = errors.New("decryption failed: authentication tag mismatch")
)

// Encrypt encrypts plaintext using AES-256-GCM with the provided key.
//
// The key must be exactly 32 bytes. A random 12-byte nonce is generated for
// each encryption and prepended to the ciphertext. The returned value is
// base64-encoded for safe storage.
//
// Format: base64(nonce || ciphertext || tag)
//
// Security notes:
//   - Never reuse a key with different encryption modes
//   - Never reuse nonces with the same key (this function handles that)
//   - Store the key securely (e.g., environment variable, secret manager)
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce for each encryption
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal appends the ciphertext to the nonce
	// Output: nonce || ciphertext || tag
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Return base64-encoded result for safe storage
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(ciphertext)))
	base64.StdEncoding.Encode(encoded, ciphertext)

	return encoded, nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM with the provided key.
//
// The key must be exactly 32 bytes and match the key used for encryption.
// The ciphertext must be in the format: base64(nonce || ciphertext || tag)
//
// Returns an error if:
//   - Key is not 32 bytes
//   - Ciphertext is too short or malformed
//   - Authentication tag verification fails (tampered data)
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}

	// Decode base64
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(ciphertext)))
	n, err := base64.StdEncoding.Decode(decoded, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	decoded = decoded[:n]

	if len(decoded) < NonceSize {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce (first 12 bytes)
	nonce := decoded[:NonceSize]
	encryptedData := decoded[NonceSize:]

	// Open verifies the authentication tag and decrypts
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	return plaintext, nil
}
