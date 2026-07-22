package middleware

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type StageFunc func(ctx context.Context, input *StageInput) (*StageOutput, error)

type StageInput struct {
	Stage  string
	Data   []byte
	Labels map[string]string
}

type StageOutput struct {
	Data   []byte
	Labels map[string]string
}

type Middleware interface {
	Name() string
	Wrap(stage string, next StageFunc) StageFunc
}

type Chain struct {
	middlewares map[string][]Middleware
}

func NewChain() *Chain {
	return &Chain{
		middlewares: make(map[string][]Middleware),
	}
}

func (c *Chain) Use(stage string, mw Middleware) {
	c.middlewares[stage] = append(c.middlewares[stage], mw)
}

func (c *Chain) Execute(stage string, input *StageInput, handler StageFunc) (*StageOutput, error) {
	return c.ExecuteContext(context.Background(), stage, input, handler)
}

func (c *Chain) ExecuteContext(ctx context.Context, stage string, input *StageInput, handler StageFunc) (*StageOutput, error) {
	mws := c.middlewares[stage]
	current := handler
	for i := len(mws) - 1; i >= 0; i-- {
		mw := mws[i]
		nextFn := current
		current = func(ctx context.Context, in *StageInput) (*StageOutput, error) {
			return mw.Wrap(in.Stage, nextFn)(ctx, in)
		}
	}
	return current(ctx, input)
}

type LogMiddleware struct {
	LogFunc func(msg string, args ...any)
}

func (l *LogMiddleware) Name() string { return "log" }

func (l *LogMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		start := time.Now()
		if l.LogFunc != nil {
			l.LogFunc("stage start", "stage", stage)
		}
		output, err := next(ctx, input)
		duration := time.Since(start)
		if l.LogFunc != nil {
			if err != nil {
				l.LogFunc("stage failed", "stage", stage, "duration", duration, "error", err)
			} else {
				l.LogFunc("stage complete", "stage", stage, "duration", duration)
			}
		}
		return output, err
	}
}

type MetricsMiddleware struct {
	RecordFunc func(stage string, duration time.Duration, err error)
}

func (m *MetricsMiddleware) Name() string { return "metrics" }

func (m *MetricsMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		start := time.Now()
		output, err := next(ctx, input)
		if m.RecordFunc != nil {
			m.RecordFunc(stage, time.Since(start), err)
		}
		return output, err
	}
}

type AuthMiddleware struct {
	ValidateToken func(token string) error
	TokenHeader   string
}

func (a *AuthMiddleware) Name() string { return "auth" }

func (a *AuthMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		if a.ValidateToken != nil && a.TokenHeader != "" {
			token := input.Labels[a.TokenHeader]
			if token == "" {
				return nil, fmt.Errorf("missing auth token in label %q", a.TokenHeader)
			}
			if err := a.ValidateToken(token); err != nil {
				return nil, fmt.Errorf("auth failed: %w", err)
			}
		}
		return next(ctx, input)
	}
}

type CacheMiddleware struct {
	Get func(key string) ([]byte, bool)
	Set func(key string, data []byte)
}

func (c *CacheMiddleware) Name() string { return "cache" }

func (c *CacheMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		key := fmt.Sprintf("%s:%x", stage, input.Data)
		if c.Get != nil {
			if cached, ok := c.Get(key); ok {
				return &StageOutput{Data: cached, Labels: input.Labels}, nil
			}
		}
		output, err := next(ctx, input)
		if err == nil && c.Set != nil && output != nil {
			c.Set(key, output.Data)
		}
		return output, err
	}
}

// RetryMiddleware retries failed stage executions with configurable max attempts and backoff.
type RetryMiddleware struct {
	MaxAttempts int
	Backoff     time.Duration
	OnRetry     func(attempt int, err error)
}

func (r *RetryMiddleware) Name() string { return "retry" }

func (r *RetryMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		maxAttempts := r.MaxAttempts
		if maxAttempts <= 0 {
			maxAttempts = 3
		}
		backoff := r.Backoff
		if backoff <= 0 {
			backoff = 100 * time.Millisecond
		}

		var lastErr error
		for attempt := 0; attempt < maxAttempts; attempt++ {
			output, err := next(ctx, input)
			if err == nil {
				return output, nil
			}
			lastErr = err
			if r.OnRetry != nil {
				r.OnRetry(attempt+1, err)
			}
			if attempt < maxAttempts-1 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
				}
			}
		}
		return nil, lastErr
	}
}

// RateLimitMiddleware rate-limits stage executions using a token bucket pattern.
type RateLimitMiddleware struct {
	Rate     int
	Burst    int
	AllowAll bool
	mu       sync.Mutex
	tokens   float64
	lastTime time.Time
}

func (rl *RateLimitMiddleware) Name() string { return "rate_limit" }

func (rl *RateLimitMiddleware) initBucket() {
	if rl.Rate <= 0 {
		rl.Rate = 1
	}
	if rl.Burst <= 0 {
		rl.Burst = rl.Rate
	}
	if rl.tokens == 0 && rl.lastTime.IsZero() {
		rl.tokens = float64(rl.Burst)
		rl.lastTime = time.Now()
	}
}

func (rl *RateLimitMiddleware) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.initBucket()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()
	rl.tokens += elapsed * float64(rl.Rate)
	if rl.tokens > float64(rl.Burst) {
		rl.tokens = float64(rl.Burst)
	}
	rl.lastTime = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

