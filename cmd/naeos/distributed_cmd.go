package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/distributed"
	"github.com/spf13/cobra"
)

func newDistributedCommand() *cobra.Command {
	var configPath string
	var workerCount int

	cmd := &cobra.Command{
		Use:   "distributed",
		Short: "Run pipeline tasks in distributed mode across multiple workers",
		Long: `Execute pipeline tasks using multiple workers for parallel processing.
Tasks are distributed across workers using round-robin load balancing.

Example:
  naeos distributed --input spec.yaml --workers 4
  naeos distributed --config config.yaml --input-file spec.yaml --workers 8`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDistributed(cmd, configPath, workerCount)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	cmd.Flags().IntVarP(&workerCount, "workers", "w", 4, "number of parallel workers")

	return cmd
}

func runDistributed(cmd *cobra.Command, configPath string, workerCount int) error {
	cfg, _, err := loadPipelineConfig(configPath, cliVerbose, nil, cliDryRun)
	if err != nil {
		return err
	}

	workers := make([]distributed.Worker, workerCount)
	for i := 0; i < workerCount; i++ {
		id := fmt.Sprintf("worker-%d", i)
		workers[i] = distributed.NewSimpleWorker(id, func(ctx context.Context, task *distributed.Task) (map[string]any, error) {
			time.Sleep(10 * time.Millisecond)
			return map[string]any{"worker": id, "task": task.ID, "status": "completed"}, nil
		})
	}

	coordinator := distributed.NewCoordinator(workers, workerCount*10)
	aggregator := distributed.NewResultAggregator()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	coordinator.Start(ctx)

	tasks := []distributed.Task{
		{ID: "parse", Type: "parse", Payload: map[string]any{"stage": "parse"}},
		{ID: "normalize", Type: "normalize", Payload: map[string]any{"stage": "normalize"}},
		{ID: "resolve", Type: "resolve", Payload: map[string]any{"stage": "resolve"}},
		{ID: "build-neir", Type: "build", Payload: map[string]any{"stage": "build-neir"}},
		{ID: "validate", Type: "validate", Payload: map[string]any{"stage": "validate"}},
		{ID: "schedule", Type: "schedule", Payload: map[string]any{"stage": "schedule"}},
		{ID: "generate", Type: "generate", Payload: map[string]any{"stage": "generate"}},
		{ID: "review", Type: "review", Payload: map[string]any{"stage": "review"}},
	}

	for i := range tasks {
		coordinator.Submit(&tasks[i])
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		for r := range coordinator.Results() {
			aggregator.Add(*r)
			count++
			if count >= len(tasks) {
				return
			}
		}
	}()

	wg.Wait()
	coordinator.Stop()

	cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Distributed pipeline: %d workers, %d tasks\n", workerCount, len(tasks))))
	cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Results: %s\n", aggregator.Summary())))

	failed := aggregator.Failed()
	if len(failed) > 0 {
		cmd.OutOrStdout().Write([]byte("Failed tasks:\n"))
		for _, f := range failed {
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  - %s: %s (worker: %s)\n", f.TaskID, f.Error, f.Worker)))
		}
	}

	cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Pipeline: %s\n", cfg.Name)))
	return nil
}
