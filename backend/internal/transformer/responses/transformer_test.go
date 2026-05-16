package responses

import (
	"testing"
)

func TestNewTransformer(t *testing.T) {
	tr := NewTransformer()
	if tr == nil {
		t.Fatal("NewTransformer returned nil")
	}
}

func TestNewStreamToJSONConverter(t *testing.T) {
	c := NewStreamToJSONConverter()
	if c == nil {
		t.Fatal("NewStreamToJSONConverter returned nil")
	}
}
