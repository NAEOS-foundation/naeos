package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Payload  map[string]any `json:"payload"`
	Priority int            `json:"priority"`
}

type TaskResult struct {
	TaskID  string         `json:"task_id"`
	Output  map[string]any `json:"output"`
	Error   string         `json:"error,omitempty"`
	Worker  string         `json:"worker"`
	Latency time.Duration  `json:"latency"`
}

type Worker interface {
	ID() string
	Execute(ctx context.Context, task *Task) (*TaskResult, error)
}

type Coordinator struct {
	workers  []Worker
	taskCh   chan *Task
	resultCh chan *TaskResult
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

func NewCoordinator(workers []Worker, queueSize int) *Coordinator {
	if queueSize <= 0 {
		queueSize = 100
	}
	return &Coordinator{
		workers:  workers,
		taskCh:   make(chan *Task, queueSize),
		resultCh: make(chan *TaskResult, queueSize),
	}
}

func (c *Coordinator) Submit(task *Task) {
	c.taskCh <- task
}

func (c *Coordinator) Start(ctx context.Context) {
	for _, w := range c.workers {
		c.wg.Add(1)
		go c.workerLoop(ctx, w)
	}
}

func (c *Coordinator) workerLoop(ctx context.Context, w Worker) {
	defer c.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-c.taskCh:
			if !ok {
				return
			}
			start := time.Now()
			result, err := w.Execute(ctx, task)
			latency := time.Since(start)
			if err != nil {
				result = &TaskResult{
					TaskID:  task.ID,
					Error:   err.Error(),
					Worker:  w.ID(),
					Latency: latency,
				}
			}
			if result != nil {
				result.Latency = latency
				result.Worker = w.ID()
			}
			c.resultCh <- result
		}
	}
}

func (c *Coordinator) Results() <-chan *TaskResult {
	return c.resultCh
}

func (c *Coordinator) Stop() {
	close(c.taskCh)
	c.wg.Wait()
	close(c.resultCh)
}

func (c *Coordinator) WorkerCount() int {
	return len(c.workers)
}

type SimpleWorker struct {
	workerID string
	handler  func(ctx context.Context, task *Task) (map[string]any, error)
}

func NewSimpleWorker(id string, handler func(ctx context.Context, task *Task) (map[string]any, error)) *SimpleWorker {
	return &SimpleWorker{workerID: id, handler: handler}
}

func (w *SimpleWorker) ID() string {
	return w.workerID
}

func (w *SimpleWorker) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	output, err := w.handler(ctx, task)
	if err != nil {
		return nil, err
	}
	return &TaskResult{
		TaskID: task.ID,
		Output: output,
	}, nil
}

type LoadBalancer struct {
	workers []Worker
	counter uint64
	mu      sync.Mutex
}

func NewLoadBalancer(workers []Worker) *LoadBalancer {
	return &LoadBalancer{workers: workers}
}

func (lb *LoadBalancer) Next() Worker {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if len(lb.workers) == 0 {
		return nil
	}
	w := lb.workers[lb.counter%uint64(len(lb.workers))]
	lb.counter++
	return w
}

func (lb *LoadBalancer) WorkerCount() int {
	return len(lb.workers)
}

type ResultAggregator struct {
	results []TaskResult
	mu      sync.Mutex
}

func NewResultAggregator() *ResultAggregator {
	return &ResultAggregator{}
}

func (ra *ResultAggregator) Add(result TaskResult) {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	ra.results = append(ra.results, result)
}

func (ra *ResultAggregator) All() []TaskResult {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	out := make([]TaskResult, len(ra.results))
	copy(out, ra.results)
	return out
}

func (ra *ResultAggregator) Failed() []TaskResult {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	var out []TaskResult
	for _, r := range ra.results {
		if r.Error != "" {
			out = append(out, r)
		}
	}
	return out
}

func (ra *ResultAggregator) Count() int {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	return len(ra.results)
}

func (ra *ResultAggregator) Summary() string {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	failed := 0
	for _, r := range ra.results {
		if r.Error != "" {
			failed++
		}
	}
	return fmt.Sprintf("%d total, %d succeeded, %d failed", len(ra.results), len(ra.results)-failed, failed)
}
