package errs

import (
	"net/http/httptest"
	"testing"
)

func TestWriteJSONError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSONError(w, "test error", 400)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	body := w.Body.String()
	if body != `{"error":"test error"}`+"\n" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestWriteJSONErrorDifferentStatus(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSONError(w, "not found", 404)

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestWriteJSONErrorEmptyMessage(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSONError(w, "", 500)

	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
