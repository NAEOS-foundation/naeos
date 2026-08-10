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

func writeTemplateRegistry(t *testing.T, dir string, templates []marketplace.TemplateEntry) string {
	t.Helper()
	data, err := json.Marshal(marketplace.TemplateList{Templates: templates})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "registry.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func sampleTemplates() []marketplace.TemplateEntry {
	return []marketplace.TemplateEntry{
		{
			Name:        "go-api",
			Version:     "1.2.0",
			Description: "Go REST API starter",
			Author:      "NAEOS Foundation",
			Tags:        []string{"go", "rest"},
			Languages:   []string{"go"},
			RepoURL:     "https://github.com/naeos-templates/go-api",
		},
		{
			Name:        "py-worker",
			Version:     "0.5.0",
			Description: "Python background worker",
			Author:      "NAEOS Foundation",
			Tags:        []string{"python", "worker"},
			Languages:   []string{"python"},
		},
	}
}

func TestTemplateListKindFilters(t *testing.T) {
	root := NewRootCommand()
	for _, tc := range []struct {
		kind string
		want string
	}{
		{"prompt-llm", "LLM Prompt Templates"},
		{"prompt-compiler", "Compiler Templates"},
		{"code", "Code Generation Templates"},
	} {
		output, err := executeCommand(root, "template", "list", "--kind", tc.kind)
		if err != nil {
			t.Fatalf("list kind %s: %v", tc.kind, err)
		}
		if !strings.Contains(output, tc.want) {
			t.Errorf("kind %s: expected %q in output, got %q", tc.kind, tc.want, output)
		}
	}

	output, err := executeCommand(root, "template", "list", "--kind", "unknown-kind")
	if err != nil {
		t.Fatalf("list unknown kind: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output for unknown kind, got %q", output)
	}
}

func TestTemplateShowPrompt(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "template", "show", "enrich-spec")
	if err != nil {
		t.Fatalf("show enrich-spec: %v", err)
	}
	if !strings.Contains(output, "Kind:        llm") {
		t.Errorf("expected llm kind, got %q", output)
	}
	if !strings.Contains(output, "System Prompt:") {
		t.Errorf("expected system prompt section, got %q", output)
	}
}

func TestTemplateShowCompiler(t *testing.T) {
	root := NewRootCommand()
	output, err := executeCommand(root, "template", "show", "copilot")
	if err != nil {
		t.Fatalf("show copilot: %v", err)
	}
	if !strings.Contains(output, "Kind:    compiler") {
		t.Errorf("expected compiler kind, got %q", output)
	}
	if !strings.Contains(output, "Output Files:") {
		t.Errorf("expected output files section, got %q", output)
	}
}

func TestTemplateShowNotFound(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "template", "show", "does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTemplateAddRemoveCustom(t *testing.T) {
	dir := t.TempDir()
	root := NewRootCommand()
	output, err := executeCommand(root, "template", "--templates-dir", dir, "add", "my-tpl", "hello {{.Name}}")
	if err != nil {
		t.Fatalf("add template: %v", err)
	}
	if !strings.Contains(output, "Added template my-tpl") {
		t.Errorf("unexpected output: %q", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "my-tpl.tmpl")); err != nil {
		t.Fatalf("template file not created: %v", err)
	}

	output, err = executeCommand(root, "template", "--templates-dir", dir, "remove", "my-tpl")
	if err != nil {
		t.Fatalf("remove template: %v", err)
	}
	if !strings.Contains(output, "Removed template my-tpl") {
		t.Errorf("unexpected output: %q", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "my-tpl.tmpl")); !os.IsNotExist(err) {
		t.Errorf("template file should be removed")
	}

	_, err = executeCommand(root, "template", "--templates-dir", dir, "remove", "my-tpl")
	if err == nil {
		t.Error("expected error removing missing template")
	}
}