func (rl *RateLimitMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		if rl.AllowAll {
			return next(ctx, input)
		}
		if !rl.allow() {
			return nil, fmt.Errorf("rate limit exceeded for stage %q", stage)
		}
		return next(ctx, input)
	}
}

// TimeoutMiddleware wraps stage execution with a context deadline.
type TimeoutMiddleware struct {
	Timeout time.Duration
}

func (t *TimeoutMiddleware) Name() string { return "timeout" }

func (t *TimeoutMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		timeout := t.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return next(ctx, input)
	}
}

// CircuitState represents the state of a circuit breaker.
type CircuitState int32

const (
	CircuitClosed   CircuitState = 0
	CircuitOpen     CircuitState = 1
	CircuitHalfOpen CircuitState = 2
)

// CircuitBreakerMiddleware tracks failures and opens the circuit after a threshold,
// with half-open recovery.
type CircuitBreakerMiddleware struct {
	Threshold     int
	ResetTimeout  time.Duration
	HalfOpenMax   int
	failureCount  int32
	successCount  int32
	state         int32
	lastFailureAt time.Time
	mu            sync.Mutex
}

func (cb *CircuitBreakerMiddleware) Name() string { return "circuit_breaker" }

func (cb *CircuitBreakerMiddleware) getState() CircuitState {
	return CircuitState(atomic.LoadInt32(&cb.state))
}

func (cb *CircuitBreakerMiddleware) getFailureCount() int {
	return int(atomic.LoadInt32(&cb.failureCount))
}

func (cb *CircuitBreakerMiddleware) getSuccessCount() int {
	return int(atomic.LoadInt32(&cb.successCount))
}

func (cb *CircuitBreakerMiddleware) reset() {
	atomic.StoreInt32(&cb.failureCount, 0)
	atomic.StoreInt32(&cb.successCount, 0)
	atomic.StoreInt32(&cb.state, int32(CircuitClosed))
}

func (cb *CircuitBreakerMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		threshold := cb.Threshold
		if threshold <= 0 {
			threshold = 5
		}
		resetTimeout := cb.ResetTimeout
		if resetTimeout <= 0 {
			resetTimeout = 30 * time.Second
		}
		halfOpenMax := cb.HalfOpenMax
		if halfOpenMax <= 0 {
			halfOpenMax = 1
		}

		cb.mu.Lock()
		switch state := cb.getState(); state {
		case CircuitOpen:
			if time.Since(cb.lastFailureAt) > resetTimeout {
				atomic.StoreInt32(&cb.state, int32(CircuitHalfOpen))
				atomic.StoreInt32(&cb.successCount, 0)
			} else {
				cb.mu.Unlock()
				return nil, fmt.Errorf("circuit breaker is open for stage %q", stage)
			}
		case CircuitHalfOpen:
			if cb.getSuccessCount() >= halfOpenMax {
				cb.mu.Unlock()
				return nil, fmt.Errorf("circuit breaker is half-open at capacity for stage %q", stage)
			}
		}
		cb.mu.Unlock()

		output, err := next(ctx, input)
		if err != nil {
			atomic.AddInt32(&cb.failureCount, 1)
			cb.mu.Lock()
			cb.lastFailureAt = time.Now()
			failures := cb.getFailureCount()
			currentState := cb.getState()
			if currentState == CircuitHalfOpen {
				atomic.StoreInt32(&cb.state, int32(CircuitOpen))
			} else if failures >= threshold {
				atomic.StoreInt32(&cb.state, int32(CircuitOpen))
			}
			cb.mu.Unlock()
			return nil, err
		}

		cb.mu.Lock()
		currentState := cb.getState()
		if currentState == CircuitHalfOpen {
			newSuccess := atomic.AddInt32(&cb.successCount, 1)
			if int(newSuccess) >= halfOpenMax {
				cb.reset()
			}
		}
		cb.mu.Unlock()

		return output, nil
	}
}

// ValidationMiddleware validates StageInput data before execution using a
// user-provided validator function.
type ValidationMiddleware struct {
	Validate func(input *StageInput) error
}

func (v *ValidationMiddleware) Name() string { return "validation" }

func (v *ValidationMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		if v.Validate != nil {
			if err := v.Validate(input); err != nil {
				return nil, fmt.Errorf("validation failed for stage %q: %w", stage, err)
			}
		}
		return next(ctx, input)
	}
}

// TransformMiddleware transforms input/output data using user-provided functions.
type TransformMiddleware struct {
	TransformInput  func(input *StageInput) (*StageInput, error)
	TransformOutput func(output *StageOutput) (*StageOutput, error)
}

func (t *TransformMiddleware) Name() string { return "transform" }

func (t *TransformMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		if t.TransformInput != nil {
			transformed, err := t.TransformInput(input)
			if err != nil {
				return nil, fmt.Errorf("input transform failed for stage %q: %w", stage, err)
			}
			input = transformed
		}
		output, err := next(ctx, input)
		if err != nil {
			return nil, err
		}
		if t.TransformOutput != nil && output != nil {
			transformed, err := t.TransformOutput(output)
			if err != nil {
				return nil, fmt.Errorf("output transform failed for stage %q: %w", stage, err)
			}
			output = transformed
		}
		return output, nil
	}
}
