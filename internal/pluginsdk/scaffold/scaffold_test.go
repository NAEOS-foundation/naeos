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
		t.Fatal(err)
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
