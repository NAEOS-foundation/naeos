package main

import (
	"strings"
	"testing"
)

func TestMarketplaceSearchCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "search", "web")
	if err != nil {
		t.Fatalf("execute marketplace search failed: %v", err)
	}
	if !strings.Contains(output, "rust-web-service") {
		t.Fatalf("expected search results, got %q", output)
	}
}

func TestMarketplaceInstallCmd(t *testing.T) {
	dir := t.TempDir()
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "install", "go-http-api")
	if err != nil {
		t.Fatalf("execute marketplace install failed: %v", err)
	}
	if !strings.Contains(output, "Installed template") {
		t.Fatalf("expected install confirmation, got %q", output)
	}
	_ = dir
}

func TestMarketplaceProfileListCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "profile", "list")
	if err != nil {
		t.Fatalf("execute marketplace profile list failed: %v", err)
	}
	if len(strings.TrimSpace(output)) == 0 {
		t.Fatal("expected profile list output")
	}
}

func TestMarketplacePluginListCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "plugin", "list")
	if err != nil {
		t.Fatalf("execute marketplace plugin list failed: %v", err)
	}
	// should not error even if empty
	_ = output
}

func TestMarketplaceHelpCmd(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "--help")
	if err != nil {
		t.Fatalf("execute marketplace --help failed: %v", err)
	}
	if !strings.Contains(output, "marketplace") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestMarketplaceSearchJSONOutput(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "search", "web", "--output", "json")
	if err != nil {
		t.Fatalf("execute marketplace search json failed: %v", err)
	}
	if !strings.Contains(output, `"results"`) {
		t.Fatalf("expected JSON output, got %q", output)
	}
}
