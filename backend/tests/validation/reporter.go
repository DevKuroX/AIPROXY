package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type ValidationReport struct {
	GeneratedAt    time.Time     `json:"generated_at"`
	FeatureSummary Summary       `json:"feature_summary"`
	TestSummary    TestSummary   `json:"test_summary"`
	ParityStatus   ParityStatus  `json:"parity_status"`
	Details        ReportDetails `json:"details"`
}

type TestSummary struct {
	TotalTests  int     `json:"total_tests"`
	Passed      int     `json:"passed"`
	Failed      int     `json:"failed"`
	Skipped     int     `json:"skipped"`
	PassRate    float64 `json:"pass_rate"`
	Duration    string  `json:"duration"`
}

type ParityStatus struct {
	OverallParity     bool              `json:"overall_parity"`
	FeatureParity     float64           `json:"feature_parity"`
	APIMethodParity   float64           `json:"api_method_parity"`
	ProviderParity    float64           `json:"provider_parity"`
	ParityByCategory  map[string]float64 `json:"parity_by_category"`
}

type ReportDetails struct {
	MissingFeatures []Feature      `json:"missing_features,omitempty"`
	FailingTests    []TestResult   `json:"failing_tests,omitempty"`
	KnownDifferences []KnownDiff   `json:"known_differences,omitempty"`
}

type KnownDiff struct {
	Feature     string `json:"feature"`
	Description string `json:"description"`
	Justification string `json:"justification"`
}

type Reporter struct {
	OutputDir string
}

func NewReporter() *Reporter {
	return &Reporter{
		OutputDir: "test-results",
	}
}

func (r *Reporter) GenerateReport(inventory *FeatureInventory, testResults *TestSuiteResult) (*ValidationReport, error) {
	report := &ValidationReport{
		GeneratedAt: time.Now(),
	}

	report.FeatureSummary = inventory.Summary

	if testResults != nil {
		report.TestSummary = TestSummary{
			TotalTests: testResults.TotalTests,
			Passed:     testResults.Passed,
			Failed:     testResults.Failed,
			Skipped:    testResults.Skipped,
			PassRate:   testResults.PassRate,
			Duration:   testResults.Duration.String(),
		}

		for _, result := range testResults.Results {
			if !result.Passed && !result.Skipped {
				report.Details.FailingTests = append(report.Details.FailingTests, result)
			}
		}
	}

	report.Details.MissingFeatures = inventory.GetMissingFeatures()

	report.ParityStatus = r.calculateParity(inventory)

	return report, nil
}

func (r *Reporter) calculateParity(inventory *FeatureInventory) ParityStatus {
	status := ParityStatus{
		ParityByCategory: make(map[string]float64),
	}

	categoryCounts := make(map[string]struct{ total, matched int })

	for _, f := range inventory.Features {
		counts := categoryCounts[f.Category]
		counts.total++
		if f.NineRouter && f.AIProxy {
			counts.matched++
		}
		categoryCounts[f.Category] = counts
	}

	for cat, counts := range categoryCounts {
		if counts.total > 0 {
			status.ParityByCategory[cat] = float64(counts.matched) / float64(counts.total) * 100
		}
	}

	if inventory.Summary.NineRouterFeatures > 0 {
		status.FeatureParity = float64(inventory.Summary.AIProxyFeatures) / float64(inventory.Summary.NineRouterFeatures) * 100
	}

	if apiParity, ok := status.ParityByCategory["api"]; ok {
		status.APIMethodParity = apiParity
	}
	if providerParity, ok := status.ParityByCategory["provider"]; ok {
		status.ProviderParity = providerParity
	}

	status.OverallParity = status.FeatureParity >= 100.0 && len(inventory.GetMissingFeatures()) == 0

	return status
}

