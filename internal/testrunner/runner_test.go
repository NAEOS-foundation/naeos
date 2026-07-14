package testrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRunnerDefaults(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{})
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
	if r.config.WorkingDir != "." {
		t.Errorf("expected default working dir '.', got %s", r.config.WorkingDir)
	}
	if r.config.Timeout != 300 {
		t.Errorf("expected default timeout 300, got %d", r.config.Timeout)
	}
}

func TestNewRunnerCustom(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{
		WorkingDir: "/tmp/test",
		Timeout:    60,
		Languages:  []string{"go"},
		Verbose:    true,
	})
	if r.config.WorkingDir != "/tmp/test" {
		t.Errorf("expected '/tmp/test', got %s", r.config.WorkingDir)
	}
	if r.config.Timeout != 60 {
		t.Errorf("expected 60, got %d", r.config.Timeout)
	}
}

func TestRunLanguageUnsupported(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{WorkingDir: "."})
	_, err := r.RunLanguage("cobol")
	if err == nil {
		t.Error("expected error for unsupported language")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected 'unsupported' in error, got: %s", err.Error())
	}
}

func TestRunAllEmpty(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{
		WorkingDir: "/nonexistent-dir",
		Languages:  []string{},
	})
	results, err := r.RunAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty languages, got %d", len(results))
	}
}

func TestFormatResultsAllPass(t *testing.T) {
	t.Parallel()
	results := []TestResult{
		{Language: "go", Passed: true, Tests: 5},
		{Language: "python", Passed: true, Tests: 3},
	}
	output := FormatResults(results)

	if !strings.Contains(output, "All tests passed!") {
		t.Error("expected 'All tests passed!'")
	}
	if !strings.Contains(output, "[PASS] go") {
		t.Error("expected PASS for go")
	}
	if !strings.Contains(output, "[PASS] python") {
		t.Error("expected PASS for python")
	}
}

func TestFormatResultsSomeFail(t *testing.T) {
	t.Parallel()
	results := []TestResult{
		{Language: "go", Passed: true, Tests: 5},
		{Language: "rust", Passed: false, Failures: 2},
	}
	output := FormatResults(results)

	if !strings.Contains(output, "Some tests failed.") {
		t.Error("expected 'Some tests failed.'")
	}
	if !strings.Contains(output, "[FAIL] rust") {
		t.Error("expected FAIL for rust")
	}
	if !strings.Contains(output, "(2 failures)") {
		t.Error("expected failure count")
	}
}

func TestParseGoOutput(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{})
	result := &TestResult{
		Output: "ok  \tpkg1\t0.01s\nok  \tpkg2\t0.02s\nFAIL\tpkg3\t0.03s",
	}
	r.parseGoOutput(result)

	if result.Tests != 2 {
		t.Errorf("expected 2 tests, got %d", result.Tests)
	}
	if result.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", result.Failures)
	}
}

func TestParseGoOutputEmpty(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{})
	result := &TestResult{Output: ""}
	r.parseGoOutput(result)

	if result.Tests != 0 {
		t.Errorf("expected 0 tests, got %d", result.Tests)
	}
	if result.Failures != 0 {
		t.Errorf("expected 0 failures, got %d", result.Failures)
	}
}

func TestDetectLanguagesGo(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{WorkingDir: "../.."})
	langs := r.detectLanguages()
	found := false
	for _, l := range langs {
		if l == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'go' to be detected (go.mod exists in project root)")
	}
}

func TestDetectLanguagesTypescript(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-ts-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	langs := r.detectLanguages()

	found := false
	for _, l := range langs {
		if l == "typescript" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'typescript' to be detected, got languages: %v", langs)
	}
}

func TestDetectLanguagesPython(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-py-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("flask\n"), 0644); err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	langs := r.detectLanguages()

	found := false
	for _, l := range langs {
		if l == "python" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'python' to be detected, got languages: %v", langs)
	}
}

