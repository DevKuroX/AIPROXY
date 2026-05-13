package crypto

import (
	"crypto/rand"
	"crypto/sha256"
)

// DeriveKey derives a 32-byte key from a secret string using SHA-256.
// This allows using arbitrary-length secrets (e.g., environment variables)
// as encryption keys for AES-256-GCM.
//
// Note: For production use, consider using a proper KDF like HKDF or Argon2
// if the secret has low entropy. SHA-256 is suitable for high-entropy secrets.
func DeriveKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// GenerateKey generates a cryptographically secure random 32-byte key.
// Use this to create new encryption keys for AES-256-GCM.
//
// The generated key should be stored securely (e.g., in a secrets manager)
// and never logged or exposed in error messages.
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