func (r *Reporter) SaveReport(report *ValidationReport, filename string) error {
	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	path := r.OutputDir + "/" + filename
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

func (r *Reporter) GenerateMarkdownReport(report *ValidationReport) string {
	var sb strings.Builder

	sb.WriteString("# Validation Report\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", report.GeneratedAt.Format(time.RFC3339)))

	sb.WriteString("## Summary\n\n")

	parityEmoji := "✅"
	if !report.ParityStatus.OverallParity {
		parityEmoji = "⚠️"
	}
	sb.WriteString(fmt.Sprintf("- **Overall Parity**: %s %.1f%%\n", parityEmoji, report.ParityStatus.FeatureParity))
	sb.WriteString(fmt.Sprintf("- **Total Features**: %d\n", report.FeatureSummary.TotalFeatures))
	sb.WriteString(fmt.Sprintf("- **9router Features**: %d\n", report.FeatureSummary.NineRouterFeatures))
	sb.WriteString(fmt.Sprintf("- **ai_proxy Features**: %d\n", report.FeatureSummary.AIProxyFeatures))
	sb.WriteString(fmt.Sprintf("- **Tested Features**: %d\n", report.FeatureSummary.TestedFeatures))
	sb.WriteString(fmt.Sprintf("- **Passing Features**: %d\n", report.FeatureSummary.PassingFeatures))

	sb.WriteString("\n## Test Results\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tests**: %d\n", report.TestSummary.TotalTests))
	sb.WriteString(fmt.Sprintf("- **Passed**: %d\n", report.TestSummary.Passed))
	sb.WriteString(fmt.Sprintf("- **Failed**: %d\n", report.TestSummary.Failed))
	sb.WriteString(fmt.Sprintf("- **Skipped**: %d\n", report.TestSummary.Skipped))
	sb.WriteString(fmt.Sprintf("- **Pass Rate**: %.1f%%\n", report.TestSummary.PassRate))
	sb.WriteString(fmt.Sprintf("- **Duration**: %s\n", report.TestSummary.Duration))

	sb.WriteString("\n## Parity by Category\n\n")
	sb.WriteString("| Category | Parity |\n")
	sb.WriteString("|----------|--------|\n")
	for cat, parity := range report.ParityStatus.ParityByCategory {
		emoji := "✅"
		if parity < 100 {
			emoji = "⚠️"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s %.1f%% |\n", cat, emoji, parity))
	}

	if len(report.Details.MissingFeatures) > 0 {
		sb.WriteString("\n## Missing Features\n\n")
		for _, f := range report.Details.MissingFeatures {
			sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", f.Name, f.Category))
		}
	}

	if len(report.Details.FailingTests) > 0 {
		sb.WriteString("\n## Failing Tests\n\n")
		for _, t := range report.Details.FailingTests {
			sb.WriteString(fmt.Sprintf("- **%s/%s**: %s\n", t.Package, t.TestName, t.Error))
		}
	}

	if len(report.Details.KnownDifferences) > 0 {
		sb.WriteString("\n## Known Differences\n\n")
		for _, d := range report.Details.KnownDifferences {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n  - Justification: %s\n", d.Feature, d.Description, d.Justification))
		}
	}

	sb.WriteString("\n---\n")
	if report.ParityStatus.OverallParity {
		sb.WriteString("✅ **100% parity achieved with 9router**\n")
	} else {
		sb.WriteString(fmt.Sprintf("⚠️ **Parity incomplete: %.1f%%**\n", report.ParityStatus.FeatureParity))
	}

	return sb.String()
}

func (r *Reporter) SaveMarkdownReport(report *ValidationReport, filename string) error {
	markdown := r.GenerateMarkdownReport(report)

	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	path := r.OutputDir + "/" + filename
	if err := os.WriteFile(path, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

func GenerateValidationReport() error {
	runner := NewTestRunner()
	reporter := NewReporter()

	inventory, err := runner.RunFeatureInventory()
	if err != nil {
		return fmt.Errorf("failed to generate feature inventory: %w", err)
	}

	testResults, err := runner.RunAll()
	if err != nil {
		fmt.Printf("Warning: some tests failed: %v\n", err)
	}

	report, err := reporter.GenerateReport(inventory, testResults)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	if err := reporter.SaveReport(report, "validation-report.json"); err != nil {
		return fmt.Errorf("failed to save JSON report: %w", err)
	}

	if err := reporter.SaveMarkdownReport(report, "validation-report.md"); err != nil {
		return fmt.Errorf("failed to save Markdown report: %w", err)
	}

	fmt.Println(reporter.GenerateMarkdownReport(report))
	return nil
}

func GenerateParityReport() error {
	runner := NewTestRunner()

	inventory, err := runner.RunFeatureInventory()
	if err != nil {
		return fmt.Errorf("failed to generate feature inventory: %w", err)
	}

	fmt.Println("=== Parity Report ===")
	fmt.Printf("Total Features: %d\n", inventory.Summary.TotalFeatures)
	fmt.Printf("9router Features: %d\n", inventory.Summary.NineRouterFeatures)
	fmt.Printf("ai_proxy Features: %d\n", inventory.Summary.AIProxyFeatures)
	fmt.Printf("Parity: %d%%\n", inventory.Summary.ParityPercentage)

	missing := inventory.GetMissingFeatures()
	if len(missing) > 0 {
		fmt.Printf("\nMissing Features (%d):\n", len(missing))
		for _, f := range missing {
			fmt.Printf("  - %s (%s)\n", f.Name, f.Category)
		}
		return fmt.Errorf("parity incomplete: %d missing features", len(missing))
	}

	fmt.Println("\n✅ 100% feature parity achieved")
	return nil
}
