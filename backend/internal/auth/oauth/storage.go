package oauth

import (
	"context"
	"time"
)

// TokenStore defines the interface for CRUD operations on encrypted OAuth tokens.
// Implementations should handle encryption/decryption transparently.
type TokenStore interface {
	// GetToken retrieves an OAuth token for the given provider and account.
	GetToken(ctx context.Context, providerID, accountID string) (*OAuthToken, error)

	// SaveToken stores an OAuth token for the given provider and account.
	// The token should be encrypted before storage.
	SaveToken(ctx context.Context, providerID, accountID string, token *OAuthToken) error

	// DeleteToken removes an OAuth token for the given provider and account.
	DeleteToken(ctx context.Context, providerID, accountID string) error
}

// TokenPairFromOAuth creates a TokenPair from OAuthToken with calculated ExpiresIn.
func TokenPairFromOAuth(token *OAuthToken) *TokenPair {
	if token == nil {
		return nil
	}
	expiresIn := int64(time.Until(token.ExpiresAt).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}
	return &TokenPair{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    int(expiresIn),
	}
}

// OAuthTokenFromPair creates an OAuthToken from a TokenPair.
func OAuthTokenFromPair(pair *TokenPair) *OAuthToken {
	if pair == nil {
		return nil
	}
	return &OAuthToken{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(pair.ExpiresIn) * time.Second),
		TokenType:    "Bearer",
	}
}
