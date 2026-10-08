// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package wasm

import (
	"context"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"
)

// TestUntrustedAdapterWASMContainmentProfile verifies the minimum runtime
// posture required when an external adapter is treated as untrusted:
// bounded memory, bounded execution, and no preopened filesystem.
func TestUntrustedAdapterWASMContainmentProfile(t *testing.T) {
	const (
		timeout = 250 * time.Millisecond
		memory  = 64 * 1024 * 1024
	)

	rt := NewWASMRuntime(timeout, memory)
	defer rt.Close()

	if rt.timeout != timeout {
		t.Fatalf("expected bounded execution timeout %v, got %v", timeout, rt.timeout)
	}
	if rt.memoryLimitPages == 0 || uint64(rt.memoryLimitPages)*65536 > uint64(memory) {
		t.Fatalf("expected bounded WASM memory at or below %d bytes, got %d pages", memory, rt.memoryLimitPages)
	}

	path := writeTempWASM(t, prestatWASM)
	plugin, err := rt.Load(path)
	if err != nil {
		t.Fatalf("load containment fixture: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mod, err := rt.rt.InstantiateModule(ctx, plugin.compiled, wazero.NewModuleConfig())
	if err != nil {
		t.Fatalf("instantiate containment fixture: %v", err)
	}
	defer mod.Close(ctx)

	results, err := mod.ExportedFunction("check").Call(ctx)
	if err != nil {
		t.Fatalf("probe WASI preopen boundary: %v", err)
	}
	if len(results) != 1 || results[0] != 8 { // WASI EBADF: fd 3 is not preopened.
		t.Fatalf("untrusted adapter received a preopened filesystem: errno=%v", results)
	}
}
