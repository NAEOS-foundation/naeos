// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyPluginValidChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin.so")
	content := []byte("fake plugin binary data")
	os.WriteFile(path, content, 0o644)

	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	if err := VerifyPlugin(path, expected); err != nil {
		t.Errorf("expected valid checksum to pass, got error: %v", err)
	}
}

func TestVerifyPluginInvalidChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin.so")
	content := []byte("fake plugin binary data")
	os.WriteFile(path, content, 0o644)

	if err := VerifyPlugin(path, "0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Error("expected error for invalid checksum")
	}
}

func TestGenerateChecksum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin.so")
	content := []byte("test content for checksum")
	os.WriteFile(path, content, 0o644)

	checksum, err := GenerateChecksum(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	if checksum != expected {
		t.Errorf("expected %s, got %s", expected, checksum)
	}
}

func TestVerifyPluginMissingFile(t *testing.T) {
	if err := VerifyPlugin("/nonexistent/path/plugin.so", "abc123"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestGenerateChecksumMissingFile(t *testing.T) {
	_, err := GenerateChecksum("/nonexistent/path/plugin.so")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
