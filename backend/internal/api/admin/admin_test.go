package admin

import (
	"testing"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler("test-secret", nil)
	if h == nil {
		t.Fatal("NewHandler returned nil")
	}
}

func TestNewAnalyticsHandler(t *testing.T) {
	h := NewAnalyticsHandler(nil)
	if h == nil {
		t.Fatal("NewAnalyticsHandler returned nil")
	}
}

func TestNewNodeHandler(t *testing.T) {
	h := NewNodeHandler(nil)
	if h == nil {
		t.Fatal("NewNodeHandler returned nil")
	}
}

func TestNewCLIHandler(t *testing.T) {
	h := NewCLIHandler()
	if h == nil {
		t.Fatal("NewCLIHandler returned nil")
	}
}

func TestUserStruct(t *testing.T) {
	u := User{
		ID:       "1",
		Username: "admin",
		Password: "hash",
		IsAdmin:  true,
	}
	if u.Username != "admin" {
		t.Fatalf("expected admin, got %s", u.Username)
	}
}
