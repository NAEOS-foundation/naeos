// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package wasm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"

	pluginapi "github.com/NAEOS-foundation/naeos/pkg/pluginapi/v1"
)

var helloWASM = []byte{
	0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x01, 0x10, 0x03, 0x60, 0x04, 0x7F, 0x7F, 0x7F,
	0x7F, 0x01, 0x7F, 0x60, 0x01, 0x7F, 0x00, 0x60, 0x00, 0x00, 0x02, 0x46, 0x02, 0x16, 0x77, 0x61,
	0x73, 0x69, 0x5F, 0x73, 0x6E, 0x61, 0x70, 0x73, 0x68, 0x6F, 0x74, 0x5F, 0x70, 0x72, 0x65, 0x76,
	0x69, 0x65, 0x77, 0x31, 0x08, 0x66, 0x64, 0x5F, 0x77, 0x72, 0x69, 0x74, 0x65, 0x00, 0x00, 0x16,
	0x77, 0x61, 0x73, 0x69, 0x5F, 0x73, 0x6E, 0x61, 0x70, 0x73, 0x68, 0x6F, 0x74, 0x5F, 0x70, 0x72,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x31, 0x09, 0x70, 0x72, 0x6F, 0x63, 0x5F, 0x65, 0x78, 0x69, 0x74,
	0x00, 0x01, 0x03, 0x02, 0x01, 0x02, 0x05, 0x03, 0x01, 0x00, 0x01, 0x07, 0x13, 0x02, 0x06, 0x6D,
	0x65, 0x6D, 0x6F, 0x72, 0x79, 0x02, 0x00, 0x06, 0x5F, 0x73, 0x74, 0x61, 0x72, 0x74, 0x00, 0x02,
	0x0A, 0x15, 0x01, 0x13, 0x00, 0x41, 0x01, 0x41, 0x80, 0x02, 0x41, 0x01, 0x41, 0x88, 0x02, 0x10,
	0x00, 0x1A, 0x41, 0x00, 0x10, 0x01, 0x0B, 0x0B, 0x30, 0x02, 0x00, 0x41, 0x00, 0x0B, 0x1C, 0x7B,
	0x22, 0x6F, 0x6B, 0x22, 0x3A, 0x74, 0x72, 0x75, 0x65, 0x2C, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6C,
	0x74, 0x22, 0x3A, 0x22, 0x68, 0x65, 0x6C, 0x6C, 0x6F, 0x22, 0x7D, 0x00, 0x41, 0x80, 0x02, 0x0B,
	0x08, 0x00, 0x00, 0x00, 0x00, 0x1C, 0x00, 0x00, 0x00,
}

// legacyResponseWASM is kept only as a compile/load fixture; its legacy
// unversioned stdout output must be rejected by the v1 response decoder.
var legacyUsageError = "must be rejected"

// prestatWASM imports WASI fd_prestat_get and calls it with fd 3. Because
// NewWASMRuntime does not preopen any filesystem, fd 3 must not be a valid
// preopened directory and WASI must return EBADF (8).
var prestatWASM = []byte{
	0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x01, 0x0B, 0x02, 0x60, 0x02, 0x7F, 0x7F,
	0x01, 0x7F, 0x60, 0x00, 0x01, 0x7F, 0x02, 0x29, 0x01, 0x16, 0x77, 0x61, 0x73, 0x69, 0x5F, 0x73,
	0x6E, 0x61, 0x70, 0x73, 0x68, 0x6F, 0x74, 0x5F, 0x70, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x31, 0x0E,
	0x66, 0x64, 0x5F, 0x70, 0x72, 0x65, 0x73, 0x74, 0x61, 0x74, 0x5F, 0x67, 0x65, 0x74, 0x00, 0x00,
	0x03, 0x02, 0x01, 0x01, 0x07, 0x09, 0x01, 0x05, 0x63, 0x68, 0x65, 0x63, 0x6B, 0x00, 0x01, 0x0A, 0x0A,
	0x01, 0x08, 0x00, 0x41, 0x03, 0x41, 0x00, 0x10, 0x00, 0x0B,
}

