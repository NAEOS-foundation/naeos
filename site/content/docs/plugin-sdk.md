---
title: Plugin SDK
description: Extend NAEOS with custom plugins, generators, and validators.
---

## Quick Start

Create a new plugin project in one command:

```bash
naeos plugin init my-validator
cd my-validator
```

This scaffolds a complete plugin with `plugin.go`, `main.go` (WASM entry), `naeos.yaml` manifest, tests, and a Makefile.

Build and install:

```bash
# Build as WASM
make build
# Or: GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .

# Install
naeos plugin install ./plugin.wasm
naeos plugin list
```

Test your plugin:

```bash
naeos plugin test ./plugin.wasm
naeos plugin execute my-validator describe
```

## Overview

NAEOS provides a Plugin SDK for extending the platform with custom functionality. Plugins integrate directly into the 9-stage pipeline and can add new code generators, validators, deployers, analyzers, and lifecycle hooks.

```
┌─────────────────────────────────────────────────────────────┐
│                     NAEOS Pipeline                          │
├───────────┬───────────┬───────────┬───────────┬─────────────┤
│  Parse    │ Normalize │  Resolve  │   Build   │  Validate   │
├───────────┴───────────┴───────────┴───────────┴──────┬──────┤
│                  Plugin Hooks ▲                      │      │
│  ┌──────────────────────────────────────┐            │      │
│  │ Plugin Manager                      │            │      │
│  │  ├─ Generator Plugin ──► Artifacts  │◄───────────┘      │
│  │  ├─ Validator Plugin ──► Issues     │◄──────────────────┘
│  │  ├─ Deployer Plugin ──► Result      │                    │
│  │  └─ Hook Plugin      ──► Events    │                    │
│  └──────────────────────────────────────┘                    │
├───────────┬───────────┬───────────┬───────────┐              │
│  Schedule │ Generate  │  Compile  │  Export   │              │
└───────────┴───────────┴───────────┴───────────┘              │
```

## Plugin Interface

Every plugin implements the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Initialize(ctx *PluginContext) error
    Execute(action string, params map[string]any) (any, error)
    Shutdown() error
}
```

### BasePlugin

The SDK provides `BasePlugin` — an embeddable struct with default implementations. Override only the methods you need:

```go
type MyPlugin struct {
    pluginhost.BasePlugin
}

