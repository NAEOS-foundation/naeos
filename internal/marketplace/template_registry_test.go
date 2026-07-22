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

func TestNewRemoteTemplateRegistryDefault(t *testing.T) {
	t.Parallel()
	r := NewRemoteTemplateRegistry("")
	if r.baseURL != DefaultTemplateRegistryURL {
		t.Errorf("expected default URL, got %q", r.baseURL)
	}
	if r.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}
}

func TestNewRemoteTemplateRegistryCustom(t *testing.T) {
	t.Parallel()
	r := NewRemoteTemplateRegistry("http://example.com/registry")
	if r.baseURL != "http://example.com/registry" {
		t.Errorf("expected custom URL, got %q", r.baseURL)
	}
}

func TestTemplateRegistryListHTTP(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "alpha", Version: "1.0.0", Description: "Alpha template"},
		{Name: "beta", Version: "2.0.0", Description: "Beta template"},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	templates, err := r.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestTemplateRegistryListFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	data := TemplateList{Templates: []TemplateEntry{
		{Name: "gamma", Version: "1.0.0", Description: "Gamma"},
	}}
	b, _ := json.Marshal(data)
	fpath := filepath.Join(tmp, "registry.json")
	os.WriteFile(fpath, b, 0o644)

	r := NewRemoteTemplateRegistry("file://" + fpath)
	templates, err := r.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(templates))
	}
}

func TestTemplateRegistryListFileNotFound(t *testing.T) {
	t.Parallel()
	r := NewRemoteTemplateRegistry("file:///nonexistent/registry.json")
	_, err := r.List()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read local registry") {
		t.Errorf("expected read local registry error, got: %v", err)
	}
}

func TestTemplateRegistryListInvalidJSON(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "registry.json"), []byte("{invalid json"), 0o644)
	r := NewRemoteTemplateRegistry("file://" + filepath.Join(tmp, "registry.json"))
	_, err := r.List()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestTemplateRegistryListNonOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	_, err := r.List()
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got: %v", err)
	}
}

func TestTemplateRegistryListEmpty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TemplateList{})
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	templates, err := r.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(templates) != 0 {
		t.Errorf("expected 0 templates, got %d", len(templates))
	}
}

func TestTemplateRegistrySearchEmptyQuery(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "alpha", Version: "1.0.0"},
		{Name: "beta", Version: "2.0.0"},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	results, err := r.Search("")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestTemplateRegistrySearchByName(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "go-micro", Version: "1.0.0", Description: "Go microservices", Tags: []string{"go"}},
		{Name: "py-web", Version: "1.0.0", Description: "Python web", Tags: []string{"python"}},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	results, err := r.Search("go")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].Name != "go-micro" {
		t.Errorf("expected go-micro, got %v", results)
	}
}

func TestTemplateRegistrySearchByDescription(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "a", Version: "1.0.0", Description: "Machine learning starter"},
		{Name: "b", Version: "1.0.0", Description: "Web API starter"},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	results, err := r.Search("learning")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].Name != "a" {
		t.Errorf("expected a, got %v", results)
	}
}

func TestTemplateRegistrySearchByTag(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "a", Version: "1.0.0", Description: "A", Tags: []string{"grpc", "go"}},
		{Name: "b", Version: "1.0.0", Description: "B", Tags: []string{"python"}},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	results, err := r.Search("grpc")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].Name != "a" {
		t.Errorf("expected a, got %v", results)
	}
}

func TestTemplateRegistrySearchNoMatch(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "alpha", Version: "1.0.0"},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	results, err := r.Search("zzz-nonexistent")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestTemplateRegistryGetFound(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "mytemplate", Version: "3.0.0"},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	entry, err := r.Get("mytemplate")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if entry.Version != "3.0.0" {
		t.Errorf("expected version 3.0.0, got %q", entry.Version)
	}
}

func TestTemplateRegistryGetNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(TemplateList{})
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestValidateTemplateManifestValid(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	content := `name: my-template
version: 1.0.0
description: A test template
author: Test
tags: [go, test]
languages: [go]
frameworks: [grpc]
`
	fpath := filepath.Join(tmp, "template.yaml")
	os.WriteFile(fpath, []byte(content), 0o644)

	m, err := ValidateTemplateManifest(fpath)
	if err != nil {
		t.Fatalf("ValidateTemplateManifest() error = %v", err)
	}
	if m.Name != "my-template" {
		t.Errorf("expected name 'my-template', got %q", m.Name)
	}
	if m.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", m.Version)
	}
}

