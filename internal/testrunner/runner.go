package testrunner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TestConfig struct {
	WorkingDir string
	Languages  []string
	Verbose    bool
	Timeout    int
	Parallel   bool
}

type TestResult struct {
	Language string
	Passed   bool
	Output   string
	Tests    int
	Failures int
}

type Runner struct {
	config TestConfig
}

func NewRunner(config TestConfig) *Runner {
	if config.WorkingDir == "" {
		config.WorkingDir = "."
	}
	if config.Timeout == 0 {
		config.Timeout = 300
	}
	return &Runner{config: config}
}

func (r *Runner) RunAll() ([]TestResult, error) {
	var results []TestResult

	languages := r.config.Languages
	if len(languages) == 0 {
		languages = r.detectLanguages()
	}

	for _, lang := range languages {
		result, err := r.RunLanguage(lang)
		if err != nil {
			result = &TestResult{
				Language: lang,
				Passed:   false,
				Output:   err.Error(),
			}
		}
		results = append(results, *result)
	}

	return results, nil
}

func (r *Runner) RunLanguage(lang string) (*TestResult, error) {
	result := &TestResult{Language: lang}

	switch lang {
	case "go":
		return r.runGoTests(result)
	case "typescript", "node":
		return r.runNodeTests(result)
	case "python":
		return r.runPythonTests(result)
	case "java":
		return r.runJavaTests(result)
	case "rust":
		return r.runRustTests(result)
	default:
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
}

func (r *Runner) runGoTests(result *TestResult) (*TestResult, error) {
	args := []string{"test", "-v", "-count=1"}
	if r.config.Timeout > 0 {
		args = append(args, fmt.Sprintf("-timeout=%ds", r.config.Timeout))
	}
	args = append(args, "./...")

	cmd := exec.Command("go", args...)
	cmd.Dir = r.config.WorkingDir
	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Passed = false
		r.parseGoOutput(result)
	} else {
		result.Passed = true
		r.parseGoOutput(result)
	}

	return result, nil
}

func (r *Runner) runNodeTests(result *TestResult) (*TestResult, error) {
	packageJSON := filepath.Join(r.config.WorkingDir, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		return nil, fmt.Errorf("no package.json found")
	}

	npmCmd := "npm"
	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "pnpm-lock.yaml")); err == nil {
			npmCmd = "pnpm"
		}

	cmd := exec.Command(npmCmd, "test")
	cmd.Dir = r.config.WorkingDir
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Passed = err == nil

	return result, nil
}

func (r *Runner) runPythonTests(result *TestResult) (*TestResult, error) {
	cmd := exec.Command("python", "-m", "pytest", "-v")
	cmd.Dir = r.config.WorkingDir
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Passed = err == nil

	return result, nil
}

func (r *Runner) runJavaTests(result *TestResult) (*TestResult, error) {
	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "pom.xml")); err == nil {
		cmd := exec.Command("mvn", "test")
		cmd.Dir = r.config.WorkingDir
		output, err := cmd.CombinedOutput()
		result.Output = string(output)
		result.Passed = err == nil
		return result, nil
	}

	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "build.gradle")); err == nil {
		cmd := exec.Command("./gradlew", "test")
		cmd.Dir = r.config.WorkingDir
		output, err := cmd.CombinedOutput()
		result.Output = string(output)
		result.Passed = err == nil
		return result, nil
	}

	return nil, fmt.Errorf("no pom.xml or build.gradle found")
}

func (r *Runner) runRustTests(result *TestResult) (*TestResult, error) {
	cmd := exec.Command("cargo", "test", "--verbose")
	cmd.Dir = r.config.WorkingDir
	output, err := cmd.CombinedOutput()
	result.Output = string(output)
	result.Passed = err == nil

	return result, nil
}

func (r *Runner) detectLanguages() []string {
	var langs []string

	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "go.mod")); err == nil {
		langs = append(langs, "go")
	}
	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "package.json")); err == nil {
		langs = append(langs, "typescript")
	}
	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "requirements.txt")); err == nil {
		langs = append(langs, "python")
	}
	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "pom.xml")); err == nil {
		langs = append(langs, "java")
	}
	if _, err := os.Stat(filepath.Join(r.config.WorkingDir, "Cargo.toml")); err == nil {
		langs = append(langs, "rust")
	}

	return langs
}

func (r *Runner) parseGoOutput(result *TestResult) {
	lines := strings.Split(result.Output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ok") && strings.Contains(line, "?") == false {
			result.Tests++
		}
		if strings.Contains(line, "FAIL") {
			result.Failures++
		}
	}
}

func FormatResults(results []TestResult) string {
	var sb strings.Builder
	sb.WriteString("Test Results:\n\n")

	allPassed := true
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
			allPassed = false
		}
		sb.WriteString(fmt.Sprintf("  [%s] %s", status, r.Language))
		if r.Tests > 0 {
			sb.WriteString(fmt.Sprintf(" (%d tests)", r.Tests))
		}
		if r.Failures > 0 {
			sb.WriteString(fmt.Sprintf(" (%d failures)", r.Failures))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	if allPassed {
		sb.WriteString("All tests passed!\n")
	} else {
		sb.WriteString("Some tests failed.\n")
	}

	return sb.String()
}
