package chat

import (
	"testing"
)

func TestNewStreamingHandler(t *testing.T) {
	h := &StreamingHandler{}
	if h == nil {
		t.Fatal("StreamingHandler nil")
	}
}

func TestNewNonStreamingHandler(t *testing.T) {
	h := &NonStreamingHandler{}
	if h == nil {
		t.Fatal("NonStreamingHandler nil")
	}
}

func TestErrorHandler(t *testing.T) {
	h := &ErrorHandler{}
	if h == nil {
		t.Fatal("ErrorHandler nil")
	}
}

func TestRequestDetail(t *testing.T) {
	rd := &RequestDetail{}
	if rd == nil {
		t.Fatal("RequestDetail nil")
	}
}
