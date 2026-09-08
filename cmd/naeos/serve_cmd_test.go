package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeCommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	out, err := executeCommand(root, "serve", "--help")
	if err != nil {
		t.Fatalf("serve --help failed: %v", err)
	}
	if !strings.Contains(out, "production daemon") {
		t.Fatalf("expected serve help text, got %q", out)
	}
}

func TestServeRunSubcommandShowsHelp(t *testing.T) {
	root := NewRootCommand()
	out, err := executeCommand(root, "serve", "run", "--help")
	if err != nil {
		t.Fatalf("serve run --help failed: %v", err)
	}
	if !strings.Contains(out, "foreground") {
		t.Fatalf("expected serve run help text, got %q", out)
	}
}

func TestServeConfigPrintsValidYAML(t *testing.T) {
	root := NewRootCommand()
	out, err := executeCommand(root, "serve", "config")
	if err != nil {
		t.Fatalf("serve config failed: %v", err)
	}
	if !strings.Contains(out, "listeners") {
		t.Fatalf("expected listeners in config output, got %q", out)
	}
	if !strings.Contains(out, ":8080") {
		t.Fatalf("expected default :8080 listener, got %q", out)
	}
}

func TestServeUninstallMissingUnit(t *testing.T) {
	root := NewRootCommand()
	out, err := executeCommand(root, "serve", "uninstall", "--user")
	if err == nil {
		t.Fatalf("expected error when unit not installed, out=%q", out)
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestServeInstallRequiresConfig(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "serve", "install")
	if err == nil {
		t.Fatal("expected error when --config missing for install")
	}
}

func TestServeInstallUserUnit(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := filepath.Join(home, "server.yaml")
	if err := os.WriteFile(cfg, []byte("listeners:\n  - addr: \":9090\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	out, err := executeCommand(root, "serve", "install", "--config", cfg)
	if err != nil {
		t.Fatalf("serve install failed: %v", err)
	}
	if !strings.Contains(out, "installed systemd unit") {
		t.Fatalf("expected install success message, got %q", out)
	}

	unitPath := filepath.Join(home, ".config", "systemd", "user", "naeos.service")
	data, err := os.ReadFile(unitPath)
	if err != nil {
		t.Fatalf("expected unit file to be written: %v", err)
	}
	if !strings.Contains(string(data), "[Service]") {
		t.Fatalf("expected systemd unit content, got %q", string(data))
	}

	// Uninstall should now remove it.
	out, err = executeCommand(root, "serve", "uninstall", "--user")
	if err != nil {
		t.Fatalf("serve uninstall failed: %v", err)
	}
	if !strings.Contains(out, "removed systemd unit") {
		t.Fatalf("expected uninstall success message, got %q", out)
	}
	if _, err := os.Stat(unitPath); !os.IsNotExist(err) {
		t.Fatal("expected unit file to be removed")
	}
}
