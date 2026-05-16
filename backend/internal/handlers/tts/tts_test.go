package tts

import (
	"testing"
)

func TestRegisterAndGetAdapter(t *testing.T) {
	RegisterAdapter("test", nil)
	a := GetAdapter("test")
	if a != nil {
		t.Logf("Got adapter for 'test'")
	}
}

func TestGetAdapterUnknown(t *testing.T) {
	a := GetAdapter("nonexistent")
	if a != nil {
		t.Fatal("expected nil for unknown adapter")
	}
}
