package pluginhost

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SandboxConfig defines security constraints for plugin execution.
type SandboxConfig struct {
	AllowedDirs []string      `json:"allowed_dirs,omitempty"`
	ExecTimeout time.Duration `json:"exec_timeout,omitempty"`
	MaxCalls    int           `json:"max_calls,omitempty"`
}

// Sandbox enforces security constraints on plugin execution.
type Sandbox struct {
	config  SandboxConfig
	mu      sync.Mutex
	callCnt map[string]int
}

// NewSandbox creates a Sandbox with the given config, applying defaults.
func NewSandbox(cfg SandboxConfig) *Sandbox {
	if cfg.ExecTimeout <= 0 {
		cfg.ExecTimeout = 30 * time.Second
	}
	if cfg.MaxCalls <= 0 {
		cfg.MaxCalls = 1000
	}
	return &Sandbox{
		config:  cfg,
		callCnt: make(map[string]int),
	}
}

// ValidatePath checks if a path is within the allowed directories.
func (s *Sandbox) ValidatePath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if len(s.config.AllowedDirs) == 0 {
		return nil
	}
	for _, dir := range s.config.AllowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		if strings.HasPrefix(abs, absDir+string(filepath.Separator)) || abs == absDir {
			return nil
		}
	}
	return fmt.Errorf("plugin path %q is outside allowed directories; add the path to sandbox.allowed_dirs in plugins.json", path)
}

// CheckRateLimit enforces per-plugin call count limits.
func (s *Sandbox) CheckRateLimit(pluginName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callCnt[pluginName]++
	if s.callCnt[pluginName] > s.config.MaxCalls {
		return fmt.Errorf("plugin %q exceeded max call limit (%d); increase sandbox.max_calls in plugins.json or reduce call frequency", pluginName, s.config.MaxCalls)
	}
	return nil
}

// ExecuteWithTimeout runs a function with a timeout enforced via context.
func (s *Sandbox) ExecuteWithTimeout(ctx context.Context, fn func() (any, error)) (any, error) {
	type result struct {
		value any
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := fn()
		ch <- result{v, err}
	}()

	timer := time.NewTimer(s.config.ExecTimeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("plugin execution canceled: %w", ctx.Err())
	case <-timer.C:
		return nil, fmt.Errorf("plugin execution timed out after %s; increase sandbox.exec_timeout in plugins.json or optimize the plugin", s.config.ExecTimeout)
	case r := <-ch:
		return r.value, r.err
	}
}
