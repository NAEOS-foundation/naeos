---
title: Installation
description: Detailed installation instructions for NAEOS on all platforms.
---

## Go Install (Recommended)

Requires Go 1.25+:

```bash
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

## Docker

```bash
docker pull ghcr.io/naeos-foundation/naeos:latest
```

Verify the installation:

```bash
docker run --rm ghcr.io/naeos-foundation/naeos:latest naeos version
```

## Build from Source

```bash
git clone https://github.com/NAEOS-foundation/naeos.git
cd naeos
make build
```

The binary will be available at `./naeos`.

## Binary Releases

Download the latest binary for your platform from the [GitHub Releases page](https://github.com/NAEOS-foundation/naeos/releases).

Supported platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Verify Installation

```bash
naeos version
```

## Configuration

NAEOS uses a YAML configuration file. Create one with:

```bash
naeos init
```

This creates `naeos.yaml` in your current directory with sensible defaults.