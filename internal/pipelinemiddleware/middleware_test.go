// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- helpers ---

func noopHandler() StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		return &StageOutput{Data: input.Data, Labels: input.Labels}, nil
	}
}

func failingHandler(msg string) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		return nil, errors.New(msg)
	}
}

func contextAwareHandler(delay time.Duration) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		select {
		case <-time.After(delay):
			return &StageOutput{Data: []byte("done")}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

type testMiddleware struct {
	name    string
	order   *[]string
	called  *bool
	callLog *[]bool
}

func (m *testMiddleware) Name() string { return m.name }

func (m *testMiddleware) Wrap(stage string, next StageFunc) StageFunc {
	return func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		if m.order != nil {
			*m.order = append(*m.order, m.name)
		}
		if m.called != nil {
			*m.called = true
		}
		if m.callLog != nil {
			*m.callLog = append(*m.callLog, true)
		}
		return next(ctx, input)
	}
}

// ==================== Chain tests ====================

func TestChainExecute(t *testing.T) {
	chain := NewChain()
	called := false
	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		called = true
		return &StageOutput{Data: input.Data}, nil
	}
	output, err := chain.Execute("parse", &StageInput{Stage: "parse", Data: []byte("hello")}, handler)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("handler not called")
	}
	if string(output.Data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(output.Data))
	}
}

func TestChainMiddlewareOrder(t *testing.T) {
	chain := NewChain()
	var order []string
	chain.Use("parse", &testMiddleware{name: "first", order: &order})
	chain.Use("parse", &testMiddleware{name: "second", order: &order})
	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		order = append(order, "handler")
		return &StageOutput{Data: input.Data}, nil
	}
	chain.Execute("parse", &StageInput{Stage: "parse"}, handler)
	expected := []string{"first", "second", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("expected order[%d]=%s, got %s", i, v, order[i])
		}
	}
}

func TestChainDifferentStages(t *testing.T) {
	chain := NewChain()
	var parseMWCalled []bool
	chain.Use("parse", &testMiddleware{name: "parse-mw", callLog: &parseMWCalled})
	parseHandler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		return &StageOutput{Data: []byte("parsed")}, nil
	}
	generateHandler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		return &StageOutput{Data: []byte("generated")}, nil
	}
	chain.Execute("parse", &StageInput{Stage: "parse"}, parseHandler)
	if len(parseMWCalled) != 1 {
		t.Errorf("parse middleware not called for parse stage, callCount=%d", len(parseMWCalled))
	}
	callsBefore := len(parseMWCalled)
	chain.Execute("generate", &StageInput{Stage: "generate"}, generateHandler)
	if len(parseMWCalled) != callsBefore {
		t.Errorf("parse middleware should not be called for generate stage")
	}
}

func TestChainNoMiddlewares(t *testing.T) {
	chain := NewChain()
	output, err := chain.Execute("noop", &StageInput{Stage: "noop", Data: []byte("x")}, noopHandler())
	if err != nil {
		t.Fatal(err)
	}
	if string(output.Data) != "x" {
		t.Errorf("expected 'x', got %q", string(output.Data))
	}
}

// ==================== LogMiddleware tests ====================

func TestLogMiddleware(t *testing.T) {
	var msgs []string
	mw := &LogMiddleware{
		LogFunc: func(msg string, args ...any) {
			msgs = append(msgs, msg)
		},
	}
	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		return &StageOutput{Data: []byte("ok")}, nil
	}
	wrapped := mw.Wrap("test", handler)
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 log messages, got %d", len(msgs))
	}
	if msgs[0] != "stage start" || msgs[1] != "stage complete" {
		t.Errorf("unexpected messages: %v", msgs)
	}
}

func TestLogMiddlewareError(t *testing.T) {
	var msgs []string
	mw := &LogMiddleware{
		LogFunc: func(msg string, args ...any) {
			msgs = append(msgs, msg)
		},
	}
	wrapped := mw.Wrap("test", failingHandler("boom"))
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[1] != "stage failed" {
		t.Errorf("expected 'stage failed', got %q", msgs[1])
	}
}

