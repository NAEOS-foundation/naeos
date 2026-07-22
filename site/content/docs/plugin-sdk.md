---
title: Plugin SDK
description: Extend NAEOS with custom plugins, generators, and validators.
---

## Overview

NAEOS provides a Plugin SDK for extending the platform with custom functionality. Plugins can add new code generators, validators, deployers, analyzers, and more. Plugins can be written in Go (native) or any language that compiles to WASM.

## Plugin Types

| Type | Description | Interface |
|------|-------------|-----------|
| **Generator** | Generate code in custom languages | `Generate(ctx, neir) → []Artifact` |
| **Validator** | Custom validation rules | `Validate(ctx, neir) → []Issue` |
| **Deployer** | Deploy to custom platforms | `Deploy(ctx, artifacts) → Result` |
| **Analyzer** | Custom analysis and reporting | `Analyze(ctx, neir) → Report` |
| **Hook** | Lifecycle hooks for pipeline stages | `OnStage(ctx, stage) → error` |

## Getting Started

### Prerequisites

- Go 1.25+ (for native plugins)
- TinyGo (for WASM plugins)

### Creating a Native Plugin

```go
package main

import (
    "github.com/NAEOS-foundation/naeos/sdk"
    "github.com/NAEOS-foundation/naeos/neir"
)

type MyGenerator struct{}

func (g *MyGenerator) Generate(ctx *sdk.Context, model *neir.Model) ([]sdk.Artifact, error) {
    var artifacts []sdk.Artifact
    for _, mod := range model.Modules {
        content := generateCode(mod)
        artifacts = append(artifacts, sdk.Artifact{
            Path:    mod.Path + "/main.go",
            Content: content,
        })
    }
    return artifacts, nil
}

func main() {
    sdk.Register(&MyGenerator{})
}
```

### Creating a WASM Plugin

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
    // Custom generation logic
    result := processModel(model)
    return sdk.WriteResult(result)
}

func main() {}
```

## Installing Plugins

```bash
# From marketplace
naeos plugin install my-generator

# From local file
naeos plugin install ./path/to/plugin.wasm

# From registry
naeos plugin install ghcr.io/naeos-foundation/plugins/my-generator:latest
```

## Managing Plugins

```bash
# List installed plugins
naeos plugin list

# Update a plugin
naeos plugin update my-generator

# Remove a plugin
naeos plugin remove my-generator

# Inspect plugin info
naeos plugin info my-generator
```

## Plugin Configuration

Plugins can receive configuration via the spec file:

```yaml
plugins:
  - name: my-generator
    config:
      template_dir: ./templates
      output_style: compact
      features: [typescript, openapi]
```

## Publishing to Marketplace

```bash
# Package your plugin
naeos plugin package ./my-generator --output my-generator.tar.gz

# Publish (requires marketplace access)
naeos marketplace publish my-generator.tar.gz
```

## SDK Reference

The Plugin SDK provides:

- `sdk.Context` — Pipeline context with config, logging, and file access
- `sdk.Register()` — Register your plugin implementation
- `sdk.ReadNEIR()` — Deserialize NEIR model from WASM memory
- `sdk.WriteResult()` — Serialize result back to WASM memory
- `sdk.Artifact` — Generated file output
- `sdk.Issue` — Validation issue with severity, location, and message

## Best Practices

- Test plugins with `naeos test --plugin my-plugin`
- Use semantic versioning for plugin releases
- Include a `plugin.yaml` manifest with metadata
- Leverage the built-in logging via `sdk.Context.Logger`
- Handle errors gracefully and return meaningful messages
