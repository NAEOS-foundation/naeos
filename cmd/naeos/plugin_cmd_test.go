package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type seededPlugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author,omitempty"`
	Path        string `json:"path,omitempty"`
	Enabled     bool   `json:"enabled"`
	Loaded      bool   `json:"loaded"`
	State       string `json:"state"`
}

func writePluginsFile(t *testing.T, dir string, plugins []seededPlugin) {
	t.Helper()
	data, err := json.Marshal(map[string]any{"plugins": plugins})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugins.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func seedTwoPlugins(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writePluginsFile(t, dir, []seededPlugin{
		{Name: "alpha", Version: "1.0.0", Description: "alpha plugin", Path: "/tmp/alpha.so", Enabled: true, State: "created"},
		{Name: "beta", Version: "2.0.0", Description: "beta plugin", Enabled: false, State: "created"},
	})
	return dir
}

func readPluginConfig(t *testing.T, dir string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "plugins.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestPluginListEmpty(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "list", "--plugin-dir", t.TempDir())
	if err != nil {
		t.Fatalf("plugin list failed: %v", err)
	}
	if !strings.Contains(output, "No plugins installed") {
		t.Fatalf("expected empty-state message, got %q", output)
	}
}

func TestPluginListTable(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "list", "--plugin-dir", dir)
	if err != nil {
		t.Fatalf("plugin list failed: %v", err)
	}
	if !strings.Contains(output, "alpha") || !strings.Contains(output, "beta") {
		t.Fatalf("expected both plugins in output, got %q", output)
	}
	if !strings.Contains(output, "disabled") || !strings.Contains(output, "enabled") {
		t.Fatalf("expected statuses in output, got %q", output)
	}
}

func TestPluginListJSON(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "list", "--plugin-dir", dir, "--output-format", "json")
	if err != nil {
		t.Fatalf("plugin list json failed: %v", err)
	}
	if !strings.Contains(output, `"name": "alpha"`) {
		t.Fatalf("expected json plugin entry, got %q", output)
	}
}

func TestPluginListYAML(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "list", "--plugin-dir", dir, "--output-format", "yaml")
	if err != nil {
		t.Fatalf("plugin list yaml failed: %v", err)
	}
	if !strings.Contains(output, "name: alpha") {
		t.Fatalf("expected yaml plugin entry, got %q", output)
	}
}

func TestPluginInfo(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "info", "alpha", "--plugin-dir", dir)
	if err != nil {
		t.Fatalf("plugin info failed: %v", err)
	}
	if !strings.Contains(output, "Name:        alpha") || !strings.Contains(output, "Version:     1.0.0") {
		t.Fatalf("expected plugin info details, got %q", output)
	}
	if !strings.Contains(output, "Status:      enabled") || !strings.Contains(output, "State:       created") {
		t.Fatalf("expected status/state lines, got %q", output)
	}
}

func TestPluginInfoNotFound(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	_, err := executeCommand(root, "plugin", "info", "ghost", "--plugin-dir", dir)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestPluginEnable(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "enable", "beta", "--plugin-dir", dir)
	if err != nil {
		t.Fatalf("plugin enable failed: %v", err)
	}
	if !strings.Contains(output, "Enabled plugin beta") {
		t.Fatalf("expected enable message, got %q", output)
	}
	plugins := readPluginConfig(t, dir)["plugins"].([]any)
	beta := plugins[1].(map[string]any)
	if beta["enabled"] != true {
		t.Fatalf("expected beta enabled=true after enable, got %v", beta["enabled"])
	}
}

func TestPluginDisable(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "disable", "alpha", "--plugin-dir", dir)
	if err != nil {
		t.Fatalf("plugin disable failed: %v", err)
	}
	if !strings.Contains(output, "Disabled plugin alpha") {
		t.Fatalf("expected disable message, got %q", output)
	}
	plugins := readPluginConfig(t, dir)["plugins"].([]any)
	alpha := plugins[0].(map[string]any)
	if alpha["enabled"] != false {
		t.Fatalf("expected alpha enabled=false after disable, got %v", alpha["enabled"])
	}
}

func TestPluginEnableNotFound(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	_, err := executeCommand(root, "plugin", "enable", "ghost", "--plugin-dir", dir)
	if err == nil {
		t.Fatal("expected error for unknown plugin")
	}
}

func TestPluginUninstall(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "uninstall", "alpha", "--plugin-dir", dir)
	if err != nil {
		t.Fatalf("plugin uninstall failed: %v", err)
	}
	if !strings.Contains(output, "Uninstalled plugin alpha") {
		t.Fatalf("expected uninstall message, got %q", output)
	}
	plugins := readPluginConfig(t, dir)["plugins"].([]any)
	if len(plugins) != 1 {
		t.Fatalf("expected one plugin remaining, got %d", len(plugins))
	}
}

func TestPluginUninstallNotFound(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	_, err := executeCommand(root, "plugin", "uninstall", "ghost", "--plugin-dir", dir)
	if err == nil {
		t.Fatal("expected error for unknown plugin")
	}
}

func TestPluginInstallBadFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "plugin", "install", filepath.Join(t.TempDir(), "missing.so"), "--plugin-dir", t.TempDir())
	if err == nil {
		t.Fatal("expected error installing nonexistent .so file")
	}
}

func TestPluginExecuteInvalidParams(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	_, err := executeCommand(root, "plugin", "execute", "alpha", "lint", "--params", "{bad json", "--plugin-dir", dir)
	if err == nil || !strings.Contains(err.Error(), "invalid params JSON") {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestPluginExecuteUnknownPlugin(t *testing.T) {
	dir := seedTwoPlugins(t)

	root := NewRootCommand()
	_, err := executeCommand(root, "plugin", "execute", "ghost", "lint", "--params", `{"file":"main.go"}`, "--plugin-dir", dir)
	if err == nil {
		t.Fatal("expected error executing unknown plugin")
	}
}

func TestPluginTestBadFile(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "test", filepath.Join(t.TempDir(), "missing.so"), "--plugin-dir", t.TempDir())
	if err != nil {
		t.Fatalf("plugin test should report failure in output, not error: %v", err)
	}
	if !strings.Contains(output, "FAIL  install/load") {
		t.Fatalf("expected FAIL install/load output, got %q", output)
	}
}

func TestPluginSearchOffline(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "plugin", "search", "lint", "--registry", "http://127.0.0.1:1/registry")
	if err == nil {
		t.Fatal("expected error when registry is unreachable")
	}
}

func TestPluginCreate(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "create", "myplug")
	if err != nil {
		t.Fatalf("plugin create failed: %v", err)
	}
	if !strings.Contains(output, "Created plugin skeleton: myplug/") {
		t.Fatalf("expected skeleton message, got %q", output)
	}
	for _, name := range []string{"naeos.yaml", "main.go"} {
		if _, err := os.Stat(filepath.Join(dir, "myplug", name)); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
}

func TestPluginInit(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	root := NewRootCommand()
	output, err := executeCommand(root, "plugin", "init", "myplug")
	if err != nil {
		t.Fatalf("plugin init failed: %v", err)
	}
	if !strings.Contains(output, "Created plugin project: myplug/") {
		t.Fatalf("expected scaffold message, got %q", output)
	}
	for _, name := range []string{"go.mod", "plugin.go", "main.go", "naeos.yaml", "Makefile", "README.md"} {
		if _, err := os.Stat(filepath.Join(dir, "myplug", name)); err != nil {
			t.Fatalf("expected %s to exist: %v", name, err)
		}
	}
}
