// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorQuickFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("pipeline:\n  name: quick-test\n  mode: development\n  output_dir: ./out\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "doctor", "--config", configPath, "--quick")
	if err != nil {
		t.Fatalf("execute doctor --quick failed: %v", err)
	}

	if !strings.Contains(output, "NAEOS Doctor") {
		t.Fatalf("expected doctor header, got %q", output)
	}
}
