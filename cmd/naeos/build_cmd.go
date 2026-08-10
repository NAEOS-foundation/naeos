package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/internal/distributed"
	"github.com/NAEOS-foundation/naeos/internal/profiling"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newBuildCommand() *cobra.Command {
	var configPath, input, inputFile, outputFormat, outputFile, profileFile, memProfileFile, schemaSource string
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
				return runBuildDistributed(cmd, configPath, input, inputFile, workerCount)
			}
			return runBuildLocal(cmd, configPath, input, inputFile, outputFormat, outputFile, languages, dryRun, profileFile != "", profileFile, memProfileFile, schemaSource)
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
	cmd.Flags().StringVar(&profileFile, "profile", "", "write pipeline performance profile to file")
	cmd.Flags().StringVar(&memProfileFile, "memprofile", "", "write memory profile (heap snapshots) to file")
	cmd.Flags().StringVar(&schemaSource, "schema-source", "", "URL or file path to JSON Schema for spec validation")

	return cmd
}

func runBuildLocal(cmd *cobra.Command, configPath, input, inputFile, outputFormat, outputFile string, languages []string, dryRun, enableProfile bool, profileFile, memProfileFile, schemaSource string) error {
	inputValue, err := loadInput(input, inputFile)
	if err != nil {
		return err
	}

	cfg, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun || dryRun, "")
	if err != nil {
		return err
	}

	if enableProfile {
		cfg.Profile = profiling.NewProfile()
	}
	if memProfileFile != "" {
		cfg.MemProfile = profiling.NewMemProfiler()
	}
	if schemaSource != "" {
		cfg.SchemaSource = schemaSource
	}

	p, err := pipeline.New(*cfg)
	if err != nil {
		return fmt.Errorf("failed to construct pipeline: %w", err)
	}

	result, err := p.Run(inputValue)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if enableProfile {
		prof := p.ProfileResult()
		if prof != nil {
			summary := prof.Summary()
			fmt.Fprint(cmd.OutOrStdout(), summary)
			if profileFile != "" {
				if err := profiling.SaveProfile(profileFile, prof); err != nil {
					return fmt.Errorf("failed to save profile: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Profile saved to %s\n", profileFile)
			}
		}
	}

	if memProfileFile != "" {
		mp := p.MemProfileResult()
		if mp != nil {
			summary := mp.Summary()
			fmt.Fprint(cmd.OutOrStdout(), summary)
			report := mp.Analyze()
			if report.Suspected {
				fmt.Fprintf(cmd.OutOrStdout(), "⚠ Memory leak suspected: %s\n", report.Details)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "✓ No memory leak detected")
			}
			if err := saveMemProfile(memProfileFile, mp); err != nil {
				return fmt.Errorf("failed to save memprofile: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Memory profile saved to %s\n", memProfileFile)
		}
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

func saveMemProfile(path string, mp *profiling.MemProfiler) error {
	snaps := mp.Snapshots()
	data, err := json.MarshalIndent(map[string]any{
		"snapshots": snaps,
		"summary":   mp.Summary(),
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func runBuildDistributed(cmd *cobra.Command, configPath, input, inputFile string, workerCount int) error {
	inputValue, err := loadInput(input, inputFile)
	if err != nil {
		return err
	}

	workers := make([]distributed.Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		id := fmt.Sprintf("builder-%d", i)
		workers[i] = distributed.NewSimpleWorker(id, func(ctx context.Context, task *distributed.Task) (map[string]any, error) {
			taskCfgPath, _ := task.Payload["config"].(string)
			taskInput, _ := task.Payload["input"].(string)
			taskModule, _ := task.Payload["module"].(string)

			pCfg, err := loadPipelineConfig(taskCfgPath, cliVerbose, nil, cliDryRun, "")
			if err != nil {
				return nil, fmt.Errorf("worker %s: load config: %w", id, err)
			}

			var specInput string
			if taskModule != "" {
				specInput = fmt.Sprintf("project: distributed-%s\nmodules:\n  - name: %s\n    path: ./%s\n", taskModule, taskModule, taskModule)
			} else {
				specInput = taskInput
			}

			p, err := pipeline.New(*pCfg)
			if err != nil {
				return nil, fmt.Errorf("worker %s: create pipeline: %w", id, err)
			}

			start := time.Now()
			result, err := p.RunContext(ctx, specInput)
			elapsed := time.Since(start)
			if err != nil {
				return nil, fmt.Errorf("worker %s: run: %w", id, err)
			}

			return map[string]any{
				"worker":    id,
				"status":    "completed",
				"module":    taskModule,
				"artifacts": len(result.Artifacts),
				"tasks":     len(result.Tasks),
				"duration":  elapsed.String(),
				"project":   result.NEIR.Project.Name,
			}, nil
		})
	}

	start := time.Now()

	coord := distributed.NewCoordinator(workers, 100)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	coord.Start(ctx)

	parsedSpec, err := parseSpecInput(inputValue)
	if err != nil {
		return fmt.Errorf("parse spec for distribution: %w", err)
	}

	var modules []string
	if specMap, ok := parsedSpec["modules"].([]any); ok && len(specMap) > 0 {
		for _, m := range specMap {
			if mod, ok := m.(map[string]any); ok {
				if name, ok := mod["name"].(string); ok {
					modules = append(modules, name)
				}
			}
		}
	}
	if len(modules) == 0 {
		modules = []string{""}
	}

	for _, mod := range modules {
		coord.Submit(&distributed.Task{
			ID:   fmt.Sprintf("build-%s", orDefault(mod, "default")),
			Type: "build",
			Payload: map[string]any{
				"config": configPath,
				"input":  inputValue,
				"module": mod,
			},
		})
	}

	var results []map[string]any
	go coord.Stop()
	for res := range coord.Results() {
		if res.Succeeded && res.Output != nil {
			results = append(results, res.Output)
		}
	}

	totalArtifacts := 0
	totalTasks := 0
	for _, r := range results {
		if arts, ok := r["artifacts"].(int); ok {
			totalArtifacts += arts
		}
		if tasks, ok := r["tasks"].(int); ok {
			totalTasks += tasks
		}
	}

	elapsed := time.Since(start)

	payload := map[string]any{
		"build":     "distributed",
		"workers":   workerCount,
		"modules":   len(modules),
		"artifacts": totalArtifacts,
		"tasks":     totalTasks,
		"duration":  elapsed.Round(time.Millisecond).String(),
	}

	rendered, err := renderOutput(payload, "text", func() []byte {
		return []byte(fmt.Sprintf("build=distributed workers=%d modules=%d completed artifacts=%d tasks=%d duration=%s\n",
			workerCount, len(modules), totalArtifacts, totalTasks, elapsed.Round(time.Millisecond)))
	})
	if err != nil {
		return err
	}

	return writeOrPrint(cmd, rendered, "")
}

func parseSpecInput(input string) (map[string]any, error) {
	var spec map[string]any
	if err := json.Unmarshal([]byte(input), &spec); err == nil {
		return spec, nil
	}
	if err := yaml.Unmarshal([]byte(input), &spec); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	return spec, nil
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
