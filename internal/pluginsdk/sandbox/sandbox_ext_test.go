package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewDefaults(t *testing.T) {
	sb := New(Config{})
	if sb.timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", sb.timeout)
	}
	if sb.maxMemory != 128*1024*1024 {
		t.Errorf("expected 128MB maxMemory, got %d", sb.maxMemory)
	}
	if sb.allowedEnv != nil {
		t.Errorf("expected nil allowedEnv, got %v", sb.allowedEnv)
	}
}

func TestNewCustomConfig(t *testing.T) {
	sb := New(Config{
		Timeout:    5 * time.Second,
		MaxMemory:  64 * 1024 * 1024,
		AllowedEnv: []string{"PATH", "HOME"},
	})
	if sb.timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", sb.timeout)
	}
	if sb.maxMemory != 64*1024*1024 {
		t.Errorf("expected 64MB maxMemory, got %d", sb.maxMemory)
	}
	if len(sb.allowedEnv) != 2 {
		t.Errorf("expected 2 allowedEnv, got %d", len(sb.allowedEnv))
	}
}

func TestBuildEnv(t *testing.T) {
	sb := New(Config{Timeout: 10 * time.Second})
	env := sb.buildEnv()

	foundSandbox := false
	foundTimeout := false
	for _, e := range env {
		if e == "NAEOS_SANDBOX=1" {
			foundSandbox = true
		}
		if strings.HasPrefix(e, "NAEOS_TIMEOUT=") {
			foundTimeout = true
		}
	}
	if !foundSandbox {
		t.Error("expected NAEOS_SANDBOX=1 in env")
	}
	if !foundTimeout {
		t.Error("expected NAEOS_TIMEOUT in env")
	}
}

func TestExecWithWASMPathLoadFailure(t *testing.T) {
	dir := t.TempDir()
	bogusWasm := filepath.Join(dir, "plugin.wasm")
	os.WriteFile(bogusWasm, []byte("not a valid wasm binary"), 0o644)

	sb := New(Config{Timeout: 5 * time.Second})
	_, err := sb.Exec(context.Background(), bogusWasm, Request{Method: "test"})
	if err == nil {
		t.Fatal("expected error for bogus wasm file")
	}
}

func TestExecWithWASMExtensionNotFound(t *testing.T) {
	sb := New(Config{Timeout: 5 * time.Second})
	_, err := sb.Exec(context.Background(), "/nonexistent/plugin.wasm", Request{Method: "test"})
	if err == nil {
		t.Fatal("expected error for nonexistent wasm plugin")
	}
}

func TestExecWithNonWASMBinaryNotFound(t *testing.T) {
	sb := New(Config{Timeout: 5 * time.Second})
	_, err := sb.Exec(context.Background(), "/nonexistent/binary", Request{Method: "test"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecWASMNotInstalled(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "plugin.wasm")
	os.WriteFile(pluginPath, []byte("dummy"), 0o644)

	sb := New(Config{Timeout: 5 * time.Second})
	_, err := sb.ExecWASM(context.Background(), pluginPath, Request{Method: "test"})
	if err == nil {
		t.Fatal("expected error (wasmtime not available)")
	}
}

func TestExecWASMBinaryNotFound(t *testing.T) {
	sb := New(Config{Timeout: 5 * time.Second})
	_, err := sb.ExecWASM(context.Background(), "/nonexistent/plugin.wasm", Request{Method: "test"})
	if err == nil {
		t.Fatal("expected error (wasmtime not available)")
	}
}

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

func TestExecWASMFullSuccess(t *testing.T) {
	dir := t.TempDir()
	wasmPath := filepath.Join(dir, "test.wasm")
	if err := os.WriteFile(wasmPath, helloWASM, 0o644); err != nil {
		t.Fatal(err)
	}

	sb := New(Config{Timeout: 5 * time.Second})
	resp, err := sb.Exec(context.Background(), wasmPath, Request{Method: "test"})
	if err != nil {
		t.Fatalf("exec wasm: %v", err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got error=%s", resp.Error)
	}
	if resp.Result != "hello" {
		t.Errorf("expected result 'hello', got %v", resp.Result)
	}
}

func TestExecWASMContextCancel(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "plugin.wasm")
	os.WriteFile(pluginPath, []byte("dummy"), 0o644)

	sb := New(Config{Timeout: 10 * time.Second})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := sb.ExecWASM(ctx, pluginPath, Request{Method: "test"})
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}