func TestValidateTemplateManifestMissingFile(t *testing.T) {
	t.Parallel()
	_, err := ValidateTemplateManifest("/nonexistent/template.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidateTemplateManifestInvalidYAML(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "template.yaml")
	os.WriteFile(fpath, []byte("{{invalid yaml"), 0o644)
	_, err := ValidateTemplateManifest(fpath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestValidateTemplateManifestMissingName(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "template.yaml")
	os.WriteFile(fpath, []byte("version: 1.0.0\ndescription: test"), 0o644)
	_, err := ValidateTemplateManifest(fpath)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestValidateTemplateManifestMissingVersion(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "template.yaml")
	os.WriteFile(fpath, []byte("name: test\ndescription: test"), 0o644)
	_, err := ValidateTemplateManifest(fpath)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestValidateTemplateManifestMissingDescription(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "template.yaml")
	os.WriteFile(fpath, []byte("name: test\nversion: 1.0.0"), 0o644)
	_, err := ValidateTemplateManifest(fpath)
	if err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestPublishTemplateWithTemplateYAML(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: my-tpl
version: 1.0.0
description: My template
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# My Template"), 0o644)

	entry, err := PublishTemplate(tmp, "")
	if err != nil {
		t.Fatalf("PublishTemplate() error = %v", err)
	}
	if entry.Name != "my-tpl" {
		t.Errorf("expected name 'my-tpl', got %q", entry.Name)
	}
}

func TestPublishTemplateWithNaeosYAML(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: alt-tpl
version: 2.0.0
description: Alt template
`
	os.WriteFile(filepath.Join(tmp, "naeos.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Alt"), 0o644)

	entry, err := PublishTemplate(tmp, "")
	if err != nil {
		t.Fatalf("PublishTemplate() error = %v", err)
	}
	if entry.Name != "alt-tpl" {
		t.Errorf("expected name 'alt-tpl', got %q", entry.Name)
	}
}

func TestPublishTemplateMissingManifest(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	_, err := PublishTemplate(tmp, "")
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestPublishTemplateMissingREADME(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: tpl
version: 1.0.0
description: desc
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)

	_, err := PublishTemplate(tmp, "")
	if err == nil {
		t.Fatal("expected error for missing README.md")
	}
}

func TestPublishTemplateWithNaeosDir(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: tpl
version: 1.0.0
description: desc
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Tpl"), 0o644)
	os.MkdirAll(filepath.Join(tmp, ".naeos"), 0o755)

	entry, err := PublishTemplate(tmp, "")
	if err != nil {
		t.Fatalf("PublishTemplate() error = %v", err)
	}
	if entry.Name != "tpl" {
		t.Errorf("expected name 'tpl', got %q", entry.Name)
	}
}

func TestPublishTemplateSubmitToRegistry(t *testing.T) {
	t.Parallel()
	var receivedEntry TemplateEntry
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&receivedEntry)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	manifest := `name: tpl
version: 1.0.0
description: desc
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Tpl"), 0o644)

	registryURL := srv.URL + "/registry.json"
	entry, err := PublishTemplate(tmp, registryURL)
	if err != nil {
		t.Fatalf("PublishTemplate() error = %v", err)
	}
	if entry.Name != "tpl" {
		t.Errorf("expected name 'tpl', got %q", entry.Name)
	}
}

func TestPublishTemplateSubmitToRegistryError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	manifest := `name: tpl
version: 1.0.0
description: desc
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Tpl"), 0o644)

	registryURL := srv.URL + "/registry.json"
	_, err := PublishTemplate(tmp, registryURL)
	if err == nil {
		t.Fatal("expected error from registry")
	}
}

func TestGenerateRegistryEntryWithTemplateYAML(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: gen-tpl
version: 1.0.0
description: Generated
author: Test
tags: [gen]
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)

	entry, err := GenerateRegistryEntry(tmp)
	if err != nil {
		t.Fatalf("GenerateRegistryEntry() error = %v", err)
	}
	if entry.Name != "gen-tpl" {
		t.Errorf("expected name 'gen-tpl', got %q", entry.Name)
	}
	if entry.UpdatedAt == "" {
		t.Error("expected non-empty UpdatedAt")
	}
}

func TestGenerateRegistryEntryWithNaeosYAML(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: alt
version: 2.0.0
description: Alt
`
	os.WriteFile(filepath.Join(tmp, "naeos.yaml"), []byte(manifest), 0o644)

	entry, err := GenerateRegistryEntry(tmp)
	if err != nil {
		t.Fatalf("GenerateRegistryEntry() error = %v", err)
	}
	if entry.Name != "alt" {
		t.Errorf("expected name 'alt', got %q", entry.Name)
	}
}

func TestGenerateRegistryEntryNotFound(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	_, err := GenerateRegistryEntry(tmp)
	if err == nil {
		t.Fatal("expected error for missing template.yaml")
	}
}

func TestDefaultTemplates(t *testing.T) {
	t.Parallel()
	templates := DefaultTemplates()
	if len(templates) == 0 {
		t.Fatal("expected non-empty default templates")
	}
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("expected non-empty name")
		}
		if tmpl.Version == "" {
			t.Errorf("expected non-empty version for %s", tmpl.Name)
		}
	}
}

func TestSubmitToRegistryCreatedStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	entry := TemplateEntry{Name: "test", Version: "1.0.0"}
	err := submitToRegistry(entry, srv.URL+"/registry.json")
	if err != nil {
		t.Fatalf("submitToRegistry() error = %v", err)
	}
}

func TestSubmitToRegistryNonOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	entry := TemplateEntry{Name: "test", Version: "1.0.0"}
	err := submitToRegistry(entry, srv.URL+"/registry.json")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestTemplateRegistrySearchCaseInsensitive(t *testing.T) {
	t.Parallel()
	list := TemplateList{Templates: []TemplateEntry{
		{Name: "GoMicro", Version: "1.0.0", Description: "A template", Tags: []string{"Go"}},
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	}))
	defer srv.Close()

	r := NewRemoteTemplateRegistry(srv.URL)
	results, err := r.Search("go")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestTemplateRegistryGetFromLocalFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	data := TemplateList{Templates: []TemplateEntry{
		{Name: "local-tpl", Version: "1.0.0"},
	}}
	b, _ := json.Marshal(data)
	fpath := filepath.Join(tmp, "registry.json")
	os.WriteFile(fpath, b, 0o644)

	r := NewRemoteTemplateRegistry("file://" + fpath)
	entry, err := r.Get("local-tpl")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if entry.Name != "local-tpl" {
		t.Errorf("expected local-tpl, got %q", entry.Name)
	}
}

func TestTemplateRegistrySearchFromLocalFile(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	data := TemplateList{Templates: []TemplateEntry{
		{Name: "a", Version: "1.0.0", Description: "Alpha"},
		{Name: "b", Version: "1.0.0", Description: "Beta"},
	}}
	b, _ := json.Marshal(data)
	fpath := filepath.Join(tmp, "registry.json")
	os.WriteFile(fpath, b, 0o644)

	r := NewRemoteTemplateRegistry("file://" + fpath)
	results, err := r.Search("Alpha")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestTemplateRegistrySearchByTagFromLocal(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	data := TemplateList{Templates: []TemplateEntry{
		{Name: "a", Version: "1.0.0", Tags: []string{"grpc", "go"}},
		{Name: "b", Version: "1.0.0", Tags: []string{"python"}},
	}}
	b, _ := json.Marshal(data)
	fpath := filepath.Join(tmp, "registry.json")
	os.WriteFile(fpath, b, 0o644)

	r := NewRemoteTemplateRegistry("file://" + fpath)
	results, err := r.Search("python")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].Name != "b" {
		t.Errorf("expected b, got %v", results)
	}
}

func TestPublishTemplateInvalidManifest(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte("{{bad yaml"), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Readme"), 0o644)

	_, err := PublishTemplate(tmp, "")
	if err == nil {
		t.Fatal("expected error for invalid manifest")
	}
}

func TestGenerateRegistryEntryInvalidManifest(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte("{{bad yaml"), 0o644)

	_, err := GenerateRegistryEntry(tmp)
	if err == nil {
		t.Fatal("expected error for invalid manifest")
	}
}

func TestGenerateRegistryEntryMissingName(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte("version: 1.0.0\ndescription: test"), 0o644)

	_, err := GenerateRegistryEntry(tmp)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestTemplateRegistrySearchFromLocalFileNotFound(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	data := TemplateList{Templates: []TemplateEntry{
		{Name: "alpha", Version: "1.0.0"},
	}}
	b, _ := json.Marshal(data)
	fpath := filepath.Join(tmp, "registry.json")
	os.WriteFile(fpath, b, 0o644)

	r := NewRemoteTemplateRegistry("file://" + fpath)
	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestTemplateRegistrySearchTagNoMatch(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	data := TemplateList{Templates: []TemplateEntry{
		{Name: "a", Version: "1.0.0", Tags: []string{"go"}},
	}}
	b, _ := json.Marshal(data)
	fpath := filepath.Join(tmp, "registry.json")
	os.WriteFile(fpath, b, 0o644)

	r := NewRemoteTemplateRegistry("file://" + fpath)
	results, err := r.Search("rust")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestPublishTemplateRegistryDefaultURL(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: tpl
version: 1.0.0
description: desc
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Tpl"), 0o644)

	entry, err := PublishTemplate(tmp, DefaultTemplateRegistryURL)
	if err != nil {
		t.Fatalf("PublishTemplate() error = %v", err)
	}
	if entry.Name != "tpl" {
		t.Errorf("expected name 'tpl', got %q", entry.Name)
	}
}

func TestPublishTemplateRegistryFileURL(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: tpl
version: 1.0.0
description: desc
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Tpl"), 0o644)

	entry, err := PublishTemplate(tmp, "file://"+filepath.Join(tmp, "registry.json"))
	if err != nil {
		t.Fatalf("PublishTemplate() error = %v", err)
	}
	if entry.Name != "tpl" {
		t.Errorf("expected name 'tpl', got %q", entry.Name)
	}
}

func TestPublishTemplateSubmitToRegistryConnectionError(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	manifest := `name: tpl
version: 1.0.0
description: desc
`
	os.WriteFile(filepath.Join(tmp, "template.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(tmp, "README.md"), []byte("# Tpl"), 0o644)

	_, err := PublishTemplate(tmp, "http://127.0.0.1:1/registry.json")
	if err == nil {
		t.Fatal("expected error for connection failure")
	}
}

func TestTemplateRegistryListBadFileURL(t *testing.T) {
	t.Parallel()
	r := NewRemoteTemplateRegistry("file://")
	_, err := r.List()
	if err == nil {
		t.Fatal("expected error for bad file URL")
	}
}
