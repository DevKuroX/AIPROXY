package middleware

import (
	"testing"
)

func TestRequireAuth(t *testing.T) {
	mw := RequireAuth("test-secret")
	if mw == nil {
		t.Fatal("RequireAuth returned nil")
	}
}

func TestRequireAPIKey(t *testing.T) {
	mw := RequireAPIKey(nil, "")
	if mw == nil {
		t.Fatal("RequireAPIKey returned nil")
	}
}

func TestMetrics(t *testing.T) {
	mw := Metrics()
	if mw == nil {
		t.Fatal("Metrics returned nil")
	}
}

