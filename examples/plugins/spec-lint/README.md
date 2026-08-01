# NAEOS Plugin — spec-lint

Lints NAEOS specification content for naming and structural conventions.
Shows how a plugin consumes pipeline input (spec text via params) and returns a
structured, actionable report.

## Checks

| Rule | Severity | Detail |
|------|----------|--------|
| `module-name-case` | warning | module `name:` must be lowercase |
| `module-name-format` | warning | module `name:` must use kebab-case (no spaces/underscores) |
| `port-leading-zero` | warning | `port:` values must not have leading zeros |

## Build

```bash
go test -race ./...
GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .
```

## Run

```bash
naeos plugin install . --name spec-lint
naeos plugin execute spec-lint lint '{"spec":"modules:\n  - name: Auth_Service\n    port: 08080\n"}'
```
