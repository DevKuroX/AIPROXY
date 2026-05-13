package errs

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Typed errors
var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidToken    = errors.New("invalid token")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
)

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSONError writes a JSON error response
func WriteJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
