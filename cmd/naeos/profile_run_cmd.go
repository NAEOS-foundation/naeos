package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/profiling"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newProfileRunCommand() *cobra.Command {
	var configPath, input, inputFile, profileOut, memprofileOut string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run pipeline with performance and memory profiling",
		Long: `Run the NAEOS pipeline with profiling enabled.

Profiles execution stages (validate, build_graph, policy_eval, schedule,
generate, review, write_artifacts) and captures heap snapshots at each
stage boundary to detect memory leaks and performance bottlenecks.

Examples:
  naeos profile run --input "project: my-app\nmodules:\n  - name: core\n    path: ./core"
  naeos profile run --input-file spec.yaml --config config.yaml
  naeos profile run --input-file spec.yaml --profile profile.json --memprofile mem.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, err := loadPipelineConfig(configPath, false, nil, false, "")
			if err != nil {
				return err
			}

			pipeProf := profiling.NewProfile()
			memProf := profiling.NewMemProfiler()
			cfg.Profile = pipeProf
			cfg.MemProfile = memProf

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Running profiled pipeline...")
			result, err := p.Run(inputValue)
			if err != nil {
				return fmt.Errorf("pipeline failed: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Pipeline complete: %d artifacts, %d tasks\n", len(result.Artifacts), len(result.Tasks))

			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprint(cmd.OutOrStdout(), pipeProf.Summary())

			if profileOut != "" {
				if err := profiling.SaveProfile(profileOut, pipeProf); err != nil {
					return fmt.Errorf("failed to save profile: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\nProfile saved to %s\n", profileOut)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprint(cmd.OutOrStdout(), memProf.Summary())

			report := memProf.Analyze()
			if report.Suspected {
				fmt.Fprintf(cmd.OutOrStdout(), "\n⚠ Memory leak suspected: %s\n", report.Details)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "\n✓ No memory leak detected")
			}

			if memprofileOut != "" {
				if err := saveMemProfileJSON(memprofileOut, memProf); err != nil {
					return fmt.Errorf("failed to save memprofile: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Memory profile saved to %s\n", memprofileOut)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to pipeline config file")
	cmd.Flags().StringVar(&input, "input", "", "specification input")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to specification file")
	cmd.Flags().StringVar(&profileOut, "profile", "", "write pipeline profile to JSON file")
	cmd.Flags().StringVar(&memprofileOut, "memprofile", "", "write memory profile to JSON file")

	return cmd
}

func saveMemProfileJSON(path string, mp *profiling.MemProfiler) error {
	data, err := json.MarshalIndent(map[string]any{
		"snapshots": mp.Snapshots(),
		"diffs":     mp.Diffs(),
		"analysis":  mp.Analyze(),
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