func (p *MyPlugin) Execute(action string, params map[string]any) (any, error) {
    switch action {
    case "validate":
        return p.doValidate(params)
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}
```

## Plugin Types

| Type | Action | Description | Output |
|------|--------|-------------|--------|
| **Generator** | `generate` | Generate code in custom languages | `[]Artifact` |
| **Validator** | `validate` | Custom validation rules | `[]Issue` |
| **Deployer** | `deploy` | Deploy to custom platforms | `Result` |
| **Analyzer** | `analyze` | Custom analysis and reporting | `Report` |
| **Hook** | — | Lifecycle hooks for pipeline stages | `error` |

## Getting Started

### Prerequisites

- Go 1.25+ (for native plugins)
- TinyGo 0.35+ (for WASM plugins)
- NAEOS CLI installed

### Creating a Native Plugin (Go)

```go
package main

import (
    "fmt"
    "github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

type MyValidator struct {
    pluginhost.BasePlugin
}

func (v *MyValidator) Initialize(ctx *pluginhost.PluginContext) error {
    ctx.Logger.Info("MyValidator initialized", nil)
    return nil
}

func (v *MyValidator) Execute(action string, params map[string]any) (any, error) {
    if action != "validate" {
        return nil, fmt.Errorf("unsupported action: %s", action)
    }
    issues := []pluginhost.Issue{
        {Severity: "warning", Message: "custom check passed"},
    }
    return issues, nil
}

func (v *MyValidator) Shutdown() error {
    return nil
}

func main() {
    plugin := &MyValidator{}
    plugin.NameVal = "my-validator"
    plugin.VersionVal = "1.0.0"
    plugin.DescriptionVal = "Custom validation rules"
    // Registration handled by the plugin host at load time
}
```

Build as a shared library:

```bash
go build -buildmode=plugin -o my-validator.so .
```

### Creating a WASM Plugin

WASM plugins work with any language that compiles to WebAssembly. Using TinyGo:

```go
//go:build wasm
// +build wasm

package main

import (
    "github.com/NAEOS-foundation/naeos/sdk"
    "github.com/NAEOS-foundation/naeos/neir"
)

//export generate
func generate(modelPtr, modelLen uint32) uint64 {
    model := sdk.ReadNEIR(modelPtr, modelLen)
    artifacts := processModel(model)
    return sdk.WriteResult(artifacts)
}

func processModel(model *neir.Model) []sdk.Artifact {
    var artifacts []sdk.Artifact
    for _, mod := range model.Modules {
        artifacts = append(artifacts, sdk.Artifact{
            Path:    mod.Path + "/generated.go",
            Content: generateCode(mod),
        })
    }
    return artifacts
}

func main() {}
```

Build:

```bash
tinygo build -o plugin.wasm -target=wasi -scheduler=none .
```

## Plugin Manifest

Every published plugin should include a `plugin.yaml` manifest:

```yaml
name: my-generator
version: "1.0.0"
description: Generate Rust service scaffolding
author: NAEOS Foundation
license: Apache-2.0
dependencies:
  - neir-schema@>=2.0
actions:
  - name: generate
    description: Generate Rust service code from NEIR model
    params:
      output_dir: string
      module: string
    returns: array
config:
  template_dir:
    type: string
    description: Path to custom templates directory
    required: false
  features:
    type: array
    description: Enabled feature flags
    required: false
```

## Plugin Context

When `Initialize` is called, plugins receive a `PluginContext` with:

| Field | Type | Description |
|-------|------|-------------|
| `ConfigDir` | `string` | Plugin configuration directory |
| `OutputDir` | `string` | Output directory for generated artifacts |
| `Verbose` | `bool` | Whether verbose logging is enabled |
| `Config` | `map[string]any` | Plugin-specific configuration from spec |
| `Logger` | `Logger` | Structured logger: `Info`, `Debug`, `Warn`, `Error` |
| `Metrics` | `MetricsCollector` | Metrics counter, gauge, timing |
| `EventBus` | `EventEmitter` | Emit and subscribe to pipeline events |

## Lifecycle Events

Plugins can subscribe to pipeline lifecycle events via the `EventBus`:

| Event | When | EventData |
|-------|------|-----------|
| `before_parse` | Before spec parsing | `PipelineID`, `SpecPath` |
| `after_parse` | After spec parsing | `PipelineID`, `AST` |
| `before_generate` | Before code generation | `PipelineID`, `NEIRModel` |
| `after_generate` | After code generation | `PipelineID`, `Artifacts` |
| `on_pipeline_complete` | Pipeline finished | `PipelineID`, `Duration`, `Result` |
| `on_pipeline_failed` | Pipeline failed | `PipelineID`, `Error` |

Example subscriber:

```go
ctx.EventBus.Subscribe(pluginhost.EventBeforeGenerate, "my-plugin",
    func(_ string, data *pluginhost.EventData) error {
        ctx.Logger.Info("Pipeline generating", map[string]any{
            "pipeline_id": data.PipelineID,
            "mod_count":   len(data.NEIRModel.Modules),
        })
        return nil
    },
)
```

## Plugin Configuration

Plugins receive configuration from the spec file:

```yaml
plugins:
  - name: my-generator
    config:
      template_dir: ./templates
      output_style: compact
      features: [typescript, openapi]
```

Access it in your plugin:

```go
func (p *MyGenerator) Initialize(ctx *pluginhost.PluginContext) error {
    tmplDir, _ := ctx.Config["template_dir"].(string)
    features, _ := ctx.Config["features"].([]any)
    return nil
}
```

## Installing Plugins

```bash
# From the marketplace
naeos plugin install my-generator

# From a local WASM file
naeos plugin install ./path/to/plugin.wasm

# From a Go shared object
naeos plugin install ./path/to/plugin.so

# From a container registry
naeos plugin install ghcr.io/naeos-foundation/plugins/my-generator:latest

# From a URL
naeos plugin install https://example.com/plugins/my-generator.wasm
```

### Managing Plugins

```bash
# List installed plugins
naeos plugin list

# Update a plugin
naeos plugin update my-generator

# Remove a plugin
naeos plugin remove my-generator

# Inspect plugin metadata
naeos plugin info my-generator

# Enable/disable
naeos plugin enable my-generator
naeos plugin disable my-generator
```

## Testing Plugins

NAEOS provides a test runner for plugins:

```bash
# Run plugin tests
naeos test --plugin my-generator

# Test with a specific spec
naeos test --plugin my-generator --input-file test-spec.yaml

# Verbose output
naeos test --plugin my-generator -v
```

Write tests in Go:

```go
func TestMyValidator(t *testing.T) {
    p := &MyValidator{}
    ctx := &pluginhost.PluginContext{
        Logger: pluginhost.NewTestLogger(t),
    }
    if err := p.Initialize(ctx); err != nil {
        t.Fatal(err)
    }
    result, err := p.Execute("validate", map[string]any{
        "spec": "project: test",
    })
    if err != nil {
        t.Fatal(err)
    }
    issues, ok := result.([]pluginhost.Issue)
    if !ok {
        t.Fatal("expected []Issue")
    }
    if len(issues) == 0 {
        t.Error("expected at least one issue")
    }
}
```

## Publishing to Marketplace

```bash
# Package your plugin
naeos plugin package ./my-generator --output my-generator.tar.gz

# Publish to marketplace
naeos marketplace publish my-generator.tar.gz

# Publish with metadata
naeos marketplace publish my-generator.tar.gz \
  --tag latest \
  --description "Rust code generator"
```

The marketplace verifies:
- SHA-256 checksum of the plugin binary
- Manifest schema conformance
- Version uniqueness
- License compatibility

## Hot-Reload

The plugin host supports hot-reloading in development:

```bash
naeos run --plugin-dir ./plugins --watch
```

When files matching `*.so` or `*.wasm` change in `--plugin-dir`, they are automatically reloaded without restarting the pipeline. This is implemented via `PluginWatcher` which uses `fsnotify` to detect changes with a debounce interval.

## SDK Reference

### Plugin Host

| Function | Description |
|----------|-------------|
| `pluginhost.NewManager(config)` | Create a new plugin manager |
| `manager.Load(path)` | Load a plugin from file |
| `manager.Register(plugin)` | Register a plugin instance |
| `manager.Execute(action, params)` | Execute an action across all plugins |
| `manager.List()` | List all registered plugins |
| `manager.Unload(name)` | Unload a specific plugin |

### WASM SDK

| Function | Description |
|----------|-------------|
| `sdk.ReadNEIR(ptr, len)` | Deserialize NEIR model from WASM memory |
| `sdk.WriteResult(data)` | Serialize result to WASM memory |
| `sdk.Log(level, msg)` | Log from WASM plugin |
| `sdk.GetConfig()` | Read plugin configuration |

### Types

| Type | Fields |
|------|--------|
| `Artifact` | `Path string`, `Content string`, `Mode os.FileMode` |
| `Issue` | `Severity string`, `Message string`, `Location string`, `Code string` |
| `Result` | `Success bool`, `Message string`, `Data map[string]any` |
| `Report` | `Title string`, `Sections []ReportSection`, `Metrics map[string]int` |
| `EventData` | `PipelineID string`, `SpecPath string`, `NEIRModel *Model`, `Artifacts []Artifact`, `Duration time.Duration`, `Error error` |

## Example: Custom Validator Plugin

Here's a complete validator that checks module naming conventions:

```go
package main

import (
    "regexp"
    "github.com/NAEOS-foundation/naeos/internal/pluginhost"
    "github.com/NAEOS-foundation/naeos/neir"
)

type NamingValidator struct {
    pluginhost.BasePlugin
    pattern *regexp.Regexp
}

func (v *NamingValidator) Initialize(ctx *pluginhost.PluginContext) error {
    v.pattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
    return nil
}

func (v *NamingValidator) Execute(action string, params map[string]any) (any, error) {
    model, ok := params["model"].(*neir.Model)
    if !ok {
        return nil, fmt.Errorf("missing model parameter")
    }
    var issues []pluginhost.Issue
    for _, mod := range model.Modules {
        if !v.pattern.MatchString(mod.Name) {
            issues = append(issues, pluginhost.Issue{
                Severity: "error",
                Code:     "INVALID_MODULE_NAME",
                Message:  "module name must be lowercase with hyphens",
                Location: fmt.Sprintf("modules[%s]", mod.Name),
            })
        }
    }
    return issues, nil
}

func main() {}
```

## Manifest Reference

The `naeos.yaml` manifest describes your plugin's metadata, capabilities, and config schema:

```yaml
name: my-validator         # unique plugin name
version: "0.1.0"          # semver
description: "Validates project conventions"
author: "Your Name"       # optional
type: wasm                # "wasm" or "native"
tags:                     # optional: for marketplace search
  - validator
  - lint
```

The manifest is automatically created by `naeos plugin init`.

## WASM Plugin Example

Plugins can run as WASM modules with an entry point that reads action/params from stdin:

```go
// main.go — WASM entry point (auto-generated by scaffold)
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        os.Exit(1)
    }

    action := os.Args[1]
    var params map[string]any
    if len(os.Args) > 2 {
        json.Unmarshal([]byte(os.Args[2]), &params)
    }

    p := New()
    p.Initialize(nil)

    switch action {
    case "ping":
        json.NewEncoder(os.Stdout).Encode(map[string]string{"status": "ok"})
    case "describe":
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "name": p.Name(), "version": p.Version(), "description": p.Description(),
        })
    default:
        result, err := p.Execute(action, params)
        if err != nil {
            json.NewEncoder(os.Stdout).Encode(map[string]any{
                "ok": false, "error": err.Error(),
            })
            return
        }
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "ok": true, "result": result,
        })
    }
}
```

Build with: `GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .`

## Best Practices

- **Start small** — Embed `BasePlugin` and implement only `Execute` first
- **Use semantic versioning** — Tag releases as `v1.0.0`, `v1.1.0`, etc.
- **Include a manifest** — Always ship `plugin.yaml` with metadata
- **Log meaningfully** — Use `ctx.Logger` with structured key-value pairs
- **Handle errors gracefully** — Return descriptive `Issue` objects instead of panicking
- **Test with `naeos test`** — Validate your plugin against real specs
- **Keep WASM lean** — WASM plugins should minimize imports for fast loading
- **Use hot-reload** — During development, use `--plugin-dir ./plugins --watch`

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Plugin not loading | Check `naeos plugin list` for errors. Verify the binary format (`.so` or `.wasm`). |
| WASM plugin crashes | Rebuild with `-scheduler=none` flag. Check memory allocation. |
| Config not passed | Verify the `plugins[].config` YAML structure matches your manifest. |
| Events not firing | Confirm you subscribed before the pipeline starts. Check event name spelling. |
| Hot-reload not working | Ensure files are in the watched directory. Use `*.so` or `*.wasm` extensions. |

## See Also

- [Plugin Marketplace](/plugins/) — Browse and install community plugins
- [Template Marketplace](/templates/) — Starter templates for new plugins
- [Pipeline Engine](/docs/pipeline-engine/) — How plugins integrate with the 9-stage DAG
- [Architecture](/docs/architecture/) — System architecture overview
