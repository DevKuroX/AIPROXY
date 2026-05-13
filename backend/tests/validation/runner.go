package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type TestRunner struct {
	OutputDir    string
	Verbose      bool
	FailFast     bool
	Timeout      time.Duration
	TestPackages []string
}

type TestResult struct {
	Package     string        `json:"package"`
	TestName    string        `json:"test_name"`
	Passed      bool          `json:"passed"`
	Duration    time.Duration `json:"duration"`
	Output      string        `json:"output,omitempty"`
	Error       string        `json:"error,omitempty"`
	Skipped     bool          `json:"skipped"`
	SkipReason  string        `json:"skip_reason,omitempty"`
}

type TestSuiteResult struct {
	Timestamp   time.Time    `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
	TotalTests  int          `json:"total_tests"`
	Passed      int          `json:"passed"`
	Failed      int          `json:"failed"`
	Skipped     int          `json:"skipped"`
	PassRate    float64      `json:"pass_rate"`
	Results     []TestResult `json:"results"`
}

func NewTestRunner() *TestRunner {
	return &TestRunner{
		OutputDir:    "test-results",
		Verbose:      true,
		FailFast:     false,
		Timeout:      5 * time.Minute,
		TestPackages: []string{"./tests/validation/...", "./tests/integration/..."},
	}
}

func (r *TestRunner) RunAll() (*TestSuiteResult, error) {
	start := time.Now()
	result := &TestSuiteResult{
		Timestamp: start,
		Results:   []TestResult{},
	}

	for _, pkg := range r.TestPackages {
		suiteResult, err := r.RunPackage(pkg)
		if err != nil {
			return nil, fmt.Errorf("failed to run package %s: %w", pkg, err)
		}

		result.TotalTests += suiteResult.TotalTests
		result.Passed += suiteResult.Passed
		result.Failed += suiteResult.Failed
		result.Skipped += suiteResult.Skipped
		result.Results = append(result.Results, suiteResult.Results...)
	}

	result.Duration = time.Since(start)
	if result.TotalTests > 0 {
		result.PassRate = float64(result.Passed) / float64(result.TotalTests) * 100
	}

	return result, nil
}

func (r *TestRunner) RunPackage(pkg string) (*TestSuiteResult, error) {
	start := time.Now()
	result := &TestSuiteResult{
		Timestamp: start,
		Results:   []TestResult{},
	}

	args := []string{"test", "-json"}
	if r.Verbose {
		args = append(args, "-v")
	}
	if r.FailFast {
		args = append(args, "-failfast")
	}
	args = append(args, pkg)

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "no test files") {
		return nil, fmt.Errorf("test command failed: %w\n%s", err, string(output))
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var event testEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
			testResult := TestResult{
				Package:  event.Package,
				TestName: event.Test,
				Duration: time.Duration(event.Elapsed * float64(time.Second)),
			}

			switch event.Action {
			case "pass":
				testResult.Passed = true
				result.Passed++
			case "fail":
				testResult.Passed = false
				testResult.Error = event.Output
				result.Failed++
			case "skip":
				testResult.Skipped = true
				testResult.SkipReason = event.Output
				result.Skipped++
			}

			result.TotalTests++
			result.Results = append(result.Results, testResult)
		}
	}

	result.Duration = time.Since(start)
	if result.TotalTests > 0 {
		result.PassRate = float64(result.Passed) / float64(result.TotalTests) * 100
	}

	return result, nil
}

func (r *TestRunner) RunFeatureInventory() (*FeatureInventory, error) {
	return GenerateFeatureInventory()
}

func (r *TestRunner) RunBenchmarks() ([]BenchmarkResult, error) {
	return []BenchmarkResult{}, nil
}

func (r *TestRunner) SaveResults(result *TestSuiteResult, filename string) error {
	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	path := r.OutputDir + "/" + filename
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	return nil
}

type testEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

func RunValidation() error {
	runner := NewTestRunner()

	fmt.Println("=== Running Feature Inventory ===")
	inventory, err := runner.RunFeatureInventory()
	if err != nil {
		return fmt.Errorf("feature inventory failed: %w", err)
	}
	fmt.Printf("Features: %d total, %d in 9router, %d in ai_proxy\n",
		inventory.Summary.TotalFeatures,
		inventory.Summary.NineRouterFeatures,
		inventory.Summary.AIProxyFeatures)

	missing := inventory.GetMissingFeatures()
	if len(missing) > 0 {
		fmt.Printf("Missing features: %d\n", len(missing))
		for _, f := range missing {
			fmt.Printf("  - %s (%s)\n", f.Name, f.Category)
		}
	}

	fmt.Println("\n=== Running Validation Tests ===")
	results, err := runner.RunAll()
	if err != nil {
		return fmt.Errorf("test run failed: %w", err)
	}

	fmt.Printf("\nTest Results:\n")
	fmt.Printf("  Total:   %d\n", results.TotalTests)
	fmt.Printf("  Passed:  %d\n", results.Passed)
	fmt.Printf("  Failed:  %d\n", results.Failed)
	fmt.Printf("  Skipped: %d\n", results.Skipped)
	fmt.Printf("  Pass Rate: %.1f%%\n", results.PassRate)
	fmt.Printf("  Duration: %v\n", results.Duration)

	if err := runner.SaveResults(results, "test-results.json"); err != nil {
		fmt.Printf("Warning: failed to save results: %v\n", err)
	}

	if err := inventory.SaveToFile(runner.OutputDir + "/feature-inventory.json"); err != nil {
		fmt.Printf("Warning: failed to save inventory: %v\n", err)
	}

	if results.Failed > 0 {
		return fmt.Errorf("%d tests failed", results.Failed)
	}

	fmt.Println("\n=== Validation Complete ===")
	return nil
}
