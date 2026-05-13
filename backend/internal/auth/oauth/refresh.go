package oauth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// TOKEN_EXPIRY_BUFFER_MS is the default buffer before token expiration.
// Tokens are refreshed if they expire within 5 minutes.
// ref: open-sse/services/tokenRefresh.js:5-6
const TOKEN_EXPIRY_BUFFER_MS = 5 * time.Minute

// Common refresh errors.
var (
	ErrRefreshInProgress    = errors.New("refresh already in progress")
	ErrRefreshTokenReused   = errors.New("refresh_token_reused")
	ErrInvalidGrant         = errors.New("invalid_grant")
	ErrUnrecoverableRefresh = errors.New("unrecoverable refresh error")
)

// isUnrecoverableError checks if the error indicates an unrecoverable refresh failure.
// ref: open-sse/services/tokenRefresh.js:15-24
func isUnrecoverableError(err error) bool {
	if err == nil {
		return false
	}
	// Check for specific error types
	var refreshErr *RefreshError
	if errors.As(err, &refreshErr) {
		return refreshErr.IsUnrecoverable()
	}
	// Check error strings for common unrecoverable patterns
	errStr := err.Error()
	return errStr == "refresh_token_reused" ||
		errStr == "invalid_grant" ||
		errStr == "invalid_request" ||
		errStr == "unrecoverable_refresh_error"
}

// refreshResult holds the result of a refresh operation.
type refreshResult struct {
	token *TokenPair
	err   error
}

// inFlightRefresh tracks an ongoing refresh operation.
type inFlightRefresh struct {
	result chan refreshResult
	done   chan struct{}
}

// RefreshManager handles deduplication of concurrent token refreshes.
// If 5 requests hit 401 simultaneously, only ONE refresh executes.
// ref: open-sse/services/tokenRefresh.js:8-12
type RefreshManager struct {
	// inFlight stores *inFlightRefresh keyed by "providerID:refreshToken"
	inFlight sync.Map
}

// NewRefreshManager creates a new RefreshManager.
func NewRefreshManager() *RefreshManager {
	return &RefreshManager{}
}

// getCacheKey generates a deduplication key.
// ref: open-sse/services/tokenRefresh.js:10-12
func getCacheKey(providerID, refreshToken string) string {
	return fmt.Sprintf("%s:%s", providerID, refreshToken)
}

// RefreshOrWait executes a token refresh with deduplication.
// If a refresh for the same providerID:refreshToken is already in progress,
// it waits for that result instead of starting a new refresh.
// ref: open-sse/services/tokenRefresh.js:8-12
func (rm *RefreshManager) RefreshOrWait(
	ctx context.Context,
	providerID string,
	refreshToken string,
	refreshFn func() (*TokenPair, error),
) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is empty")
	}

	key := getCacheKey(providerID, refreshToken)

	// Try to create a new in-flight entry
	newFlight := &inFlightRefresh{
		result: make(chan refreshResult, 1),
		done:   make(chan struct{}),
	}

	// LoadOrStore returns existing value if present, or stores new value
	existing, loaded := rm.inFlight.LoadOrStore(key, newFlight)
	flight := existing.(*inFlightRefresh)

	if loaded {
		// Another goroutine is already refreshing - wait for result
		select {
		case result := <-flight.result:
			return result.token, result.err
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-flight.done:
			// Refresh completed but result already consumed
			return nil, ErrRefreshInProgress
		}
	}

	// We're the first - do the refresh
	defer func() {
		close(newFlight.done)
		rm.inFlight.Delete(key)
	}()

	result := refreshResult{}

	// Execute the refresh function
	token, err := refreshFn()
	result.token = token
	result.err = err

	// Check for unrecoverable errors (e.g., refresh_token_reused)
	// ref: open-sse/services/tokenRefresh.js:15-24
	if err != nil && isUnrecoverableError(err) {
		// Wrap as unrecoverable error for caller to handle
		result.err = fmt.Errorf("%w: %v", ErrUnrecoverableRefresh, err)
	}

	// Send result to any waiters
	select {
	case newFlight.result <- result:
	default:
		// No waiters
	}

	return result.token, result.err
}

// IsRefreshing checks if a refresh is currently in progress for the given key.
func (rm *RefreshManager) IsRefreshing(providerID, refreshToken string) bool {
	key := getCacheKey(providerID, refreshToken)
	_, ok := rm.inFlight.Load(key)
	return ok
}

// Clear removes any in-flight refresh entry (for cleanup or forced reset).
func (rm *RefreshManager) Clear(providerID, refreshToken string) {
	key := getCacheKey(providerID, refreshToken)
	rm.inFlight.Delete(key)
}
