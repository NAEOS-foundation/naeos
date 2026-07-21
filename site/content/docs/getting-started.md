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
naeos compile --all --input-file spec.yaml
```

## Next Steps

- Explore the [CLI Reference](/docs/cli-reference/) for all available commands
- Read about the [Architecture](/docs/architecture/) to understand how NAEOS works
- Check out the [Features](/features/) page for a complete overview