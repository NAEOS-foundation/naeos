package distributed

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func newTestWorker(id string) *SimpleWorker {
	return NewSimpleWorker(id, func(ctx context.Context, task *Task) (map[string]any, error) {
		return map[string]any{"result": "ok"}, nil
	})
}

func TestCoordinatorBasic(t *testing.T) {
	workers := []Worker{newTestWorker("w1"), newTestWorker("w2")}
	c := NewCoordinator(workers, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)
	c.Submit(&Task{ID: "t1", Type: "test"})
	c.Submit(&Task{ID: "t2", Type: "test"})
	var results []*TaskResult
	timeout := time.After(2 * time.Second)
	for len(results) < 2 {
		select {
		case r := <-c.Results():
			results = append(results, r)
		case <-timeout:
			t.Fatal("timeout waiting for results")
		}
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	m := c.Metrics()
	if m.TasksCompleted != 2 {
		t.Errorf("expected 2 completed, got %d", m.TasksCompleted)
	}
	c.Stop()
}

func TestCoordinatorRetry(t *testing.T) {
	var attempts atomic.Int32
	worker := NewSimpleWorker("retry-w", func(ctx context.Context, task *Task) (map[string]any, error) {
		count := attempts.Add(1)
		if count < 3 {
			return nil, fmt.Errorf("transient error")
		}
		return map[string]any{"ok": true}, nil
	})
	c := NewCoordinator([]Worker{worker}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)
	c.Submit(&Task{ID: "retry-t", Type: "test", MaxRetry: 3})
	var result *TaskResult
	timeout := time.After(5 * time.Second)
	select {
	case result = <-c.Results():
	case <-timeout:
		t.Fatal("timeout waiting for retry result")
	}
	if !result.Succeeded {
		t.Errorf("expected success after retries, got error: %s", result.Error)
	}
	if result.Attempt != 3 {
		t.Errorf("expected 3 attempts, got %d", result.Attempt)
	}
	if result.Retries != 2 {
		t.Errorf("expected 2 retries, got %d", result.Retries)
	}
	c.Stop()
}

func TestCoordinatorTimeout(t *testing.T) {
	worker := NewSimpleWorker("slow-w", func(ctx context.Context, task *Task) (map[string]any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Second):
			return map[string]any{"done": true}, nil
		}
	})
	c := NewCoordinator([]Worker{worker}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)
	c.Submit(&Task{ID: "timeout-t", Type: "test", Timeout: 50 * time.Millisecond, MaxRetry: 0})
	var result *TaskResult
	timeout := time.After(2 * time.Second)
	select {
	case result = <-c.Results():
	case <-timeout:
		t.Fatal("timeout waiting for result")
	}
	if result.Error == "" {
		t.Error("expected timeout error")
	}
	c.Stop()
}

func TestCoordinatorDrain(t *testing.T) {
	worker := NewSimpleWorker("drain-w", func(ctx context.Context, task *Task) (map[string]any, error) {
		time.Sleep(10 * time.Millisecond)
		return map[string]any{"ok": true}, nil
	})
	c := NewCoordinator([]Worker{worker}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)
	for i := 0; i < 5; i++ {
		c.Submit(&Task{ID: fmt.Sprintf("d%d", i), Type: "test"})
	}
	time.Sleep(20 * time.Millisecond)
	c.Drain()
}

func TestLoadBalancer(t *testing.T) {
	workers := []Worker{newTestWorker("w1"), newTestWorker("w2"), newTestWorker("w3")}
	lb := NewLoadBalancer(workers)
	if lb.WorkerCount() != 3 {
		t.Errorf("expected 3 workers, got %d", lb.WorkerCount())
	}
	w1 := lb.Next()
	w2 := lb.Next()
	w3 := lb.Next()
	if w1.ID() == w2.ID() || w2.ID() == w3.ID() {
		t.Error("expected round-robin distribution")
	}
	lbEmpty := NewLoadBalancer(nil)
	if lbEmpty.Next() != nil {
		t.Error("expected nil from empty load balancer")
	}
}

