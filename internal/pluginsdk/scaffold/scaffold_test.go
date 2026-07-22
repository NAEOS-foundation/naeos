package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAll(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := Files{
		Dir:     filepath.Join(dir, "myplugin"),
		Module:  "github.com/example/myplugin",
		Name:    "myplugin",
		Author:  "testuser",
		Desc:    "A test plugin",
		UseWASM: true,
	}

	if err := f.WriteAll(); err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	expectedFiles := []string{
		"naeos.yaml",
		"plugin.go",
		"plugin_test.go",
		"main.go",
		"Makefile",
		".github/workflows/ci.yml",
		"README.md",
		"go.mod",
	}

	for _, name := range expectedFiles {
		path := filepath.Join(f.Dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("missing file %s: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty", name)
		}
	}
}

func TestWriteAllContainsModuleName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := Files{
		Dir:     filepath.Join(dir, "testmod"),
		Module:  "github.com/example/testmod",
		Name:    "testmod",
		Author:  "author",
		Desc:    "desc",
		UseWASM: false,
	}

	if err := f.WriteAll(); err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	gomod, err := os.ReadFile(filepath.Join(f.Dir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !containsStr(string(gomod), f.Module) {
		t.Errorf("go.mod missing module name %q", f.Module)
	}

	yaml, err := os.ReadFile(filepath.Join(f.Dir, "naeos.yaml"))
	if err != nil {
		t.Fatalf("read naeos.yaml: %v", err)
	}
	content := string(yaml)
	if !containsStr(content, f.Name) {
		t.Errorf("naeos.yaml missing plugin name %q", f.Name)
	}
	if !containsStr(content, f.Desc) {
		t.Errorf("naeos.yaml missing description %q", f.Desc)
	}
	if !containsStr(content, f.Author) {
		t.Errorf("naeos.yaml missing author %q", f.Author)
	}
}

func TestWriteAllPluginGo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := Files{
		Dir:     filepath.Join(dir, "plug"),
		Module:  "github.com/test/plug",
		Name:    "plug",
		Author:  "tester",
		Desc:    "test plugin",
		UseWASM: true,
	}

	if err := f.WriteAll(); err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(f.Dir, "plugin.go"))
	if err != nil {
		t.Fatalf("read plugin.go: %v", err)
	}
	content := string(data)
	if !containsStr(content, "package main") {
		t.Error("plugin.go missing package main")
	}
	if !containsStr(content, `"plug"`) {
		t.Error("plugin.go missing plugin name")
	}
	if !containsStr(content, `"test plugin"`) {
		t.Error("plugin.go missing plugin description")
	}
}

func TestWriteAllREADME(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := Files{
		Dir:     filepath.Join(dir, "readme-test"),
		Module:  "github.com/test/readme-test",
		Name:    "readme-test",
		Author:  "tester",
		Desc:    "A readme test plugin",
		UseWASM: true,
	}

	if err := f.WriteAll(); err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(f.Dir, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	content := string(data)
	if !containsStr(content, "readme-test") {
		t.Error("README.md missing plugin name")
	}
	if !containsStr(content, "A readme test plugin") {
		t.Error("README.md missing description")
	}
}

func TestWriteAllInvalidDir(t *testing.T) {
	t.Parallel()

	f := Files{
		Dir:    "/nonexistent/deeply/nested/dir",
		Module: "test",
		Name:   "test",
		Author: "test",
		Desc:   "test",
	}

	err := f.WriteAll()
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

func TestGoModContent(t *testing.T) {
	t.Parallel()

	f := Files{Module: "github.com/example/mymod"}
	got := f.goMod()
	if !containsStr(got, "module github.com/example/mymod") {
		t.Errorf("goMod() missing module declaration, got:\n%s", got)
	}
}

func TestNaeosYAMLContent(t *testing.T) {
	t.Parallel()

	f := Files{Name: "my-plugin", Desc: "my desc", Author: "me"}
	got := f.naeosYAML()
	if !containsStr(got, "name: my-plugin") {
		t.Errorf("naeosYAML() missing name, got:\n%s", got)
	}
	if !containsStr(got, "my desc") {
		t.Errorf("naeosYAML() missing description, got:\n%s", got)
	}
	if !containsStr(got, "me") {
		t.Errorf("naeosYAML() missing author, got:\n%s", got)
	}
}

func TestWriteAllGitHubWorkflow(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := Files{
		Dir:     filepath.Join(dir, "ci-test"),
		Module:  "github.com/test/ci-test",
		Name:    "ci-test",
		Author:  "tester",
		Desc:    "ci test",
		UseWASM: true,
	}

	if err := f.WriteAll(); err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(f.Dir, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("read ci.yml: %v", err)
	}
	content := string(data)
	if !containsStr(content, "name: CI") {
		t.Error("ci.yml missing CI name")
	}
	if !containsStr(content, "go test") {
		t.Error("ci.yml missing go test command")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
