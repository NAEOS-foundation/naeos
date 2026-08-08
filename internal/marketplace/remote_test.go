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

	rr := NewRemoteRegistry(server.URL+"/plugins", t.TempDir())
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

	rr := NewRemoteRegistry(server.URL+"/plugins", t.TempDir())

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

func TestRemoteRegistrySearchFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list := RemotePluginList{
			Plugins: []RemotePlugin{
				{Name: "plug-a", Version: "1.0.0", Description: "alpha", Author: "team-a", Platform: "linux/amd64", Tags: []string{"go"}},
				{Name: "plug-b", Version: "2.0.0", Description: "beta", Author: "team-b", Platform: "linux/arm64", Tags: []string{"python"}},
				{Name: "plug-c", Version: "1.5.0", Description: "gamma", Author: "team-a", Platform: "linux/amd64", Tags: []string{"go", "http"}},
			},
		}
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	rr := NewRemoteRegistry(server.URL, t.TempDir())

	t.Run("filter by author", func(t *testing.T) {
		results, err := rr.SearchFilter(RemoteSearchFilter{Author: "team-a"})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 2 {
			t.Errorf("expected 2, got %d", len(results))
		}
	})

	t.Run("filter by platform", func(t *testing.T) {
		results, err := rr.SearchFilter(RemoteSearchFilter{Platform: "linux/arm64"})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 || results[0].Name != "plug-b" {
			t.Errorf("expected plug-b, got %v", results)
		}
	})

	t.Run("filter by tags", func(t *testing.T) {
		results, err := rr.SearchFilter(RemoteSearchFilter{Tags: []string{"http"}})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 || results[0].Name != "plug-c" {
			t.Errorf("expected plug-c, got %v", results)
		}
	})

	t.Run("filter by query and author", func(t *testing.T) {
		results, err := rr.SearchFilter(RemoteSearchFilter{Query: "alpha", Author: "team-a"})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 || results[0].Name != "plug-a" {
			t.Errorf("expected plug-a, got %v", results)
		}
	})

	t.Run("no match", func(t *testing.T) {
		results, err := rr.SearchFilter(RemoteSearchFilter{Platform: "windows/amd64"})
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0, got %d", len(results))
		}
	})
}

func TestRemoteRegistryInstall(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/plugins/registry.json":
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
	rr := NewRemoteRegistry(server.URL+"/plugins/registry.json", installDir)

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

	rr := NewRemoteRegistry(server.URL+"/plugins", t.TempDir())
	_, err := rr.Install("nonexistent", "")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestRemoteRegistryListStaticJSONURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/plugins/registry.json" {
			t.Errorf("expected /plugins/registry.json, got %s", r.URL.Path)
		}
		list := RemotePluginList{
			Plugins: []RemotePlugin{
				{Name: "static", Version: "2.0.0", Description: "static registry plugin", Platforms: []string{"linux/amd64"}},
			},
		}
		json.NewEncoder(w).Encode(list)
	}))
	defer server.Close()

	rr := NewRemoteRegistry(server.URL+"/plugins/registry.json", t.TempDir())
	plugins, err := rr.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 1 || plugins[0].Name != "static" {
		t.Errorf("expected static registry entry, got %v", plugins)
	}
	if len(plugins[0].Platforms) != 1 || plugins[0].Platforms[0] != "linux/amd64" {
		t.Errorf("expected platforms array to be decoded, got %v", plugins[0].Platforms)
	}
}

func TestSupportsPlatform(t *testing.T) {
	tests := []struct {
		name     string
		plugin   RemotePlugin
		platform string
		want     bool
	}{
		{"no platform declared matches any", RemotePlugin{}, "linux/amd64", true},
		{"exact single platform", RemotePlugin{Platform: "linux/amd64"}, "linux/amd64", true},
		{"mismatched single platform", RemotePlugin{Platform: "darwin/arm64"}, "linux/amd64", false},
		{"any wildcard matches", RemotePlugin{Platform: "any"}, "windows/amd64", true},
		{"platforms array match", RemotePlugin{Platforms: []string{"linux/amd64", "linux/arm64"}}, "linux/arm64", true},
		{"platforms array no match", RemotePlugin{Platforms: []string{"linux/amd64"}}, "linux/arm64", false},
		{"platforms case insensitive", RemotePlugin{Platforms: []string{"Linux/amd64"}}, "linux/amd64", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := supportsPlatform(tt.plugin, tt.platform); got != tt.want {
				t.Errorf("supportsPlatform(%v, %q) = %v, want %v", tt.plugin, tt.platform, got, tt.want)
			}
		})
	}
}
