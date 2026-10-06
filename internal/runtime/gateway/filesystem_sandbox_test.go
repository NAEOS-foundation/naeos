// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSandboxFilesystemWriteIsObservable(t *testing.T) {
	root := t.TempDir()
	sb := NewDefaultSandbox(SandboxConfig{FilesystemRoot: root})
	req := ToolRequest{
		Tool:    "filesystem",
		Action:  "write",
		Payload: map[string]any{
			"path":    "observed.txt",
			"content": "naeos",
		},
	}

	out, err := sb.Execute(req)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out == "" {
		t.Fatal("expected execution output")
	}

	data, err := os.ReadFile(filepath.Join(root, "observed.txt"))
	if err != nil {
		t.Fatalf("read side effect: %v", err)
	}
	if string(data) != "naeos" {
		t.Fatalf("unexpected content: %q", data)
	}
}

func TestDefaultSandboxFilesystemWriteCannotEscapeRoot(t *testing.T) {
	root := t.TempDir()
	sb := NewDefaultSandbox(SandboxConfig{FilesystemRoot: root})

	_, err := sb.Execute(ToolRequest{
		Tool:   "filesystem",
		Action: "write",
		Payload: map[string]any{"path": "../escape.txt", "content": "no"},
	})
	if err == nil {
		t.Fatal("expected sandbox escape to be denied")
	}
}
