package config

import (
	"os"

	"github.com/DevKuroX/AIPROXY/internal/auth/crypto"
)

// GetEncryptionKey retrieves the encryption key from ENCRYPTION_KEY env var.
// It panics if the key is not set, as tokens cannot be stored safely without encryption.
//
// The key is derived using SHA-256 to produce a 32-byte key for AES-256-GCM.
// Set ENCRYPTION_KEY to a high-entropy secret (at least 32 random characters).
func GetEncryptionKey() []byte {
	secret := os.Getenv("ENCRYPTION_KEY")
	if secret == "" {
		panic("ENCRYPTION_KEY environment variable is not set - tokens cannot be stored safely")
	}
	return crypto.DeriveKey(secret)
}
