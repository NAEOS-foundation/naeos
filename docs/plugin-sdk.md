# NAEOS Plugin SDK

> **Status: in-repository / experimental.** This document describes the
> current implementation, not a stable external Go SDK. The examples,
> scaffold, and native plugin interface import `internal/pluginhost`; Go's
> `internal` rule prevents an unrelated module from importing that package.
> The generated project therefore expects a NAEOS source checkout and a local
> `replace` directive. The WASM wire envelope now uses the public
> `pkg/pluginapi/v1` DTO package, but host-side protocol compatibility,
> manifest negotiation, and compatibility tests are still incomplete; do not
> promise stability yet.

## Overview

The current plugin host loads Go plugins implementing the internal
`pluginhost.Plugin` interface. WASM modules receive one JSON request on stdin
and return a JSON response on stdout; native `.so` loading is an
in-process, platform-specific option. Neither the internal Go interface nor
the current JSON-over-stdio shape is a versioned external extension API.

For experiments inside a NAEOS checkout, the scaffold generator is available:

```bash
naeos plugin init my-plugin
```

This produces a project (`plugin.go`, `main.go` entry point, `naeos.yaml`,
tests, Makefile) that imports `internal/pluginhost`. Its generated `go.mod`
requires a local `replace` to a matching NAEOS checkout. See
`examples/plugins/` for in-repository examples.

## Current WASM wire format

The host sends one JSON object on stdin using the public v1 envelope types
from `pkg/pluginapi/v1` (`Request`, `Response`, `Manifest`):

```json
{
  "protocol_version": "1.0",
  "request_id": "8c50f92c",
  "method": "execute",
  "action": "ping",
  "params": {}
}
```

The plugin writes exactly one JSON response object to stdout (anything after
the first object makes the response invalid). A successful response echoes
the protocol version and request ID:

```json
{"protocol_version":"1.0","request_id":"8c50f92c","ok":true,"result":{"status":"ok"}}
```

A failed response uses a documented v1 error code finite set
(`invalid_request`, `unsupported_version`, `invalid_params`, `denied`,
`timeout`, `resource_limit`, `plugin_error`, `invalid_response`):

```json
{"protocol_version":"1.0","request_id":"8c50f92c","ok":false,"error":{"code":"invalid_params","message":"name is required"}}
```

Rules currently enforced by the host: request IDs are generated host-side and
must be echoed untouched; stdout is reserved for exactly one response object
(diagnostics go to stderr); oversized responses and uncorrelated or
unversioned responses are rejected as structured host errors; elapsed-time
measurement stays host-side. The exact protocol-version selection helper
(`Manifest.SupportsProtocol`) exists in the public DTO package, but
host-side manifest loading and negotiation are not implemented yet. This
remains an experimental implementation detail until the compatibility policy
is published.

WASM execution currently uses wazero with a timeout and memory limit. The host
does not configure filesystem preopens in the shown execution path. Review the
runtime implementation and threat model before relying on any broader sandbox
or isolation guarantee.

## Creating a Plugin

1. **Implement the Plugin interface**

   ```go
   import (
       "github.com/NAEOS-foundation/naeos/internal/pluginhost"
   )

   type MyPlugin struct {
       pluginhost.BasePlugin // provides default Name, Version, Description
       // add any fields you need (e.g., config, logger)
   }

   func New() *MyPlugin {
       return &MyPlugin{
           BasePlugin: pluginhost.BasePlugin{
               Name:        "my-plugin",
               Version:     "v0.1.0",
               Description: "Does something useful",
           },
       }
   }

   func (p *MyPlugin) Initialize(ctx *pluginhost.PluginContext) error {
       // Store context for later use
       p.Logger().Info("plugin initialized")
       return nil
   }

   func (p *MyPlugin) Execute(action string, params map[string]any) (any, error) {
       switch action {
       case "do-something":
           // implement action
           return result, nil
       default:
           return nil, fmt.Errorf("unknown action: %s", action)
       }
   }

   func (p *MyPlugin) Shutdown() error {
       p.Logger().Info("plugin shutting down")
       return nil
   }
   ```