func TestLogMiddlewareNilLogFunc(t *testing.T) {
	mw := &LogMiddleware{}
	wrapped := mw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogMiddlewareName(t *testing.T) {
	mw := &LogMiddleware{}
	if mw.Name() != "log" {
		t.Errorf("expected 'log', got %q", mw.Name())
	}
}

// ==================== MetricsMiddleware tests ====================

func TestMetricsMiddleware(t *testing.T) {
	var recordedStage string
	var recordedDuration time.Duration
	var recordedErr error
	mw := &MetricsMiddleware{
		RecordFunc: func(stage string, duration time.Duration, err error) {
			recordedStage = stage
			recordedDuration = duration
			recordedErr = err
		},
	}
	wrapped := mw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if recordedStage != "test" {
		t.Errorf("expected stage 'test', got %q", recordedStage)
	}
	if recordedDuration <= 0 {
		t.Error("expected positive duration")
	}
	if recordedErr != nil {
		t.Errorf("expected nil error, got %v", recordedErr)
	}
}

func TestMetricsMiddlewareRecordsError(t *testing.T) {
	var recordedErr error
	mw := &MetricsMiddleware{
		RecordFunc: func(stage string, duration time.Duration, err error) {
			recordedErr = err
		},
	}
	wrapped := mw.Wrap("test", failingHandler("fail"))
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	if recordedErr == nil {
		t.Error("expected error to be recorded")
	}
}

func TestMetricsMiddlewareName(t *testing.T) {
	mw := &MetricsMiddleware{}
	if mw.Name() != "metrics" {
		t.Errorf("expected 'metrics', got %q", mw.Name())
	}
}

// ==================== AuthMiddleware tests ====================

func TestAuthMiddleware(t *testing.T) {
	mw := &AuthMiddleware{
		ValidateToken: func(token string) error {
			if token != "valid" {
				return fmt.Errorf("invalid token")
			}
			return nil
		},
		TokenHeader: "auth_token",
	}
	handler := noopHandler()
	wrapped := mw.Wrap("test", handler)

	_, err := wrapped(context.Background(), &StageInput{
		Stage: "test", Labels: map[string]string{"auth_token": "valid"},
	})
	if err != nil {
		t.Errorf("expected no error for valid token, got %v", err)
	}

	_, err = wrapped(context.Background(), &StageInput{
		Stage: "test", Labels: map[string]string{"auth_token": "bad"},
	})
	if err == nil {
		t.Error("expected error for invalid token")
	}

	_, err = wrapped(context.Background(), &StageInput{
		Stage: "test", Labels: map[string]string{},
	})
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestAuthMiddlewareNilValidator(t *testing.T) {
	mw := &AuthMiddleware{TokenHeader: "x"}
	wrapped := mw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Errorf("expected no error with nil validator, got %v", err)
	}
}

func TestAuthMiddlewareEmptyHeader(t *testing.T) {
	mw := &AuthMiddleware{
		ValidateToken: func(token string) error { return nil },
	}
	wrapped := mw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Errorf("expected no error with empty header, got %v", err)
	}
}

func TestAuthMiddlewareName(t *testing.T) {
	mw := &AuthMiddleware{}
	if mw.Name() != "auth" {
		t.Errorf("expected 'auth', got %q", mw.Name())
	}
}

// ==================== CacheMiddleware tests ====================

func TestCacheMiddleware(t *testing.T) {
	cache := make(map[string][]byte)
	var mu sync.Mutex
	mw := &CacheMiddleware{
		Get: func(key string) ([]byte, bool) {
			mu.Lock()
			defer mu.Unlock()
			v, ok := cache[key]
			return v, ok
		},
		Set: func(key string, data []byte) {
			mu.Lock()
			defer mu.Unlock()
			cache[key] = data
		},
	}
	callCount := 0
	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		callCount++
		return &StageOutput{Data: []byte("result")}, nil
	}
	wrapped := mw.Wrap("test", handler)

	_, _ = wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("key1")})
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	_, _ = wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("key1")})
	if callCount != 1 {
		t.Errorf("expected cache hit (1 call), got %d", callCount)
	}
}

func TestCacheMiddlewareMissDifferentKeys(t *testing.T) {
	cache := make(map[string][]byte)
	var mu sync.Mutex
	mw := &CacheMiddleware{
		Get: func(key string) ([]byte, bool) {
			mu.Lock()
			defer mu.Unlock()
			v, ok := cache[key]
			return v, ok
		},
		Set: func(key string, data []byte) {
			mu.Lock()
			defer mu.Unlock()
			cache[key] = data
		},
	}
	callCount := 0
	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		callCount++
		return &StageOutput{Data: []byte("result")}, nil
	}
	wrapped := mw.Wrap("test", handler)

	_, _ = wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("key1")})
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("key2")})
	if callCount != 2 {
		t.Errorf("expected 2 calls for different keys, got %d", callCount)
	}
}

