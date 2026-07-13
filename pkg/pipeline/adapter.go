package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/eventsourcing"
	pm "github.com/NAEOS-foundation/naeos/internal/pipelinemiddleware"
)

type PipelineAdapter struct {
	pipeline     *Pipeline
	middleware   *pm.Chain
	eventStore   eventsourcing.EventStore
	runID        string
	telemetryFn  func(stage string, duration time.Duration, err error)
}

func NewAdapter(p *Pipeline) *PipelineAdapter {
	return &PipelineAdapter{
		pipeline:   p,
		middleware: pm.NewChain(),
		eventStore: eventsourcing.NewInMemoryStore(),
	}
}

func (a *PipelineAdapter) UseMiddleware(stage string, mw pm.Middleware) {
	a.middleware.Use(stage, mw)
}

func (a *PipelineAdapter) OnTelemetryRecord(fn func(stage string, duration time.Duration, err error)) {
	a.telemetryFn = fn
}

func (a *PipelineAdapter) RunWithMiddleware(ctx context.Context, input string) (*Result, error) {
	a.runID = fmt.Sprintf("run-%d", time.Now().UnixNano())

	snap := eventsourcing.NewPipelineRun(a.runID, a.pipelineName())
	snap.Started()

	if err := a.recordEvent("pipeline.started", map[string]any{"name": a.pipelineName()}); err != nil {
		return nil, err
	}

	start := time.Now()

	out, err := a.middleware.Execute("pre-process", &pm.StageInput{
		Stage:  "pre-process",
		Data:   []byte(input),
		Labels: map[string]string{"run_id": a.runID},
	}, func(ctx context.Context, in *pm.StageInput) (*pm.StageOutput, error) {
		return &pm.StageOutput{Data: in.Data, Labels: in.Labels}, nil
	})
	if err != nil {
		snap.Failed(err)
		return nil, fmt.Errorf("middleware pre-process: %w", err)
	}

	result, err := a.pipeline.RunContext(ctx, string(out.Data))

	if a.telemetryFn != nil {
		a.telemetryFn("full_pipeline", time.Since(start), err)
	}

	if err != nil {
		snap.Failed(err)
		_ = a.recordEvent("pipeline.failed", map[string]any{"error": err.Error()})
		return nil, err
	}

	artifactCount := len(result.Artifacts)
	snap.Completed(artifactCount)
	_ = a.recordEvent("pipeline.completed", map[string]any{
		"artifacts": artifactCount,
		"tasks":     len(result.Tasks),
		"reviews":   len(result.Reviews),
		"duration":  time.Since(start).String(),
	})

	return result, nil
}

func (a *PipelineAdapter) RunSnapshot() *eventsourcing.PipelineRunSnapshot {
	events, _ := a.eventStore.Load(a.runID)
	if events == nil {
		return nil
	}
	return eventsourcing.RebuildFromEvents(a.runID, events)
}

func (a *PipelineAdapter) EventCount() int {
	store := a.eventStore.(*eventsourcing.InMemoryStore)
	return store.EventCount(a.runID)
}

func (a *PipelineAdapter) RunID() string {
	return a.runID
}

func (a *PipelineAdapter) pipelineName() string {
	if a.pipeline != nil {
		return a.pipeline.Name()
	}
	return "unknown"
}

func (a *PipelineAdapter) recordEvent(eventType string, data map[string]any) error {
	return a.eventStore.Append(a.runID, []eventsourcing.Event{
		{Type: eventType, Data: data},
	})
}