2. **Expose a constructor function**

   The plugin host expects a `New()` function that returns an instance of your plugin (or a struct that implements `Plugin`). If you use the `main` package pattern, the host will look for a `Plugin` variable; however the recommended way is to provide a `New` function as shown.

3. **Build the plugin**

   ```bash
   GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .
   ```

   The resulting `.wasm` file can be installed with `naeos plugin install`.
   The source still has to satisfy the Go `internal` import rule described
   above. Native in-process builds
   (`go build -buildmode=plugin -o plugin.so`) remain an
   implementation-specific option for development.

## PluginContext

During `Initialize`, you receive a `*pluginhost.PluginContext` containing:

- `ConfigDir`: directory for plugin‑specific configuration.
- `OutputDir`: directory where generated code is written (useful for post‑processing plugins).
- `Verbose`: flag indicating verbose logging.
- `Config`: map of plugin configuration from `naeos.yaml` or CLI flags.
- `Logger`: structured logger (`pluginhost.Logger`).
- `Metrics`: metrics collector.
- `EventBus`: event emitter for publishing custom events.

## Example: golangci‑lint plugin

Below is a complete example plugin that adds golangci-lint configuration to generated Go projects:

```go
package myplugin

import (
    "path/filepath"

    "github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

type GolangciLintPlugin struct {
    pluginhost.BasePlugin
    outputDir string
}

func New() *GolangciLintPlugin {
    return &GolangciLintPlugin{
        BasePlugin: pluginhost.BasePlugin{
            NameVal:        "golangci-lint",
            VersionVal:     "v0.1.0",
            DescriptionVal: "Adds golangci-lint configuration to generated Go projects.",
        },
    }
}

func (p *GolangciLintPlugin) Initialize(ctx *pluginhost.PluginContext) error {
    p.outputDir = ctx.OutputDir
    p.Logger().Info("golangci-lint plugin initialized")
    return nil
}

func (p *GolangciLintPlugin) Execute(action string, params map[string]any) (any, error) {
    switch action {
    case "add-config":
        path := filepath.Join(p.outputDir, ".golangci.yml")
        content := []byte(`linters:
  enable:
    - govet
    - staticcheck
`)
        return map[string]string{"path": path}, nil
    default:
        return nil, nil
    }
}
```

## Testing your plugin

You can implement a simple test harness:

```go
func TestMyPlugin(t *testing.T) {
    p := New()
    ctx := &pluginhost.PluginContext{
        OutputDir: t.TempDir(),
        Logger:    pluginhost.NewSimpleLogger(os.Stdout, pluginhost.LevelInfo),
    }
    if err := p.Initialize(ctx); err != nil {
        t.Fatalf("Init failed: %v", err)
    }
    _, err := p.Execute("do-something", map[string]any{})
    if err != nil {
        t.Errorf("Execute failed: %v", err)
    }
    if err := p.Shutdown(); err != nil {
        t.Errorf("Shutdown failed: %v", err)
    }
}
```

## Publishing

1. Build the WASM module (`GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .`) and include a `naeos.yaml` manifest with metadata.
2. Publish to a plugin registry via `naeos marketplace publish` or share privately.
3. Consumers install with `naeos plugin install <name>`.

## Further Reading

- `internal/pluginhost/pluginhost.go` – full interface definition.
- `internal/pluginhost/manager.go` – how plugins are loaded and managed.
- `internal/pluginsdk/wasm/wasm.go` – WASM execution bridge.
- `internal/pluginsdk/scaffold/scaffold.go` – plugin project scaffold (`naeos plugin init`).
- `internal/pluginsdk/sandbox/sandbox.go` – plugin sandboxing.
- `examples/plugins/README.md` – official example plugins.
- `docs/NES-053-WASMPlugin.md` – WASM plugin sandbox specification.
- `docs/architecture/PLUGIN-API-DESIGN.md` – proposed public protocol boundary; not implemented or stable.

Happy hacking!
