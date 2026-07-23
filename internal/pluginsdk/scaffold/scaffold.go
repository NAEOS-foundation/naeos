package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

type Files struct {
	Dir     string
	Module  string
	Name    string
	Author  string
	Desc    string
	UseWASM bool
}

func (f Files) goMod() string {
	return fmt.Sprintf(`module %s

go 1.25.0
`, f.Module)
}

func (f Files) naeosYAML() string {
	return fmt.Sprintf(`name: %s
version: 0.1.0
description: %s
author: %s
type: wasm
tags: []
`, f.Name, f.Desc, f.Author)
}

func (f Files) pluginGo() string {
	return fmt.Sprintf(`package main

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

type Plugin struct {
	pluginhost.BasePlugin
}

func New() *Plugin {
	return &Plugin{
		BasePlugin: pluginhost.BasePlugin{
			Name:        %q,
			Version:     "0.1.0",
			Description: %q,
		},
	}
}

func (p *Plugin) Initialize(ctx *pluginhost.PluginContext) error {
	return nil
}

func (p *Plugin) Execute(action string, params map[string]any) (any, error) {
	switch action {
	case "ping":
		return map[string]string{"status": "ok"}, nil
	case "describe":
		return map[string]any{
			"name":        p.Name(),
			"version":     p.Version(),
			"description": p.Description(),
		}, nil
	default:
		return nil, fmt.Errorf("unknown action: %%s", action)
	}
}

func (p *Plugin) Shutdown() error {
	return nil
}
`, f.Name, f.Desc)
}

func (f Files) wasmMainGo() string {
	bt := "`"
	return fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(%[1]s{"error":"missing action","ok":false}%[1]s)
		os.Exit(1)
	}

	action := os.Args[1]
	var params map[string]any
	if len(os.Args) > 2 {
		json.Unmarshal([]byte(os.Args[2]), &params)
	}

	switch action {
	case "ping":
		fmt.Println(%[1]s{"ok":true,"result":{"status":"ok"}}%[1]s)
	case "describe":
		fmt.Printf(%[1]s{"ok":true,"result":{"name":"%[2]s","version":"0.1.0"}}%[1]s)
	default:
		fmt.Printf(%[1]s{"error":"unknown action: %%s","ok":false}%[1]s, action)
		os.Exit(1)
	}
}
`, bt, f.Name)
}

func (f Files) pluginTestGo() string {
	return fmt.Sprintf(`package main

import (
	"testing"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

func TestPluginPing(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}

	ctx := &pluginhost.PluginContext{}
	if err := p.Initialize(ctx); err != nil {
		t.Fatalf("Initialize: %%v", err)
	}
	defer p.Shutdown()

	result, err := p.Execute("ping", nil)
	if err != nil {
		t.Fatalf("Execute ping: %%v", err)
	}

	r, ok := result.(map[string]string)
	if !ok || r["status"] != "ok" {
		t.Errorf("expected {status: ok}, got %%v", result)
	}
}

func TestPluginDescribe(t *testing.T) {
	p := New()
	result, err := p.Execute("describe", nil)
	if err != nil {
		t.Fatalf("Execute describe: %%v", err)
	}

	r, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %%T", result)
	}
	if r["name"] != %q {
		t.Errorf("expected name %%q, got %%v", %q, r["name"])
	}
}
`, f.Name, f.Name)
}

func (f Files) makefile() string {
	return `# NAEOS Plugin — ` + f.Name + `
.PHONY: build test clean

build:
	go build -o plugin.wasm -buildmode=c-shared .

build-plugin:
	go build -buildmode=plugin -o plugin.so .

test:
	go test -v -race ./...

clean:
	rm -f plugin.wasm plugin.so

all: test build
`
}

func (f Files) ciYML() string {
	return fmt.Sprintf(`name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Test
        run: go test -v -race ./...
      - name: Build (WASM)
        run: GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .
      - name: Build (plugin mode)
        run: go build -buildmode=plugin -o plugin.so .
`)
}

func (f Files) readme() string {
	t := "`"
	return fmt.Sprintf(`# %[1]s

%[2]s

## Usage

%[3]sbash
# Install the plugin
naeos plugin install ./plugin.so

# Run actions
naeos plugin execute %[4]s ping
naeos plugin execute %[4]s describe
%[3]s

## Development

%[3]sbash
# Run tests
make test

# Build
make build
make build-plugin
%[3]s

## License

Apache 2.0
`, f.Name, f.Desc, t, f.Name)
}

func (f Files) WriteAll() error {
	dirs := []string{
		f.Dir,
		filepath.Join(f.Dir, ".github", "workflows"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	files := map[string]string{
		"naeos.yaml":               f.naeosYAML(),
		"plugin.go":                f.pluginGo(),
		"plugin_test.go":           f.pluginTestGo(),
		"main.go":                  f.wasmMainGo(),
		"Makefile":                 f.makefile(),
		".github/workflows/ci.yml": f.ciYML(),
		"README.md":                f.readme(),
		"go.mod":                   f.goMod(),
	}

	for path, content := range files {
		full := filepath.Join(f.Dir, path)
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
}
