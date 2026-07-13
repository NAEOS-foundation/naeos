package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/version"
	"github.com/spf13/cobra"
)

func newHealthCommand() *cobra.Command {
	outputFormat := "text"

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Run system health checks and diagnostics",
		Long: `Perform comprehensive health checks on the NAEOS installation,
configuration, and dependencies.

Example:
  naeos health
  naeos health --output json
  naeos health --output yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := runHealthChecks()
			return renderHealthReport(cmd, report, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json, yaml")
	return cmd
}

type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthReport struct {
	Status   string        `json:"status"`
	Version  string        `json:"version"`
	Go       string        `json:"go_version"`
	Platform string        `json:"platform"`
	Checks   []HealthCheck `json:"checks"`
}

func runHealthChecks() *HealthReport {
	report := &HealthReport{
		Version:  "0.6.0",
		Go:       runtime.Version(),
		Platform: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	checks := []HealthCheck{
		checkGoBinary(),
		checkGitBinary(),
		checkConfigDir(),
		checkVersionFile(),
	}

	allHealthy := true
	for _, c := range checks {
		if c.Status != "healthy" {
			allHealthy = false
		}
	}
	if allHealthy {
		report.Status = "healthy"
	} else {
		report.Status = "degraded"
	}
	report.Checks = checks
	return report
}

func checkGoBinary() HealthCheck {
	_, err := exec.LookPath("go")
	if err != nil {
		return HealthCheck{Name: "go_binary", Status: "unhealthy", Message: "go not found in PATH"}
	}
	return HealthCheck{Name: "go_binary", Status: "healthy"}
}

func checkGitBinary() HealthCheck {
	_, err := exec.LookPath("git")
	if err != nil {
		return HealthCheck{Name: "git_binary", Status: "unhealthy", Message: "git not found in PATH"}
	}
	return HealthCheck{Name: "git_binary", Status: "healthy"}
}

func checkConfigDir() HealthCheck {
	home, err := os.UserHomeDir()
	if err != nil {
		return HealthCheck{Name: "config_dir", Status: "degraded", Message: "cannot determine home directory"}
	}
	configDir := home + "/.config/naeos"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return HealthCheck{Name: "config_dir", Status: "degraded", Message: "will be created on first use"}
	}
	return HealthCheck{Name: "config_dir", Status: "healthy"}
}

func checkVersionFile() HealthCheck {
	return HealthCheck{Name: "version", Status: "healthy", Message: version.String()}
}

func renderHealthReport(cmd *cobra.Command, report *HealthReport, format string) error {
	switch format {
	case "json":
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf(`{"status":"%s","version":"%s","go_version":"%s","platform":"%s","checks":[`, report.Status, report.Version, report.Go, report.Platform)))
		for i, c := range report.Checks {
			if i > 0 {
				cmd.OutOrStdout().Write([]byte(","))
			}
			msg := ""
			if c.Message != "" {
				msg = fmt.Sprintf(`,"message":"%s"`, c.Message)
			}
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf(`{"name":"%s","status":"%s"%s}`, c.Name, c.Status, msg)))
		}
		cmd.OutOrStdout().Write([]byte("]}\n"))
	case "yaml":
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("status: %s\nversion: %s\ngo_version: %s\nplatform: %s\nchecks:\n", report.Status, report.Version, report.Go, report.Platform)))
		for _, c := range report.Checks {
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  - name: %s\n    status: %s\n", c.Name, c.Status)))
			if c.Message != "" {
				cmd.OutOrStdout().Write([]byte(fmt.Sprintf("    message: %s\n", c.Message)))
			}
		}
	default:
		cmd.OutOrStdout().Write([]byte("NAEOS Health Report\n"))
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Status: %s | Version: %s | Go: %s | %s\n", report.Status, report.Version, report.Go, report.Platform)))
		cmd.OutOrStdout().Write([]byte(strings.Repeat("─", 45) + "\n"))
		for _, c := range report.Checks {
			icon := "✓"
			if c.Status == "degraded" {
				icon = "⚠"
			} else if c.Status == "unhealthy" {
				icon = "✗"
			}
			msg := ""
			if c.Message != "" {
				msg = fmt.Sprintf(" — %s", c.Message)
			}
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %s %s%s\n", icon, c.Name, msg)))
		}
	}
	return nil
}
