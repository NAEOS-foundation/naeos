// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"context"
	"os"
	"os/exec"
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

// buildHelloWASM compiles examples/plugins/hello to a wasip1 module so the
// sandbox wrapper exercises the real pluginapi v1 wire path. Static response
// fixtures cannot echo the host-generated request ID.
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

func TestExecWASMFullSuccess(t *testing.T) {
	wasmPath := buildHelloWASM(t)

	sb := New(Config{Timeout: 60 * time.Second})
	resp, err := sb.Exec(context.Background(), wasmPath, Request{Method: "ping"})
	if err != nil {
		t.Fatalf("exec wasm: %v", err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got error=%s", resp.Error)
	}
	if resp.Result == nil {
		t.Fatal("expected decoded result, got nil")
	}
	if result, ok := resp.Result.(map[string]any); !ok || result["status"] != "ok" {
		t.Errorf("expected ping result, got %#v", resp.Result)
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
		t.Fatal("expected error with canceled context")
	}
}
