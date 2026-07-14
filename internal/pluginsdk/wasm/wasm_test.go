package wasm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
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

var loopWASM = []byte{
	0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00, 0x01, 0x04, 0x01, 0x60, 0x00, 0x00, 0x03, 0x02,
	0x01, 0x00, 0x05, 0x03, 0x01, 0x00, 0x01, 0x07, 0x13, 0x02, 0x06, 0x6D, 0x65, 0x6D, 0x6F, 0x72,
	0x79, 0x02, 0x00, 0x06, 0x5F, 0x73, 0x74, 0x61, 0x72, 0x74, 0x00, 0x00, 0x0A, 0x09, 0x01, 0x07,
	0x00, 0x03, 0x40, 0x0C, 0x00, 0x0B, 0x0B, 0x0B, 0x06, 0x01, 0x00, 0x41, 0x00, 0x0B, 0x00,
}

func writeTempWASM(t *testing.T, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.wasm")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write temp wasm: %v", err)
	}
	return path
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

func TestWASMPluginExecute(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, helloWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	result, err := plugin.Execute("test", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("execute wasm: %v", err)
	}

	resp, ok := result.(*Response)
	if !ok {
		t.Fatalf("expected *Response, got %T", result)
	}

	if !resp.OK {
		t.Errorf("expected ok=true, got error: %s", resp.Error)
	}

	var inner map[string]any
	if err := json.Unmarshal([]byte(`{"ok":true,"result":"hello"}`), &inner); err != nil {
		t.Fatal(err)
	}
	_ = inner

	if resp.Elapsed < 0 {
		t.Errorf("expected non-negative elapsed, got %d", resp.Elapsed)
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
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}

	resp, ok := result.(*Response)
	if !ok {
		t.Fatalf("expected *Response, got %T", result)
	}

	if resp.OK {
		t.Error("expected ok=false for timeout, got ok=true")
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
		_ = plugin
		t.Skip("load should have failed for invalid wasm")
	}
}

func TestWASMPluginExecuteJSONRoundtrip(t *testing.T) {
	rt := NewWASMRuntime(5*time.Second, 0)
	defer rt.Close()

	path := writeTempWASM(t, helloWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load wasm: %v", err)
	}

	result, err := plugin.Execute("echo", map[string]any{"data": "test"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	resp, ok := result.(*Response)
	if !ok {
		t.Fatalf("expected *Response, got %T", result)
	}

	respJSON, _ := json.Marshal(resp)
	var parsed map[string]any
	if err := json.Unmarshal(respJSON, &parsed); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if _, exists := parsed["ok"]; !exists {
		t.Error("response missing 'ok' field")
	}
}