var loopWASM = []byte{
	0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x01, 0x04, 0x01, 0x60, 0x00, 0x00, 0x03, 0x02,
	0x01, 0x00, 0x05, 0x03, 0x01, 0x00, 0x01, 0x07, 0x13, 0x02, 0x06, 0x6D, 0x65, 0x6D, 0x6F,
	0x72, 0x79, 0x02, 0x00, 0x06, 0x5F, 0x73, 0x74, 0x61, 0x72, 0x74, 0x00, 0x00, 0x0A, 0x09, 0x01, 0x07,
	0x00, 0x03, 0x40, 0x0C, 0x00, 0x0B, 0x0B, 0x0B, 0x06, 0x01, 0x00, 0x41, 0x00, 0x0B, 0x00,
}

var _ = legacyUsageError

func writeTempWASM(t *testing.T, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.wasm")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write temp wasm: %v", err)
	}
	return path
}

func TestWASMRuntimeHasNoPreopenedFilesystem(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, prestatWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mod, err := rt.rt.InstantiateModule(ctx, plugin.compiled, wazero.NewModuleConfig())
	if err != nil {
		t.Fatalf("instantiate wasm: %v", err)
	}
	defer mod.Close(ctx)

	results, err := mod.ExportedFunction("check").Call(ctx)
	if err != nil {
		t.Fatalf("call fd_prestat_get: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one errno result, got %d", len(results))
	}
	if results[0] != 8 { // WASI EBADF: no preopened directory is available.
		t.Fatalf("expected EBADF (8) for fd_prestat_get on fd 3, got %d", results[0])
	}
}

func TestWASMRuntimeCreation(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 64*1024*1024)
	defer rt.Close()

	if rt == nil {
		t.Fatal("expected non-nil WASMRuntime")
	}
	if rt.timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", rt.timeout)
	}
	if rt.maxMemory != 64*1024*1024 {
		t.Errorf("expected maxMemory 64MB, got %v", rt.maxMemory)
	}
	if rt.memoryLimitPages != 1024 {
		t.Errorf("expected memory limit 1024 pages, got %d", rt.memoryLimitPages)
	}
}

func TestWASMRuntimeDefaults(t *testing.T) {
	rt := NewWASMRuntime(0, 0)
	defer rt.Close()

	if rt.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", rt.timeout)
	}
	if rt.maxMemory != 128*1024*1024 {
		t.Errorf("expected default maxMemory 128MB, got %v", rt.maxMemory)
	}
}

func TestWASMRuntimeLoad(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, helloWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	if plugin.Name() != "test" {
		t.Errorf("expected name 'test', got %q", plugin.Name())
	}
}

func TestWASMRuntimeLoadInvalid(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, []byte{0x00, 0x61, 0x73, 0x6D, 0xFF, 0xFF, 0xFF, 0xFF})
	_, err := rt.Load(path)
	if err == nil {
		t.Fatal("expected error loading invalid wasm")
	}
}

func TestWASMRuntimeLoadMissing(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	_, err := rt.Load("/nonexistent/plugin.wasm")
	if err == nil {
		t.Fatal("expected error loading missing wasm file")
	}
}

func TestWASMPluginExecuteTimeout(t *testing.T) {
	rt := NewWASMRuntime(200*time.Millisecond, 0)
	defer rt.Close()

	path := writeTempWASM(t, loopWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	result, err := plugin.Execute("test", nil)
	if err == nil {
		t.Fatalf("expected timeout error, got result %v", result)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout message, got %v", err)
	}
}

func TestWASMPluginExecuteContextCancellation(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, loopWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := plugin.ExecuteContext(ctx, "test", nil)
		done <- err
	}()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WASM execution did not stop after context cancellation")
	}
}

func TestIsWASMPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"plugin.wasm", true},
		{"path/to/plugin.wasm", true},
		{"plugin.so", false},
		{"plugin.exe", false},
		{"plugin.wasmb", false},
	}

	for _, tt := range tests {
		if got := IsWASMPath(tt.path); got != tt.want {
			t.Errorf("IsWASMPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestWASMPluginVersionDescription(t *testing.T) {
	p := &WASMPlugin{version: "1.0.0", description: "test plugin"}
	if p.Version() != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", p.Version())
	}
	if p.Description() != "test plugin" {
		t.Errorf("expected description %q, got %q", "test plugin", p.Description())
	}
}

func TestWASMPluginExecuteMarshalError(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, helloWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	_, err = plugin.Execute("test", map[string]any{"ch": make(chan int)})
	if err == nil {
		t.Fatal("expected marshal error for unserializable params")
	}
}

func TestWASMPluginInitializeShutdown(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, helloWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	if err := plugin.Initialize(nil); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	if err := plugin.Shutdown(); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestWASMPluginExecuteInvalidWASM(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, []byte{0xDE, 0xAD, 0xBE, 0xEF})
	plugin, err := rt.Load(path)
	if err == nil {
		_, err = plugin.Execute("", nil)
	}
	if err == nil {
		t.Fatal("expected error for invalid wasm")
	}
}

func TestNewRequestIDUnique(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 100; i++ {
		id := newRequestID()
		if id == "" {
			t.Fatal("expected non-empty request ID")
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate request ID %q", id)
		}
		seen[id] = struct{}{}
	}
}

func validRequest() pluginapi.Request {
	return pluginapi.Request{
		ProtocolVersion: pluginapi.ProtocolVersion,
		RequestID:       "8c50f92c",
		Method:          "execute",
		Action:          "ping",
	}
}

func TestParsePluginResponseParsesSuccess(t *testing.T) {
	req := validRequest()
	raw := []byte(`{"protocol_version":"1.0","request_id":"8c50f92c","ok":true,"result":{"status":"ok"}}`)

	resp, err := parsePluginResponse(raw, req)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !resp.OK {
		t.Fatal("expected successful response")
	}
	result, err := decodeResponseResult(resp)
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}
	decoded, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected decoded map, got %T", result)
	}
	if decoded["status"] != "ok" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestParsePluginResponseAcceptsTrailingWhitespaceOnly(t *testing.T) {
	req := validRequest()
	raw := []byte(`{"protocol_version":"1.0","request_id":"8c50f92c","ok":true,"result":null}` + "\n\t \n")

	resp, err := parsePluginResponse(raw, req)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result, err := decodeResponseResult(resp)
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result for missing result field, got %v", result)
	}
}

func TestParsePluginResponseParsesPluginError(t *testing.T) {
	req := validRequest()
	raw := []byte(`{"protocol_version":"1.0","request_id":"8c50f92c","ok":false,"error":{"code":"invalid_params","message":"name is required"}}`)

	resp, err := parsePluginResponse(raw, req)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.OK {
		t.Fatal("expected failed response")
	}
	failure := responseFailure(req, resp)
	if !strings.Contains(failure.Error(), "code=invalid_params") {
		t.Fatalf("expected error code in message, got %v", failure)
	}
	if !strings.Contains(failure.Error(), "name is required") {
		t.Fatalf("expected plugin message in error, got %v", failure)
	}
}

func TestParsePluginResponseRejects(t *testing.T) {
	req := validRequest()
	responses := map[string]string{
		"legacy unversioned response": `{"ok":true,"result":"hello"}`,
		"wrong protocol version":      `{"protocol_version":"9.9","request_id":"8c50f92c","ok":true,"result":null}`,
		"wrong request ID":            `{"protocol_version":"1.0","request_id":"00000000","ok":true,"result":null}`,
		"ok with error":               `{"protocol_version":"1.0","request_id":"8c50f92c","ok":true,"error":{"code":"invalid_params","message":"x"}}`,
		"failure without code":        `{"protocol_version":"1.0","request_id":"8c50f92c","ok":false,"error":{"code":"","message":"x"}}`,
		"failure without error":       `{"protocol_version":"1.0","request_id":"8c50f92c","ok":false}`,
		"undocumented error code":     `{"protocol_version":"1.0","request_id":"8c50f92c","ok":false,"error":{"code":"banana","message":"x"}}`,
		"malformed json":              `{"protocol_version":"1.0",`,
		"array response":              `[{"ok":true}]`,
		"multiple objects":            `{"protocol_version":"1.0","request_id":"8c50f92c","ok":true,"result":null} {"ok":true}`,
		"trailing garbage":            `{"protocol_version":"1.0","request_id":"8c50f92c","ok":true,"result":null} garbage`,
		"empty output":                ``,
	}

	for name, raw := range responses {
		resp, err := parsePluginResponse([]byte(raw), req)
		if resp != nil {
			t.Errorf("%s: expected nil response, got %v", name, resp)
		}
		if err == nil {
			t.Errorf("%s: expected parse rejection", name)
		}
	}

	if _, err := parsePluginResponse(make([]byte, MaxResponseBytes+1), req); err == nil || !strings.Contains(err.Error(), "limit") {
		t.Errorf("expected size-limit rejection, got %v", err)
	}
}

