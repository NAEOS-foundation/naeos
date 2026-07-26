package marketplace

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewRemoteTemplateRegistry(t *testing.T) {
	t.Run("default URL", func(t *testing.T) {
		r := NewRemoteTemplateRegistry("")
		if r.baseURL != DefaultTemplateRegistryURL {
			t.Errorf("expected %q, got %q", DefaultTemplateRegistryURL, r.baseURL)
		}
	})

	t.Run("custom URL", func(t *testing.T) {
		r := NewRemoteTemplateRegistry("https://example.com/registry.json")
		if r.baseURL != "https://example.com/registry.json" {
			t.Errorf("expected custom URL, got %q", r.baseURL)
		}
		if r.httpClient == nil {
			t.Error("httpClient should not be nil")
		}
	})
}

func TestRemoteTemplateRegistry_List(t *testing.T) {
	t.Run("file protocol valid", func(t *testing.T) {
		dir := t.TempDir()
		regPath := filepath.Join(dir, "registry.json")
		entries := TemplateList{
			Templates: []TemplateEntry{
				{Name: "test-tmpl", Version: "1.0.0", Description: "test", Author: "tester"},
			},
		}
		data, _ := json.Marshal(entries)
		if err := os.WriteFile(regPath, data, 0o600); err != nil {
			t.Fatal(err)
		}
		r := NewRemoteTemplateRegistry("file://" + regPath)
		results, err := r.List()
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 template, got %d", len(results))
		}
		if results[0].Name != "test-tmpl" {
			t.Errorf("expected name 'test-tmpl', got %q", results[0].Name)
		}
	})

	t.Run("file protocol invalid path", func(t *testing.T) {
		r := NewRemoteTemplateRegistry("file:///nonexistent/path/registry.json")
		_, err := r.List()
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		dir := t.TempDir()
		regPath := filepath.Join(dir, "registry.json")
		if err := os.WriteFile(regPath, []byte("not json"), 0o600); err != nil {
			t.Fatal(err)
		}
		r := NewRemoteTemplateRegistry("file://" + regPath)
		_, err := r.List()
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("HTTP registry valid", func(t *testing.T) {
		entries := TemplateList{
			Templates: []TemplateEntry{
				{Name: "http-tmpl", Version: "2.0.0", Description: "http test", Author: "naeos"},
			},
		}
		data, _ := json.Marshal(entries)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(data)
		}))
		defer ts.Close()

		reg := NewRemoteTemplateRegistry(ts.URL + "/registry.json")
		results, err := reg.List()
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if len(results) != 1 || results[0].Name != "http-tmpl" {
			t.Errorf("unexpected results: %+v", results)
		}
	})

	t.Run("HTTP registry error status", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		reg := NewRemoteTemplateRegistry(ts.URL + "/registry.json")
		_, err := reg.List()
		if err == nil {
			t.Fatal("expected error for 500 status")
		}
	})
}

func TestRemoteTemplateRegistry_Search(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.json")
	entries := TemplateList{
		Templates: []TemplateEntry{
			{Name: "go-api", Description: "Go REST API", Author: "a", Tags: []string{"go", "rest"}},
			{Name: "py-ml", Description: "Python ML", Author: "b", Tags: []string{"python", "ml"}},
			{Name: "fullstack", Description: "JS fullstack", Author: "c", Tags: []string{"javascript", "node"}},
		},
	}
	data, _ := json.Marshal(entries)
	if err := os.WriteFile(regPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	r := NewRemoteTemplateRegistry("file://" + regPath)

	tests := []struct {
		name     string
		query    string
		want     int
		wantName string
	}{
		{"empty query", "", 3, ""},
		{"match name prefix", "go", 1, "go-api"},
		{"match description", "ML", 1, "py-ml"},
		{"match tag", "node", 1, "fullstack"},
		{"case insensitive", "GO", 1, "go-api"},
		{"no match", "nonexistent", 0, ""},
		{"partial tag", "jav", 1, "fullstack"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := r.Search(tt.query)
			if err != nil {
				t.Fatalf("Search(%q) error: %v", tt.query, err)
			}
			if len(results) != tt.want {
				t.Errorf("Search(%q) got %d results, want %d", tt.query, len(results), tt.want)
			}
			if tt.wantName != "" && len(results) > 0 && results[0].Name != tt.wantName {
				t.Errorf("Search(%q) first result name = %q, want %q", tt.query, results[0].Name, tt.wantName)
			}
		})
	}
}

