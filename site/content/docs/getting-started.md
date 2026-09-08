---
title: Getting Started
description: Install NAEOS and run your first pipeline in minutes.
---

## Prerequisites

- Go 1.25+ (for `go install` method)
- A terminal with basic command-line knowledge

## Installation

Choose one of these methods:

### Go Install

```bash
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

### Docker

```bash
docker pull ghcr.io/naeos-foundation/naeos:latest
docker run --rm -v $(pwd):/workspace ghcr.io/naeos-foundation/naeos:latest naeos version
```

### Build from Source

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
go build ./cmd/naeos/
```

## Your First Pipeline

### 1. Create a specification file

Create `spec.yaml`:

```yaml
project: my-app
modules:
  - name: auth
    path: ./auth
  - name: api
    path: ./api
    dependencies: [auth]
services:
  - name: gateway
    kind: http
    port: 8080
generation:
  languages: [go, typescript]
```

### 2. Initialize configuration

```bash
naeos init
```

### 3. Run the pipeline

```bash
naeos run --input-file spec.yaml
```

### 4. Generate AI context

```bash
naeos context --input-file spec.yaml
```

### 5. Compile for AI assistants

```bash
naeos ai compile --input-file spec.yaml --target opencode
```

## Verify the result

After `naeos run`, the generated project is written to the configured output
directory (typically `./generated`). Confirm the pipeline produced the expected
structure:

```bash
find generated -maxdepth 2 -type f | sort
```

You should see a project README, language-specific files, and configuration
artifacts. The exact file list depends on the languages and modules in your
specification.

For a complete, isolated example that validates output automatically, run the
[CLI demo](https://github.com/NAEOS-foundation/naeos/tree/main/examples/demo-cli):

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

The demo writes `context.md`, `summary.md`, and generated artifacts under
`examples/demo-cli/.run/`.

## Troubleshooting

- If `naeos` is not found after `go install`, add `$(go env GOPATH)/bin` to
  your `PATH`, or use the binary built from source.
- If validation fails, check the reported field and compare the specification
  with the [Spec Language guide](/docs/spec-language/).
- If generation succeeds but files are not where expected, inspect
  `naeos.yaml` and its `output_dir` value.

## Next Steps

- Explore the [CLI Reference](/docs/cli-reference/) for all available commands
- Read about the [Architecture](/docs/architecture/) to understand how NAEOS works
- Check out the [Features](/features/) page for a complete overview

## Download

- [Getting Started PDF](/downloads/naeos-getting-started.pdf)