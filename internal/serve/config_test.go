package serve

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Listeners) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(cfg.Listeners))
	}
	if cfg.Listeners[0].Addr != ":8080" {
		t.Fatalf("expected :8080, got %s", cfg.Listeners[0].Addr)
	}
	if !cfg.Listeners[0].API {
		t.Fatal("expected default listener to be an API listener")
	}
	if cfg.ShutdownTimeout != "30s" {
		t.Fatalf("expected shutdown timeout 30s, got %s", cfg.ShutdownTimeout)
	}
}

func TestDefaultConfigValidates(t *testing.T) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf("default config should validate: %v", err)
	}
}

func TestLoadConfigValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.yaml")
	content := `
listeners:
  - addr: ":9443"
    name: api
    api: true
  - addr: "127.0.0.1:9444"
    name: dashboard
shutdown_timeout: "45s"
log_level: "debug"
api_keys:
  test-key: 10
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Listeners) != 2 {
		t.Fatalf("expected 2 listeners, got %d", len(cfg.Listeners))
	}
	if cfg.Listeners[1].Addr != "127.0.0.1:9444" {
		t.Fatalf("unexpected second listener: %s", cfg.Listeners[1].Addr)
	}
	if cfg.Listeners[0].IsTLS() {
		t.Fatal("expected first listener to have TLS disabled")
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("unexpected log level %s", cfg.LogLevel)
	}
	if cfg.APIKeys["test-key"] != 10 {
		t.Fatalf("unexpected api key limit: %d", cfg.APIKeys["test-key"])
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
	if !naeoserr.Is(err, naeoserr.ErrConfig) {
		t.Fatalf("expected ErrConfig code, got %v", err)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("listeners: [not: : valid"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestValidateRejectsEmptyAddr(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Listeners = []Listener{{Addr: "  ", API: true}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty addr")
	}
}

func TestValidateRejectsPartialTLS(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Listeners = []Listener{{Addr: ":8443", TLSCert: "cert.pem"}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for TLS cert without key")
	}
}

func TestValidateRejectsBadLogLevel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.LogLevel = "loud"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for bad log level")
	}
}

func TestListenerIsTLS(t *testing.T) {
	if (Listener{Addr: ":8080"}).IsTLS() {
		t.Fatal("expected IsTLS false with no material")
	}
	if !(Listener{Addr: ":8443", TLSCert: "c", TLSKey: "k"}).IsTLS() {
		t.Fatal("expected IsTLS true with cert+key")
	}
}

func TestParseLogLevel(t *testing.T) {
	if parseLogLevel("debug") != slog.LevelDebug {
		t.Fatal("debug should map to LevelDebug")
	}
	if parseLogLevel("info") != slog.LevelInfo {
		t.Fatal("info should map to LevelInfo")
	}
	if parseLogLevel("unknown") != slog.LevelInfo {
		t.Fatal("unknown should default to info")
	}
}

func TestListenersReturnsCopy(t *testing.T) {
	cfg := DefaultConfig()
	srv, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := srv.Listeners()
	got[0].Addr = ":9999"
	if srv.cfg.Listeners[0].Addr == ":9999" {
		t.Fatal("Listeners should return a copy, not a live reference")
	}
}
