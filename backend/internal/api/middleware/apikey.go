package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/auth"
	"github.com/DevKuroX/AIPROXY/internal/errs"
	"github.com/DevKuroX/AIPROXY/internal/models"
)

type apiKeyCtxKey string

const APIKeyIDKey apiKeyCtxKey = "apiKeyID"

// KeyStore defines the interface for API key storage operations.
type KeyStore interface {
	GetAPIKeyByKey(ctx context.Context, key string) (*models.APIKey, error)
	UpdateAPIKeyLastUsed(ctx context.Context, id string) error
}

// RequireAPIKey returns a middleware that validates API keys for /v1/* routes.
// It extracts the Bearer token from the Authorization header, verifies the HMAC signature,
// and injects the apiKeyID into the request context.
func RequireAPIKey(keyStore KeyStore, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errs.WriteJSONError(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				errs.WriteJSONError(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			apiKey := parts[1]

			// Look up the API key in the database
			storedKey, err := keyStore.GetAPIKeyByKey(r.Context(), apiKey)
			if err != nil {
				errs.WriteJSONError(w, "invalid API key", http.StatusUnauthorized)
				return
			}

			// Verify the HMAC signature and hash
			if !auth.VerifyAPIKey(apiKey, storedKey.KeyHash, secret) {
				errs.WriteJSONError(w, "invalid API key", http.StatusUnauthorized)
				return
			}

			// Update last used timestamp (non-blocking, ignore errors)
			go func() {
				_ = keyStore.UpdateAPIKeyLastUsed(context.Background(), storedKey.ID)
			}()

			// Inject apiKeyID into context
			ctx := context.WithValue(r.Context(), APIKeyIDKey, storedKey.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAPIKeyID extracts the API key ID from the context.
func GetAPIKeyID(ctx context.Context) string {
	if apiKeyID, ok := ctx.Value(APIKeyIDKey).(string); ok {
		return apiKeyID
	}
	return ""
}
