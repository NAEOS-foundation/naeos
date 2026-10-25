# Official Example Plugins

This directory contains NAEOS OSS–maintained example packages within the
repository's Go module. Each demonstrates a specific plugin implementation
pattern; they are not standalone external Go modules.

| Plugin | Pattern demonstrated |
|--------|----------------------|
| [`hello/`](hello/) | Canonical minimal plugin: `BasePlugin`, action dispatch, JSON-over-stdio entry point |
| [`spec-lint/`](spec-lint/) | Consuming pipeline input (spec text) and returning a structured report |
| [`artifact-stats/`](artifact-stats/) | Processing batched inputs (artifact files) and aggregating results |
| [`trivy-config/`](trivy-config/) | Native integration that returns structured Trivy configuration findings |

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

The examples use the in-repository `internal/pluginhost` implementation and
are not standalone external Go modules. The current WASM entry point exchanges
one JSON request/response over stdio; this wire shape is experimental and has
no published compatibility guarantee.

## Installing Locally

```bash
naeos plugin install examples/plugins/hello --name hello
naeos plugin execute hello ping
```

## Publishing to the Registry

```bash
naeos marketplace plugin publish examples/plugins/hello \
  --registry https://naeos-foundation.github.io/naeos
```

Check the registry and install implementation for the integrity guarantees
available in your selected distribution path. Repository documentation alone
does not establish publisher-signature verification or a trust policy.

## Writing Your Own

Use the scaffold generator — it produces a complete plugin project with
tests, Makefile, WASM build, and CI workflow:

```bash
naeos plugin init my-plugin --author "You" --desc "My first plugin"
```

The `trivy-config` example is a native Go plugin because it invokes the
external `trivy` executable. It is a partner-pilot reference, not a WASM
plugin, and requires Trivy on `PATH`.
