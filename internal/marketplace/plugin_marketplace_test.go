package marketplace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewPluginMarketplace(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	if m == nil {
		t.Fatal("expected non-nil")
	}
}

func TestPluginMarketplacePublishAndGet(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	entry := PluginEntry{Name: "my-plugin", Version: "1.0.0", Description: "test"}
	if err := m.Publish(entry); err != nil {
		t.Fatalf("publish: %v", err)
	}
	got, err := m.Get("my-plugin")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", got.Version)
	}
}

func TestPluginMarketplacePublishUpdate(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	m.Publish(PluginEntry{Name: "p", Version: "1.0.0"})
	m.Publish(PluginEntry{Name: "p", Version: "2.0.0"})
	got, _ := m.Get("p")
	if got.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", got.Version)
	}
}

func TestPluginMarketplaceGetNotFound(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	_, err := m.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPluginMarketplaceList(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	entries, err := m.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 4 {
		t.Errorf("expected 4 default plugins, got %d", len(entries))
	}
}

func TestPluginMarketplaceSearch(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	results, err := m.Search("lint", nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'lint'")
	}
}

func TestPluginMarketplaceSearchByTag(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	results, err := m.Search("", []string{"testing"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for tag 'testing'")
	}
}

func TestPluginMarketplaceSearchNoMatch(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	results, err := m.Search("zzz-no-match", nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0, got %d", len(results))
	}
}

func TestPluginMarketplaceInstall(t *testing.T) {
	cacheDir := t.TempDir()
	installDir := t.TempDir()
	m := NewPluginMarketplace(cacheDir, installDir)

	if err := m.Install("naeos-lint"); err != nil {
		t.Fatalf("install: %v", err)
	}
	if !m.IsInstalled("naeos-lint") {
		t.Error("expected plugin to be installed")
	}
	pluginDir := filepath.Join(installDir, "naeos-lint")
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Error("plugin dir not created")
	}
}

func TestPluginMarketplaceInstallAlreadyInstalled(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	m.Install("naeos-lint")
	err := m.Install("naeos-lint")
	if err == nil {
		t.Error("expected error for already installed")
	}
}

func TestPluginMarketplaceInstallInvalidName(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	err := m.Install("../../evil")
	if err == nil {
		t.Error("expected error for invalid name")
	}
}

func TestPluginMarketplaceInstallNotFound(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	err := m.Install("nonexistent-plugin")
	if err == nil {
		t.Error("expected error")
	}
}

func TestPluginMarketplaceInstallWithDependencies(t *testing.T) {
	cacheDir := t.TempDir()
	installDir := t.TempDir()

	m := NewPluginMarketplace(cacheDir, installDir)
	m.Publish(PluginEntry{Name: "dep-a", Version: "1.0.0"})
	m.Publish(PluginEntry{
		Name:         "main-plugin",
		Version:      "1.0.0",
		Dependencies: []PluginDependency{{Name: "dep-a", Version: "1.0.0"}},
	})

	if err := m.Install("main-plugin"); err != nil {
		t.Fatalf("install with deps: %v", err)
	}
	if !m.IsInstalled("main-plugin") {
		t.Error("main-plugin not installed")
	}
}

func TestPluginMarketplaceUninstall(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	m.Install("naeos-lint")
	if err := m.Uninstall("naeos-lint"); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if m.IsInstalled("naeos-lint") {
		t.Error("expected plugin to be uninstalled")
	}
}

func TestPluginMarketplaceUninstallInvalidName(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	err := m.Uninstall("")
	if err == nil {
		t.Error("expected error")
	}
}

func TestPluginMarketplaceListInstalled(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	m.Install("naeos-lint")
	installed, err := m.ListInstalled()
	if err != nil {
		t.Fatalf("list installed: %v", err)
	}
	if len(installed) != 1 {
		t.Errorf("expected 1 installed, got %d", len(installed))
	}
}

func TestPluginMarketplaceVersionHistory(t *testing.T) {
	cacheDir := t.TempDir()
	installDir := t.TempDir()
	m := NewPluginMarketplace(cacheDir, installDir)
	m.Publish(PluginEntry{Name: "test-p", Version: "1.0.0"})
	m.Install("test-p")

	history, err := m.VersionHistory("test-p")
	if err != nil {
		t.Fatalf("version history: %v", err)
	}
	if len(history) == 0 {
		t.Error("expected non-empty history")
	}
}

func TestPluginMarketplaceVersionHistoryNotFound(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	_, err := m.VersionHistory("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestPluginMarketplaceVersionHistoryInvalidName(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	_, err := m.VersionHistory("../bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestPluginMarketplaceRollback(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())

	m.Publish(PluginEntry{Name: "test-p", Version: "1.0.0"})
	m.Install("test-p")

	entries, _ := m.loadPlugins()
	for i, e := range entries {
		if e.Name == "test-p" {
			entries[i].Version = "2.0.0"
			entries[i].VersionHistory = append(entries[i].VersionHistory,
				VersionEntry{Version: "2.0.0", Installed: time.Now()})
			break
		}
	}
	m.savePlugins(entries)

	if err := m.Rollback("test-p", ""); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	entry, _ := m.Get("test-p")
	if entry.Version != "1.0.0" {
		t.Errorf("expected 1.0.0 after rollback, got %s", entry.Version)
	}
}

func TestPluginMarketplaceRollbackSpecificVersion(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())

	m.Publish(PluginEntry{Name: "test-p", Version: "1.0.0"})
	m.Install("test-p")

	entries, _ := m.loadPlugins()
	for i, e := range entries {
		if e.Name == "test-p" {
			entries[i].Version = "3.0.0"
			entries[i].VersionHistory = append(entries[i].VersionHistory,
				VersionEntry{Version: "2.0.0", Installed: time.Now()},
				VersionEntry{Version: "3.0.0", Installed: time.Now()})
			break
		}
	}
	m.savePlugins(entries)
	m.Install("test-p")

	if err := m.Rollback("test-p", "1.0.0"); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	entry, _ := m.Get("test-p")
	if entry.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", entry.Version)
	}
}

