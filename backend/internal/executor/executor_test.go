package executor

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	list := ListProviders()
	if len(list) == 0 {
		t.Fatal("ListProviders should return at least 1 provider")
	}
	t.Logf("Registered executors: %v", list)
}

func TestGetExisting(t *testing.T) {
	exec, ok := Get("opencode")
	if !ok {
		t.Fatal("opencode executor should exist")
	}
	if exec == nil {
		t.Fatal("opencode executor should not be nil")
	}
}

func TestGetNonExistent(t *testing.T) {
	_, ok := Get("nonexistent-executor-xyz")
	if ok {
		t.Fatal("nonexistent executor should not exist")
	}
}

func TestMustGetExisting(t *testing.T) {
	exec, err := MustGet("opencode")
	if err != nil {
		t.Fatalf("MustGet opencode failed: %v", err)
	}
	if exec == nil {
		t.Fatal("MustGet opencode returned nil")
	}
}

func TestBaseExecutor(t *testing.T) {
	b := NewBaseExecutor("test-provider")
	if b.Provider() != "test-provider" {
		t.Fatalf("expected test-provider, got %s", b.Provider())
	}
}

func TestNewOpenCodeExecutor(t *testing.T) {
	e := NewOpenCodeExecutor()
	if e == nil {
		t.Fatal("NewOpenCodeExecutor returned nil")
	}
}

func TestNewGitHubExecutor(t *testing.T) {
	e := NewGitHubExecutor()
	if e == nil {
		t.Fatal("NewGitHubExecutor returned nil")
	}
}
