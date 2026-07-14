package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/pluginsdk/wasm"
)

type Sandbox struct {
	timeout    time.Duration
	maxMemory  int64
	allowedEnv []string
	mu         sync.Mutex
}

type Config struct {
	Timeout    time.Duration
	MaxMemory  int64
	AllowedEnv []string
}

type Request struct {
	Method string                 `json:"method"`
	Params map[string]any `json:"params"`
}

type Response struct {
	OK      bool        `json:"ok"`
	Result  any `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
	Elapsed int64       `json:"elapsed_ms"`
}

func New(cfg Config) *Sandbox {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxMemory == 0 {
		cfg.MaxMemory = 128 * 1024 * 1024
	}
	return &Sandbox{
		timeout:    cfg.Timeout,
		maxMemory:  cfg.MaxMemory,
		allowedEnv: cfg.AllowedEnv,
	}
}

func (s *Sandbox) Exec(ctx context.Context, pluginPath string, req Request) (*Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if wasm.IsWASMPath(pluginPath) {
		return s.execWASM(ctx, pluginPath, req)
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cmd := exec.CommandContext(reqCtx, pluginPath)
	cmd.Stdin = bytes.NewReader(data)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = s.buildEnv()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start plugin: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-reqCtx.Done():
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return &Response{
			OK:    false,
			Error: fmt.Sprintf("plugin execution timed out after %s", s.timeout),
		}, nil
	case err := <-done:
		if err != nil {
			return &Response{
				OK:    false,
				Error: fmt.Sprintf("plugin error: %v, stderr: %s", err, stderr.String()),
			}, nil
		}
	}

	var resp Response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return &Response{
			OK:    false,
			Error: fmt.Sprintf("invalid plugin response: %v, raw: %s", err, stdout.String()),
		}, nil
	}

	return &resp, nil
}

func (s *Sandbox) execWASM(ctx context.Context, wasmPath string, req Request) (*Response, error) {
	rt := wasm.NewWASMRuntime(s.timeout, 0)
	defer rt.Close()

	plugin, err := rt.Load(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("load wasm plugin: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	type result struct {
		value any
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := plugin.Execute(req.Method, req.Params)
		ch <- result{v, err}
	}()

	select {
	case <-reqCtx.Done():
		return &Response{
			OK:    false,
			Error: fmt.Sprintf("wasm plugin execution timed out after %s", s.timeout),
		}, nil
	case r := <-ch:
		if r.err != nil {
			return &Response{
				OK:    false,
				Error: fmt.Sprintf("wasm plugin error: %v", r.err),
			}, nil
		}
		if wasmResp, ok := r.value.(*wasm.Response); ok {
			return &Response{
				OK:      wasmResp.OK,
				Result:  wasmResp.Result,
				Error:   wasmResp.Error,
				Elapsed: wasmResp.Elapsed,
			}, nil
		}
		return &Response{
			OK:     true,
			Result: r.value,
		}, nil
	}
}

func (s *Sandbox) ExecWASM(ctx context.Context, wasmPath string, req Request) (*Response, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "naeos-wasm-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input.json")
	outputPath := filepath.Join(tmpDir, "output.json")

	if err := os.WriteFile(inputPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("write input: %w", err)
	}

	cmd := exec.CommandContext(reqCtx, "wasmtime", wasmPath,
		"--dir", tmpDir,
		"--env", fmt.Sprintf("INPUT=%s", inputPath),
		"--env", fmt.Sprintf("OUTPUT=%s", outputPath),
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wasm execution: %w, stderr: %s", err, stderr.String())
	}

	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("read output: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(outputData, &resp); err != nil {
		return nil, fmt.Errorf("parse output: %w", err)
	}

	return &resp, nil
}

func (s *Sandbox) buildEnv() []string {
	env := []string{
		"NAEOS_SANDBOX=1",
		fmt.Sprintf("NAEOS_TIMEOUT=%s", s.timeout),
	}
	env = append(env, os.Environ()...)
	return env
}