func TestPluginMarketplaceRollbackNoHistory(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	err := m.Rollback("naeos-lint", "")
	if err == nil {
		t.Error("expected error with no history")
	}
}

func TestPluginMarketplaceRollbackInvalidName(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	err := m.Rollback("", "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestResolveDependencies(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	m.Publish(PluginEntry{Name: "lib-a", Version: "1.0.0"})

	resolved, err := m.ResolveDependencies([]PluginDependency{{Name: "lib-a"}}, nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(resolved) != 1 {
		t.Errorf("expected 1, got %d", len(resolved))
	}
}

func TestResolveDependenciesCircular(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	m.Publish(PluginEntry{Name: "a", Version: "1.0.0", Dependencies: []PluginDependency{{Name: "b"}}})
	m.Publish(PluginEntry{Name: "b", Version: "1.0.0", Dependencies: []PluginDependency{{Name: "a"}}})

	_, err := m.ResolveDependencies([]PluginDependency{{Name: "a"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected circular dependency error, got: %v", err)
	}
}

func TestResolveDependenciesMissing(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	_, err := m.ResolveDependencies([]PluginDependency{{Name: "missing-dep"}}, nil)
	if err == nil {
		t.Error("expected error")
	}
}

func TestResolveDependenciesNested(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	m.Publish(PluginEntry{Name: "leaf", Version: "1.0.0"})
	m.Publish(PluginEntry{Name: "mid", Version: "1.0.0", Dependencies: []PluginDependency{{Name: "leaf"}}})
	m.Publish(PluginEntry{Name: "top", Version: "1.0.0", Dependencies: []PluginDependency{{Name: "mid"}}})

	resolved, err := m.ResolveDependencies([]PluginDependency{{Name: "top"}}, nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(resolved) < 2 {
		t.Errorf("expected at least 2 resolved, got %d", len(resolved))
	}
}

func TestVersionMatchExact(t *testing.T) {
	if !versionMatch("1.0.0", "1.0.0") {
		t.Error("expected exact match")
	}
}

func TestVersionMatchEmpty(t *testing.T) {
	if !versionMatch("", "1.0.0") {
		t.Error("expected empty required to match")
	}
}

func TestVersionMatchWildcard(t *testing.T) {
	if !versionMatch("*", "1.0.0") {
		t.Error("expected wildcard to match")
	}
}

func TestVersionMatchMismatch(t *testing.T) {
	if versionMatch("1.0.0", "2.0.0") {
		t.Error("expected mismatch")
	}
}

func TestPluginMarketplaceIsNotInstalled(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	if m.IsInstalled("nonexistent") {
		t.Error("expected false")
	}
}

func TestPluginMarketplaceIsInstalledInvalidName(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	if m.IsInstalled("../bad") {
		t.Error("expected false for invalid name")
	}
}

func TestSetCreatedAtPreserved(t *testing.T) {
	m := NewPluginMarketplace(t.TempDir(), t.TempDir())
	now := time.Now()
	entry := PluginEntry{Name: "p", Version: "1.0.0", CreatedAt: now}
	m.Publish(entry)
	got, _ := m.Get("p")
	if !got.CreatedAt.Equal(now) {
		t.Error("created_at should be preserved")
	}
}