func TestCacheMiddlewareSkipsErrorResults(t *testing.T) {
	cache := make(map[string][]byte)
	var mu sync.Mutex
	mw := &CacheMiddleware{
		Get: func(key string) ([]byte, bool) {
			mu.Lock()
			defer mu.Unlock()
			v, ok := cache[key]
			return v, ok
		},
		Set: func(key string, data []byte) {
			mu.Lock()
			defer mu.Unlock()
			cache[key] = data
		},
	}
	wrapped := mw.Wrap("test", failingHandler("fail"))
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("key1")})
	mu.Lock()
	_, ok := cache["test:6b657931"]
	mu.Unlock()
	if ok {
		t.Error("expected error result not to be cached")
	}
}

func TestCacheMiddlewareName(t *testing.T) {
	mw := &CacheMiddleware{}
	if mw.Name() != "cache" {
		t.Errorf("expected 'cache', got %q", mw.Name())
	}
}

// ==================== RetryMiddleware tests ====================

func TestRetryMiddlewareSucceedsFirstAttempt(t *testing.T) {
	mw := &RetryMiddleware{MaxAttempts: 3, Backoff: time.Millisecond}
	calls := 0
	wrapped := mw.Wrap("test", func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		return &StageOutput{Data: []byte("ok")}, nil
	})
	out, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
	if string(out.Data) != "ok" {
		t.Errorf("expected 'ok', got %q", string(out.Data))
	}
}

func TestRetryMiddlewareRetriesThenSucceeds(t *testing.T) {
	mw := &RetryMiddleware{MaxAttempts: 3, Backoff: time.Millisecond}
	calls := 0
	wrapped := mw.Wrap("test", func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("transient")
		}
		return &StageOutput{Data: []byte("ok")}, nil
	})
	out, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
	if string(out.Data) != "ok" {
		t.Errorf("expected 'ok', got %q", string(out.Data))
	}
}

func TestRetryMiddlewareExhaustsAttempts(t *testing.T) {
	mw := &RetryMiddleware{MaxAttempts: 3, Backoff: time.Millisecond}
	calls := 0
	wrapped := mw.Wrap("test", func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		return nil, errors.New("permanent")
	})
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryMiddlewareOnRetryCallback(t *testing.T) {
	var retries []int
	mw := &RetryMiddleware{
		MaxAttempts: 3,
		Backoff:     time.Millisecond,
		OnRetry: func(attempt int, err error) {
			retries = append(retries, attempt)
		},
	}
	calls := 0
	wrapped := mw.Wrap("test", func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		return nil, errors.New("fail")
	})
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	if len(retries) != 3 {
		t.Fatalf("expected 3 retry callbacks, got %d", len(retries))
	}
	if retries[0] != 1 || retries[1] != 2 || retries[2] != 3 {
		t.Errorf("expected retry attempts [1,2,3], got %v", retries)
	}
}

