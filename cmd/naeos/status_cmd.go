package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/pipelinecache"
	"github.com/NAEOS-foundation/naeos/internal/version"
	cfgpkg "github.com/NAEOS-foundation/naeos/pkg/config"
)

func newStatusCommand() *cobra.Command {
	var configPath string
	var metricsPort string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show current pipeline, system and project status",
		Long: `Display the current status of the NAEOS project and pipeline configuration.

Example:
  naeos status
  naeos status --config config.yaml
  naeos status -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			fileCfg, err := cfgpkg.LoadFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			startTime := time.Now()

			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			cacheDir := fileCfg.Pipeline.OutputDir
			if cacheDir == "" {
				cacheDir = ".naeos/cache"
			}
			cache := pipelinecache.New(cacheDir, 100)
			cacheStats := cache.Stats()

			type statusOutput struct {
				Config     string                   `json:"config" yaml:"config"`
				Pipeline   string                   `json:"pipeline" yaml:"pipeline"`
				Mode       string                   `json:"mode" yaml:"mode"`
				OutputDir  string                   `json:"output_dir" yaml:"output_dir"`
				Languages  []string                 `json:"languages" yaml:"languages"`
				Verbose    bool                     `json:"verbose" yaml:"verbose"`
				Version    string                   `json:"version" yaml:"version"`
				GoVersion  string                   `json:"go_version" yaml:"go_version"`
				Platform   string                   `json:"platform" yaml:"platform"`
				StartTime  string                   `json:"start_time" yaml:"start_time"`
				Uptime     string                   `json:"uptime" yaml:"uptime"`
				Goroutines int                      `json:"goroutines" yaml:"goroutines"`
				AllocMB    float64                  `json:"alloc_mb" yaml:"alloc_mb"`
				Cache      pipelinecache.CacheStats `json:"cache" yaml:"cache"`
				CheckedAt  string                   `json:"checked_at" yaml:"checked_at"`
			}

			status := statusOutput{
				Config:     resolved,
				Pipeline:   fileCfg.Pipeline.Name,
				Mode:       fileCfg.Pipeline.Mode,
				OutputDir:  fileCfg.Pipeline.OutputDir,
				Languages:  fileCfg.Pipeline.Language,
				Verbose:    fileCfg.Pipeline.Verbose,
				Version:    version.String(),
				GoVersion:  runtime.Version(),
				Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				StartTime:  startTime.Format(time.RFC3339),
				Uptime:     time.Since(startTime).Round(time.Second).String(),
				Goroutines: runtime.NumGoroutine(),
				AllocMB:    float64(m.Alloc) / 1024 / 1024,
				Cache:      cacheStats,
				CheckedAt:  time.Now().Format(time.RFC3339),
			}

			_ = metricsPort

			switch cliOutputFormat {
			case "json", "yaml":
				return FormatOutput(cmd.OutOrStdout(), status, cliOutputFormat)
			default:
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "NAEOS Status\n")
				fmt.Fprintf(out, "%s\n", "================================")
				fmt.Fprintf(out, "Version:       %s\n", status.Version)
				fmt.Fprintf(out, "Go:            %s\n", status.GoVersion)
				fmt.Fprintf(out, "Platform:      %s\n", status.Platform)
				fmt.Fprintf(out, "Uptime:        %s\n", status.Uptime)
				fmt.Fprintf(out, "Goroutines:    %d\n", status.Goroutines)
				fmt.Fprintf(out, "Alloc:         %.1f MB\n", status.AllocMB)
				fmt.Fprintf(out, "Config:        %s\n", status.Config)
				fmt.Fprintf(out, "Pipeline:      %s\n", status.Pipeline)
				fmt.Fprintf(out, "Mode:          %s\n", status.Mode)
				fmt.Fprintf(out, "Output Dir:    %s\n", status.OutputDir)
				fmt.Fprintf(out, "Languages:     %s\n", joinStrings(status.Languages))
				fmt.Fprintf(out, "Verbose:       %t\n", status.Verbose)
				fmt.Fprintf(out, "Cache:         %d/%d entries\n", status.Cache.Size, status.Cache.MaxSize)
				fmt.Fprintf(out, "Checked At:    %s\n", status.CheckedAt)
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&metricsPort, "metrics-port", "", "prometheus metrics endpoint (e.g. :9090)")
	return cmd
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	return strings.Join(ss, ", ")
}
