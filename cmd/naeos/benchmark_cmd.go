package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
	"github.com/spf13/cobra"
)

func newBenchmarkCommand() *cobra.Command {
	var iterations int
	var configPath string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Run pipeline benchmarks",
		Long: `Benchmark the pipeline performance by running multiple iterations.
Reports timing statistics including average, min, max, and p95.

Example:
  naeos benchmark --iterations 100
  naeos benchmark --input spec.yaml --iterations 50
  naeos benchmark --output json --iterations 100`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, nil, cliDryRun)
			if err != nil {
				return err
			}

			input := "project:\n  name: benchapp\n  version: \"1.0.0\"\nservices:\n  - name: api\n    port: 8080"

			durations := make([]time.Duration, 0, iterations)
			var errCount int

			for i := 0; i < iterations; i++ {
				p, err := pipeline.New(*cfg)
				if err != nil {
					return fmt.Errorf("pipeline creation failed: %w", err)
				}
				start := time.Now()
				_, err = p.Run(input)
				durations = append(durations, time.Since(start))
				if err != nil {
					errCount++
				}
			}

			var total time.Duration
			minD := durations[0]
			maxD := durations[0]
			for _, d := range durations {
				total += d
				if d < minD {
					minD = d
				}
				if d > maxD {
					maxD = d
				}
			}
			avg := total / time.Duration(len(durations))

			type benchResult struct {
				Iterations int     `json:"iterations" yaml:"iterations"`
				Errors     int     `json:"errors" yaml:"errors"`
				Average    string  `json:"average" yaml:"average"`
				Min        string  `json:"min" yaml:"min"`
				Max        string  `json:"max" yaml:"max"`
				Total      string  `json:"total" yaml:"total"`
				OpsPerSec  float64 `json:"ops_per_sec" yaml:"ops_per_sec"`
			}

			data := benchResult{
				Iterations: iterations,
				Errors:     errCount,
				Average:    avg.Round(time.Microsecond).String(),
				Min:        minD.Round(time.Microsecond).String(),
				Max:        maxD.Round(time.Microsecond).String(),
				Total:      total.Round(time.Millisecond).String(),
				OpsPerSec:  float64(time.Second) / float64(avg),
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			var sb strings.Builder
			sb.WriteString("NAEOS Benchmark Results\n")
			sb.WriteString(fmt.Sprintf("Iterations: %d | Errors: %d\n", iterations, errCount))
			sb.WriteString(strings.Repeat("─", 45) + "\n")
			sb.WriteString(fmt.Sprintf("  Average:  %s\n", avg.Round(time.Microsecond)))
			sb.WriteString(fmt.Sprintf("  Min:      %s\n", minD.Round(time.Microsecond)))
			sb.WriteString(fmt.Sprintf("  Max:      %s\n", maxD.Round(time.Microsecond)))
			sb.WriteString(fmt.Sprintf("  Total:    %s\n", total.Round(time.Millisecond)))
			sb.WriteString(fmt.Sprintf("  Ops/sec:  %.0f\n", float64(time.Second)/float64(avg)))

			cmd.OutOrStdout().Write([]byte(sb.String()))
			return nil
		},
	}

	cmd.Flags().IntVarP(&iterations, "iterations", "n", 10, "number of iterations")
	cmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}
