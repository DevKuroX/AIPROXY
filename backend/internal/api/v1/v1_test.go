package v1

import (
	"testing"
)

func TestNewModelsStore(t *testing.T) {
	s := NewModelsStore()
	if s == nil {
		t.Fatal("NewModelsStore returned nil")
	}
}

func TestSetSearchEnabled(t *testing.T) {
	SetSearchEnabled(true)
	SetSearchEnabled(false)
}

func TestSetImageEnabled(t *testing.T) {
	SetImageEnabled(true)
	SetImageEnabled(false)
}

func TestSetTTSEnabled(t *testing.T) {
	SetTTSEnabled(true)
	SetTTSEnabled(false)
}

func TestSetSearchProviderStore(t *testing.T) {
	SetSearchProviderStore(nil)
}

func TestSetModelsStore(t *testing.T) {
	s := NewModelsStore()
	SetModelsStore(s)
}
