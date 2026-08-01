# Official Example Plugins

This directory contains NAEOS Foundation–maintained example plugins. Each is a
self-contained Go module demonstrating a specific plugin SDK pattern.

| Plugin | Pattern demonstrated |
|--------|----------------------|
| [`hello/`](hello/) | Canonical minimal plugin: `BasePlugin`, action dispatch, JSON-over-stdio entry point |
| [`spec-lint/`](spec-lint/) | Consuming pipeline input (spec text) and returning a structured report |
| [`artifact-stats/`](artifact-stats/) | Processing batched inputs (artifact files) and aggregating results |

## Requirements

- Go 1.25+
- NAEOS CLI (any version with the unified plugin system, v0.5.0+)

## Building & Testing

The example plugins are packages of the NAEOS module, so they build and test
with the standard toolchain from the repo root or from their own directory:

```bash
go test -race ./examples/plugins/...
cd examples/plugins/hello
GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .
```

The WASM build is import-free (JSON-over-stdio protocol), so the resulting
`plugin.wasm` runs standalone in the plugin sandbox.

## Installing Locally

```bash
naeos plugin install examples/plugins/hello --name hello
naeos plugin execute hello ping
```

## Publishing to the Registry

```bash
naeos marketplace plugin publish examples/plugins/hello \
  --registry https://registry.naeos.dev
```

The plugin registry verifies publisher signatures and plugin hashes
(SHA-256) on install, so published plugins are tamper-evident.

## Writing Your Own

Use the scaffold generator — it produces a complete plugin project with
tests, Makefile, WASM build, and CI workflow:

```bash
naeos plugin init my-plugin --author "You" --desc "My first plugin"
```