func TestRetryMiddlewareContextCancellation(t *testing.T) {
	mw := &RetryMiddleware{MaxAttempts: 5, Backoff: 5 * time.Second}
	wrapped := mw.Wrap("test", failingHandler("fail"))
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	_, err := wrapped(ctx, &StageInput{Stage: "test"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetryMiddlewareDefaults(t *testing.T) {
	mw := &RetryMiddleware{}
	wrapped := mw.Wrap("test", failingHandler("fail"))
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRetryMiddlewareName(t *testing.T) {
	mw := &RetryMiddleware{}
	if mw.Name() != "retry" {
		t.Errorf("expected 'retry', got %q", mw.Name())
	}
}

// ==================== RateLimitMiddleware tests ====================

func TestRateLimitMiddlewareAllowsUnderLimit(t *testing.T) {
	rl := &RateLimitMiddleware{Rate: 100, Burst: 10}
	wrapped := rl.Wrap("test", noopHandler())
	for i := 0; i < 10; i++ {
		_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
		if err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}
}

func TestRateLimitMiddlewareRejectsOverLimit(t *testing.T) {
	rl := &RateLimitMiddleware{Rate: 1, Burst: 2}
	wrapped := rl.Wrap("test", noopHandler())
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err == nil {
		t.Error("expected rate limit error")
	}
}

func TestRateLimitMiddlewareAllowAll(t *testing.T) {
	rl := &RateLimitMiddleware{Rate: 0, Burst: 0, AllowAll: true}
	wrapped := rl.Wrap("test", noopHandler())
	for i := 0; i < 100; i++ {
		_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
		if err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}
}

func TestRateLimitMiddlewareName(t *testing.T) {
	rl := &RateLimitMiddleware{}
	if rl.Name() != "rate_limit" {
		t.Errorf("expected 'rate_limit', got %q", rl.Name())
	}
}

// ==================== TimeoutMiddleware tests ====================

func TestTimeoutMiddlewareCompletesInTime(t *testing.T) {
	tmw := &TimeoutMiddleware{Timeout: time.Second}
	wrapped := tmw.Wrap("test", contextAwareHandler(10*time.Millisecond))
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTimeoutMiddlewareTimesOut(t *testing.T) {
	tmw := &TimeoutMiddleware{Timeout: 10 * time.Millisecond}
	wrapped := tmw.Wrap("test", contextAwareHandler(time.Second))
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTimeoutMiddlewareDefault(t *testing.T) {
	tmw := &TimeoutMiddleware{Timeout: 0}
	wrapped := tmw.Wrap("test", contextAwareHandler(10*time.Millisecond))
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatalf("expected no error with default timeout, got %v", err)
	}
}

func TestTimeoutMiddlewareName(t *testing.T) {
	tmw := &TimeoutMiddleware{}
	if tmw.Name() != "timeout" {
		t.Errorf("expected 'timeout', got %q", tmw.Name())
	}
}

// ==================== CircuitBreakerMiddleware tests ====================

func TestCircuitBreakerClosedNormalFlow(t *testing.T) {
	cb := &CircuitBreakerMiddleware{Threshold: 3}
	wrapped := cb.Wrap("test", noopHandler())
	for i := 0; i < 3; i++ {
		_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if cb.getState() != CircuitClosed {
		t.Error("expected circuit to be closed")
	}
}

func TestCircuitBreakerOpensAfterThreshold(t *testing.T) {
	cb := &CircuitBreakerMiddleware{Threshold: 2, ResetTimeout: time.Hour}
	wrapped := cb.Wrap("test", failingHandler("fail"))
	for i := 0; i < 2; i++ {
		_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	}
	if cb.getState() != CircuitOpen {
		t.Error("expected circuit to be open")
	}
	_, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err == nil || !strings.Contains(err.Error(), "circuit breaker is open") {
		t.Errorf("expected circuit breaker open error, got %v", err)
	}
}

func TestCircuitBreakerHalfOpenRecovery(t *testing.T) {
	cb := &CircuitBreakerMiddleware{
		Threshold:    1,
		ResetTimeout: 10 * time.Millisecond,
		HalfOpenMax:  1,
	}
	wrapped := cb.Wrap("test", failingHandler("fail"))
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	if cb.getState() != CircuitOpen {
		t.Fatal("expected open state")
	}

	time.Sleep(20 * time.Millisecond)
	calls := 0
	wrapped = cb.Wrap("test", func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		return &StageOutput{Data: []byte("recovered")}, nil
	})
	out, err := wrapped(context.Background(), &StageInput{Stage: "test"})
	if err != nil {
		t.Fatalf("expected recovery, got %v", err)
	}
	if string(out.Data) != "recovered" {
		t.Errorf("expected 'recovered', got %q", string(out.Data))
	}
	if cb.getState() != CircuitClosed {
		t.Errorf("expected closed after recovery, got %d", cb.getState())
	}
}

func TestCircuitBreakerHalfOpenFailsResetsToOpen(t *testing.T) {
	cb := &CircuitBreakerMiddleware{
		Threshold:    1,
		ResetTimeout: 10 * time.Millisecond,
		HalfOpenMax:  1,
	}
	wrapped := cb.Wrap("test", failingHandler("fail"))
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	time.Sleep(20 * time.Millisecond)
	wrapped = cb.Wrap("test", failingHandler("fail again"))
	_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	if cb.getState() != CircuitOpen {
		t.Error("expected circuit to re-open after half-open failure")
	}
}

func TestCircuitBreakerDefaults(t *testing.T) {
	cb := &CircuitBreakerMiddleware{}
	calls := 0
	wrapped := cb.Wrap("test", func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		if calls <= 5 {
			return nil, errors.New("fail")
		}
		return &StageOutput{Data: []byte("ok")}, nil
	})
	for i := 0; i < 5; i++ {
		_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	}
	if cb.getState() != CircuitOpen {
		t.Error("expected open after 5 failures with default threshold")
	}
}

func TestCircuitBreakerSuccessResetsCount(t *testing.T) {
	cb := &CircuitBreakerMiddleware{Threshold: 5}
	calls := 0
	wrapped := cb.Wrap("test", func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		if calls%2 == 1 {
			return nil, errors.New("fail")
		}
		return &StageOutput{Data: []byte("ok")}, nil
	})
	for i := 0; i < 4; i++ {
		_, _ = wrapped(context.Background(), &StageInput{Stage: "test"})
	}
	if cb.getFailureCount() != 2 {
		t.Errorf("expected 2 failures after interleaved successes, got %d", cb.getFailureCount())
	}
	if cb.getState() != CircuitClosed {
		t.Error("expected circuit to remain closed")
	}
}

func TestCircuitBreakerName(t *testing.T) {
	cb := &CircuitBreakerMiddleware{}
	if cb.Name() != "circuit_breaker" {
		t.Errorf("expected 'circuit_breaker', got %q", cb.Name())
	}
}

// ==================== ValidationMiddleware tests ====================

func TestValidationMiddlewarePassesValidInput(t *testing.T) {
	vw := &ValidationMiddleware{
		Validate: func(input *StageInput) error {
			if len(input.Data) == 0 {
				return errors.New("empty data")
			}
			return nil
		},
	}
	wrapped := vw.Wrap("test", noopHandler())
	out, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("ok")})
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "ok" {
		t.Errorf("expected 'ok', got %q", string(out.Data))
	}
}

func TestValidationMiddlewareRejectsInvalidInput(t *testing.T) {
	vw := &ValidationMiddleware{
		Validate: func(input *StageInput) error {
			if len(input.Data) == 0 {
				return errors.New("empty data")
			}
			return nil
		},
	}
	wrapped := vw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte{}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidationMiddlewareNilValidator(t *testing.T) {
	vw := &ValidationMiddleware{}
	wrapped := vw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("ok")})
	if err != nil {
		t.Fatalf("expected no error with nil validator, got %v", err)
	}
}

