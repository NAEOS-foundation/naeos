package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testFiles() Files {
	return Files{
		Dir:     "/tmp/test-plugin",
		Module:  "github.com/test/plugin",
		Name:    "test-plugin",
		Author:  "test-author",
		Desc:    "A test plugin",
		UseWASM: true,
	}
}

func TestGoMod(t *testing.T) {
	f := testFiles()
	content := f.goMod()
	if !strings.Contains(content, "module github.com/test/plugin") {
		t.Error("expected module declaration")
	}
	if !strings.Contains(content, "go 1.25.0") {
		t.Error("expected go version")
	}
}

func TestNaeosYAML(t *testing.T) {
	f := testFiles()
	content := f.naeosYAML()
	if !strings.Contains(content, "name: test-plugin") {
		t.Error("expected plugin name")
	}
	if !strings.Contains(content, "author: test-author") {
		t.Error("expected author")
	}
	if !strings.Contains(content, "type: wasm") {
		t.Error("expected type wasm")
	}
	if !strings.Contains(content, "version: 0.1.0") {
		t.Error("expected version")
	}
}

func TestPluginGo(t *testing.T) {
	f := testFiles()
	content := f.pluginGo()
	if !strings.Contains(content, "package main") {
		t.Error("expected package main")
	}
	if !strings.Contains(content, `"test-plugin"`) {
		t.Error("expected plugin name in code")
	}
	if !strings.Contains(content, `"A test plugin"`) {
		t.Error("expected description in code")
	}
	if !strings.Contains(content, "func (p *Plugin) Initialize") {
		t.Error("expected Initialize method")
	}
	if !strings.Contains(content, "func (p *Plugin) Execute") {
		t.Error("expected Execute method")
	}
	if !strings.Contains(content, "func (p *Plugin) Shutdown") {
		t.Error("expected Shutdown method")
	}
	if !strings.Contains(content, `"ping"`) {
		t.Error("expected ping action handler")
	}
	if !strings.Contains(content, `"describe"`) {
		t.Error("expected describe action handler")
	}
}

func TestWasmMainGo(t *testing.T) {
	f := testFiles()
	content := f.wasmMainGo()
	if !strings.Contains(content, "package main") {
		t.Error("expected package main")
	}
	if !strings.Contains(content, `"ping"`) {
		t.Error("expected ping action")
	}
	if !strings.Contains(content, `"describe"`) {
		t.Error("expected describe action")
	}
	if !strings.Contains(content, "os.Exit(1)") {
		t.Error("expected error handling")
	}
}

func TestPluginTestGo(t *testing.T) {
	f := testFiles()
	content := f.pluginTestGo()
	if !strings.Contains(content, "func TestPluginPing") {
		t.Error("expected TestPluginPing")
	}
	if !strings.Contains(content, "func TestPluginDescribe") {
		t.Error("expected TestPluginDescribe")
	}
	if !strings.Contains(content, `"test-plugin"`) {
		t.Error("expected plugin name")
	}
}

func TestMakefile(t *testing.T) {
	f := testFiles()
	content := f.makefile()
	if !strings.Contains(content, "test-plugin") {
		t.Error("expected plugin name in comment")
	}
	if !strings.Contains(content, "build:") {
		t.Error("expected build target")
	}
	if !strings.Contains(content, "test:") {
		t.Error("expected test target")
	}
	if !strings.Contains(content, "clean:") {
		t.Error("expected clean target")
	}
}

func TestCIYML(t *testing.T) {
	f := testFiles()
	content := f.ciYML()
	if !strings.Contains(content, "go test -v -race ./...") {
		t.Error("expected test command")
	}
	if !strings.Contains(content, "GOOS=wasip1") {
		t.Error("expected WASM build")
	}
	if !strings.Contains(content, "buildmode=plugin") {
		t.Error("expected plugin build mode")
	}
}

func TestReadme(t *testing.T) {
	f := testFiles()
	content := f.readme()
	if !strings.Contains(content, "# test-plugin") {
		t.Error("expected plugin name heading")
	}
	if !strings.Contains(content, "A test plugin") {
		t.Error("expected description")
	}
	if !strings.Contains(content, "naeos plugin install") {
		t.Error("expected install instruction")
	}
	if !strings.Contains(content, "Apache 2.0") {
		t.Error("expected license")
	}
}

func TestWriteAll(t *testing.T) {
	dir := t.TempDir()
	f := Files{
		Dir:     dir,
		Module:  "github.com/test/plugin",
		Name:    "test-plugin",
		Author:  "test-author",
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
		"README.md",
		"go.mod",
		".github/workflows/ci.yml",
	}

	for _, path := range expectedFiles {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", path)
		}
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

func TestWriteAllError(t *testing.T) {
	f := Files{
		Dir:    "/nonexistent/path/that/should/fail",
		Module: "test",
		Name:   "test",
	}
	if err := f.WriteAll(); err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestFilesDifferentName(t *testing.T) {
	f := Files{
		Module: "github.com/custom/plugin",
		Name:   "my-custom-plugin",
		Author: "dev",
		Desc:   "Custom plugin",
	}

	pluginContent := f.pluginGo()
	if !strings.Contains(pluginContent, `"my-custom-plugin"`) {
		t.Error("expected custom plugin name")
	}

	yamlContent := f.naeosYAML()
	if !strings.Contains(yamlContent, "name: my-custom-plugin") {
		t.Error("expected custom name in yaml")
	}
}

func TestFilesWithoutWASM(t *testing.T) {
	f := Files{
		Module:  "test",
		Name:    "test",
		UseWASM: false,
	}

	mf := f.makefile()
	if !strings.Contains(mf, "build-plugin:") {
		t.Error("expected build-plugin target")
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
