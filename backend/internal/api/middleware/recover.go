package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recover middleware catches panics in HTTP handlers,
// logs the error with stack trace server-side, and returns 500
// to the client. This prevents a single panic from crashing the
// entire server.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\nStack trace:\n%s", err, debug.Stack())
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
