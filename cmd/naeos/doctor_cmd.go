package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/lint"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

type checkResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func newDoctorCommand() *cobra.Command {
	var configPath string
	var specFile string
	var quick bool

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics on the NAEOS environment and configuration",
		Long: `Run comprehensive diagnostics to check the health of your NAEOS setup.

Checks include:
  - Go toolchain and version
  - Language runtimes (Node, Python, Java, Rust)
  - Docker and container tools
  - Git version
  - NAEOS configuration
  - Spec validation (if spec provided)
  - Network connectivity`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			var results []checkResult

			// System info
			fmt.Fprintf(out, "NAEOS Doctor v0.2.0\n")
			fmt.Fprintf(out, "Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Fprintf(out, "Go: %s\n\n", runtime.Version())

			// 1. Go toolchain
			results = append(results, checkGo())

			// 2. Language runtimes (skip in quick mode)
			if !quick {
				results = append(results, checkNode())
				results = append(results, checkPython())
				results = append(results, checkJava())
				results = append(results, checkRust())
			}

			// 3. Docker
			results = append(results, checkDocker())

			// 4. Git
			results = append(results, checkGit())

			// 5. NAEOS config
			results = append(results, checkConfig(configPath))

			// 6. Spec validation
			if specFile != "" {
				results = append(results, checkSpec(specFile))
			}

			// 7. Network (skip in quick mode)
			if !quick {
				results = append(results, checkNetwork())
			}

			// Print results
			passed := 0
			warned := 0
			failed := 0
			for _, r := range results {
				icon := "  OK"
				switch r.Status {
				case "warn":
					icon = " WARN"
					warned++
				case "fail":
					icon = " FAIL"
					failed++
				default:
					passed++
				}
				if r.Detail != "" {
					fmt.Fprintf(out, "[%s] %s — %s\n", icon, r.Name, r.Detail)
				} else {
					fmt.Fprintf(out, "[%s] %s\n", icon, r.Name)
				}
			}

			fmt.Fprintf(out, "\nResults: %d passed, %d warnings, %d failed\n", passed, warned, failed)
			if failed > 0 {
				return fmt.Errorf("%d check(s) failed", failed)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	cmd.Flags().StringVar(&specFile, "spec", "", "path to spec file for validation")
	cmd.Flags().BoolVar(&quick, "quick", false, "skip language runtime and network checks")
	return cmd
}

func checkGo() checkResult {
	path, err := exec.LookPath("go")
	if err != nil {
		return checkResult{Name: "Go toolchain", Status: "fail", Detail: "go not found in PATH"}
	}
	out, err := exec.Command(path, "version").Output()
	if err != nil {
		return checkResult{Name: "Go toolchain", Status: "fail", Detail: "go version failed"}
	}
	return checkResult{Name: "Go toolchain", Status: "pass", Detail: strings.TrimSpace(string(out))}
}

func checkNode() checkResult {
	path, err := exec.LookPath("node")
	if err != nil {
		return checkResult{Name: "Node.js", Status: "warn", Detail: "not installed (optional for TypeScript/JS projects)"}
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return checkResult{Name: "Node.js", Status: "warn", Detail: "installed but version check failed"}
	}
	detail := strings.TrimSpace(string(out))

	npmPath, _ := exec.LookPath("npm")
	if npmPath != "" {
		detail += " (npm available)"
	}

	return checkResult{Name: "Node.js", Status: "pass", Detail: detail}
}

func checkPython() checkResult {
	for _, bin := range []string{"python3", "python"} {
		path, err := exec.LookPath(bin)
		if err != nil {
			continue
		}
		out, err := exec.Command(path, "--version").Output()
		if err != nil {
			continue
		}
		detail := strings.TrimSpace(string(out))

		pipPath, _ := exec.LookPath("pip3")
		if pipPath == "" {
			pipPath, _ = exec.LookPath("pip")
		}
		if pipPath != "" {
			detail += " (pip available)"
		}

		return checkResult{Name: "Python", Status: "pass", Detail: detail}
	}
	return checkResult{Name: "Python", Status: "warn", Detail: "not installed (optional for Python projects)"}
}

func checkJava() checkResult {
	path, err := exec.LookPath("java")
	if err != nil {
		return checkResult{Name: "Java", Status: "warn", Detail: "not installed (optional for Java projects)"}
	}
	out, err := exec.Command(path, "-version").Output()
	if err != nil {
		return checkResult{Name: "Java", Status: "warn", Detail: "installed but version check failed"}
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	detail := lines[0]
	if len(lines) > 0 {
		detail = strings.TrimPrefix(lines[0], "java version ")
		detail = strings.Trim(detail, "\"")
	}

	mvnPath, _ := exec.LookPath("mvn")
	gradlePath, _ := exec.LookPath("gradle")
	var tools []string
	if mvnPath != "" {
		tools = append(tools, "maven")
	}
	if gradlePath != "" {
		tools = append(tools, "gradle")
	}
	if len(tools) > 0 {
		detail += " (" + strings.Join(tools, ", ") + " available)"
	}

	return checkResult{Name: "Java", Status: "pass", Detail: detail}
}

func checkRust() checkResult {
	path, err := exec.LookPath("rustc")
	if err != nil {
		return checkResult{Name: "Rust", Status: "warn", Detail: "not installed (optional for Rust projects)"}
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return checkResult{Name: "Rust", Status: "warn", Detail: "installed but version check failed"}
	}
	detail := strings.TrimSpace(string(out))

	cargoPath, _ := exec.LookPath("cargo")
	if cargoPath != "" {
		detail += " (cargo available)"
	}

	return checkResult{Name: "Rust", Status: "pass", Detail: detail}
}

func checkDocker() checkResult {
	path, err := exec.LookPath("docker")
	if err != nil {
		return checkResult{Name: "Docker", Status: "warn", Detail: "not installed (optional for containerized deployment)"}
	}
	out, err := exec.Command(path, "version", "--format", "{{.Server.Version}}").Output()
	if err != nil {
		return checkResult{Name: "Docker", Status: "warn", Detail: "installed but daemon not running"}
	}
	return checkResult{Name: "Docker", Status: "pass", Detail: "v" + strings.TrimSpace(string(out))}
}

func checkGit() checkResult {
	path, err := exec.LookPath("git")
	if err != nil {
		return checkResult{Name: "Git", Status: "fail", Detail: "git not found in PATH"}
	}
	out, err := exec.Command(path, "version").Output()
	if err != nil {
		return checkResult{Name: "Git", Status: "fail", Detail: "git version failed"}
	}
	return checkResult{Name: "Git", Status: "pass", Detail: strings.TrimSpace(string(out))}
}

func checkConfig(configPath string) checkResult {
	resolved, err := resolveConfigPath(configPath)
	if err != nil {
		return checkResult{Name: "NAEOS Config", Status: "warn", Detail: "no config found (using defaults)"}
	}

	cfg, err := pipeline.ConfigFromFile(resolved)
	if err != nil {
		return checkResult{Name: "NAEOS Config", Status: "fail", Detail: fmt.Sprintf("invalid config: %v", err)}
	}

	detail := fmt.Sprintf("name=%s mode=%s", cfg.Name, cfg.Mode)
	if cfg.OutputDir != "" {
		if _, err := os.Stat(cfg.OutputDir); err == nil {
			detail += fmt.Sprintf(" output_dir=%s [exists]", cfg.OutputDir)
		} else {
			detail += fmt.Sprintf(" output_dir=%s [missing]", cfg.OutputDir)
		}
	}

	return checkResult{Name: "NAEOS Config", Status: "pass", Detail: detail}
}

func checkSpec(specFile string) checkResult {
	content, err := os.ReadFile(specFile)
	if err != nil {
		return checkResult{Name: "Spec Validation", Status: "fail", Detail: fmt.Sprintf("cannot read %s", specFile)}
	}

	issues := lint.ValidateSpec(string(content))
	if len(issues) == 0 {
		return checkResult{Name: "Spec Validation", Status: "pass", Detail: "valid"}
	}

	errors := 0
	warnings := 0
	for _, issue := range issues {
		if issue.Severity == lint.SeverityError {
			errors++
		} else {
			warnings++
		}
	}

	detail := fmt.Sprintf("%d error(s), %d warning(s)", errors, warnings)
	if errors > 0 {
		for _, issue := range issues {
			if issue.Severity == lint.SeverityError {
				detail += "\n    " + issue.Message
				break
			}
		}
		return checkResult{Name: "Spec Validation", Status: "fail", Detail: detail}
	}
	return checkResult{Name: "Spec Validation", Status: "warn", Detail: detail}
}

func checkNetwork() checkResult {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://github.com")
	if err != nil {
		return checkResult{Name: "Network", Status: "warn", Detail: "cannot reach github.com"}
	}
	resp.Body.Close()
	return checkResult{Name: "Network", Status: "pass", Detail: "github.com reachable"}
}
