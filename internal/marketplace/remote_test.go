package marketplace

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRemoteRegistryList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list := RemotePluginList{
			Plugins: []RemotePlugin{
				{Name: "go-http-api", Version: "1.0.0", Description: "Go HTTP API", Platform: "linux/amd64", DownloadURL: "http://example.com/plugin.so"},
				{Name: "python-ml", Version: "0.5.0", Description: "Python ML plugin", Platform: "linux/amd64", DownloadURL: "http://example.com/ml.so"},
			},
		}
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	rr := NewRemoteRegistry(server.URL, t.TempDir())
	plugins, err := rr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}
}

func TestRemoteRegistrySearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list := RemotePluginList{
			Plugins: []RemotePlugin{
				{Name: "go-http-api", Version: "1.0.0", Description: "Go HTTP API", Tags: []string{"go", "http"}},
				{Name: "python-ml", Version: "0.5.0", Description: "Python ML plugin", Tags: []string{"python", "ml"}},
				{Name: "rust-web", Version: "0.1.0", Description: "Rust web service", Tags: []string{"rust", "web"}},
			},
		}
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	rr := NewRemoteRegistry(server.URL, t.TempDir())

	results, err := rr.Search("python")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "python-ml" {
		t.Errorf("expected python-ml, got %v", results)
	}

	results, err = rr.Search("http")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "go-http-api" {
		t.Errorf("expected go-http-api by tag/desc, got %v", results)
	}
}

func TestRemoteRegistryInstall(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/plugins":
			list := RemotePluginList{
				Plugins: []RemotePlugin{
					{Name: "test-plugin", Version: "1.0.0", Description: "test", Platform: "linux/amd64", DownloadURL: serverURL + "/download"},
				},
			}
			json.NewEncoder(w).Encode(list)
		case "/download":
			w.Write([]byte("fake plugin binary"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	installDir := t.TempDir()
	rr := NewRemoteRegistry(server.URL, installDir)

	path, err := rr.Install("test-plugin", "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("plugin file not installed")
	}

	metaPath := filepath.Join(installDir, "test-plugin.meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		t.Error("meta file not created")
	}

	installed, err := rr.Installed()
	if err != nil {
		t.Fatal(err)
	}
	if len(installed) != 1 {
		t.Errorf("expected 1 installed, got %d", len(installed))
	}
}

func TestRemoteRegistryUninstall(t *testing.T) {
	installDir := t.TempDir()
	os.WriteFile(filepath.Join(installDir, "test.so"), []byte("binary"), 0o644)
	os.WriteFile(filepath.Join(installDir, "test.meta.json"), []byte("{}"), 0o644)

	rr := NewRemoteRegistry("http://unused", installDir)
	if err := rr.Uninstall("test"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(installDir, "test.so")); !os.IsNotExist(err) {
		t.Error("expected .so to be removed")
	}
	if _, err := os.Stat(filepath.Join(installDir, "test.meta.json")); !os.IsNotExist(err) {
		t.Error("expected .meta.json to be removed")
	}
}

func TestRemoteRegistryNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list := RemotePluginList{Plugins: []RemotePlugin{}}
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	rr := NewRemoteRegistry(server.URL, t.TempDir())
	_, err := rr.Install("nonexistent", "")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}