func TestRemoteTemplateRegistry_Get(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.json")
	entries := TemplateList{
		Templates: []TemplateEntry{
			{Name: "tmpl-one", Version: "1.0.0", Description: "first"},
			{Name: "tmpl-two", Version: "2.0.0", Description: "second"},
		},
	}
	data, _ := json.Marshal(entries)
	if err := os.WriteFile(regPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	r := NewRemoteTemplateRegistry("file://" + regPath)

	t.Run("found", func(t *testing.T) {
		entry, err := r.Get("tmpl-one")
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}
		if entry.Name != "tmpl-one" || entry.Version != "1.0.0" {
			t.Errorf("unexpected entry: %+v", entry)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := r.Get("nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent template")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected 'not found' error, got: %v", err)
		}
	})
}

func TestValidateTemplateManifest(t *testing.T) {
	t.Run("valid manifest", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "template.yaml")
		content := []byte("name: my-template\nversion: \"1.0.0\"\ndescription: A test template\nauthor: tester\n")
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
		m, err := ValidateTemplateManifest(path)
		if err != nil {
			t.Fatalf("ValidateTemplateManifest() error: %v", err)
		}
		if m.Name != "my-template" || m.Version != "1.0.0" || m.Author != "tester" {
			t.Errorf("unexpected manifest: %+v", m)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := ValidateTemplateManifest("/nonexistent/path.yaml")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "template.yaml")
		if err := os.WriteFile(path, []byte(": : invalid"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := ValidateTemplateManifest(path)
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "template.yaml")
		if err := os.WriteFile(path, []byte("version: \"1.0.0\"\ndescription: test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := ValidateTemplateManifest(path)
		if err == nil || !strings.Contains(err.Error(), "name is required") {
			t.Errorf("expected 'name is required' error, got: %v", err)
		}
	})

	t.Run("missing version", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "template.yaml")
		if err := os.WriteFile(path, []byte("name: test\ndescription: test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := ValidateTemplateManifest(path)
		if err == nil || !strings.Contains(err.Error(), "version is required") {
			t.Errorf("expected 'version is required' error, got: %v", err)
		}
	})

	t.Run("missing description", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "template.yaml")
		if err := os.WriteFile(path, []byte("name: test\nversion: \"1.0.0\"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := ValidateTemplateManifest(path)
		if err == nil || !strings.Contains(err.Error(), "description is required") {
			t.Errorf("expected 'description is required' error, got: %v", err)
		}
	})
}

func TestPublishTemplate(t *testing.T) {
	t.Run("valid publish", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: my-tmpl\nversion: \"1.0.0\"\ndescription: test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# my-tmpl"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "src.go"), []byte("package main"), 0o600); err != nil {
			t.Fatal(err)
		}

		entry, err := PublishTemplate(dir, "")
		if err != nil {
			t.Fatalf("PublishTemplate() error: %v", err)
		}
		if entry.Name != "my-tmpl" || entry.Version != "1.0.0" {
			t.Errorf("unexpected entry: %+v", entry)
		}
	})

	t.Run("missing manifest", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "somefile.txt"), []byte("content"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := PublishTemplate(dir, "")
		if err == nil || !strings.Contains(err.Error(), "no template.yaml") {
			t.Errorf("expected 'no template.yaml' error, got: %v", err)
		}
	})

	t.Run("missing README", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: test\nversion: \"1.0.0\"\ndescription: test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := PublishTemplate(dir, "")
		if err == nil || !strings.Contains(err.Error(), "README.md") {
			t.Errorf("expected 'README.md' error, got: %v", err)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		dir := t.TempDir()
		_, err := PublishTemplate(dir, "")
		if err == nil || !strings.Contains(err.Error(), "empty") {
			t.Errorf("expected 'empty' error, got: %v", err)
		}
	})

	t.Run("publish with alt manifest name", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "naeos.yaml"), []byte("name: alt-tmpl\nversion: \"2.0.0\"\ndescription: alt test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# alt"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o600); err != nil {
			t.Fatal(err)
		}
		entry, err := PublishTemplate(dir, "")
		if err != nil {
			t.Fatalf("PublishTemplate() with naeos.yaml: %v", err)
		}
		if entry.Name != "alt-tmpl" {
			t.Errorf("expected name 'alt-tmpl', got %q", entry.Name)
		}
	})

	t.Run("publish to remote registry", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/api/publish") {
				w.WriteHeader(http.StatusCreated)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: remote-tmpl\nversion: \"1.0.0\"\ndescription: remote test\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# remote"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "app.go"), []byte("package main"), 0o600); err != nil {
			t.Fatal(err)
		}

		entry, err := PublishTemplate(dir, ts.URL+"/registry.json")
		if err != nil {
			t.Fatalf("PublishTemplate() to remote: %v", err)
		}
		if entry.Name != "remote-tmpl" {
			t.Errorf("expected 'remote-tmpl', got %q", entry.Name)
		}
	})

	t.Run("publish to remote registry error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		}))
		defer ts.Close()

		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: fail-tmpl\nversion: \"1.0.0\"\ndescription: fail\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# fail"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "fail.go"), []byte("package main"), 0o600); err != nil {
			t.Fatal(err)
		}

		_, err := PublishTemplate(dir, ts.URL+"/registry.json")
		if err == nil {
			t.Fatal("expected error for failed remote publish")
		}
	})
}

