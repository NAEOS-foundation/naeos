package gateway

import (
	"fmt"
	"strings"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Restriction defines a policy-based command restriction. When a request
// matches a restriction it is always denied, regardless of the control
// plane decision.
type Restriction struct {
	// Tool is the tool name pattern to restrict. Empty matches all tools.
	Tool string
	// Action is the action pattern to restrict. Empty matches all actions.
	Action string
	// Environment is the environment to restrict. Empty matches all.
	Environment string
	// Reason explains why the restriction exists.
	Reason string
}

// Matches reports whether the given request matches this restriction.
// Empty fields act as wildcards.
func (r Restriction) Matches(req ToolRequest) bool {
	if r.Tool != "" && !matchPattern(r.Tool, req.Tool) {
		return false
	}
	if r.Action != "" && !matchPattern(r.Action, req.Action) {
		return false
	}
	if r.Environment != "" && !matchPattern(r.Environment, req.Environment) {
		return false
	}
	return true
}

func matchPattern(pattern, value string) bool {
	if pattern == value {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(value, strings.TrimSuffix(pattern, "*"))
	}
	return false
}

// SandboxConfig configures the DefaultSandbox.
type SandboxConfig struct {
	MaxOutputBytes int
	Timeout        time.Duration
	AllowedTools   []string
	BlockedTools   []string
}

// DefaultSandbox is a baseline sandbox implementation that validates
// requests and captures output. In production, this would be replaced
// with process-level or container-level isolation.
type DefaultSandbox struct {
	mu     sync.RWMutex
	config SandboxConfig
	count  int
}

// NewDefaultSandbox creates a DefaultSandbox with the given config.
func NewDefaultSandbox(cfg SandboxConfig) *DefaultSandbox {
	if cfg.MaxOutputBytes == 0 {
		cfg.MaxOutputBytes = 1024 * 1024 // 1 MiB default
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &DefaultSandbox{config: cfg}
}

// Execute runs the tool request inside the sandbox and returns the output.
func (s *DefaultSandbox) Execute(req ToolRequest) (string, error) {
	if req.Tool == "" {
		return "", naeoserr.New(naeoserr.ErrValidation, "tool name must not be empty")
	}

	s.mu.Lock()
	s.count++
	s.mu.Unlock()

	// Check blocked tools.
	for _, blocked := range s.config.BlockedTools {
		if matchPattern(blocked, req.Tool) {
			return "", naeoserr.New(naeoserr.ErrPermDenied, fmt.Sprintf("tool %s is blocked by sandbox policy", req.Tool))
		}
	}

	// Check allowed tools (empty list = all allowed).
	if len(s.config.AllowedTools) > 0 {
		allowed := false
		for _, a := range s.config.AllowedTools {
			if matchPattern(a, req.Tool) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", naeoserr.New(naeoserr.ErrPermDenied, fmt.Sprintf("tool %s is not in the allowed list", req.Tool))
		}
	}

	// Build a deterministic output that reflects the execution context.
	output := fmt.Sprintf("executed %s/%s on %s", req.Tool, req.Action, req.Resource)
	if len(req.Payload) > 0 {
		output += fmt.Sprintf(" [%d payload keys]", len(req.Payload))
	}
	return output, nil
}

// ExecutedCount returns the number of executions.
func (s *DefaultSandbox) ExecutedCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}
