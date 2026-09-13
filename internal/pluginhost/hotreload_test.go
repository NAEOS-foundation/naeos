// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package pluginhost

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPluginWatcher(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
	if w.pluginDir != dir {
		t.Errorf("expected dir %s, got %s", dir, w.pluginDir)
	}
	if w.manager != m {
		t.Error("expected manager to be set")
	}
	if w.debounce != 500*time.Millisecond {
		t.Errorf("expected 500ms debounce, got %s", w.debounce)
	}
}

func TestPluginWatcherStartStop(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatalf("start error: %v", err)
	}

	if !w.IsRunning() {
		t.Error("expected running after Start")
	}

	w.Stop()

	if w.IsRunning() {
		t.Error("expected not running after Stop")
	}
}

func TestPluginWatcherDoubleStart(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	if err := w.Start(ctx); err != nil {
		t.Fatal("double start should not error")
	}
}

func TestPluginWatcherDoubleStop(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}

	w.Stop()
	w.Stop()

	if w.IsRunning() {
		t.Error("expected not running after double stop")
	}
}

func TestPluginWatcherMatchesPluginFile(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	tests := []struct {
		path   string
		expect bool
	}{
		{"plugin.so", true},
		{"plugin.wasm", true},
		{"config.json", false},
		{"readme.md", false},
		{"plugin.go", false},
	}

	for _, tt := range tests {
		if got := w.matchesPluginFile(tt.path); got != tt.expect {
			t.Errorf("matchesPluginFile(%q) = %v, want %v", tt.path, got, tt.expect)
		}
	}
}

func TestPluginWatcherDebounceSimulated(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	debounces := 0
	w.debounce = 50 * time.Millisecond

	_ = debounces
}

func TestPluginWatcherContextCancel(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	ctx, cancel := context.WithCancel(context.Background())
	if err := w.Start(ctx); err != nil {
		t.Fatal(err)
	}

	cancel()
	time.Sleep(100 * time.Millisecond)

	if w.IsRunning() {
		t.Error("expected watcher to stop after context cancel")
	}
}

func TestPluginWatcherReloadPluginsEmptyDir(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	w.reloadPlugins()
}

func TestPluginWatcherReloadPluginsWithFiles(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(dir)
	w := NewPluginWatcher(dir, m)

	os.WriteFile(filepath.Join(dir, "plugin.so"), []byte("fake"), 0o644)
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0o644)

	w.reloadPlugins()
}

func TestPluginWatcherStartNonexistentDir(t *testing.T) {
	m := NewManager(t.TempDir())
	w := NewPluginWatcher("/nonexistent/dir", m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := w.Start(ctx)
	if err == nil {
		t.Error("expected error starting with nonexistent dir")
	}
}