func TestValidErrorCode(t *testing.T) {
	for code, want := range map[string]bool{
		pluginapi.ErrorCodeInvalidRequest:     true,
		pluginapi.ErrorCodeUnsupportedVersion: true,
		pluginapi.ErrorCodePluginError:        true,
		pluginapi.ErrorCodeInvalidResponse:    true,
		"banana":                              false,
		"":                                    false,
	} {
		if got := pluginapi.ValidErrorCode(code); got != want {
			t.Errorf("ValidErrorCode(%q) = %v, want %v", code, got, want)
		}
	}
}

// buildHelloWASM compiles examples/plugins/hello to a wasip1 module using the
// local Go toolchain. It exercises the real v1 host-side wire path instead of
// a static response fixture, because valid responses must echo the
// host-generated request ID.
func buildHelloWASM(t *testing.T) string {
	t.Helper()
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skipf("go toolchain unavailable: %v", err)
	}
	out := filepath.Join(t.TempDir(), "hello.wasm")
	buildDir, err := filepath.Abs("../../../examples/plugins/hello")
	if err != nil {
		t.Fatalf("resolve example dir: %v", err)
	}
	cmd := exec.Command(goBin, "build", "-o", out, ".")
	cmd.Dir = buildDir
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build example plugin: %v\n%s", err, output)
	}
	return out
}

func TestWASMPluginV1EnvelopeRoundTrip(t *testing.T) {
	rt := NewWASMRuntime(30*time.Second, 0)
	defer rt.Close()

	path := buildHelloWASM(t)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	result, err := plugin.Execute("ping", nil)
	if err != nil {
		t.Fatalf("execute ping: %v", err)
	}
	decoded, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected decoded map, got %T", result)
	}
	if decoded["status"] != "ok" {
		t.Fatalf("unexpected ping result: %v", result)
	}

	greet, err := plugin.Execute("greet", map[string]any{"name": "NAEOS"})
	if err != nil {
		t.Fatalf("execute greet: %v", err)
	}
	greetMap, ok := greet.(map[string]any)
	if !ok {
		t.Fatalf("expected decoded map, got %T", greet)
	}
	if greetMap["message"] != "Hello, NAEOS!" {
		t.Fatalf("unexpected greet result: %v", greet)
	}
}

func TestWASMPluginV1EnvelopeErrorMapping(t *testing.T) {
	rt := NewWASMRuntime(30*time.Second, 0)
	defer rt.Close()

	plugin, err := rt.Load(buildHelloWASM(t))
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	result, err := plugin.Execute("nope", nil)
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
	if !strings.Contains(err.Error(), "code=plugin_error") {
		t.Fatalf("expected plugin error code, got %v", err)
	}
	if !strings.Contains(err.Error(), "unknown action") {
		t.Fatalf("expected plugin message, got %v", err)
	}
}

func TestDecodePreservesJSONNumbers(t *testing.T) {
	req := validRequest()
	raw := []byte(`{"protocol_version":"1.0","request_id":"8c50f92c","ok":true,"result":{"count":7,"ok_flag":true}}`)

	resp, err := parsePluginResponse(raw, req)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	result, err := decodeResponseResult(resp)
	if err != nil {
		t.Fatalf("decode result: %v", err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("re-encode: %v", err)
	}
	if !bytes.Contains(encoded, []byte(`"count":7`)) {
		t.Fatalf("expected exact JSON number 7, got %s", encoded)
	}
}