func TestValidationMiddlewareName(t *testing.T) {
	vw := &ValidationMiddleware{}
	if vw.Name() != "validation" {
		t.Errorf("expected 'validation', got %q", vw.Name())
	}
}

// ==================== TransformMiddleware tests ====================

func TestTransformMiddlewareTransformsInput(t *testing.T) {
	tw := &TransformMiddleware{
		TransformInput: func(input *StageInput) (*StageInput, error) {
			input.Data = append(input.Data, []byte("_transformed")...)
			return input, nil
		},
	}
	wrapped := tw.Wrap("test", noopHandler())
	out, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("original")})
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "original_transformed" {
		t.Errorf("expected 'original_transformed', got %q", string(out.Data))
	}
}

func TestTransformMiddlewareTransformsOutput(t *testing.T) {
	tw := &TransformMiddleware{
		TransformOutput: func(output *StageOutput) (*StageOutput, error) {
			output.Data = append(output.Data, []byte("_out")...)
			return output, nil
		},
	}
	wrapped := tw.Wrap("test", noopHandler())
	out, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("data")})
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "data_out" {
		t.Errorf("expected 'data_out', got %q", string(out.Data))
	}
}

func TestTransformMiddlewareBothTransforms(t *testing.T) {
	tw := &TransformMiddleware{
		TransformInput: func(input *StageInput) (*StageInput, error) {
			input.Data = []byte("in")
			return input, nil
		},
		TransformOutput: func(output *StageOutput) (*StageOutput, error) {
			output.Data = append(output.Data, []byte("_out")...)
			return output, nil
		},
	}
	wrapped := tw.Wrap("test", noopHandler())
	out, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("orig")})
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "in_out" {
		t.Errorf("expected 'in_out', got %q", string(out.Data))
	}
}

