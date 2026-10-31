// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
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

require github.com/NAEOS-foundation/naeos v0.0.0

// NAEOS plugins import internal packages (Go internal rule), so the module
// must stay under the github.com/NAEOS-foundation/naeos prefix and resolve
// naeos to the source checkout matching your CLI:
replace github.com/NAEOS-foundation/naeos => /path/to/naeos
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
			NameVal:        %[1]q,
			VersionVal:     "0.1.0",
			DescriptionVal: %[2]q,
		},
	}
}

var (
	pluginName        = %[1]q
	pluginVersion     = "0.1.0"
	pluginDescription = %[2]q
	pluginAuthor      = %[3]q

	PluginName        = &pluginName
	PluginVersion     = &pluginVersion
	PluginDescription = &pluginDescription
	PluginAuthor      = &pluginAuthor
)

var _ pluginhost.Plugin = (*Plugin)(nil)

var NaeosPlugin pluginhost.Plugin = New()

func (p *Plugin) Initialize(ctx *pluginhost.PluginContext) error {
	return nil
}

func (p *Plugin) Execute(action string, params map[string]any) (any, error) {
	switch action {
	case "ping":
		return map[string]string{"status": "ok"}, nil
	case "health":
		return map[string]string{"status": "healthy"}, nil
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
`, f.Name, f.Desc, f.Author)
}

func (f Files) wasmMainGo() string {
	bt := "`"
	return fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"os"

	pluginapi "github.com/NAEOS-foundation/naeos/pkg/pluginapi/v1"
)

func respond(resp pluginapi.Response) {
	// stdout is reserved for exactly one response object; diagnostics go to stderr.
	out, err := json.Marshal(resp)
	if err != nil {
		resp = pluginapi.Response{
			ProtocolVersion: pluginapi.ProtocolVersion,
			OK:              false,
			Error:           &pluginapi.Error{Code: pluginapi.ErrorCodeInvalidResponse, Message: "encode response"},
		}
		out, _ = json.Marshal(resp)
	}
	fmt.Println(string(out))
}

func fail(req pluginapi.Request, code, message string) {
	respond(pluginapi.Response{
		ProtocolVersion: req.ProtocolVersion,
		RequestID:       req.RequestID,
		OK:              false,
		Error:           &pluginapi.Error{Code: code, Message: message},
	})
	os.Exit(1)
}

func main() {
	var req pluginapi.Request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fail(pluginapi.Request{ProtocolVersion: pluginapi.ProtocolVersion}, pluginapi.ErrorCodeInvalidRequest, "invalid request: "+err.Error())
	}
	if err := req.Validate(); err != nil {
		fail(req, pluginapi.ErrorCodeInvalidRequest, err.Error())
	}

	switch req.Action {
	case "ping":
		respond(pluginapi.Response{
			ProtocolVersion: req.ProtocolVersion,
			RequestID:       req.RequestID,
			OK:              true,
			Result:          json.RawMessage("{\"status\":\"ok\"}"),
		})
	case "health":
		respond(pluginapi.Response{
			ProtocolVersion: req.ProtocolVersion,
			RequestID:       req.RequestID,
			OK:              true,
			Result:          json.RawMessage("{\"status\":\"healthy\"}"),
		})
	case "describe":
		respond(pluginapi.Response{
			ProtocolVersion: req.ProtocolVersion,
			RequestID:       req.RequestID,
			OK:              true,
			Result:          json.RawMessage(fmt.Sprintf("{\"name\":%%q,\"version\":\"0.1.0\"}", %[2]q)),
		})
	default:
		fail(req, pluginapi.ErrorCodePluginError, fmt.Sprintf("unknown action: %%s", req.Action))
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
	GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .

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

## Prerequisites

Native (.so) plugins import NAEOS internal packages, which Go only allows
for modules under the `+"`github.com/NAEOS-foundation/naeos`"+` prefix. Point
`+"`replace`"+` in `+"`go.mod`"+` to the NAEOS source checkout matching your CLI
before building.

## Usage

%[3]sbash
# Build the native plugin
make build-plugin

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

# Build (WASM)
make build
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
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create dir %s", d)
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
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "write %s", path)
		}
	}

	return nil
}
