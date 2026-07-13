package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/watch"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newWatchCommand() *cobra.Command {
	var configPath, input, inputFile string
	var languages []string

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch for specification changes and re-run the pipeline",
		Long: `Watch for specification file changes and automatically re-run the pipeline.

Only .yaml, .yml, and .json files trigger a re-run.
The watcher watches the directory containing the input spec file.

Example:
  naeos watch --config config.yaml --input spec.yaml
  naeos watch --config config.yaml --input-file spec.yaml --language go`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveInputFile(input, inputFile)
			if err != nil {
				return err
			}

			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun)
			if err != nil {
				return err
			}

			specDir := filepath.Dir(resolved)
			if specDir == "" {
				specDir = "."
			}

			runPipeline := func() error {
				data, err := os.ReadFile(resolved)
				if err != nil {
					return fmt.Errorf("read spec: %w", err)
				}
				p, err := pipeline.New(*cfg)
				if err != nil {
					return fmt.Errorf("construct pipeline: %w", err)
				}
				result, err := p.Run(string(data))
				if err != nil {
					return fmt.Errorf("pipeline run: %w", err)
				}
				fmt.Fprintf(os.Stderr, "[naeos] pipeline complete: %d artifacts\n", len(result.Artifacts))
				return nil
			}

			if err := runPipeline(); err != nil {
				fmt.Fprintf(os.Stderr, "[naeos] initial run failed: %v\n", err)
			}

			watcher := watch.NewWatcher(500*time.Millisecond, func(path string) {
				ext := filepath.Ext(path)
				switch ext {
				case ".yaml", ".yml", ".json":
					fmt.Fprintf(os.Stderr, "[naeos] spec change detected: %s\n", path)
				}
			})

			if err := watcher.AddDirectory(specDir); err != nil {
				return fmt.Errorf("watch spec directory: %w", err)
			}

			fmt.Fprintf(os.Stderr, "[naeos] watching %s for changes...\n", specDir)
			return watcher.Run(runPipeline)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language for code generation")
	return cmd
}

func resolveInputFile(input, inputFile string) (string, error) {
	if inputFile != "" {
		if _, err := os.Stat(inputFile); err != nil {
			return "", fmt.Errorf("input file not found: %w", err)
		}
		return inputFile, nil
	}
	if input != "" {
		if _, err := os.Stat(input); err == nil {
			return input, nil
		}
		return "", nil
	}
	return "", fmt.Errorf("specify --input or --input-file")
}
