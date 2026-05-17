package storage

import (
	"context"
	"os"
	"testing"
)

func testDB(t *testing.T) *DB {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://aiproxy:aiproxy123@localhost:5432/aiproxy?sslmode=disable"
	}
	db, err := New(context.Background(), dbURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewDB(t *testing.T) {
	db := testDB(t)
	if db == nil {
		t.Fatal("DB is nil")
	}
}

func TestGetSettings(t *testing.T) {
	db := testDB(t)
	val, err := db.GetSetting(context.Background(), "test_key")
	if err != nil && err != ErrSettingNotFound {
		t.Fatalf("GetSetting failed: %v", err)
	}
	t.Logf("GetSetting returned: %q, err: %v", val, err)
}

func TestSetAndGetSettings(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	err := db.SetSetting(ctx, "test_key_1", "test_value_1")
	if err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	val, err := db.GetSetting(ctx, "test_key_1")
	if err != nil {
		t.Fatalf("GetSetting failed: %v", err)
	}
	if val != "test_value_1" {
		t.Fatalf("expected test_value_1, got %s", val)
	}
}

func TestGetSettingNotFound(t *testing.T) {
	db := testDB(t)
	_, err := db.GetSetting(context.Background(), "nonexistent_key_xyz")
	if err != ErrSettingNotFound {
		t.Fatalf("expected ErrSettingNotFound, got %v", err)
	}
}

func TestGetUsers(t *testing.T) {
	db := testDB(t)
	user, err := db.GetUserByUsername(context.Background(), "admin")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if user.Username != "admin" {
		t.Fatalf("expected admin, got %s", user.Username)
	}
}

func TestGetAPIKeys(t *testing.T) {
	db := testDB(t)
	keys, err := db.ListAPIKeys(context.Background())
	if err != nil {
		t.Fatalf("ListAPIKeys failed: %v", err)
	}
	t.Logf("Found %d API keys", len(keys))
}

func TestGetProviders(t *testing.T) {
	db := testDB(t)
	providers, err := db.ListProviders(context.Background())
	if err != nil {
		t.Fatalf("ListProviders failed: %v", err)
	}
	if len(providers) == 0 {
		t.Fatal("expected at least 1 provider")
	}
	t.Logf("Found %d providers", len(providers))
	for _, p := range providers {
		t.Logf("  - %s (type=%s)", p.Name, p.Type)
	}
}

func TestGetProviderAccounts(t *testing.T) {
	db := testDB(t)
	providers, err := db.ListProviders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) > 0 {
		accounts, err := db.ListProviderAccounts(context.Background(), providers[0].ID)
		if err != nil {
			t.Fatalf("ListProviderAccounts failed: %v", err)
		}
		t.Logf("Provider %s has %d accounts", providers[0].Name, len(accounts))
	}
}
