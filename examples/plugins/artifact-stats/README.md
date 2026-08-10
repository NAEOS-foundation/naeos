# NAEOS Plugin — artifact-stats

Aggregates line and byte statistics over generated artifacts. Demonstrates
processing batched inputs (file contents via params) and returning a
JSON-friendly report.

## Action: `stats`

Params: `files` — array of `{"path": string, "content": string}`.

Returns:

```json
{
  "files": 3,
  "lines": 42,
  "bytes": 512,
  "by_ext": {
    ".go": {"files": 1, "lines": 20, "bytes": 200},
    ".ts": {"files": 1, "lines": 20, "bytes": 300},
    "(none)": {"files": 1, "lines": 2, "bytes": 12}
  }
}
```

## Build

```bash
go test -race ./...
GOOS=wasip1 GOARCH=wasm go build -o plugin.wasm .
```

## Run

```bash
naeos plugin install . --name artifact-stats
naeos plugin execute artifact-stats stats '{"files":[{"path":"main.go","content":"package main"}]}'
```
