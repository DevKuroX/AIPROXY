package validation

import (
	"encoding/json"
	"testing"
)

func TestGenerateFeatureInventory(t *testing.T) {
	inventory, err := GenerateFeatureInventory()
	if err != nil {
		t.Fatalf("GenerateFeatureInventory() error = %v", err)
	}
	if inventory == nil {
		t.Fatal("GenerateFeatureInventory() returned nil")
	}
	if len(inventory.Features) == 0 {
		t.Fatal("GenerateFeatureInventory() returned empty features list")
	}

	t.Run("has_api_features", func(t *testing.T) {
		apiFeatures := inventory.GetFeaturesByCategory("api")
		if len(apiFeatures) == 0 {
			t.Error("expected API features to be present")
		}
	})

	t.Run("has_provider_features", func(t *testing.T) {
		providerFeatures := inventory.GetFeaturesByCategory("provider")
		if len(providerFeatures) == 0 {
			t.Error("expected provider features to be present")
		}
	})

	t.Run("has_rtk_features", func(t *testing.T) {
		rtkFeatures := inventory.GetFeaturesByCategory("rtk")
		if len(rtkFeatures) == 0 {
			t.Error("expected RTK features to be present")
		}
	})

	t.Run("summary_calculated", func(t *testing.T) {
		if inventory.Summary.TotalFeatures != len(inventory.Features) {
			t.Errorf("summary TotalFeatures = %d, want %d", inventory.Summary.TotalFeatures, len(inventory.Features))
		}
		if inventory.Summary.ParityPercentage > 100 {
			t.Errorf("parity percentage %d exceeds 100", inventory.Summary.ParityPercentage)
		}
	})
}

func TestFeatureInventory_MarkTested(t *testing.T) {
	inventory, _ := GenerateFeatureInventory()

	featureName := "POST /v1/chat/completions"
	inventory.MarkTested(featureName, true)

	found := false
	for _, f := range inventory.Features {
		if f.Name == featureName {
			found = true
			if !f.Tested {
				t.Error("feature not marked as tested")
			}
			if !f.Passing {
				t.Error("feature not marked as passing")
			}
			break
		}
	}
	if !found {
		t.Errorf("feature %s not found in inventory", featureName)
	}
}

func TestFeatureInventory_GetMissingFeatures(t *testing.T) {
	inventory := &FeatureInventory{
		Features: []Feature{
			{Name: "feature1", NineRouter: true, AIProxy: true},
			{Name: "feature2", NineRouter: true, AIProxy: false},
			{Name: "feature3", NineRouter: false, AIProxy: true},
		},
	}

	missing := inventory.GetMissingFeatures()
	if len(missing) != 1 {
		t.Errorf("GetMissingFeatures() returned %d features, want 1", len(missing))
	}
	if len(missing) > 0 && missing[0].Name != "feature2" {
		t.Errorf("missing feature = %s, want feature2", missing[0].Name)
	}
}

func TestFeatureInventory_GetFailingTests(t *testing.T) {
	inventory := &FeatureInventory{
		Features: []Feature{
			{Name: "feature1", Tested: true, Passing: true},
			{Name: "feature2", Tested: true, Passing: false},
			{Name: "feature3", Tested: false, Passing: false},
		},
	}

	failing := inventory.GetFailingTests()
	if len(failing) != 1 {
		t.Errorf("GetFailingTests() returned %d features, want 1", len(failing))
	}
	if len(failing) > 0 && failing[0].Name != "feature2" {
		t.Errorf("failing feature = %s, want feature2", failing[0].Name)
	}
}

func TestFeatureInventory_JSON(t *testing.T) {
	inventory, _ := GenerateFeatureInventory()

	jsonStr, err := inventory.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}
	if jsonStr == "" {
		t.Error("ToJSON() returned empty string")
	}

	loaded, err := LoadFromJSONString(jsonStr)
	if err != nil {
		t.Fatalf("LoadFromJSONString() error = %v", err)
	}
	if len(loaded.Features) != len(inventory.Features) {
		t.Errorf("loaded features = %d, want %d", len(loaded.Features), len(inventory.Features))
	}
}

func LoadFromJSONString(jsonStr string) (*FeatureInventory, error) {
	var inventory FeatureInventory
	if err := json.Unmarshal([]byte(jsonStr), &inventory); err != nil {
		return nil, err
	}
	return &inventory, nil
}

func TestValidateFeatureInventory(t *testing.T) {
	t.Run("no_errors_when_complete", func(t *testing.T) {
		inventory := &FeatureInventory{
			Features: []Feature{
				{Name: "feature1", NineRouter: true, AIProxy: true},
				{Name: "feature2", NineRouter: true, AIProxy: true},
			},
		}
		errors := ValidateFeatureInventory(inventory)
		if len(errors) != 0 {
			t.Errorf("ValidateFeatureInventory() returned %d errors, want 0", len(errors))
		}
	})

	t.Run("errors_on_missing_features", func(t *testing.T) {
		inventory := &FeatureInventory{
			Features: []Feature{
				{Name: "feature1", NineRouter: true, AIProxy: false},
			},
		}
		errors := ValidateFeatureInventory(inventory)
		if len(errors) == 0 {
			t.Error("ValidateFeatureInventory() should return errors for missing features")
		}
	})

	t.Run("errors_on_failing_tests", func(t *testing.T) {
		inventory := &FeatureInventory{
			Features: []Feature{
				{Name: "feature1", NineRouter: true, AIProxy: true, Tested: true, Passing: false},
			},
		}
		errors := ValidateFeatureInventory(inventory)
		if len(errors) == 0 {
			t.Error("ValidateFeatureInventory() should return errors for failing tests")
		}
	})
}

func TestGetAPIEndpointFeatures(t *testing.T) {
	features := getAPIEndpointFeatures()

	requiredEndpoints := []string{
		"POST /v1/chat/completions",
		"GET /v1/models",
		"POST /v1/embeddings",
		"POST /v1/images/generations",
		"POST /v1/audio/speech",
		"POST /v1/search",
		"POST /v1/fetch",
	}

	for _, endpoint := range requiredEndpoints {
		found := false
		for _, f := range features {
			if f.Name == endpoint {
				found = true
				if !f.NineRouter || !f.AIProxy {
					t.Errorf("endpoint %s: NineRouter=%v AIProxy=%v, want both true", endpoint, f.NineRouter, f.AIProxy)
				}
				break
			}
		}
		if !found {
			t.Errorf("required endpoint %s not found in features", endpoint)
		}
	}
}

func TestGetProviderFeatures(t *testing.T) {
	features := getProviderFeatures()

	t.Logf("Got %d features from getProviderFeatures()", len(features))
	for i, f := range features {
		t.Logf("  Feature %d: Name=%q Category=%q", i, f.Name, f.Category)
	}

	requiredProviders := []string{
		"OpenAI provider",
		"Claude provider",
		"Gemini provider",
		"Ollama provider",
	}

	for _, provider := range requiredProviders {
		found := false
		for _, f := range features {
			if f.Name == provider {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required provider %s not found in features", provider)
		}
	}
}

func TestGetRTKFeatures(t *testing.T) {
	features := getRTKFeatures()

	requiredFilters := []string{
		"RTK dedup_log filter",
		"RTK find filter",
		"RTK grep filter",
		"RTK smart_truncate filter",
	}

	for _, filter := range requiredFilters {
		found := false
		for _, f := range features {
			if f.Name == filter {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required RTK filter %s not found in features", filter)
		}
	}
}
