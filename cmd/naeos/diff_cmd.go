package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/diff"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newDiffCommand() *cobra.Command {
	var configPath, input, inputFile, outputDir, format string
	var visual bool

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare generated artifacts with existing output directory",
		Long: `Compare current pipeline output with existing files in the output directory.

Example:
  naeos diff --config config.yaml --input spec.yaml
  naeos diff --config config.yaml --input spec.yaml --output-dir ./out
  naeos diff --config config.yaml --input spec.yaml --format unified
  naeos diff --config config.yaml --input spec.yaml --visual`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" && inputFile == "" {
				return fmt.Errorf("missing required --input or --input-file")
			}

			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, err := loadPipelineConfig(configPath, cliVerbose, nil, cliDryRun)
			if err != nil {
				return err
			}
			if outputDir == "" {
				outputDir = cfg.OutputDir
			}

			specInput, err := resolveInput(inputValue)
			if err != nil {
				return fmt.Errorf("resolve input: %w", err)
			}

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			result, err := p.Run(specInput)
			if err != nil {
				return fmt.Errorf("diff run failed: %w", err)
			}

			var diffs []*diff.FileDiff
			for _, artifact := range result.Artifacts {
				oldPath := filepath.Join(outputDir, artifact.Path)
				oldContent := ""
				if data, err := os.ReadFile(oldPath); err == nil {
					oldContent = string(data)
				}
				newContent := string(artifact.Content)
				diffs = append(diffs, diff.ComputeDiff(oldContent, newContent, artifact.Path))
			}

			for _, d := range diffs {
				if visual {
					fmt.Fprint(cmd.OutOrStdout(), diff.FormatSideBySide(d, 0))
				} else {
					fmt.Fprint(cmd.OutOrStdout(), diff.FormatDiff(d))
				}
			}

			added, removed, modified, unchanged := diff.Summary(diffs)
			fmt.Fprintf(cmd.OutOrStdout(), "\nSummary: %d added, %d removed, %d modified, %d unchanged\n",
				added, removed, modified, unchanged)
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "existing output directory to compare against")
	cmd.Flags().StringVar(&format, "format", "unified", "diff format: unified")
	cmd.Flags().BoolVar(&visual, "visual", false, "show side-by-side tree view")
	return cmd
}
