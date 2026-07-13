package middleware

import (
	"context"
	"fmt"
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
	mws := c.middlewares[stage]
	current := handler
	for i := len(mws) - 1; i >= 0; i-- {
		mw := mws[i]
		nextFn := current
		current = func(ctx context.Context, in *StageInput) (*StageOutput, error) {
			return mw.Wrap(in.Stage, nextFn)(ctx, in)
		}
	}
	return current(context.Background(), input)
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
