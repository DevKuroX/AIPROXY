package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/auth"
	"github.com/DevKuroX/AIPROXY/internal/errs"
)

type ctxKey string

const claimsKey ctxKey = "claims"

func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
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

			tokenString := parts[1]
			claims, err := auth.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				errs.WriteJSONError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaimsFromContext(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(claimsKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
