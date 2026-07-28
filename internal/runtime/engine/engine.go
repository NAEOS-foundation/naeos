package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Artifact struct {
	Path    string
	Content []byte
}

type ExecutionResult struct {
	Artifact Artifact
	Status   string
	Output   string
	Error    error
}

type Engine struct {
	mu        sync.RWMutex
	history   []ExecutionResult
	executed  map[string]bool
	outputDir string
}

func NewEngine() *Engine {
	return &Engine{
		executed: make(map[string]bool),
	}
}

func (e *Engine) SetOutputDir(dir string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.outputDir = dir
}

func (e *Engine) Run(artifact any) error {
	if artifact == nil {
		return naeoserr.New(naeoserr.ErrValidation, "artifact is nil")
	}
	return nil
}

func (e *Engine) Execute(artifact Artifact) (*ExecutionResult, error) {
	if artifact.Path == "" {
		return nil, naeoserr.New(naeoserr.ErrValidation, "artifact path must not be empty")
	}

	if err := e.Validate(artifact); err != nil {
		return nil, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	result := &ExecutionResult{
		Artifact: artifact,
		Status:   "completed",
	}

	if e.executed[artifact.Path] {
		result.Status = "skipped"
		result.Output = "already executed"
		e.history = append(e.history, *result)
		return result, nil
	}

	if e.outputDir != "" {
		fullPath := filepath.Join(e.outputDir, artifact.Path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			result.Status = "failed"
			result.Error = naeoserr.Wrapf(err, naeoserr.ErrNetwork, "create directory %s", dir)
			e.history = append(e.history, *result)
			return result, result.Error
		}
		if err := os.WriteFile(fullPath, artifact.Content, 0o600); err != nil {
			result.Status = "failed"
			result.Error = naeoserr.Wrapf(err, naeoserr.ErrNetwork, "write file %s", fullPath)
			e.history = append(e.history, *result)
			return result, result.Error
		}
		result.Output = fmt.Sprintf("wrote %s (%d bytes)", fullPath, len(artifact.Content))
	} else {
		result.Output = fmt.Sprintf("executed %s (%d bytes)", artifact.Path, len(artifact.Content))
	}

	e.executed[artifact.Path] = true
	e.history = append(e.history, *result)
	return result, nil
}

func (e *Engine) ExecuteAll(artifacts []Artifact) ([]ExecutionResult, error) {
	if len(artifacts) == 0 {
		return nil, naeoserr.New(naeoserr.ErrValidation, "no artifacts to execute")
	}

	var results []ExecutionResult
	for _, artifact := range artifacts {
		result, err := e.Execute(artifact)
		if err != nil {
			return results, naeoserr.Wrapf(err, naeoserr.ErrPipeline, "failed to execute %s", artifact.Path)
		}
		results = append(results, *result)
	}
	return results, nil
}

func (e *Engine) Validate(artifact Artifact) error {
	if artifact.Path == "" {
		return naeoserr.New(naeoserr.ErrValidation, "artifact path must not be empty")
	}

	ext := filepath.Ext(artifact.Path)
	switch ext {
	case ".go":
		if len(artifact.Content) == 0 {
			return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("go file %s has no content", artifact.Path))
		}
		content := string(artifact.Content)
		if !strings.Contains(content, "package ") {
			return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("go file %s missing package declaration", artifact.Path))
		}
	case ".yaml", ".yml":
		if len(artifact.Content) == 0 {
			return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("yaml file %s has no content", artifact.Path))
		}
	case ".md":
		if len(artifact.Content) == 0 {
			return naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("markdown file %s has no content", artifact.Path))
		}
	}

	return nil
}

func (e *Engine) History() []ExecutionResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]ExecutionResult, len(e.history))
	copy(result, e.history)
	return result
}

func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.history = nil
	e.executed = make(map[string]bool)
}

func (e *Engine) ExecutedCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.executed)
}

func (e *Engine) FailedCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	count := 0
	for _, r := range e.history {
		if r.Status == "failed" {
			count++
		}
	}
	return count
}
