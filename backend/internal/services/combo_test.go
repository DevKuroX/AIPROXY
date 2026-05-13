package services

import (
	"testing"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

func TestNewComboService(t *testing.T) {
	svc := NewComboService()
	if svc == nil {
		t.Error("expected non-nil service")
	}
	if svc.rotationMap == nil {
		t.Error("expected initialized rotation map")
	}
}

func TestNormalizeStickyLimit(t *testing.T) {
	tests := []struct {
		name       string
		input      int
		wantResult int
	}{
		{"positive value returns as-is", 5, 5},
		{"zero returns 1", 0, 1},
		{"negative returns 1", -1, 1},
		{"negative large returns 1", -100, 1},
		{"large positive returns as-is", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeStickyLimit(tt.input)
			if got != tt.wantResult {
				t.Errorf("normalizeStickyLimit(%d) = %d, want %d", tt.input, got, tt.wantResult)
			}
		})
	}
}

func TestRotateModelsFromIndex(t *testing.T) {
	tests := []struct {
		name       string
		models     []string
		index      int
		wantResult []string
	}{
		{
			name:       "index 0 returns original",
			models:     []string{"a", "b", "c"},
			index:      0,
			wantResult: []string{"a", "b", "c"},
		},
		{
			name:       "negative index returns original",
			models:     []string{"a", "b", "c"},
			index:      -1,
			wantResult: []string{"a", "b", "c"},
		},
		{
			name:       "index equals length returns original",
			models:     []string{"a", "b", "c"},
			index:      3,
			wantResult: []string{"a", "b", "c"},
		},
		{
			name:       "index 1 rotates",
			models:     []string{"a", "b", "c"},
			index:      1,
			wantResult: []string{"b", "c", "a"},
		},
		{
			name:       "index 2 rotates",
			models:     []string{"a", "b", "c"},
			index:      2,
			wantResult: []string{"c", "a", "b"},
		},
		{
			name:       "single element returns same",
			models:     []string{"a"},
			index:      0,
			wantResult: []string{"a"},
		},
		{
			name:       "empty slice returns empty",
			models:     []string{},
			index:      0,
			wantResult: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rotateModelsFromIndex(tt.models, tt.index)

			if len(got) != len(tt.wantResult) {
				t.Errorf("length mismatch: got %d, want %d", len(got), len(tt.wantResult))
				return
			}

			for i := range got {
				if got[i] != tt.wantResult[i] {
					t.Errorf("rotateModelsFromIndex()[%d] = %q, want %q", i, got[i], tt.wantResult[i])
				}
			}
		})
	}
}

func TestGetRotatedModels_NoRotation(t *testing.T) {
	svc := NewComboService()

	tests := []struct {
		name     string
		models   []string
		strategy string
	}{
		{"empty models", []string{}, "round-robin"},
		{"single model", []string{"model-a"}, "round-robin"},
		{"not round-robin strategy", []string{"a", "b"}, "fallback"},
		{"empty strategy", []string{"a", "b"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.GetRotatedModels(tt.models, "test-combo", tt.strategy, 1)

			if len(got) != len(tt.models) {
				t.Errorf("length mismatch: got %d, want %d", len(got), len(tt.models))
			}
		})
	}
}

