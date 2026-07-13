package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/generation/adapters/container"
	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	"github.com/NAEOS-foundation/naeos/internal/neir/model/project"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newExportComposeCommand() *cobra.Command {
	var configPath, input, outputDir string
	var languages []string

	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Generate docker-compose.yaml and Dockerfile from spec",
		Long: `Generate Docker Compose configuration and Dockerfile from a specification.

Reads the spec, runs the pipeline to produce NEIR, then generates
a docker-compose.yaml and Dockerfile in the output directory.

Example:
  naeos export compose --input spec.yaml
  naeos export compose --input spec.yaml --output-dir ./docker`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("missing required --input")
			}

			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun)
			if err != nil {
				return err
			}
			if outputDir == "" {
				outputDir = "."
			}

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("construct pipeline: %w", err)
			}

			result, err := p.Run(input)
			if err != nil {
				return fmt.Errorf("pipeline run: %w", err)
			}

			neir := resultToNEIR(result)

			gen := container.NewGenerator()
			artifacts := gen.Generate(neir)

			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return fmt.Errorf("create output dir: %w", err)
			}

			for _, a := range artifacts {
				outPath := filepath.Join(outputDir, a.Path)
				if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(outPath, a.Content, 0o644); err != nil {
					return fmt.Errorf("write %s: %w", a.Path, err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "generated: %s\n", outPath)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%d files generated in %s\n", len(artifacts), outputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	cmd.Flags().StringVar(&input, "input", "", "specification input or file path")
	cmd.Flags().StringVar(&outputDir, "output-dir", ".", "output directory")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language")
	return cmd
}

func resultToNEIR(result *pipeline.Result) *model.NEIR {
	if result.NEIR != nil {
		return result.NEIR
	}
	return &model.NEIR{
		Project: &project.Project{Name: "app"},
	}
}