func TestResultAggregator(t *testing.T) {
	agg := NewResultAggregator()
	agg.Add(TaskResult{TaskID: "t1", Succeeded: true})
	agg.Add(TaskResult{TaskID: "t2", Error: "failed", Succeeded: false})
	agg.Add(TaskResult{TaskID: "t3", Succeeded: true})
	if agg.Count() != 3 {
		t.Errorf("expected 3, got %d", agg.Count())
	}
	if len(agg.Failed()) != 1 {
		t.Errorf("expected 1 failed, got %d", len(agg.Failed()))
	}
	if len(agg.Succeeded()) != 2 {
		t.Errorf("expected 2 succeeded, got %d", len(agg.Succeeded()))
	}
	summary := agg.Summary()
	if len(summary) == 0 {
		t.Error("expected non-empty summary")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	if cb.State() != "closed" {
		t.Errorf("expected closed, got %s", cb.State())
	}
	if !cb.Allow() {
		t.Error("expected allowed when closed")
	}
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != "open" {
		t.Errorf("expected open after 3 failures, got %s", cb.State())
	}
	if cb.Allow() {
		t.Error("expected rejected when open")
	}
	time.Sleep(150 * time.Millisecond)
	if !cb.Allow() {
		t.Error("expected allowed after reset timeout")
	}
	if cb.State() != "half-open" {
		t.Errorf("expected half-open, got %s", cb.State())
	}
	cb.RecordSuccess()
	if cb.State() != "closed" {
		t.Errorf("expected closed after success, got %s", cb.State())
	}
}

func TestCircuitBreakerWorker(t *testing.T) {
	worker := NewSimpleWorker("cb-w", func(ctx context.Context, task *Task) (map[string]any, error) {
		return nil, fmt.Errorf("always fails")
	})
	cb := NewCircuitBreaker(2, 100*time.Millisecond)
	cbw := NewCircuitBreakerWorker(worker, cb)
	ctx := context.Background()
	_, err := cbw.Execute(ctx, &Task{ID: "t1"})
	if err == nil {
		t.Error("expected error")
	}
	_, err = cbw.Execute(ctx, &Task{ID: "t2"})
	if err == nil {
		t.Error("expected error")
	}
	_, err = cbw.Execute(ctx, &Task{ID: "t3"})
	if err == nil {
		t.Error("expected circuit breaker error")
	}
}

func TestHealthChecker(t *testing.T) {
	workers := []Worker{newTestWorker("hw1"), newTestWorker("hw2")}
	hc := NewHealthChecker(workers, 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hc.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	status := hc.Status()
	if len(status) != 2 {
		t.Errorf("expected 2 health statuses, got %d", len(status))
	}
	for _, h := range status {
		if !h.Healthy {
			t.Errorf("expected worker %s to be healthy", h.WorkerID)
		}
	}
	healthy := hc.HealthyWorkers()
	if len(healthy) != 2 {
		t.Errorf("expected 2 healthy workers, got %d", len(healthy))
	}
	hc.Stop()
}

func TestComputeBackoff(t *testing.T) {
	b1 := computeBackoff(1)
	b2 := computeBackoff(2)
	b3 := computeBackoff(3)
	if b1 >= b2 || b2 >= b3 {
		t.Error("expected increasing backoff")
	}
	if b3 > 5*time.Second {
		t.Error("expected max 5s backoff")
	}
}

func TestSubmitPriorityOrder(t *testing.T) {
	worker := NewSimpleWorker("p-w", func(ctx context.Context, task *Task) (map[string]any, error) {
		return map[string]any{"id": task.ID}, nil
	})
	c := NewCoordinator([]Worker{worker}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)

	c.Submit(&Task{ID: "low", Priority: 1})
	c.SubmitPriority(&Task{ID: "high", Priority: 10})
	c.SubmitPriority(&Task{ID: "mid", Priority: 5})

	timeout := time.After(2 * time.Second)
	var ids []string
	for len(ids) < 3 {
		select {
		case r := <-c.Results():
			if r.Succeeded {
				ids = append(ids, r.Output["id"].(string))
			}
		case <-timeout:
			t.Fatal("timeout waiting for results")
		}
	}

	if len(ids) != 3 {
		t.Fatalf("expected 3 results, got %d", len(ids))
	}
	if ids[0] != "high" {
		t.Errorf("expected first result to be 'high' (priority 10), got %s", ids[0])
	}
	if ids[1] != "mid" {
		t.Errorf("expected second result to be 'mid' (priority 5), got %s", ids[1])
	}
	if ids[2] != "low" {
		t.Errorf("expected third result to be 'low' (priority 1), got %s", ids[2])
	}
	c.Stop()
}

func TestSubmitPriorityDefault(t *testing.T) {
	worker := NewSimpleWorker("pd-w", func(ctx context.Context, task *Task) (map[string]any, error) {
		return map[string]any{"id": task.ID}, nil
	})
	c := NewCoordinator([]Worker{worker}, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)

	c.Submit(&Task{ID: "a"})
	c.Submit(&Task{ID: "b"})
	c.Submit(&Task{ID: "c"})

	timeout := time.After(2 * time.Second)
	var count int
	for count < 3 {
		select {
		case <-c.Results():
			count++
		case <-timeout:
			t.Fatal("timeout waiting for results")
		}
	}
	c.Stop()
}

func TestRegisterAgent(t *testing.T) {
	c := NewCoordinator(nil, 10)

	agent := &Agent{ID: "agent-1", Type: "builder"}
	err := c.RegisterAgent(agent)
	if err != nil {
		t.Fatalf("RegisterAgent failed: %v", err)
	}

	got := c.Agent("agent-1")
	if got == nil {
		t.Fatal("expected to find agent-1")
	}
	if got.Status != AgentStatusOnline {
		t.Errorf("expected online, got %s", got.Status)
	}
	if got.RegisteredAt.IsZero() {
		t.Error("expected non-zero RegisteredAt")
	}
}

func TestRegisterAgentEmptyID(t *testing.T) {
	c := NewCoordinator(nil, 10)
	err := c.RegisterAgent(&Agent{ID: ""})
	if err == nil {
		t.Fatal("expected error for empty agent ID")
	}
}

func TestUnregisterAgent(t *testing.T) {
	c := NewCoordinator(nil, 10)
	c.RegisterAgent(&Agent{ID: "agent-1"})

	err := c.UnregisterAgent("agent-1")
	if err != nil {
		t.Fatalf("UnregisterAgent failed: %v", err)
	}

	got := c.Agent("agent-1")
	if got != nil {
		t.Error("expected agent to be removed")
	}
}

func TestUnregisterAgentNotFound(t *testing.T) {
	c := NewCoordinator(nil, 10)
	err := c.UnregisterAgent("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
}

func TestRecordHeartbeat(t *testing.T) {
	c := NewCoordinator(nil, 10)
	c.RegisterAgent(&Agent{ID: "agent-1"})

	original := c.Agent("agent-1").LastHeartbeat
	time.Sleep(time.Millisecond)

	err := c.RecordHeartbeat("agent-1")
	if err != nil {
		t.Fatalf("RecordHeartbeat failed: %v", err)
	}

	updated := c.Agent("agent-1").LastHeartbeat
	if !updated.After(original) {
		t.Error("expected heartbeat to be updated")
	}
}

func TestRecordHeartbeatNotFound(t *testing.T) {
	c := NewCoordinator(nil, 10)
	err := c.RecordHeartbeat("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
}

func TestListAgents(t *testing.T) {
	c := NewCoordinator(nil, 10)
	c.RegisterAgent(&Agent{ID: "a1"})
	c.RegisterAgent(&Agent{ID: "a2"})
	c.RegisterAgent(&Agent{ID: "a3"})

	agents := c.ListAgents()
	if len(agents) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(agents))
	}

	ids := make(map[string]bool)
	for _, a := range agents {
		ids[a.ID] = true
	}
	if !ids["a1"] || !ids["a2"] || !ids["a3"] {
		t.Error("missing expected agent IDs")
	}
}

func TestAgentStatusBusy(t *testing.T) {
	agent := &Agent{ID: "busy-agent"}
	agent.Status = AgentStatusBusy
	if agent.Status != AgentStatusBusy {
		t.Errorf("expected busy status")
	}
}

func TestCoordinatorMultipleWorkers(t *testing.T) {
	workers := []Worker{newTestWorker("w1"), newTestWorker("w2"), newTestWorker("w3")}
	c := NewCoordinator(workers, 20)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx)

	for i := 0; i < 6; i++ {
		c.Submit(&Task{ID: fmt.Sprintf("t%d", i), Type: "test"})
	}

	timeout := time.After(2 * time.Second)
	var results []*TaskResult
	for len(results) < 6 {
		select {
		case r := <-c.Results():
			results = append(results, r)
		case <-timeout:
			t.Fatal("timeout waiting for all results")
		}
	}
	if len(results) != 6 {
		t.Fatalf("expected 6 results, got %d", len(results))
	}
	c.Stop()
}
