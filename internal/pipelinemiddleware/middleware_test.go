package middleware

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

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
		t.Errorf("parse middleware should not be called for generate stage, callCount went from %d to %d", callsBefore, len(parseMWCalled))
	}
}

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

	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		return &StageOutput{Data: []byte("ok")}, nil
	}

	wrapped := mw.Wrap("test", handler)

	_, err := wrapped(context.Background(), &StageInput{
		Stage:  "test",
		Labels: map[string]string{"auth_token": "valid"},
	})
	if err != nil {
		t.Errorf("expected no error for valid token, got %v", err)
	}

	_, err = wrapped(context.Background(), &StageInput{
		Stage:  "test",
		Labels: map[string]string{"auth_token": "bad"},
	})
	if err == nil {
		t.Error("expected error for invalid token")
	}

	_, err = wrapped(context.Background(), &StageInput{
		Stage:  "test",
		Labels: map[string]string{},
	})
	if err == nil {
		t.Error("expected error for missing token")
	}
}

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

	_, err := wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("key1")})
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	_, err = wrapped(context.Background(), &StageInput{Stage: "test", Data: []byte("key1")})
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Errorf("expected cache hit (1 call), got %d", callCount)
	}
}

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

	handler := func(ctx context.Context, input *StageInput) (*StageOutput, error) {
		return &StageOutput{Data: []byte("ok")}, nil
	}

	wrapped := mw.Wrap("test", handler)
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
