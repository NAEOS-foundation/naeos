// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package observability

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveInputFile(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(existing, []byte("project: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := resolveInputFile("", ""); err == nil {
		t.Error("expected error when no flags given")
	}

	if _, err := resolveInputFile("", filepath.Join(dir, "missing.yaml")); err == nil {
		t.Error("expected error for missing input file")
	}

	path, err := resolveInputFile("", existing)
	if err != nil || path != existing {
		t.Errorf("expected %q, got %q (err %v)", existing, path, err)
	}

	path, err = resolveInputFile(filepath.Join(dir, "ghost.yaml"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path for missing --input, got %q", path)
	}

	path, err = resolveInputFile(existing, "")
	if err != nil || path != existing {
		t.Errorf("expected %q from --input, got %q (err %v)", existing, path, err)
	}
}