func TestGetRotatedModels_RoundRobin(t *testing.T) {
	svc := NewComboService()
	models := []string{"model-a", "model-b", "model-c"}

	t.Run("first call returns original order", func(t *testing.T) {
		got := svc.GetRotatedModels(models, "combo1", "round-robin", 1)
		want := []string{"model-a", "model-b", "model-c"}

		if len(got) != len(want) {
			t.Errorf("length mismatch: got %d, want %d", len(got), len(want))
			return
		}

		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("second call returns rotated", func(t *testing.T) {
		got := svc.GetRotatedModels(models, "combo1", "round-robin", 1)
		want := []string{"model-b", "model-c", "model-a"}

		if len(got) != len(want) {
			t.Errorf("length mismatch: got %d, want %d", len(got), len(want))
			return
		}

		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("third call returns rotated again", func(t *testing.T) {
		got := svc.GetRotatedModels(models, "combo1", "round-robin", 1)
		want := []string{"model-c", "model-a", "model-b"}

		if len(got) != len(want) {
			t.Errorf("length mismatch: got %d, want %d", len(got), len(want))
			return
		}

		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})
}

func TestGetRotatedModels_StickyLimit(t *testing.T) {
	svc := NewComboService()
	models := []string{"model-a", "model-b", "model-c"}

	t.Run("sticky limit 2 stays on first model", func(t *testing.T) {
		got1 := svc.GetRotatedModels(models, "sticky-test", "round-robin", 2)
		got2 := svc.GetRotatedModels(models, "sticky-test", "round-robin", 2)

		want := []string{"model-a", "model-b", "model-c"}

		for i := range got1 {
			if got1[i] != want[i] {
				t.Errorf("first call got1[%d] = %q, want %q", i, got1[i], want[i])
			}
		}

		for i := range got2 {
			if got2[i] != want[i] {
				t.Errorf("second call got2[%d] = %q, want %q", i, got2[i], want[i])
			}
		}
	})

	t.Run("third call rotates after sticky limit reached", func(t *testing.T) {
		got := svc.GetRotatedModels(models, "sticky-test", "round-robin", 2)
		want := []string{"model-b", "model-c", "model-a"}

		for i := range got {
			if got[i] != want[i] {
				t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})
}

func TestGetRotatedModels_DifferentCombos(t *testing.T) {
	svc := NewComboService()
	models := []string{"model-a", "model-b", "model-c"}

	got1 := svc.GetRotatedModels(models, "combo-alpha", "round-robin", 1)
	got2 := svc.GetRotatedModels(models, "combo-beta", "round-robin", 1)

	want := []string{"model-a", "model-b", "model-c"}

	for i := range got1 {
		if got1[i] != want[i] {
			t.Errorf("combo-alpha got1[%d] = %q, want %q", i, got1[i], want[i])
		}
	}

	for i := range got2 {
		if got2[i] != want[i] {
			t.Errorf("combo-beta got2[%d] = %q, want %q", i, got2[i], want[i])
		}
	}
}

func TestGetRotatedModels_EmptyComboName(t *testing.T) {
	svc := NewComboService()
	models := []string{"model-a", "model-b"}

	got1 := svc.GetRotatedModels(models, "", "round-robin", 1)
	got2 := svc.GetRotatedModels(models, "", "round-robin", 1)

	want1 := []string{"model-a", "model-b"}
	want2 := []string{"model-b", "model-a"}

	for i := range got1 {
		if got1[i] != want1[i] {
			t.Errorf("first call got1[%d] = %q, want %q", i, got1[i], want1[i])
		}
	}

	for i := range got2 {
		if got2[i] != want2[i] {
			t.Errorf("second call got2[%d] = %q, want %q", i, got2[i], want2[i])
		}
	}
}

func TestResetComboRotation(t *testing.T) {
	svc := NewComboService()
	models := []string{"model-a", "model-b", "model-c"}

	svc.GetRotatedModels(models, "reset-test", "round-robin", 1)
	svc.ResetComboRotation("reset-test")

	got := svc.GetRotatedModels(models, "reset-test", "round-robin", 1)
	want := []string{"model-a", "model-b", "model-c"}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("after reset got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResetComboRotation_All(t *testing.T) {
	svc := NewComboService()
	models := []string{"model-a", "model-b"}

	svc.GetRotatedModels(models, "combo1", "round-robin", 1)
	svc.GetRotatedModels(models, "combo2", "round-robin", 1)

	svc.ResetComboRotation("")

	got1 := svc.GetRotatedModels(models, "combo1", "round-robin", 1)
	got2 := svc.GetRotatedModels(models, "combo2", "round-robin", 1)

	want := []string{"model-a", "model-b"}

	for i := range got1 {
		if got1[i] != want[i] {
			t.Errorf("combo1 got[%d] = %q, want %q", i, got1[i], want[i])
		}
	}

	for i := range got2 {
		if got2[i] != want[i] {
			t.Errorf("combo2 got[%d] = %q, want %q", i, got2[i], want[i])
		}
	}
}

func TestGetComboModelsFromData(t *testing.T) {
	combos := []models.Combo{
		{Name: "fast-models", Models: []string{"gpt-4o-mini", "claude-3-haiku"}},
		{Name: "smart-models", Models: []string{"gpt-4", "claude-3-opus"}},
	}

	tests := []struct {
		name       string
		modelStr   string
		wantResult []string
	}{
		{
			name:       "found combo returns models",
			modelStr:   "fast-models",
			wantResult: []string{"gpt-4o-mini", "claude-3-haiku"},
		},
		{
			name:       "another found combo",
			modelStr:   "smart-models",
			wantResult: []string{"gpt-4", "claude-3-opus"},
		},
		{
			name:       "not found returns nil",
			modelStr:   "unknown-combo",
			wantResult: nil,
		},
		{
			name:       "provider/model format returns nil",
			modelStr:   "openai/gpt-4",
			wantResult: nil,
		},
		{
			name:       "empty string returns nil",
			modelStr:   "",
			wantResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetComboModelsFromData(tt.modelStr, combos)

			if tt.wantResult == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}

			if len(got) != len(tt.wantResult) {
				t.Errorf("length mismatch: got %d, want %d", len(got), len(tt.wantResult))
				return
			}

			for i := range got {
				if got[i] != tt.wantResult[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.wantResult[i])
				}
			}
		})
	}
}

func TestGetComboModelsFromData_EmptyModels(t *testing.T) {
	combos := []models.Combo{
		{Name: "empty-combo", Models: []string{}},
	}

	got := GetComboModelsFromData("empty-combo", combos)

	if got != nil {
		t.Errorf("expected nil for empty models, got %v", got)
	}
}

func TestGetComboModelsFromData_EmptyCombos(t *testing.T) {
	got := GetComboModelsFromData("any-combo", []models.Combo{})

	if got != nil {
		t.Errorf("expected nil for empty combos, got %v", got)
	}
}
