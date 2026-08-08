package configreload

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

func TestGetStringTypeMismatch(t *testing.T) {
	c := New("")
	c.Set("port", 8080)
	if got := c.GetString("port", "fallback"); got != "fallback" {
		t.Errorf("expected fallback for non-string, got %q", got)
	}
	if got := c.GetString("missing", "fallback"); got != "fallback" {
		t.Errorf("expected fallback for missing key, got %q", got)
	}
	c.Set("name", "naeos")
	if got := c.GetString("name", "fallback"); got != "naeos" {
		t.Errorf("expected 'naeos', got %q", got)
	}
}

func TestGetIntTypeMismatch(t *testing.T) {
	c := New("")
	c.Set("port", 8080)
	if got := c.GetInt("port", -1); got != 8080 {
		t.Errorf("expected 8080, got %d", got)
	}
	c.Set("ratio", 2.5)
	if got := c.GetInt("ratio", -1); got != 2 {
		t.Errorf("expected 2 from float64, got %d", got)
	}
	c.Set("name", "naeos")
	if got := c.GetInt("name", -1); got != -1 {
		t.Errorf("expected default for non-number, got %d", got)
	}
	if got := c.GetInt("missing", -1); got != -1 {
		t.Errorf("expected default for missing key, got %d", got)
	}
}

func TestGetBoolTypeMismatch(t *testing.T) {
	c := New("")
	c.Set("verbose", true)
	if got := c.GetBool("verbose", false); !got {
		t.Error("expected true")
	}
	c.Set("name", "naeos")
	if got := c.GetBool("name", true); !got {
		t.Error("expected default for non-bool")
	}
	if got := c.GetBool("missing", true); !got {
		t.Error("expected default for missing key")
	}
}

func TestLastModified(t *testing.T) {
	c := New("")
	if !c.LastModified().IsZero() {
		t.Error("expected zero time before any mutation")
	}
	c.Set("key", "value")
	if c.LastModified().IsZero() {
		t.Error("expected non-zero time after Set")
	}
	before := c.LastModified()
	time.Sleep(5 * time.Millisecond)
	c.SetAll(map[string]any{"other": 1})
	if !c.LastModified().After(before) {
		t.Error("expected LastModified to advance after SetAll")
	}
}

func TestHotReloaderStartTwice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("name: test\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	c := New(path)
	hr := NewHotReloader(c)

	if err := hr.Start(); err != nil {
		t.Fatalf("first start: %v", err)
	}
	defer hr.Stop()

	err := hr.Start()
	if err == nil {
		t.Fatal("expected error on second Start")
	}
	var ne *naeoserr.NaeosError
	if !errors.As(err, &ne) {
		t.Errorf("expected naeoserr, got %T", err)
	}
}

func TestHotReloaderStartInvalidDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-dir", "config.yaml")
	c := New(path)
	hr := NewHotReloader(c)

	err := hr.Start()
	if err == nil {
		t.Fatal("expected error watching nonexistent directory")
	}
	if hr.IsRunning() {
		t.Error("expected not running after failed start")
	}
}

func TestHotReloaderStopNotRunning(t *testing.T) {
	c := New("")
	hr := NewHotReloader(c)
	hr.Stop() // must not panic
	if hr.IsRunning() {
		t.Error("expected not running")
	}
}

func TestHotReloaderLoopErrorChannel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("name: test\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	c := New(path)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("fsnotify watcher: %v", err)
	}
	defer w.Close()
	if err := w.Add(dir); err != nil {
		t.Fatalf("watch dir: %v", err)
	}

	hr := NewHotReloader(c)
	hr.watcher = w
	hr.stopCh = make(chan struct{})
	go hr.loop()

	w.Errors <- fmt.Errorf("simulated watcher failure")

	// Give the loop time to consume the error, then stop.
	time.Sleep(50 * time.Millisecond)
	close(hr.stopCh)
	time.Sleep(20 * time.Millisecond)
}

func TestHotReloaderLoopIgnoresNonMatchingEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("name: test\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	c := New(path)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("fsnotify watcher: %v", err)
	}
	defer w.Close()
	if err := w.Add(dir); err != nil {
		t.Fatalf("watch dir: %v", err)
	}

	hr := NewHotReloader(c)
	hr.watcher = w
	hr.stopCh = make(chan struct{})
	go hr.loop()

	other := filepath.Join(dir, "unrelated.tmp")
	if err := os.WriteFile(other, []byte("x"), 0o600); err != nil {
		t.Fatalf("write unrelated file: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if c.Version() != 0 {
		t.Errorf("expected no reload for unrelated file, version=%d", c.Version())
	}

	close(hr.stopCh)
	time.Sleep(20 * time.Millisecond)
}

func TestHotReloaderMatchesConfigFile(t *testing.T) {
	c := New("/etc/naeos/config.yaml")
	hr := NewHotReloader(c)

	if !hr.matchesConfigFile("/etc/naeos/config.yaml") {
		t.Error("expected exact path to match")
	}
	if !hr.matchesConfigFile("/etc/naeos/other/config.yaml") {
		t.Error("expected same basename to match")
	}
	if hr.matchesConfigFile("/etc/naeos/config.json") {
		t.Error("expected different basename not to match")
	}
}

func TestHotReloaderLoopClosedEventChannel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	c := New(path)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatalf("fsnotify watcher: %v", err)
	}
	if err := w.Add(dir); err != nil {
		t.Fatalf("watch dir: %v", err)
	}

	hr := NewHotReloader(c)
	hr.watcher = w
	hr.stopCh = make(chan struct{})
	done := make(chan struct{})
	go func() {
		hr.loop()
		close(done)
	}()

	// Closing the watcher closes both the Events and Errors channels,
	// which must terminate loop().
	if err := w.Close(); err != nil {
		t.Fatalf("close watcher: %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not exit after watcher closed")
	}
}
