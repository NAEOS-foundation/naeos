package marketplace

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient(t.TempDir())
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestSearchDefaultEntries(t *testing.T) {
	client := NewClient(t.TempDir())
	results, err := client.Search(SearchFilter{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) < 3 {
		t.Errorf("expected at least 3 default entries, got %d", len(results))
	}
}

func TestSearchWithQuery(t *testing.T) {
	client := NewClient(t.TempDir())
	results, err := client.Search(SearchFilter{Query: "go", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'go' query")
	}
}

func TestSearchWithTags(t *testing.T) {
	client := NewClient(t.TempDir())
	results, err := client.Search(SearchFilter{Tags: []string{"python"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'python' tag")
	}
}

func TestGet(t *testing.T) {
	client := NewClient(t.TempDir())
	entry, err := client.Get("go-http-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "go-http-api" {
		t.Errorf("expected go-http-api, got %s", entry.Name)
	}
}

func TestGetNotFound(t *testing.T) {
	client := NewClient(t.TempDir())
	_, err := client.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent entry")
	}
}

func TestPublish(t *testing.T) {
	cacheDir := t.TempDir()
	client := NewClient(cacheDir)
	entry := RegistryEntry{Name: "test-spec", Version: "1.0.0", Description: "Test"}
	if err := client.Publish(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it was published
	data, err := os.ReadFile(filepath.Join(cacheDir, "registry.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty registry file")
	}
}

func TestInstall(t *testing.T) {
	client := NewClient(t.TempDir())
	targetDir := t.TempDir()
	if err := client.Install("go-http-api", targetDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	specPath := filepath.Join(targetDir, "spec.yaml")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("spec.yaml not created: %v", err)
	}
	content := string(data)
	if len(content) == 0 {
		t.Error("expected non-empty spec.yaml")
	}
	// Verify key fields appear in the generated spec
	for _, want := range []string{"go-http-api", "project:"} {
		if !contains(content, want) {
			t.Errorf("spec.yaml content missing %q", want)
		}
	}
}

func TestInstallNotFound(t *testing.T) {
	client := NewClient(t.TempDir())
	if err := client.Install("nonexistent-entry", t.TempDir()); err == nil {
		t.Fatal("expected error for nonexistent entry")
	}
}

func TestPublishUpdate(t *testing.T) {
	cacheDir := t.TempDir()
	client := NewClient(cacheDir)
	entry := RegistryEntry{Name: "my-spec", Version: "1.0.0", Description: "first"}
	if err := client.Publish(entry); err != nil {
		t.Fatalf("first publish: %v", err)
	}
	updated := RegistryEntry{Name: "my-spec", Version: "2.0.0", Description: "updated"}
	if err := client.Publish(updated); err != nil {
		t.Fatalf("second publish: %v", err)
	}
	got, err := client.Get("my-spec")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0 after update, got %s", got.Version)
	}
	if got.Description != "updated" {
		t.Errorf("expected description 'updated', got %s", got.Description)
	}
}

func TestSearchLimitExact(t *testing.T) {
	client := NewClient(t.TempDir())
	results, err := client.Search(SearchFilter{Limit: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected exactly 1 result with limit=1, got %d", len(results))
	}
}

func TestSearchNoMatch(t *testing.T) {
	client := NewClient(t.TempDir())
	results, err := client.Search(SearchFilter{Query: "zzz-nonexistent-query-zzz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", len(results))
	}
}

func TestSearchMultipleTags(t *testing.T) {
	cacheDir := t.TempDir()
	// Write a clean registry with only our test entries (avoiding defaults)
	data := []byte(`[
	  {"name": "alpha", "version": "1.0.0", "description": "A", "tags": ["go", "web"]},
	  {"name": "beta", "version": "1.0.0", "description": "B", "tags": ["python", "ml"]},
	  {"name": "gamma", "version": "1.0.0", "description": "C", "tags": ["go", "ml"]}
]`)
	os.WriteFile(filepath.Join(cacheDir, "registry.json"), data, 0o600)

	client := NewClient(cacheDir)

	// Search for entries that have the "go" tag
	results, err := client.Search(SearchFilter{Tags: []string{"go"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for tag 'go', got %d", len(results))
	}

	// Search for entries that have the "ml" tag
	results, err = client.Search(SearchFilter{Tags: []string{"ml"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results for tag 'ml', got %d", len(results))
	}

	// Search with multiple tags uses OR semantics: match any of the filter tags
	results, err = client.Search(SearchFilter{Tags: []string{"go", "ml"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results for tags 'go' OR 'ml', got %d", len(results))
	}
}

func TestContainsExactMatch(t *testing.T) {
	if !contains("abc", "abc") {
		t.Error("expected contains(\"abc\", \"abc\") = true")
	}
}

func TestContainsSubstrMatch(t *testing.T) {
	if !contains("abcde", "bcd") {
		t.Error("expected contains(\"abcde\", \"bcd\") = true")
	}
}

func TestContainsNoMatch(t *testing.T) {
	if contains("abc", "xyz") {
		t.Error("expected contains(\"abc\", \"xyz\") = false")
	}
}

func TestContainsEmptyQuery(t *testing.T) {
	if !contains("abc", "") {
		t.Error("expected contains(\"abc\", \"\") = true")
	}
}

func TestLoadCacheCorrupted(t *testing.T) {
	cacheDir := t.TempDir()
	// Write invalid JSON to registry.json
	if err := os.WriteFile(filepath.Join(cacheDir, "registry.json"), []byte("{invalid json!!"), 0o600); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}
	client := NewClient(cacheDir)
	_, err := client.Search(SearchFilter{})
	if err == nil {
		t.Fatal("expected error for corrupted cache, got nil")
	}
}

func TestDefaultEntriesVersion(t *testing.T) {
	client := NewClient(t.TempDir())
	entry, err := client.Get("go-http-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Version == "" {
		t.Error("expected non-empty version for go-http-api")
	}
	entry2, err := client.Get("rust-web-service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry2.Version == "" {
		t.Error("expected non-empty version for rust-web-service")
	}
}

func TestFetchPluginNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	rr := NewRemoteRegistry(srv.URL+"/plugins", t.TempDir())
	_, err := rr.List()
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("expected 404 in error, got: %v", err)
	}
}

func TestFetchPluginTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"plugins":[]}`))
	}))
	defer srv.Close()

	rr := &RemoteRegistry{
		baseURL:    srv.URL + "/plugins",
		installDir: t.TempDir(),
		httpClient: &http.Client{Timeout: 50 * time.Millisecond},
	}
	_, err := rr.List()
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestFetchPluginInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	rr := NewRemoteRegistry(srv.URL+"/plugins", t.TempDir())
	_, err := rr.List()
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("expected decode error, got: %v", err)
	}
}

func TestSearchPluginsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"plugins":[]}`))
	}))
	defer srv.Close()

	rr := NewRemoteRegistry(srv.URL+"/plugins", t.TempDir())
	results, err := rr.Search("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestContainsSubstr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		s, substr string
		want      bool
	}{
		{"hello", "ell", true},
		{"hello", "xyz", false},
		{"a", "a", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			t.Parallel()
			if got := containsSubstr(tt.s, tt.substr); got != tt.want {
				t.Errorf("containsSubstr(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestSearchWithQueryInDescription(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	data := []byte(`[
	  {"name": "alpha", "version": "1.0.0", "description": "Go HTTP API", "tags": ["web"]}
]`)
	os.WriteFile(filepath.Join(cacheDir, "registry.json"), data, 0o644)

	client := NewClient(cacheDir)
	results, err := client.Search(SearchFilter{Query: "HTTP"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestSearchWithQueryInTag(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	data := []byte(`[
	  {"name": "beta", "version": "1.0.0", "description": "Service", "tags": ["grpc", "go"]}
]`)
	os.WriteFile(filepath.Join(cacheDir, "registry.json"), data, 0o644)

	client := NewClient(cacheDir)
	results, err := client.Search(SearchFilter{Query: "grpc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestGetWithCorruptedCache(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	os.WriteFile(filepath.Join(cacheDir, "registry.json"), []byte("bad json"), 0o644)

	client := NewClient(cacheDir)
	_, err := client.Get("test")
	if err == nil {
		t.Fatal("expected error for corrupted cache")
	}
}

func TestPublishWithCorruptedCache(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	os.WriteFile(filepath.Join(cacheDir, "registry.json"), []byte("bad json"), 0o644)

	client := NewClient(cacheDir)
	err := client.Publish(RegistryEntry{Name: "test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, err := client.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if entry.Name != "test" {
		t.Errorf("expected 'test', got %q", entry.Name)
	}
}

func TestFetchPluginNonOKStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer srv.Close()

	rr := NewRemoteRegistry(srv.URL+"/plugins", t.TempDir())
	_, err := rr.List()
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected 403 in error, got: %v", err)
	}
}

func TestNewRemoteRegistryDefaultURL(t *testing.T) {
	t.Parallel()

	rr := NewRemoteRegistry("", t.TempDir())
	if rr.baseURL != DefaultRegistryURL {
		t.Errorf("expected default URL, got %q", rr.baseURL)
	}
}

func TestRemoteRegistryInstallNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RemotePluginList{Plugins: []RemotePlugin{}})
	}))
	defer srv.Close()

	rr := NewRemoteRegistry(srv.URL+"/plugins", t.TempDir())
	_, err := rr.Install("nonexistent", "")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestRemoteRegistryInstallWrongVersion(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RemotePluginList{
			Plugins: []RemotePlugin{
				{Name: "test", Version: "1.0.0", Platform: "linux/amd64"},
			},
		})
	}))
	defer srv.Close()

	rr := NewRemoteRegistry(srv.URL+"/plugins", t.TempDir())
	_, err := rr.Install("test", "2.0.0")
	if err == nil {
		t.Fatal("expected error for wrong version")
	}
}

func TestRemoteRegistryInstalledEmptyDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rr := NewRemoteRegistry("http://unused", dir)
	plugins, err := rr.Installed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 installed, got %d", len(plugins))
	}
}

func TestRemoteRegistryInstalledNonExistentDir(t *testing.T) {
	t.Parallel()

	rr := NewRemoteRegistry("http://unused", "/nonexistent/dir")
	plugins, err := rr.Installed()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 installed, got %d", len(plugins))
	}
}