func TestGenerateRegistryEntry(t *testing.T) {
	t.Run("valid entry", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: gen-tmpl\nversion: \"3.0.0\"\ndescription: generated\nauthor: gen\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		entry, err := GenerateRegistryEntry(dir)
		if err != nil {
			t.Fatalf("GenerateRegistryEntry() error: %v", err)
		}
		if entry.Name != "gen-tmpl" || entry.Version != "3.0.0" || entry.Author != "gen" {
			t.Errorf("unexpected entry: %+v", entry)
		}
		if entry.RepoURL != "" || entry.DownloadURL != "" {
			t.Errorf("expected empty URLs for generated entry")
		}
	})

	t.Run("missing manifest", func(t *testing.T) {
		dir := t.TempDir()
		_, err := GenerateRegistryEntry(dir)
		if err == nil {
			t.Fatal("expected error for missing manifest")
		}
	})

	t.Run("alt manifest name", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "naeos.yaml"), []byte("name: naeos-tmpl\nversion: \"1.0.0\"\ndescription: from naeos.yaml\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		entry, err := GenerateRegistryEntry(dir)
		if err != nil {
			t.Fatalf("GenerateRegistryEntry() with naeos.yaml: %v", err)
		}
		if entry.Name != "naeos-tmpl" {
			t.Errorf("expected 'naeos-tmpl', got %q", entry.Name)
		}
	})
}

func TestDefaultTemplates(t *testing.T) {
	templates := DefaultTemplates()

	if len(templates) != 6 {
		t.Fatalf("expected 6 default templates, got %d", len(templates))
	}

	expectedNames := []string{
		"microservices-go",
		"serverless-ts",
		"web-api-py",
		"rust-cli-tool",
		"fullstack-js",
		"event-driven-java",
	}

	for _, name := range expectedNames {
		found := false
		for _, tmpl := range templates {
			if tmpl.Name == name {
				found = true
				if tmpl.Version == "" {
					t.Errorf("template %q has empty version", name)
				}
				if tmpl.Description == "" {
					t.Errorf("template %q has empty description", name)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected template %q not found in defaults", name)
		}
	}
}
