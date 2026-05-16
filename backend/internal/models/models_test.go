package models

import (
	"testing"
	"time"
)

func TestAPIKey(t *testing.T) {
	k := APIKey{
		ID:        "1",
		Key:       "sk-test123",
		Name:      "Test Key",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	if k.Key != "sk-test123" {
		t.Fatalf("expected sk-test123, got %s", k.Key)
	}
	if !k.IsActive {
		t.Fatal("expected active")
	}
}

func TestAPIKeyInactive(t *testing.T) {
	k := APIKey{ID: "2", Key: "sk-test456", Name: "Inactive", IsActive: false}
	if k.IsActive {
		t.Fatal("expected inactive")
	}
}

func TestUser(t *testing.T) {
	u := User{
		ID:           1,
		Username:     "admin",
		PasswordHash: "hash123",
		IsAdmin:      true,
	}
	if u.Username != "admin" {
		t.Fatalf("expected admin, got %s", u.Username)
	}
	if !u.IsAdmin {
		t.Fatal("expected admin")
	}
}

func TestProvider(t *testing.T) {
	p := Provider{
		ID:      "1",
		Name:    "Test Provider",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Enabled: true,
	}
	if p.Type != "openai" {
		t.Fatalf("expected openai, got %s", p.Type)
	}
}

func TestProviderAccount(t *testing.T) {
	a := ProviderAccount{
		ID:         "1",
		ProviderID: "2",
		Name:       "test@test.com",
		APIKey:     "sk-key123",
		IsActive:   true,
	}
	if a.Name != "test@test.com" {
		t.Fatalf("expected test@test.com, got %s", a.Name)
	}
}

func TestCombo(t *testing.T) {
	c := Combo{
		Name:    "test-combo",
		Models:  []string{"model-a", "model-b"},
		Strategy: "fallback",
	}
	if len(c.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(c.Models))
	}
}

func TestModelAlias(t *testing.T) {
	a := ModelAlias{
		ID:          "1",
		NodeID:      "node1",
		Alias:       "gpt-test",
		TargetModel: "gpt-4",
	}
	if a.Alias != "gpt-test" {
		t.Fatalf("expected gpt-test, got %s", a.Alias)
	}
}

func TestProviderNode(t *testing.T) {
	n := ProviderNode{
		ID:       "node1",
		Name:     "Custom Node",
		BaseURL:  "https://custom.api.com",
		Enabled:  true,
	}
	if !n.Enabled {
		t.Fatal("expected enabled")
	}
}
