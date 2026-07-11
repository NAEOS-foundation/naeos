package main

import (
	"fmt"
	"os"
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

Example:
  naeos watch --config config.yaml --input spec.yaml
  naeos watch --config config.yaml --input-file spec.yaml --language go`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, _, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun)
			if err != nil {
				return err
			}

			runPipeline := func() error {
				p, err := pipeline.New(*cfg)
				if err != nil {
					return fmt.Errorf("failed to construct pipeline: %w", err)
				}
				result, err := p.Run(inputValue)
				if err != nil {
					return fmt.Errorf("pipeline run failed: %w", err)
				}
				fmt.Fprintf(os.Stderr, "[naeos] pipeline complete: %d artifacts\n", len(result.Artifacts))
				return nil
			}

			watcher := watch.NewWatcher(500*time.Millisecond, func(path string) {
				fmt.Fprintf(os.Stderr, "[naeos] file changed: %s\n", path)
			})

			watcher.AddDirectory(".")
			return watcher.Run(runPipeline)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language for code generation")
	return cmd
}
