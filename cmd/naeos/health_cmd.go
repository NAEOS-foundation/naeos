// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/version"
)

func newHealthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Run system health checks and diagnostics",
		Long: `Perform comprehensive health checks on the NAEOS installation,
configuration, and dependencies.

Example:
  naeos health
  naeos health -o json
  naeos health -o yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := runHealthChecks()
			return renderHealthReport(cmd, report, cliOutputFormat)
		},
	}

	return cmd
}

type HealthCheck struct {
	Name    string `json:"name" yaml:"name"`
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
}

type HealthReport struct {
	Status   string        `json:"status" yaml:"status"`
	Version  string        `json:"version" yaml:"version"`
	Go       string        `json:"go_version" yaml:"go_version"`
	Platform string        `json:"platform" yaml:"platform"`
	Checks   []HealthCheck `json:"checks" yaml:"checks"`
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
		return FormatOutput(cmd.OutOrStdout(), report, "json")
	case "yaml":
		return FormatOutput(cmd.OutOrStdout(), report, "yaml")
	default:
		_, _ = cmd.OutOrStdout().Write([]byte("NAEOS Health Report\n"))
		fmt.Fprintf(cmd.OutOrStdout(), "Status: %s | Version: %s | Go: %s | %s\n", report.Status, report.Version, report.Go, report.Platform)
		_, _ = cmd.OutOrStdout().Write([]byte(strings.Repeat("─", 45) + "\n"))
		for _, c := range report.Checks {
			icon := "✓"
			switch c.Status {
			case "degraded":
				icon = "⚠"
			case "unhealthy":
				icon = "✗"
			}
			msg := ""
			if c.Message != "" {
				msg = fmt.Sprintf(" — %s", c.Message)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s %s%s\n", icon, c.Name, msg)
		}
	}
	return nil
}