func TestTemplatePromptCreate(t *testing.T) {
	dir := t.TempDir()
	root := NewRootCommand()
	output, err := executeCommand(root, "template", "--templates-dir", dir, "prompt-create", "custom-prompt",
		"--system", "You are helpful", "--user", "Analyze: {{.SpecContent}}", "--provider", "anthropic", "--description", "custom desc")
	if err != nil {
		t.Fatalf("prompt-create: %v", err)
	}
	if !strings.Contains(output, "Created prompt template") {
		t.Errorf("unexpected output: %q", output)
	}
	data, err := os.ReadFile(filepath.Join(dir, "prompts", "custom-prompt.yaml"))
	if err != nil {
		t.Fatalf("prompt file not created: %v", err)
	}
	if !strings.Contains(string(data), "provider: anthropic") {
		t.Errorf("expected provider in prompt file, got %q", string(data))
	}

	_, err = executeCommand(NewRootCommand(), "template", "--templates-dir", dir, "prompt-create", "no-user")
	if err == nil {
		t.Error("expected error when --user is missing")
	}

	output, err = executeCommand(root, "template", "--templates-dir", dir, "prompt-remove", "custom-prompt")
	if err != nil {
		t.Fatalf("prompt-remove: %v", err)
	}
	if !strings.Contains(output, "Removed prompt template: custom-prompt") {
		t.Errorf("unexpected output: %q", output)
	}

	_, err = executeCommand(root, "template", "--templates-dir", dir, "prompt-remove", "custom-prompt")
	if err == nil {
		t.Error("expected error removing missing prompt")
	}
}

func TestTemplatePublishErrors(t *testing.T) {
	root := NewRootCommand()

	_, err := executeCommand(root, "template", "publish", "/nonexistent/template-dir")
	if err == nil {
		t.Error("expected error for missing directory")
	}

	file := filepath.Join(t.TempDir(), "not-a-dir.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = executeCommand(root, "template", "publish", file)
	if err == nil {
		t.Error("expected error for file argument")
	}

	dir := t.TempDir()
	_, err = executeCommand(root, "template", "publish", dir)
	if err == nil {
		t.Error("expected error for missing manifest")
	}

	if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: t\nversion: 1.0.0\ndescription: d\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = executeCommand(root, "template", "publish", dir)
	if err == nil {
		t.Error("expected error for missing README")
	}

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# T\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = executeCommand(root, "template", "publish", dir)
	if err == nil {
		t.Error("expected error for invalid manifest (missing version/description)")
	}
}

func TestTemplatePublishLocalAndJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "naeos.yaml"), []byte("name: my-starter\nversion: 0.1.0\ndescription: starter project\nauthor: me\nlanguages: [go]\ntags: [api]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# My Starter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRootCommand()
	output, err := executeCommand(root, "template", "publish", dir, "--registry", "file://"+filepath.Join(t.TempDir(), "reg.json"))
	if err != nil {
		t.Fatalf("publish local: %v", err)
	}
	if !strings.Contains(output, "published") || !strings.Contains(output, "my-starter") {
		t.Errorf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "add to site/static/templates/registry.json") {
		t.Errorf("expected PR hint for file:// registry, got %q", output)
	}

	output, err = executeCommand(root, "template", "publish", dir, "--json")
	if err != nil {
		t.Fatalf("publish json: %v", err)
	}
	var entry marketplace.TemplateEntry
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if entry.Name != "my-starter" || entry.Version != "0.1.0" {
		t.Errorf("unexpected entry: %+v", entry)
	}
}

