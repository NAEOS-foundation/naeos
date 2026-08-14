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

NAEOS provides a Plugin SDK for extending the platform with custom functionality. Plugins integrate directly into the 11-stage pipeline and can add new code generators, validators, deployers, analyzers, and lifecycle hooks.

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

Native plugins are Go `main` packages built with `-buildmode=plugin`. The
plugin must export the `PluginName` metadata variables and a `NaeosPlugin`
value implementing `pluginhost.Plugin`:

```go
package main

import (
    "fmt"

    "github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

type MyValidator struct {
    pluginhost.BasePlugin
}

func New() *MyValidator {
    return &MyValidator{
        BasePlugin: pluginhost.BasePlugin{
            NameVal:        "my-validator",
            VersionVal:     "1.0.0",
            DescriptionVal: "Custom validation rules",
        },
    }
}

var (
    pluginName        = "my-validator"
    pluginVersion     = "1.0.0"
    pluginDescription = "Custom validation rules"
    pluginAuthor      = "you"

    PluginName        = &pluginName
    PluginVersion     = &pluginVersion
    PluginDescription = &pluginDescription
    PluginAuthor      = &pluginAuthor
)

var NaeosPlugin pluginhost.Plugin = New()

func (v *MyValidator) Initialize(ctx *pluginhost.PluginContext) error {
    ctx.Logger.Info("MyValidator initialized")
    return nil
}

func (v *MyValidator) Execute(action string, params map[string]any) (any, error) {
    if action != "validate" {
        return nil, fmt.Errorf("unsupported action: %s", action)
    }
    return []map[string]string{
        {"severity": "warning", "message": "custom check passed"},
    }, nil
}

func (v *MyValidator) Shutdown() error {
    return nil
}
```

Build as a shared library:

```bash
go build -buildmode=plugin -o my-validator.so .
```

Because the plugin imports NAEOS internal packages, Go requires the plugin
module to live under the `github.com/NAEOS-foundation/naeos` prefix and resolve
`naeos` to the source checkout matching your CLI:

```
replace github.com/NAEOS-foundation/naeos => /path/to/naeos
```

`naeos plugin init` sets all of this up for you.

### Creating a WASM Plugin

WASM plugins run as WASI modules. At execution the host feeds the request
(`{"method": "<action>", "params": {...}}`) on stdin and expects a
`{"ok": true, "result": ...}` or `{"ok": false, "error": "..."}` JSON response
on stdout. Any language that compiles to WASI can implement this protocol.
Using TinyGo:

```go
package main

import (
    "encoding/json"
    "io"
    "os"
)

func main() {
    var req struct {
        Method string `json:"method"`
        Params map[string]any `json:"params"`
    }
    if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "ok": false, "error": "invalid request: " + err.Error(),
        })
        os.Exit(1)
    }

    switch req.Method {
    case "ping":
        json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "result": "pong"})
    default:
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "ok": false, "error": "unknown action: " + req.Method,
        })
        os.Exit(1)
    }
}
```

Build:

```bash
tinygo build -o plugin.wasm -target=wasi -scheduler=none .
```

## Plugin Manifest

Every plugin includes a `naeos.yaml` manifest (created automatically by
`naeos plugin init`):

```yaml
name: my-generator
version: "0.1.0"
description: Generate Rust service scaffolding
author: NAEOS Foundation
type: wasm              # "wasm" or "native"
tags: []
```

