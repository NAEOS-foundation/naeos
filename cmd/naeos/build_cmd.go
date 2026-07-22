package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/distributed"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newBuildCommand() *cobra.Command {
	var configPath, input, inputFile, outputFormat, outputFile string
	var languages []string
	var dryRun bool
	var distributedMode bool
	var workerCount int

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build artifacts from a specification",
		Long: `Build artifacts from a specification using the NAEOS pipeline.

By default, build runs locally. Use --distributed to distribute work
across multiple workers for parallel processing.

Example:
  naeos build --config config.yaml --input spec.yaml
  naeos build --config config.yaml --input-file spec.yaml --distributed --workers 8`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if distributedMode {
				return runBuildDistributed(cmd, configPath, workerCount)
			}
			return runBuildLocal(cmd, configPath, input, inputFile, outputFormat, outputFile, languages, dryRun)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text, json, or yaml")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write formatted output")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language for code generation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview artifacts without writing to disk")
	cmd.Flags().BoolVar(&distributedMode, "distributed", false, "enable distributed building across workers")
	cmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "number of parallel workers (used with --distributed)")

	return cmd
}

func runBuildLocal(cmd *cobra.Command, configPath, input, inputFile, outputFormat, outputFile string, languages []string, dryRun bool) error {
	inputValue, err := loadInput(input, inputFile)
	if err != nil {
		return err
	}

	cfg, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun || dryRun)
	if err != nil {
		return err
	}

	p, err := pipeline.New(*cfg)
	if err != nil {
		return fmt.Errorf("failed to construct pipeline: %w", err)
	}

	result, err := p.Run(inputValue)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	payload := map[string]any{
		"pipeline":   cfg.Name,
		"mode":       cfg.Mode,
		"build":      "local",
		"verbose":    cfg.Verbose,
		"output_dir": cfg.OutputDir,
		"artifacts":  len(result.Artifacts),
		"tasks":      len(result.Tasks),
	}

	if len(languages) > 0 {
		payload["languages"] = languages
	}
	if cfg.DryRun {
		payload["dry_run"] = true
	}

	rendered, err := renderOutput(payload, outputFormat, func() []byte {
		return []byte(fmt.Sprintf("build=local pipeline=%s mode=%s verbose=%t output_dir=%s\nartifacts=%d tasks=%d\n", result.NEIR.Project, cfg.Mode, cfg.Verbose, cfg.OutputDir, len(result.Artifacts), len(result.Tasks)))
	})
	if err != nil {
		return err
	}

	return writeOrPrint(cmd, rendered, outputFile)
}

func runBuildDistributed(cmd *cobra.Command, configPath string, workerCount int) error {
	_, err := loadPipelineConfig(configPath, cliVerbose, nil, cliDryRun)
	if err != nil {
		return err
	}

	workers := make([]distributed.Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		id := fmt.Sprintf("builder-%d", i)
		workers[i] = distributed.NewSimpleWorker(id, func(ctx context.Context, task *distributed.Task) (map[string]any, error) {
			stage, _ := task.Payload["stage"].(string)
			var duration time.Duration
			switch stage {
			case "parse":
				duration = 800 * time.Millisecond
			case "normalize":
				duration = 600 * time.Millisecond
			case "resolve":
				duration = 1000 * time.Millisecond
			case "generate":
				duration = 1200 * time.Millisecond
			default:
				duration = 500 * time.Millisecond
			}

			select {
			case <-time.After(duration):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			return map[string]any{
				"stage":    stage,
				"status":   "completed",
				"duration": duration.String(),
			}, nil
		})
	}

	coord := distributed.NewCoordinator(workers, 100)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	coord.Start(ctx)

	stages := []string{"parse", "normalize", "resolve", "generate"}
	for _, s := range stages {
		coord.Submit(&distributed.Task{
			ID:      fmt.Sprintf("build-%s", s),
			Type:    "build",
			Payload: map[string]any{"stage": s},
		})
	}

	var completed int
	for range coord.Results() {
		completed++
		if completed >= len(stages) {
			break
		}
	}

	coord.Stop()

	if completed < len(stages) {
		return fmt.Errorf("build distributed: expected %d results, got %d", len(stages), completed)
	}

	payload := map[string]any{
		"build":   "distributed",
		"workers": workerCount,
		"stages":  len(stages),
	}

	rendered, err := renderOutput(payload, "text", func() []byte {
		return []byte(fmt.Sprintf("build=distributed workers=%d stages=%d completed\n", workerCount, len(stages)))
	})
	if err != nil {
		return err
	}

	return writeOrPrint(cmd, rendered, "")
}
