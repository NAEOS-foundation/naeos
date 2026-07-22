package marketplace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewProfileMarketplace(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	if m == nil {
		t.Fatal("expected non-nil")
	}
}

func TestProfileMarketplacePublishAndGet(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	entry := ProfileEntry{Name: "my-profile", Version: "1.0.0", Description: "test"}
	if err := m.Publish(entry); err != nil {
		t.Fatalf("publish: %v", err)
	}
	got, err := m.Get("my-profile")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", got.Version)
	}
}

func TestProfileMarketplacePublishUpdate(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	m.Publish(ProfileEntry{Name: "p", Version: "1.0.0"})
	m.Publish(ProfileEntry{Name: "p", Version: "2.0.0"})
	got, _ := m.Get("p")
	if got.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", got.Version)
	}
}

func TestProfileMarketplaceGetNotFound(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	_, err := m.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProfileMarketplaceList(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	entries, err := m.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 9 {
		t.Errorf("expected 9 default profiles, got %d", len(entries))
	}
}

func TestProfileMarketplaceSearch(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	results, err := m.Search("saas", nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'saas'")
	}
}

func TestProfileMarketplaceSearchByIndustry(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	results, err := m.Search("fintech", nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'fintech'")
	}
}

func TestProfileMarketplaceSearchByTag(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	results, err := m.Search("", []string{"hipaa"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for tag 'hipaa'")
	}
}

func TestProfileMarketplaceSearchNoMatch(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	results, err := m.Search("zzz-no-match-here", nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0, got %d", len(results))
	}
}

func TestProfileMarketplaceDownload(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	targetDir := t.TempDir()

	if err := m.Download("saas-starter", targetDir); err != nil {
		t.Fatalf("download: %v", err)
	}

	profileFile := filepath.Join(targetDir, ".naeos", "profiles", "saas-starter.json")
	if _, err := os.Stat(profileFile); os.IsNotExist(err) {
		t.Error("profile file not created")
	}
}

func TestProfileMarketplaceDownloadNotFound(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	err := m.Download("nonexistent", t.TempDir())
	if err == nil {
		t.Error("expected error")
	}
}

func TestProfileMarketplaceUpload(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	dir := t.TempDir()
	entry := ProfileEntry{Name: "uploaded-profile", Version: "1.0.0", Description: "uploaded"}
	data, _ := json.Marshal(entry)
	filePath := filepath.Join(dir, "profile.json")
	os.WriteFile(filePath, data, 0o600)

	if err := m.Upload(filePath); err != nil {
		t.Fatalf("upload: %v", err)
	}
	got, err := m.Get("uploaded-profile")
	if err != nil {
		t.Fatalf("get after upload: %v", err)
	}
	if got.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", got.Version)
	}
}

func TestProfileMarketplaceUploadInvalidFile(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	err := m.Upload("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error")
	}
}

func TestProfileMarketplaceRemove(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	m.Publish(ProfileEntry{Name: "to-remove", Version: "1.0.0"})
	if err := m.Remove("to-remove"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	_, err := m.Get("to-remove")
	if err == nil {
		t.Error("expected error after remove")
	}
}

func TestProfileMarketplaceRemoveNotFound(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	err := m.Remove("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestProfileMarketplaceIncrementDownloads(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	m.Publish(ProfileEntry{Name: "test-p", Version: "1.0.0"})
	if err := m.IncrementDownloads("test-p"); err != nil {
		t.Fatalf("increment: %v", err)
	}
	entry, _ := m.Get("test-p")
	if entry.Downloads != 1 {
		t.Errorf("expected 1 download, got %d", entry.Downloads)
	}
}

func TestProfileMarketplaceIncrementDownloadsNotFound(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	err := m.IncrementDownloads("nonexistent")
	if err == nil {
		t.Error("expected error")
	}
}

func TestContainsStr(t *testing.T) {
	if !containsStr("hello world", "world") {
		t.Error("expected substring match")
	}
	if containsStr("hello", "xyz") {
		t.Error("expected no match")
	}
}

func TestProfileMarketplaceSearchQueryInName(t *testing.T) {
	m := NewProfileMarketplace(t.TempDir())
	results, err := m.Search("saas-starter", nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for exact name match")
	}
}