The manifest is required for `naeos marketplace publish` (see [Manifest
Reference](#manifest-reference) below). Actions are declared in code via the
`Execute` method, not in the manifest.

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

Load, initialize, and health-check a plugin with the plugin test runner:

```bash
# Load and health-check a plugin
naeos plugin test my-generator

# Point at a custom plugin directory
naeos plugin test my-generator --plugin-dir ./my-plugins
```

`naeos test` also runs tests for your generated code per language (Go, TypeScript, Python, Java, Rust).

Write Go tests against the plugin host API:

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
    issues, ok := result.([]map[string]string)
    if !ok {
        t.Fatal("expected []map[string]string")
    }
    if len(issues) == 0 {
        t.Error("expected at least one issue")
    }
}
```

## Publishing to Marketplace

```bash
# Publish your plugin package (directory containing naeos.yaml) to the marketplace
naeos marketplace publish ./my-generator
```

The publish command validates that the package contains a `naeos.yaml` manifest
with `name`, `version`, and `type` fields before publishing.

## Hot-Reload

The plugin host provides a library-level watcher for development. `PluginWatcher` uses `fsnotify` to watch a plugin directory and reload changed `*.so` or `*.wasm` plugins automatically (500 ms debounce):

```go
pw := pluginhost.NewPluginWatcher("./plugins", manager)
if err := pw.Start(ctx); err != nil {
    // handle error
}
```

At the CLI level, use `naeos plugin test --plugin-dir ./plugins` to re-check a plugin after rebuilding it.

## SDK Reference

### Plugin Host

| Function | Description |
|----------|-------------|
| `pluginhost.NewManager(pluginDir)` | Create a new plugin manager |
| `manager.Install(path)` | Install a `.so` plugin (reads exported metadata) |
| `manager.LoadAll(ctx)` | Load and initialize all installed plugins |
| `manager.Register(plugin)` | Register an in-process plugin instance |
| `manager.Execute(ctx, name, action, params)` | Execute one plugin's action |
| `manager.List()` | List all installed plugins |
| `manager.GetInfo(name)` | Get metadata for one plugin |
| `manager.Cleanup()` | Shutdown all plugins and release resources |

### WASM Protocol

| Direction | Payload |
|-----------|---------|
| Request (stdin) | `{"method": "<action>", "params": {}}` |
| Success (stdout) | `{"ok": true, "result": <any>}` |
| Failure (stdout) | `{"ok": false, "error": "<message>"}` |

### Types

| Type | Fields |
|------|--------|
| `EventData` | `PipelineID string`, `Stage string`, `Artifacts int`, `Duration string`, `Error string`, `Extra map[string]any` |
| `ActionManifest` | `Name string`, `Description string`, `Params map[string]string`, `Returns string` |
| `ConfigField` | `Type string`, `Description string`, `Required bool`, `Default string` |
| `PluginInfo` | `Name string`, `Version string`, `Description string`, `Author string`, `Path string`, `Enabled bool`, `Loaded bool`, `State PluginState` |

## Example: Custom Validator Plugin

Here's a complete validator that checks module naming conventions:

```go
package main

import (
    "fmt"
    "regexp"

    "github.com/NAEOS-foundation/naeos/internal/pluginhost"
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
    if action != "validate" || params["name"] == nil {
        return nil, fmt.Errorf("missing name parameter")
    }
    name, _ := params["name"].(string)
    if !v.pattern.MatchString(name) {
        return []map[string]string{
            {"severity": "error", "code": "INVALID_MODULE_NAME", "message": "module name must be lowercase with hyphens"},
        }, nil
    }
    return nil, nil
}

var _ pluginhost.Plugin = (*NamingValidator)(nil)

var NaeosPlugin pluginhost.Plugin = New()

func New() *NamingValidator {
    v := &NamingValidator{}
    v.NameVal = "naming-validator"
    v.VersionVal = "0.1.0"
    v.DescriptionVal = "Validates module naming conventions"
    return v
}
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

WASM modules read the request JSON (`{"method": "<action>", "params": {...}}`)
from stdin and write the response JSON on stdout:

```go
// main.go — WASM entry point (auto-generated by scaffold)
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

type request struct {
    Method string         `json:"method"`
    Params map[string]any `json:"params"`
}

func main() {
    var req request
    if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
        fmt.Fprintf(os.Stdout, `{"ok":false,"error":%q}`+"\n", "invalid request: "+err.Error())
        os.Exit(1)
    }

    switch req.Method {
    case "ping":
        json.NewEncoder(os.Stdout).Encode(map[string]any{"ok": true, "result": "pong"})
    case "describe":
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "ok": true, "result": map[string]string{"name": "my-plugin", "version": "0.1.0"},
        })
    default:
        json.NewEncoder(os.Stdout).Encode(map[string]any{
            "ok": false, "error": "unknown action: " + req.Method,
        })
    }
}
```

Build with: `GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .`

## Best Practices

- **Start small** — Embed `BasePlugin` and implement only `Execute` first
- **Use semantic versioning** — Tag releases as `v1.0.0`, `v1.1.0`, etc.
- **Include a manifest** — Always ship `naeos.yaml` with metadata
- **Log meaningfully** — Use `ctx.Logger` with structured key-value pairs
- **Handle errors gracefully** — Return descriptive result structures instead of panicking
- **Test with `naeos test`** — Validate your plugin against real specs
- **Keep WASM lean** — WASM plugins should minimize imports for fast loading
- **Use hot-reload** — During development, re-check plugins after rebuilds with `naeos plugin test --plugin-dir ./plugins`

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
- [Pipeline Engine](/docs/pipeline-engine/) — How plugins integrate with the 11-stage DAG
- [Architecture](/docs/architecture/) — System architecture overview
