# NAEOS Plugin SDK

## Overview

NAEOS plugins extend the platform by providing custom actions that can be executed during the generation pipeline or via the CLI. Plugins are Go packages that implement the `pluginhost.Plugin` interface and are compiled to WebAssembly (`GOOS=wasip1 GOARCH=wasm`) for sandboxed execution. In-process execution of native `.so` plugins is also supported via `pluginhost`.

The quickest way to start is the scaffold generator:

```bash
naeos plugin init my-plugin
```

This produces a complete WASM plugin project (`plugin.go`, `main.go` WASM entry, `naeos.yaml` manifest, tests, Makefile). See the official examples in `examples/plugins/` (`hello`, `spec-lint`, `artifact-stats`).

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

   The resulting `.wasm` file can be placed in the plugin directory configured via NAEOS (see `naeos plugin install`). Native in-process builds (`go build -buildmode=plugin -o plugin.so`) remain supported for development.

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

Happy hacking!