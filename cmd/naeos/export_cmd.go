package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newExportCommand() *cobra.Command {
	var configPath, input, outputDir string
	var languages []string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export generated artifacts to a directory",
		Long: `Export all generated artifacts to a directory.

Example:
  naeos export --config config.yaml --input spec.yaml
  naeos export --config config.yaml --input spec.yaml --output-dir ./generated
  naeos export --config config.yaml --input spec.yaml --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("missing required --input")
			}

			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun || dryRun)
			if err != nil {
				return err
			}
			if outputDir == "" {
				outputDir = cfg.OutputDir
			}
			if outputDir == "" {
				return fmt.Errorf("missing output directory")
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
				return fmt.Errorf("export run failed: %w", err)
			}

			if dryRun || cfg.DryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Export Preview (%d artifacts)\n", len(result.Artifacts))
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("=", 50))
				totalSize := 0
				for _, artifact := range result.Artifacts {
					fmt.Fprintf(cmd.OutOrStdout(), "  %-40s %6d bytes\n", artifact.Path, len(artifact.Content))
					totalSize += len(artifact.Content)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("=", 50))
				fmt.Fprintf(cmd.OutOrStdout(), "Total: %d artifacts, %d bytes\n", len(result.Artifacts), totalSize)
				return nil
			}

			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return fmt.Errorf("create export dir: %w", err)
			}
			for _, artifact := range result.Artifacts {
				artifactPath := filepath.Join(outputDir, artifact.Path)
				if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
					return fmt.Errorf("create artifact dir: %w", err)
				}
				if err := os.WriteFile(artifactPath, artifact.Content, 0o600); err != nil {
					return fmt.Errorf("write artifact %s: %w", artifact.Path, err)
				}
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "exported %d artifacts to %s\n", len(result.Artifacts), outputDir)
			if len(languages) > 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "languages: %s\n", strings.Join(languages, ", "))
			}
			return err
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input or file path to process")
	cmd.Flags().StringVar(&outputDir, "output-dir", "", "directory to write exported artifacts")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language for code generation (go, typescript, python, java, rust)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview artifacts without writing to disk")
	return cmd
}
