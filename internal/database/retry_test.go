package database

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWithRetrySuccess(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), 3, 10*time.Millisecond, func(ctx context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetrySuccessAfterRetries(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), 3, 10*time.Millisecond, func(ctx context.Context) error {
		calls++
		if calls < 3 {
			return fmt.Errorf("connection refused")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetryMaxRetries(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), 2, 10*time.Millisecond, func(ctx context.Context) error {
		calls++
		return fmt.Errorf("connection refused")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", calls)
	}
}

func TestWithRetryNonTransientError(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), 3, 10*time.Millisecond, func(ctx context.Context) error {
		calls++
		return fmt.Errorf("syntax error")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retries for non-transient), got %d", calls)
	}
}

func TestWithRetryContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := WithRetry(ctx, 5, 10*time.Millisecond, func(ctx context.Context) error {
		calls++
		cancel()
		return fmt.Errorf("connection refused")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWithRetryContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	calls := 0
	err := WithRetry(ctx, 100, 100*time.Millisecond, func(ctx context.Context) error {
		calls++
		return fmt.Errorf("connection refused")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIsTransientError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"connection refused", fmt.Errorf("connection refused"), true},
		{"connection reset", fmt.Errorf("connection reset by peer"), true},
		{"broken pipe", fmt.Errorf("broken pipe"), true},
		{"unexpected EOF", fmt.Errorf("unexpected EOF"), true},
		{"i/o timeout", fmt.Errorf("i/o timeout"), true},
		{"syntax error", fmt.Errorf("syntax error"), false},
		{"table not found", fmt.Errorf("table not found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTransientError(tt.err)
			if got != tt.want {
				t.Errorf("isTransientError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
