package database

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"
)

func WithRetry(ctx context.Context, maxRetries int, baseDelay time.Duration, fn func(ctx context.Context) error) error {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := fn(ctx); err != nil {
			lastErr = err
			if !isTransientError(err) {
				return err
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			exp := 1
			if attempt < 30 {
				exp = 1 << uint(attempt)
			}
			delay := baseDelay * time.Duration(exp)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		} else {
			return nil
		}
	}
	return lastErr
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return false
	}

	errStr := strings.ToLower(err.Error())
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"eof",
		"timeout",
		"i/o timeout",
		"connection timed out",
		"no connection",
		"bad connection",
		"invalid connection",
	}
	for _, pattern := range transientPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
