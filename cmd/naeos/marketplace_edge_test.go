package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/marketplace"
)

func TestMarketplaceSearchNoResults(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "search", "zzz-no-match")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(output, "No results found") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplaceSearchJSONEmpty(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "search", "zzz-no-match", "--output", "json")
	if err != nil {
		t.Fatalf("search json: %v", err)
	}
	if !strings.Contains(output, `"count": 0`) {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplaceInstallNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "marketplace", "install", "no-such-template")
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
}

func TestMarketplaceProfileSearch(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "profile", "search", "fintech")
	if err != nil {
		t.Fatalf("profile search: %v", err)
	}
	if !strings.Contains(output, "fintech-core") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplaceProfileDownloadNotFound(t *testing.T) {
	dir := t.TempDir()
	root := NewRootCommand()
	_, err := executeCommand(root, "marketplace", "--cache-dir", dir, "profile", "download", "ghost-profile")
	if err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestMarketplaceProfilePublishErrors(t *testing.T) {
	dir := t.TempDir()
	root := NewRootCommand()

	_, err := executeCommand(root, "marketplace", "--cache-dir", dir, "profile", "publish", "/nonexistent/profile.json")
	if err == nil {
		t.Fatal("expected error for missing profile file")
	}

	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = executeCommand(root, "marketplace", "--cache-dir", dir, "profile", "publish", bad)
	if err == nil {
		t.Fatal("expected error for invalid profile JSON")
	}
}

func TestMarketplacePluginListLocal(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal([]marketplace.PluginEntry{{
		Name: "auth-plugin", Version: "1.0.0", Type: "auth", Description: "Auth helpers",
	}})
	if err := os.WriteFile(filepath.Join(cacheDir, "plugins.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "--cache-dir", cacheDir, "plugin", "list")
	if err != nil {
		t.Fatalf("plugin list: %v", err)
	}
	if !strings.Contains(output, "auth-plugin") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplacePluginSearchLocal(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal([]marketplace.PluginEntry{{
		Name: "web-plugin", Version: "2.0.0", Type: "web", Description: "Web helpers",
	}})
	if err := os.WriteFile(filepath.Join(cacheDir, "plugins.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "--cache-dir", cacheDir, "plugin", "search", "web")
	if err != nil {
		t.Fatalf("plugin search: %v", err)
	}
	if !strings.Contains(output, "web-plugin") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplacePluginInstallUninstall(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal([]marketplace.PluginEntry{{
		Name: "clip-plugin", Version: "1.0.0", Type: "cli", Description: "Clipboard helpers",
	}})
	if err := os.WriteFile(filepath.Join(cacheDir, "plugins.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	workdir := t.TempDir()
	t.Chdir(workdir)

	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "--cache-dir", cacheDir, "plugin", "install", "clip-plugin")
	if err != nil {
		t.Fatalf("plugin install: %v", err)
	}
	if !strings.Contains(output, "Installed plugin clip-plugin") {
		t.Errorf("unexpected output: %q", output)
	}

	output, err = executeCommand(root, "marketplace", "--cache-dir", cacheDir, "plugin", "list")
	if err != nil {
		t.Fatalf("plugin list: %v", err)
	}
	if !strings.Contains(output, "[installed]") {
		t.Errorf("expected installed marker, got %q", output)
	}

	output, err = executeCommand(root, "marketplace", "--cache-dir", cacheDir, "plugin", "uninstall", "clip-plugin")
	if err != nil {
		t.Fatalf("plugin uninstall: %v", err)
	}
	if !strings.Contains(output, "Uninstalled plugin clip-plugin") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplacePluginRemoteList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/plugins" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"plugins":[{"name":"remote-plugin","version":"3.0.0","platform":"linux","description":"Remote plugin"}]}`))
	}))
	defer server.Close()

	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "plugin", "list", "--registry", server.URL)
	if err != nil {
		t.Fatalf("remote plugin list: %v", err)
	}
	if !strings.Contains(output, "remote-plugin") || !strings.Contains(output, "linux") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplacePluginRemoteSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/plugins" {
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"plugins":[{"name":"search-plugin","version":"1.0.0","description":"Searchable plugin"}]}`))
	}))
	defer server.Close()

	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "plugin", "search", "search", "--registry", server.URL)
	if err != nil {
		t.Fatalf("remote plugin search: %v", err)
	}
	if !strings.Contains(output, "search-plugin") || !strings.Contains(output, "any") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestMarketplacePluginRemoteError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	root := NewRootCommand()
	_, err := executeCommand(root, "marketplace", "plugin", "list", "--registry", server.URL)
	if err == nil {
		t.Fatal("expected error from failing registry")
	}
}

func TestMarketplacePublishMissingManifest(t *testing.T) {
	dir := t.TempDir()
	root := NewRootCommand()
	_, err := executeCommand(root, "marketplace", "publish", dir)
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
	if !strings.Contains(err.Error(), "no naeos.yaml manifest") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMarketplacePublishSuccess(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "naeos.yaml"), []byte("name: pkg\nversion: 1.0.0\ntype: template\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := NewRootCommand()
	output, err := executeCommand(root, "marketplace", "publish", dir)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if !strings.Contains(output, "Publishing package") || !strings.Contains(output, "published") {
		t.Errorf("unexpected output: %q", output)
	}
}
