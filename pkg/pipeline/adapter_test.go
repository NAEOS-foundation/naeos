package pipeline

import (
	"context"
	"testing"
	"time"

	pm "github.com/NAEOS-foundation/naeos/internal/pipelinemiddleware"
)

func TestAdapterNew(t *testing.T) {
	cfg := Config{Name: "test-pipeline", OutputDir: t.TempDir()}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	adapter := NewAdapter(p)
	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}
	if adapter.RunID() != "" {
		t.Errorf("expected empty RunID before run, got %q", adapter.RunID())
	}
}

func TestAdapterRunWithMiddleware(t *testing.T) {
	cfg := Config{Name: "test-pipeline", OutputDir: t.TempDir()}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	adapter := NewAdapter(p)

	var middlewareCalled bool
	adapter.UseMiddleware("pre-process", &testMW{
		name: "test-mw",
		fn: func(next func(ctx context.Context, in *pm.StageInput) (*pm.StageOutput, error)) pm.StageFunc {
			return func(ctx context.Context, in *pm.StageInput) (*pm.StageOutput, error) {
				middlewareCalled = true
				return next(ctx, in)
			}
		},
	})

	result, err := adapter.RunWithMiddleware(context.Background(), "project: test\nversion: 1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !middlewareCalled {
		t.Error("middleware was not called")
	}
	if adapter.RunID() == "" {
		t.Error("expected non-empty RunID after run")
	}
}

func TestAdapterEventSourcing(t *testing.T) {
	cfg := Config{Name: "test-pipeline", OutputDir: t.TempDir()}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	adapter := NewAdapter(p)

	_, err = adapter.RunWithMiddleware(context.Background(), "project: test\nversion: 1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	snap := adapter.RunSnapshot()
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snap.Name != "test-pipeline" {
		t.Errorf("expected name 'test-pipeline', got %q", snap.Name)
	}
	if snap.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", snap.Status)
	}
	if adapter.EventCount() < 2 {
		t.Errorf("expected >= 2 events, got %d", adapter.EventCount())
	}
}

func TestAdapterTelemetry(t *testing.T) {
	cfg := Config{Name: "test-pipeline", OutputDir: t.TempDir()}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	adapter := NewAdapter(p)

	var recordedStage string
	adapter.OnTelemetryRecord(func(stage string, duration time.Duration, err error) {
		recordedStage = stage
	})

	_, err = adapter.RunWithMiddleware(context.Background(), "project: test\nversion: 1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	if recordedStage != "full_pipeline" {
		t.Errorf("expected stage 'full_pipeline', got %q", recordedStage)
	}
}

func TestAdapterPipelineName(t *testing.T) {
	cfg := Config{Name: "my-pipeline", OutputDir: t.TempDir()}
	p, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "my-pipeline" {
		t.Errorf("expected 'my-pipeline', got %q", p.Name())
	}
}

type testMW struct {
	name string
	fn   func(next func(ctx context.Context, in *pm.StageInput) (*pm.StageOutput, error)) pm.StageFunc
}

func (m *testMW) Name() string { return m.name }

func (m *testMW) Wrap(stage string, next pm.StageFunc) pm.StageFunc {
	return m.fn(next)
}
