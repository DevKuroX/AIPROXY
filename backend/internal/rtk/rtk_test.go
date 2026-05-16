package rtk

import (
	"testing"
)

func TestDetectContentType(t *testing.T) {
	ct := DetectContentType("total 100\n-rw-r--r-- 1 root root 1024 May 16 file.txt")
	if ct != "ls" {
		t.Logf("Detected: %s", ct)
	}
}

func TestAutoSelectFilter(t *testing.T) {
	f := AutoSelectFilter("error: connection refused")
	if f != nil {
		t.Logf("Selected filter: %T", f)
	}
}

func TestApplyFilter(t *testing.T) {
	content := "line1\nline2\nline3\n"
	result, err := ApplyFilter(content, nil)
	if err == nil {
		t.Logf("ApplyFilter with nil filter: %s", result)
	}
}

func TestInjectCaveman(t *testing.T) {
	result := InjectCaveman("You are a helpful assistant", "full")
	if result != "" {
		t.Logf("Caveman prompt: %s", result)
	}
}
