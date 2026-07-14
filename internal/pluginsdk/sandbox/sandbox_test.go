package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTestPlugin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "test-plugin")
	content := `#!/usr/bin/env python3
import json, sys
req = json.loads(sys.stdin.read())
resp = {"ok": True, "result": {"method": req["method"]}, "elapsed_ms": 0}
print(json.dumps(resp))
`
	if err := os.WriteFile(pluginPath, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return pluginPath
}

func TestSandboxExec(t *testing.T) {
	pluginPath := createTestPlugin(t)
	sb := New(Config{Timeout: 5 * time.Second})

	resp, err := sb.Exec(context.Background(), pluginPath, Request{
		Method: "test",
		Params: map[string]any{"key": "value"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !resp.OK {
		t.Errorf("expected ok=true, error=%s", resp.Error)
	}
}

func TestSandboxExecTimeout(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "slow-plugin")
	content := `#!/usr/bin/env python3
import time, json, sys
json.loads(sys.stdin.read())
time.sleep(10)
print(json.dumps({"ok": true}))
`
	if err := os.WriteFile(pluginPath, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	sb := New(Config{Timeout: 100 * time.Millisecond})

	_, err := sb.Exec(context.Background(), pluginPath, Request{Method: "slow"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSandboxExecInvalidResponse(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "bad-plugin")
	content := `#!/usr/bin/env python3
import sys
json.loads(sys.stdin.read())
print("not json")
`
	if err := os.WriteFile(pluginPath, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	sb := New(Config{Timeout: 5 * time.Second})

	resp, err := sb.Exec(context.Background(), pluginPath, Request{Method: "test"})
	if err != nil {
		t.Fatal(err)
	}

	if resp.OK {
		t.Error("expected ok=false for invalid response")
	}
}

func TestSandboxExecPluginError(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "error-plugin")
	content := `#!/usr/bin/env python3
import json, sys
json.loads(sys.stdin.read())
print(json.dumps({"ok": False, "error": "something went wrong"}))
`
	if err := os.WriteFile(pluginPath, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	sb := New(Config{Timeout: 5 * time.Second})

	resp, err := sb.Exec(context.Background(), pluginPath, Request{Method: "test"})
	if err != nil {
		t.Fatal(err)
	}

	if resp.OK {
		t.Error("expected ok=false")
	}
	if resp.Error != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", resp.Error)
	}
}

func TestSandboxExecNoPlugin(t *testing.T) {
	sb := New(Config{Timeout: 5 * time.Second})
	_, err := sb.Exec(context.Background(), "/nonexistent/plugin", Request{Method: "test"})
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestResponseJSON(t *testing.T) {
	resp := Response{
		OK:      true,
		Result:  map[string]string{"key": "value"},
		Elapsed: 42,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if !decoded.OK || decoded.Elapsed != 42 {
		t.Errorf("unexpected decoded response: %+v", decoded)
	}
}

func TestSandboxConcurrentExecution(t *testing.T) {
	pluginPath := createTestPlugin(t)
	sb := New(Config{Timeout: 5 * time.Second})

	const goroutines = 5
	errs := make(chan error, goroutines)
	done := make(chan struct{}, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer func() { done <- struct{}{} }()
			resp, err := sb.Exec(context.Background(), pluginPath, Request{
				Method: fmt.Sprintf("concurrent-%d", n),
			})
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %v", n, err)
				return
			}
			if !resp.OK {
				errs <- fmt.Errorf("goroutine %d: expected ok=true, error=%s", n, resp.Error)
				return
			}
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestSandboxExecuteContextCancellation(t *testing.T) {
	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "slow-plugin")
	content := `#!/usr/bin/env python3
import time, json, sys
json.loads(sys.stdin.read())
time.sleep(10)
print(json.dumps({"ok": true}))
`
	if err := os.WriteFile(pluginPath, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}

	sb := New(Config{Timeout: 10 * time.Second})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		resp, err := sb.Exec(ctx, pluginPath, Request{Method: "cancel"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp.OK {
			t.Error("expected ok=false after cancellation")
		}
		done <- struct{}{}
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done
}
