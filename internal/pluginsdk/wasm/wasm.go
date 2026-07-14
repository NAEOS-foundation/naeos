package wasm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

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

type WASMRuntime struct {
	timeout   time.Duration
	maxMemory int64
	rt        wazero.Runtime
}

func NewWASMRuntime(timeout time.Duration, maxMemory int64) *WASMRuntime {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if maxMemory == 0 {
		maxMemory = 128 * 1024 * 1024
	}
	ctx := context.Background()
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCloseOnContextDone(true))
	wasi_snapshot_preview1.Instantiate(ctx, rt)

	return &WASMRuntime{
		timeout:   timeout,
		maxMemory: maxMemory,
		rt:        rt,
	}
}

type WASMPlugin struct {
	wasmRuntime *WASMRuntime
	wasmPath    string
	name        string
	version     string
	description string
	compiled    wazero.CompiledModule
}

func (w *WASMRuntime) Load(wasmPath string) (*WASMPlugin, error) {
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("read wasm file: %w", err)
	}

	ctx := context.Background()
	compiled, err := w.rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile wasm module: %w", err)
	}

	name := filepath.Base(wasmPath)
	name = name[:len(name)-len(filepath.Ext(name))]

	return &WASMPlugin{
		wasmRuntime: w,
		wasmPath:    wasmPath,
		name:        name,
		compiled:    compiled,
	}, nil
}

func (p *WASMPlugin) Name() string        { return p.name }
func (p *WASMPlugin) Version() string     { return p.version }
func (p *WASMPlugin) Description() string { return p.description }

func (p *WASMPlugin) Initialize(_ any) error {
	return nil
}

func (p *WASMPlugin) Execute(action string, params map[string]any) (any, error) {
	req := Request{
		Method: action,
		Params: params,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.wasmRuntime.timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer

	cfg := wazero.NewModuleConfig().
		WithStdin(bytes.NewReader(reqBytes)).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithName("").
		WithStartFunctions("_start")

	start := time.Now()
	mod, err := p.wasmRuntime.rt.InstantiateModule(ctx, p.compiled, cfg)
	elapsed := time.Since(start)

	if err != nil {
		return &Response{
			OK:      false,
			Error:   fmt.Sprintf("wasm execution failed: %v, stderr: %s", err, stderr.String()),
			Elapsed: elapsed.Milliseconds(),
		}, nil
	}
	defer mod.Close(ctx)

	var resp Response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return &Response{
			OK:      false,
			Error:   fmt.Sprintf("invalid wasm response: %v, raw: %s", err, stdout.String()),
			Elapsed: elapsed.Milliseconds(),
		}, nil
	}
	resp.Elapsed = elapsed.Milliseconds()

	return &resp, nil
}

func (p *WASMPlugin) Shutdown() error {
	return nil
}

func (w *WASMRuntime) Close() error {
	return w.rt.Close(context.Background())
}

func IsWASMPath(path string) bool {
	return filepath.Ext(path) == ".wasm"
}