func TestDetectLanguagesJava(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-java-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte("<project></project>"), 0644); err != nil {
		t.Fatalf("failed to write pom.xml: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	langs := r.detectLanguages()

	found := false
	for _, l := range langs {
		if l == "java" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'java' to be detected, got languages: %v", langs)
	}
}

func TestDetectLanguagesRust(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-rust-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte("[package]\nname = \"test\""), 0644); err != nil {
		t.Fatalf("failed to write Cargo.toml: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	langs := r.detectLanguages()

	found := false
	for _, l := range langs {
		if l == "rust" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'rust' to be detected, got languages: %v", langs)
	}
}

func TestDetectLanguagesNone(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-empty-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	langs := r.detectLanguages()

	if len(langs) != 0 {
		t.Errorf("expected 0 languages detected in empty dir, got %v", langs)
	}
}

func TestRunLanguageNodeNotFound(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-node-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	_, err = r.RunLanguage("node")
	if err == nil {
		t.Fatal("expected error when no package.json found")
	}
	if !strings.Contains(err.Error(), "no package.json found") {
		t.Errorf("expected 'no package.json found' in error, got: %s", err.Error())
	}
}

func TestRunLanguageJavaNotFound(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-java-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	_, err = r.RunLanguage("java")
	if err == nil {
		t.Fatal("expected error when no pom.xml or build.gradle found")
	}
	if !strings.Contains(err.Error(), "no pom.xml or build.gradle found") {
		t.Errorf("expected 'no pom.xml or build.gradle found' in error, got: %s", err.Error())
	}
}

func TestRunAllWithErrors(t *testing.T) {
	t.Parallel()
	r := NewRunner(TestConfig{
		WorkingDir: ".",
		Languages:  []string{"cobol"},
	})
	results, err := r.RunAll()
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Passed {
		t.Error("expected result.Passed to be false")
	}
	if results[0].Language != "cobol" {
		t.Errorf("expected language 'cobol', got '%s'", results[0].Language)
	}
	if !strings.Contains(results[0].Output, "unsupported") {
		t.Errorf("expected 'unsupported' in output, got: %s", results[0].Output)
	}
}

func TestRunGoTests(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-gotest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a minimal go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module testpkg\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Write a simple passing test
	if err := os.WriteFile(filepath.Join(tmpDir, "sample_test.go"), []byte(`package testpkg

import "testing"

func TestSample(t *testing.T) {
	t.Parallel()
	if true != true {
		t.Fatal("unexpected")
	}
}
`), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir, Timeout: 60})
	result, err := r.RunLanguage("go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Language != "go" {
		t.Errorf("expected language 'go', got '%s'", result.Language)
	}
	// The test should pass (it's a trivially passing test)
	if !result.Passed {
		t.Errorf("expected tests to pass, output: %s", result.Output)
	}
}

func TestRunNodeTestsSuccess(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-nodetest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a package.json with a passing test script
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name":"test","scripts":{"test":"echo ok"}}`), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	result, err := r.RunLanguage("node")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Language != "node" {
		t.Errorf("expected language 'node', got '%s'", result.Language)
	}
	if !result.Passed {
		t.Errorf("expected test to pass, output: %s", result.Output)
	}
}

func TestRunNodeTestsPnpm(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-pnpmtest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write package.json and pnpm-lock.yaml
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name":"test","scripts":{"test":"echo pnpm-ok"}}`), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte("lockfileVersion: '9.0'\n"), 0644); err != nil {
		t.Fatalf("failed to write pnpm-lock.yaml: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	result, err := r.RunLanguage("node")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Language != "node" {
		t.Errorf("expected language 'node', got '%s'", result.Language)
	}
}

func TestRunPythonTests(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunnerpytest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	result, err := r.RunLanguage("python")
	// python may or may not be available; the function should still return a result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Language != "python" {
		t.Errorf("expected language 'python', got '%s'", result.Language)
	}
}

func TestRunJavaTestsWithPom(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-maventest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a minimal pom.xml
	if err := os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte("<project></project>"), 0644); err != nil {
		t.Fatalf("failed to write pom.xml: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	result, err := r.RunLanguage("java")
	// mvn may not be available; the function should still return a result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Language != "java" {
		t.Errorf("expected language 'java', got '%s'", result.Language)
	}
}

func TestRunJavaTestsWithGradle(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-gradletest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write a minimal build.gradle
	if err := os.WriteFile(filepath.Join(tmpDir, "build.gradle"), []byte("apply plugin: 'java'\n"), 0644); err != nil {
		t.Fatalf("failed to write build.gradle: %v", err)
	}

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	result, err := r.RunLanguage("java")
	// gradle may not be available; the function should still return a result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Language != "java" {
		t.Errorf("expected language 'java', got '%s'", result.Language)
	}
}

func TestRunRustTests(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "testrunner-rusttest-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	r := NewRunner(TestConfig{WorkingDir: tmpDir})
	result, err := r.RunLanguage("rust")
	// cargo may not be available; the function should still return a result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Language != "rust" {
		t.Errorf("expected language 'rust', got '%s'", result.Language)
	}
}
