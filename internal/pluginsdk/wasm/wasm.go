// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package wasm

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	pluginapi "github.com/NAEOS-foundation/naeos/pkg/pluginapi/v1"
)

// MaxResponseBytes is the provisional host-side output size limit for a
// single plugin response. The default value is still an open design decision
// and may change before the protocol is declared stable.
const MaxResponseBytes = 10 << 20

type WASMRuntime struct {
	timeout          time.Duration
	maxMemory        int64
	memoryLimitPages uint32
	rt               wazero.Runtime
}

func NewWASMRuntime(timeout time.Duration, maxMemory int64) *WASMRuntime {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if maxMemory == 0 {
		maxMemory = 128 * 1024 * 1024
	}
	ctx := context.Background()
	memoryLimitPages := uint32(maxMemory / 65536)
	if maxMemory%65536 != 0 {
		memoryLimitPages++
	}
	if memoryLimitPages == 0 {
		memoryLimitPages = 1
	}
	if uint64(memoryLimitPages)*65536 > uint64(65536)*65536 {
		memoryLimitPages = 65536
	}
	runtimeConfig := wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithMemoryLimitPages(memoryLimitPages)
	rt := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)
	_, _ = wasi_snapshot_preview1.Instantiate(ctx, rt)

	return &WASMRuntime{
		timeout:          timeout,
		maxMemory:        maxMemory,
		memoryLimitPages: memoryLimitPages,
		rt:               rt,
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
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPlugin, "read wasm file")
	}

	ctx := context.Background()
	compiled, err := w.rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPlugin, "compile wasm module")
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
	result, err := p.ExecuteContext(context.Background(), action, params)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, naeoserr.New(naeoserr.ErrTimeout, fmt.Sprintf("wasm execution timed out after %s", p.wasmRuntime.timeout))
	}
	return result, err
}

// ExecuteContext runs one action until completion or caller/runtime deadline.
// Each invocation uses the public pluginapi v1 protocol envelope: the host
// generates a request ID and requires the module to echo it in exactly one
// JSON response object on stdout.
func (p *WASMPlugin) ExecuteContext(ctx context.Context, action string, params map[string]any) (any, error) {
	if ctx == nil {
		return nil, fmt.Errorf("WASM execution context must not be nil")
	}
	rawParams, err := toRawParams(params)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPlugin, "marshal request params")
	}
	req := pluginapi.Request{
		ProtocolVersion: pluginapi.ProtocolVersion,
		RequestID:       newRequestID(),
		Method:          "execute",
		Action:          action,
		Params:          rawParams,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrPlugin, "marshal request")
	}

	execCtx, cancel := context.WithTimeout(ctx, p.wasmRuntime.timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer

	cfg := wazero.NewModuleConfig().
		WithStdin(bytes.NewReader(reqBytes)).
		WithStdout(&stdout).
		WithStderr(&stderr).
		WithName("").
		WithStartFunctions("_start")

	start := time.Now()
	mod, err := p.wasmRuntime.rt.InstantiateModule(execCtx, p.compiled, cfg)
	elapsed := time.Since(start)

	_ = elapsed

	if mod != nil {
		defer mod.Close(context.Background())
	}

	// Interpret a structured v1 response if the module produced one, even
	// when the module exited non-zero after writing it.
	resp, parseErr := parsePluginResponse(stdout.Bytes(), req)
	if parseErr == nil {
		if resp.OK {
			result, decodeErr := decodeResponseResult(resp)
			if decodeErr != nil {
				return nil, naeoserr.Wrap(naeoserr.ErrPlugin, "invalid plugin response", decodeErr)
			}
			if err == nil {
				return result, nil
			}
		} else {
			return nil, responseFailure(req, resp)
		}
	}

	if err != nil {
		if execCtx.Err() != nil {
			return nil, execCtx.Err()
		}
		return nil, naeoserr.New(naeoserr.ErrPlugin, fmt.Sprintf("wasm execution failed: %v, stderr: %s", err, stderr.String()))
	}
	return nil, naeoserr.Wrap(naeoserr.ErrPlugin, "invalid plugin response", parseErr)
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

// newRequestID generates a fresh host-side request identifier for one
// invocation. It is not derived from plugin state so responses must echo it
// untouched.
func newRequestID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		// Fall back to a non-cryptographic but unique-enough identifier;
		// correlation, not secrecy, is the requirement.
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}

// toRawParams marshals each parameter value separately so the request stays
// a single JSON object with raw JSON parameter bodies.
func toRawParams(params map[string]any) (map[string]json.RawMessage, error) {
	if params == nil {
		return nil, nil
	}
	raw := make(map[string]json.RawMessage, len(params))
	for key, value := range params {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("param %q: %w", key, err)
		}
		raw[key] = encoded
	}
	return raw, nil
}

// parsePluginResponse enforces the v1 response rules: stdout must carry
// exactly one JSON object that echoes the protocol version and request ID,
// and success/error framing must be consistent. Raw untrusted output is
// never included in error messages.
func parsePluginResponse(raw []byte, req pluginapi.Request) (*pluginapi.Response, error) {
	if len(raw) > MaxResponseBytes {
		return nil, fmt.Errorf("plugin response exceeds the %d-byte limit", MaxResponseBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var resp pluginapi.Response
	if err := decoder.Decode(&resp); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, fmt.Errorf("plugin stdout must contain exactly one response object")
		}
		return nil, fmt.Errorf("decode plugin response: %w", err)
	}
	// Only whitespace may remain after the single response object.
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("plugin stdout must contain exactly one response object")
	}
	if err := resp.Validate(req); err != nil {
		return nil, err
	}
	return &resp, nil
}

// decodeResponseResult decodes the raw JSON result body of a successful
// response into a plain Go value.
func decodeResponseResult(resp *pluginapi.Response) (any, error) {
	if resp.Result == nil {
		return nil, nil
	}
	var result any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("decode plugin result: %w", err)
	}
	return result, nil
}

// responseFailure converts a valid failed response into a host error. The
// plugin message is wrapped but not treated as authorization or policy
// evidence.
func responseFailure(req pluginapi.Request, resp *pluginapi.Response) error {
	return naeoserr.New(naeoserr.ErrPlugin, fmt.Sprintf("plugin action %q failed (code=%s): %s", req.Action, resp.Error.Code, resp.Error.Message))
}
