package services

import (
	"testing"
	"time"
)

func TestCacheConstants(t *testing.T) {
	if CacheTTL != 1*time.Hour {
		t.Errorf("CacheTTL = %v, want 1 hour", CacheTTL)
	}

	if PendingTTL != 2*time.Minute {
		t.Errorf("PendingTTL = %v, want 2 minutes", PendingTTL)
	}

	if CleanupInterval != 10*time.Minute {
		t.Errorf("CleanupInterval = %v, want 10 minutes", CleanupInterval)
	}
}

func TestPlatformConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		want     int
	}{
		{"PlatformUnspecified", PlatformUnspecified, 0},
		{"PlatformDarwinAMD64", PlatformDarwinAMD64, 1},
		{"PlatformDarwinARM64", PlatformDarwinARM64, 2},
		{"PlatformLinuxAMD64", PlatformLinuxAMD64, 3},
		{"PlatformLinuxARM64", PlatformLinuxARM64, 4},
		{"PlatformWindowsAMD64", PlatformWindowsAMD64, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.want)
			}
		})
	}
}

func TestIDEConstants(t *testing.T) {
	if IDETypeAntigravity != 9 {
		t.Errorf("IDETypeAntigravity = %d, want 9", IDETypeAntigravity)
	}

	if PluginTypeGemini != 2 {
		t.Errorf("PluginTypeGemini = %d, want 2", PluginTypeGemini)
	}
}

func TestCloudCodeEndpoints(t *testing.T) {
	if CloudCodeLoadCodeAssist != "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist" {
		t.Errorf("CloudCodeLoadCodeAssist has unexpected value")
	}

	if CloudCodeOnboardUser != "https://cloudcode-pa.googleapis.com/v1internal:onboardUser" {
		t.Errorf("CloudCodeOnboardUser has unexpected value")
	}
}

func TestCacheEntry(t *testing.T) {
	now := time.Now()
	entry := cacheEntry{
		ProjectID: "test-project-123",
		FetchedAt: now,
	}

	if entry.ProjectID != "test-project-123" {
		t.Errorf("ProjectID = %q, want test-project-123", entry.ProjectID)
	}

	if !entry.FetchedAt.Equal(now) {
		t.Errorf("FetchedAt mismatch")
	}
}

func TestPendingFetch(t *testing.T) {
	resultChan := make(chan string)
	cancel := func() {}

	pending := pendingFetch{
		Result:    resultChan,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}

	if pending.Result == nil {
		t.Error("Result channel should not be nil")
	}

	if pending.Cancel == nil {
		t.Error("Cancel function should not be nil")
	}
}

func TestMetadataRequest(t *testing.T) {
	req := metadataRequest{
		IDEType:    IDETypeAntigravity,
		Platform:   PlatformLinuxAMD64,
		PluginType: PluginTypeGemini,
	}

	if req.IDEType != IDETypeAntigravity {
		t.Errorf("IDEType = %d, want %d", req.IDEType, IDETypeAntigravity)
	}

	if req.Platform != PlatformLinuxAMD64 {
		t.Errorf("Platform = %d, want %d", req.Platform, PlatformLinuxAMD64)
	}

	if req.PluginType != PluginTypeGemini {
		t.Errorf("PluginType = %d, want %d", req.PluginType, PluginTypeGemini)
	}
}

func TestLoadCodeAssistRequest(t *testing.T) {
	req := loadCodeAssistRequest{
		Metadata: metadataRequest{
			IDEType:    IDETypeAntigravity,
			Platform:   PlatformDarwinARM64,
			PluginType: PluginTypeGemini,
		},
	}

	if req.Metadata.IDEType != IDETypeAntigravity {
		t.Errorf("Metadata.IDEType = %d, want %d", req.Metadata.IDEType, IDETypeAntigravity)
	}
}
