package distributed

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoordinatorWorkers(t *testing.T) {
	workers := []Worker{
		NewSimpleWorker("w1", func(ctx context.Context, task *Task) (map[string]any, error) {
			return map[string]any{"result": "ok"}, nil
		}),
		NewSimpleWorker("w2", func(ctx context.Context, task *Task) (map[string]any, error) {
			return map[string]any{"result": "ok"}, nil
		}),
	}

	c := NewCoordinator(workers, 10)
	if c.WorkerCount() != 2 {
		t.Errorf("expected 2 workers, got %d", c.WorkerCount())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.Start(ctx)

	c.Submit(&Task{ID: "t1", Type: "test"})
	c.Submit(&Task{ID: "t2", Type: "test"})

	var results []TaskResult
	for r := range c.Results() {
		results = append(results, *r)
		if len(results) == 2 {
			break
		}
	}

	c.Stop()

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestCoordinatorErrorHandling(t *testing.T) {
	worker := NewSimpleWorker("w1", func(ctx context.Context, task *Task) (map[string]any, error) {
		return nil, fmt.Errorf("task failed")
	})

	c := NewCoordinator([]Worker{worker}, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c.Start(ctx)
	c.Submit(&Task{ID: "t1", Type: "test"})

	r := <-c.Results()
	c.Stop()

	if r.Error != "task failed" {
		t.Errorf("expected error 'task failed', got %q", r.Error)
	}
	if r.Worker != "w1" {
		t.Errorf("expected worker 'w1', got %q", r.Worker)
	}
}

func TestLoadBalancer(t *testing.T) {
	var counts [3]int64
	var mu sync.Mutex

	workers := make([]Worker, 3)
	for i := range workers {
		idx := i
		workers[i] = NewSimpleWorker(fmt.Sprintf("w%d", i), func(ctx context.Context, task *Task) (map[string]any, error) {
			mu.Lock()
			counts[idx]++
			mu.Unlock()
			return nil, nil
		})
	}

	lb := NewLoadBalancer(workers)
	if lb.WorkerCount() != 3 {
		t.Errorf("expected 3 workers, got %d", lb.WorkerCount())
	}

	for i := 0; i < 6; i++ {
		w := lb.Next()
		if w == nil {
			t.Fatal("expected non-nil worker")
		}
	}

	if lb.Next() == nil {
		t.Error("expected non-nil worker")
	}
}

func TestResultAggregator(t *testing.T) {
	agg := NewResultAggregator()

	agg.Add(TaskResult{TaskID: "t1", Output: map[string]any{"ok": true}})
	agg.Add(TaskResult{TaskID: "t2", Error: "failed"})
	agg.Add(TaskResult{TaskID: "t3", Output: map[string]any{"ok": true}})

	if agg.Count() != 3 {
		t.Errorf("expected 3 results, got %d", agg.Count())
	}

	failed := agg.Failed()
	if len(failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(failed))
	}

	all := agg.All()
	if len(all) != 3 {
		t.Errorf("expected 3 all, got %d", len(all))
	}

	summary := agg.Summary()
	if summary != "3 total, 2 succeeded, 1 failed" {
		t.Errorf("unexpected summary: %s", summary)
	}
}

func TestSimpleWorker(t *testing.T) {
	w := NewSimpleWorker("test", func(ctx context.Context, task *Task) (map[string]any, error) {
		return map[string]any{"echo": task.ID}, nil
	})

	if w.ID() != "test" {
		t.Errorf("expected ID 'test', got %q", w.ID())
	}

	result, err := w.Execute(context.Background(), &Task{ID: "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.TaskID != "t1" {
		t.Errorf("expected task ID 't1', got %q", result.TaskID)
	}
}

func TestCoordinatorConcurrency(t *testing.T) {
	var counter int64
	workers := []Worker{
		NewSimpleWorker("w1", func(ctx context.Context, task *Task) (map[string]any, error) {
			atomic.AddInt64(&counter, 1)
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		}),
		NewSimpleWorker("w2", func(ctx context.Context, task *Task) (map[string]any, error) {
			atomic.AddInt64(&counter, 1)
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		}),
	}

	c := NewCoordinator(workers, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c.Start(ctx)

	for i := 0; i < 10; i++ {
		c.Submit(&Task{ID: fmt.Sprintf("t%d", i), Type: "test"})
	}

	done := make(chan struct{})
	go func() {
		count := 0
		for range c.Results() {
			count++
			if count == 10 {
				break
			}
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for results")
	}

	c.Stop()

	if atomic.LoadInt64(&counter) != 10 {
		t.Errorf("expected 10 tasks executed, got %d", counter)
	}
}