func TestTemplatePublishRemote(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte("name: remote-starter\nversion: 1.0.0\ndescription: remote starter\nauthor: me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/publish" {
			http.Error(w, "wrong path", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	root := NewRootCommand()
	output, err := executeCommand(root, "template", "publish", dir, "--registry", server.URL+"/registry.json")
	if err != nil {
		t.Fatalf("publish remote: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 publish call, got %d", calls)
	}
	if !strings.Contains(output, "Registry:    "+server.URL+"/registry.json") {
		t.Errorf("expected registry URL in output, got %q", output)
	}

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer badServer.Close()
	_, err = executeCommand(root, "template", "publish", dir, "--registry", badServer.URL+"/registry.json")
	if err == nil {
		t.Error("expected error when registry rejects publish")
	}
}

func TestTemplateSearchFileRegistry(t *testing.T) {
	regPath := writeTemplateRegistry(t, t.TempDir(), sampleTemplates())
	root := NewRootCommand()

	output, err := executeCommand(root, "template", "search", "go", "--registry", "file://"+regPath)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(output, "Found 1 template(s)") || !strings.Contains(output, "go-api") {
		t.Errorf("unexpected output: %q", output)
	}

	output, err = executeCommand(root, "template", "search", "python", "--registry", "file://"+regPath, "--output", "json")
	if err != nil {
		t.Fatalf("search json: %v", err)
	}
	if !strings.Contains(output, `"query": "python"`) || !strings.Contains(output, "py-worker") {
		t.Errorf("unexpected json output: %q", output)
	}

	output, err = executeCommand(NewRootCommand(), "template", "search", "nothing-matches", "--registry", "file://"+regPath, "--output", "text")
	if err != nil {
		t.Fatalf("search no results: %v", err)
	}
	if !strings.Contains(output, "No templates found") {
		t.Errorf("unexpected no-result output: %q", output)
	}

	_, err = executeCommand(root, "template", "search", "x", "--registry", "file://"+filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Error("expected error when registry file missing")
	}
}

func TestTemplateSearchRegistryHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	root := NewRootCommand()
	_, err := executeCommand(root, "template", "search", "go", "--registry", server.URL+"/registry.json")
	if err == nil {
		t.Error("expected error from failing registry")
	}
}

func TestTemplateInit(t *testing.T) {
	regPath := writeTemplateRegistry(t, t.TempDir(), sampleTemplates())
	root := NewRootCommand()

	output, err := executeCommand(root, "template", "init", "go-api", "--registry", "file://"+regPath, "--output", "./my-project")
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if !strings.Contains(output, "git clone https://github.com/naeos-templates/go-api ./my-project") {
		t.Errorf("expected clone instructions, got %q", output)
	}

	_, err = executeCommand(root, "template", "init", "py-worker", "--registry", "file://"+regPath)
	if err == nil {
		t.Error("expected error when template has no download URL")
	}

	_, err = executeCommand(root, "template", "init", "missing-tpl", "--registry", "file://"+regPath)
	if err == nil {
		t.Error("expected error when template not in registry")
	}
}

func TestWorkspaceInitAddListRemoveLock(t *testing.T) {
	root := t.TempDir()
	cli := NewRootCommand()

	output, err := executeCommand(cli, "workspace", "--root", root, "init", "my-ws")
	if err != nil {
		t.Fatalf("workspace init: %v", err)
	}
	if !strings.Contains(output, "Initialized workspace my-ws") {
		t.Errorf("unexpected output: %q", output)
	}

	_, err = executeCommand(cli, "workspace", "--root", root, "init", "")
	if err == nil {
		t.Error("expected error for empty workspace name")
	}

	output, err = executeCommand(cli, "workspace", "--root", root, "add", "core", "./core")
	if err != nil {
		t.Fatalf("workspace add: %v", err)
	}
	if !strings.Contains(output, "Added module core") {
		t.Errorf("unexpected output: %q", output)
	}

	output, err = executeCommand(cli, "workspace", "--root", root, "list")
	if err != nil {
		t.Fatalf("workspace list: %v", err)
	}
	if !strings.Contains(output, "core") || !strings.Contains(output, "my-ws") {
		t.Errorf("expected modules in list, got %q", output)
	}

	output, err = executeCommand(cli, "workspace", "--root", root, "info")
	if err != nil {
		t.Fatalf("workspace info: %v", err)
	}
	if !strings.Contains(output, "Workspace Information") || !strings.Contains(output, "Modules:  2") {
		t.Errorf("unexpected info output: %q", output)
	}

	output, err = executeCommand(cli, "workspace", "--root", root, "lock")
	if err != nil {
		t.Fatalf("workspace lock: %v", err)
	}
	if !strings.Contains(output, "2 modules locked") {
		t.Errorf("unexpected lock output: %q", output)
	}
	lockData, err := os.ReadFile(filepath.Join(root, "naeos.lock"))
	if err != nil {
		t.Fatalf("lockfile not created: %v", err)
	}
	if !strings.Contains(string(lockData), "# NAEOS Workspace Lockfile") {
		t.Errorf("unexpected lockfile content: %q", string(lockData))
	}

	output, err = executeCommand(cli, "workspace", "--root", root, "remove", "core")
	if err != nil {
		t.Fatalf("workspace remove: %v", err)
	}
	if !strings.Contains(output, "Removed module core") {
		t.Errorf("unexpected output: %q", output)
	}

	_, err = executeCommand(cli, "workspace", "--root", root, "remove", "ghost")
	if err == nil {
		t.Error("expected error removing unknown module")
	}

	output, err = executeCommand(cli, "workspace", "--root", root, "list")
	if err != nil {
		t.Fatalf("workspace list after remove: %v", err)
	}
	if strings.Contains(output, "core") {
		t.Errorf("removed module still listed: %q", output)
	}
}

func TestWorkspaceListEmpty(t *testing.T) {
	cli := NewRootCommand()
	output, err := executeCommand(cli, "workspace", "--root", t.TempDir(), "list")
	if err != nil {
		t.Fatalf("workspace list empty: %v", err)
	}
	if !strings.Contains(output, "No modules in workspace") {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestWorkspaceInfoNoModules(t *testing.T) {
	cli := NewRootCommand()
	output, err := executeCommand(cli, "workspace", "--root", t.TempDir(), "info")
	if err != nil {
		t.Fatalf("workspace info: %v", err)
	}
	if !strings.Contains(output, "Modules:  0") {
		t.Errorf("expected zero modules, got %q", output)
	}
}

func TestWorkspaceLockEmpty(t *testing.T) {
	dir := t.TempDir()
	cli := NewRootCommand()
	output, err := executeCommand(cli, "workspace", "--root", dir, "lock")
	if err != nil {
		t.Fatalf("workspace lock: %v", err)
	}
	if !strings.Contains(output, "0 modules locked") {
		t.Errorf("unexpected output: %q", output)
	}
	if _, err := os.Stat(filepath.Join(dir, "naeos.lock")); err != nil {
		t.Fatalf("lockfile not created: %v", err)
	}
}

func TestResolveInputFile(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(existing, []byte("project: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := resolveInputFile("", ""); err == nil {
		t.Error("expected error when no flags given")
	}

	if _, err := resolveInputFile("", filepath.Join(dir, "missing.yaml")); err == nil {
		t.Error("expected error for missing input file")
	}

	path, err := resolveInputFile("", existing)
	if err != nil || path != existing {
		t.Errorf("expected %q, got %q (err %v)", existing, path, err)
	}

	path, err = resolveInputFile(filepath.Join(dir, "ghost.yaml"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path for missing --input, got %q", path)
	}

	path, err = resolveInputFile(existing, "")
	if err != nil || path != existing {
		t.Errorf("expected %q from --input, got %q (err %v)", existing, path, err)
	}
}

func TestWatchNoInputFlag(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "watch")
	if err == nil {
		t.Fatal("expected error when watch invoked without --input/--input-file")
	}
	if !strings.Contains(err.Error(), "specify --input or --input-file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWatchMissingInputFile(t *testing.T) {
	root := NewRootCommand()
	_, err := executeCommand(root, "watch", "--input-file", "/nonexistent/spec.yaml")
	if err == nil {
		t.Fatal("expected error for missing input file")
	}
	if !strings.Contains(err.Error(), "input file not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWatchBadConfig(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "spec.yaml")
	if err := os.WriteFile(spec, []byte("project: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := NewRootCommand()
	_, err := executeCommand(root, "watch", "--input-file", spec, "--config", "/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !strings.Contains(err.Error(), "config") {
		t.Errorf("unexpected error: %v", err)
	}
}