func TestTransformMiddlewareInputError(t *testing.T) {
	tw := &TransformMiddleware{
		TransformInput: func(input *StageInput) (*StageInput, error) {
			return nil, errors.New("bad input")
		},
	}
	wrapped := tw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("x")})
	if err == nil {
		t.Fatal("expected input transform error")
	}
	if !strings.Contains(err.Error(), "input transform failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTransformMiddlewareOutputError(t *testing.T) {
	tw := &TransformMiddleware{
		TransformOutput: func(output *StageOutput) (*StageOutput, error) {
			return nil, errors.New("bad output")
		},
	}
	wrapped := tw.Wrap("test", noopHandler())
	_, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("x")})
	if err == nil {
		t.Fatal("expected output transform error")
	}
	if !strings.Contains(err.Error(), "output transform failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTransformMiddlewareNilFunctions(t *testing.T) {
	tw := &TransformMiddleware{}
	wrapped := tw.Wrap("test", noopHandler())
	out, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("x")})
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "x" {
		t.Errorf("expected 'x', got %q", string(out.Data))
	}
}

func TestTransformMiddlewareName(t *testing.T) {
	tw := &TransformMiddleware{}
	if tw.Name() != "transform" {
		t.Errorf("expected 'transform', got %q", tw.Name())
	}
}

// ==================== Integration / chaining tests ====================

func TestChainWithRetryAndTimeout(t *testing.T) {
	chain := NewChain()
	var retryCount int32

	chain.Use("stage", &RetryMiddleware{
		MaxAttempts: 3,
		Backoff:     time.Millisecond,
		OnRetry: func(attempt int, err error) {
			atomic.AddInt32(&retryCount, 1)
		},
	})
	chain.Use("stage", &TimeoutMiddleware{Timeout: time.Second})

	calls := 0
	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		calls++
		if calls < 2 {
			return nil, errors.New("transient")
		}
		return &StageOutput{Data: []byte("ok")}, nil
	}

	out, err := chain.Execute("stage", &StageInput{Stage: "stage"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "ok" {
		t.Errorf("expected 'ok', got %q", string(out.Data))
	}
	if atomic.LoadInt32(&retryCount) != 1 {
		t.Errorf("expected 1 retry, got %d", retryCount)
	}
}

func TestChainWithValidationAndTransform(t *testing.T) {
	chain := NewChain()

	chain.Use("stage", &ValidationMiddleware{
		Validate: func(input *StageInput) error {
			if len(input.Data) == 0 {
				return errors.New("empty")
			}
			return nil
		},
	})
	chain.Use("stage", &TransformMiddleware{
		TransformOutput: func(output *StageOutput) (*StageOutput, error) {
			output.Data = append(output.Data, []byte("_done")...)
			return output, nil
		},
	})

	out, err := chain.Execute("stage", &StageInput{Stage: "stage", Data: []byte("hello")}, noopHandler())
	if err != nil {
		t.Fatal(err)
	}
	if string(out.Data) != "hello_done" {
		t.Errorf("expected 'hello_done', got %q", string(out.Data))
	}
}

func TestChainValidationRejectsBeforeTransform(t *testing.T) {
	chain := NewChain()
	transformCalled := false

	chain.Use("stage", &ValidationMiddleware{
		Validate: func(input *StageInput) error {
			return errors.New("reject")
		},
	})
	chain.Use("stage", &TransformMiddleware{
		TransformInput: func(input *StageInput) (*StageInput, error) {
			transformCalled = true
			return input, nil
		},
	})

	_, err := chain.Execute("stage", &StageInput{Stage: "stage", Data: []byte("x")}, noopHandler())
	if err == nil {
		t.Fatal("expected validation error")
	}
	if transformCalled {
		t.Error("transform should not be called after validation failure")
	}
}

func TestChainWithAllMiddlewares(t *testing.T) {
	chain := NewChain()
	var order []string

	chain.Use("stage", &ValidationMiddleware{
		Validate: func(input *StageInput) error {
			order = append(order, "validation")
			return nil
		},
	})
	chain.Use("stage", &LogMiddleware{
		LogFunc: func(msg string, args ...any) {
			if strings.Contains(msg, "start") {
				order = append(order, "log")
			}
		},
	})
	chain.Use("stage", &MetricsMiddleware{
		RecordFunc: func(stage string, duration time.Duration, err error) {
			order = append(order, "metrics")
		},
	})
	chain.Use("stage", &TransformMiddleware{
		TransformInput: func(input *StageInput) (*StageInput, error) {
			order = append(order, "transform")
			return input, nil
		},
	})

	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		order = append(order, "handler")
		return &StageOutput{Data: []byte("ok")}, nil
	}

	_, err := chain.Execute("stage", &StageInput{Stage: "stage", Data: []byte("x")}, handler)
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"validation", "log", "transform", "handler", "metrics"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d]: expected %q, got %q", i, v, order[i])
		}
	}
}
