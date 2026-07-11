package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newPreviewCommand() *cobra.Command {
	var configPath, input string

	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview generated artifacts without writing them",
		Long: `Preview what artifacts would be generated without writing to disk.

Example:
  naeos preview --config config.yaml --input spec.yaml
  naeos preview --config config.yaml --input-file spec.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("missing required --input")
			}

			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, nil, true)
			if err != nil {
				return err
			}

			specInput, err := resolveInput(input)
			if err != nil {
				return fmt.Errorf("resolve input: %w", err)
			}

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			result, err := p.Run(specInput)
			if err != nil {
				return fmt.Errorf("preview run failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Preview for %s\n", cfg.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("-", 50))
			totalSize := 0
			for _, artifact := range result.Artifacts {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-40s %6d bytes\n", artifact.Path, len(artifact.Content))
				totalSize += len(artifact.Content)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("-", 50))
			fmt.Fprintf(cmd.OutOrStdout(), "Total: %d artifacts, %d bytes\n", len(result.Artifacts), totalSize)
			if len(result.Reviews) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Reviews: %d issues found\n", len(result.Reviews))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input or file path to process")
	return cmd
}
