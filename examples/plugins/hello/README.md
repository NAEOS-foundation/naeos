# NAEOS Plugin — hello

Canonical minimal NAEOS plugin. Demonstrates the plugin SDK contract:
`pluginhost.Plugin` interface, `BasePlugin` embedding, action dispatch, and the
JSON-over-stdio WASM entry point.

## Actions

| Action | Params | Returns |
|--------|--------|---------|
| `ping` | — | `{"status": "ok"}` |
| `describe` | — | name, version, description |
| `greet` | `name: string` | `{"message": "Hello, <name>!"}` |

## Build

```bash
go test -race ./...
GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .
```

## Install & Run

```bash
naeos plugin install . --name hello
naeos plugin execute hello ping
naeos plugin execute hello greet '{"name":"NAEOS"}'
```

## Publish

```bash
naeos marketplace plugin publish . --registry https://naeos-foundation.github.io/naeos
```
