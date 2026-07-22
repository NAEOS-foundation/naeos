---
title: Troubleshooting
description: Common issues and solutions when working with NAEOS.
weight: 16
---

This page covers common issues encountered when using NAEOS and how to resolve them.

## Installation

### `naeos: command not found`

After installing with `go install`, ensure your Go bin directory is in your PATH:

```bash
# Check if the binary exists
ls ~/go/bin/naeos

# Add to PATH (add to your shell profile)
export PATH="$HOME/go/bin:$PATH"
```

### Permission denied on macOS

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ~/go/bin/naeos
```

### Version mismatch

If you see version conflicts between CLI and API:

```bash
# Check CLI version
naeos version

# Ensure latest version
go install github.com/NAEOS-foundation/naeos/cmd/naeos@latest
```

## Specification Parsing

### `yaml: unmarshal errors`

This usually means your spec YAML has structural issues:

```yaml
# WRONG — services should be a list
services:
  api:
    kind: rest

# CORRECT
services:
  - name: api
    kind: rest
```

### `module not found in dependency graph`

A module references a dependency that doesn't exist:

```yaml
# WRONG — user-service depends on cache-service, but cache-service isn't defined
modules:
  - name: user-service
    dependencies: [cache-service]

# CORRECT — define cache-service first
modules:
  - name: cache-service
    path: ./cache
  - name: user-service
    path: ./users
    dependencies: [cache-service]
```

### Circular dependency detected

NAEOS doesn't allow circular dependencies between modules:

```
Error: circular dependency detected: A → B → C → A
```

Resolve by extracting the shared functionality into a separate module:

```
# BEFORE (circular)
A depends on B
B depends on C
C depends on A

# AFTER (resolved)
A depends on D, B
B depends on D, C
C depends on D
D (shared module, no dependencies)
```

## Generation

### `permission denied` when writing output

The output directory may have restrictive permissions:

```bash
# Fix permissions
chmod -R u+w ./output

# Or use a different output directory
naeos run --input-file spec.yaml --output-dir /tmp/output
```

### Generated code doesn't compile

This is rare but can happen with custom generation targets. Try:

```bash
# Re-run with verbose output to see what was generated
naeos run --input-file spec.yaml --verbose

# Check the generated files
find ./generated -name "*.go" -o -name "*.ts" | head -20
```

### Missing language adapter

If you specify a language that isn't installed:

```
Error: language adapter "rust" not found
```

Check available adapters:

```bash
naeos gen --list
```

Install missing adapters via the plugin system:

```bash
naeos plugin install rust-adapter
```

## AI Compiler

### `LLM API key not configured`

The AI compiler needs an API key for your chosen provider:

```bash
# Set the environment variable
export NAEOS_LLM_API_KEY="your-api-key"

# Or pass it directly
naeos compile --target copilot --api-key "your-api-key"
```

### Context bundle is empty

This usually means your spec doesn't have enough detail:

```yaml
# TOO SIMPLE — minimal context generated
project: my-app

# BETTER — more context in the bundle
project: my-app
description: E-commerce platform with event-driven microservices
modules:
  - name: api-gateway
    path: ./gateway
    description: Reverse proxy with rate limiting and JWT validation
    dependencies: [user-service, order-service]
```

### Compilation takes too long

Large specs with many modules can take time. Use incremental compilation:

```bash
# Only compile changed modules
naeos compile --incremental --input-file spec.yaml
```

## API Server

### `address already in use`

Another process is using port 8080:

```bash
# Find the process
lsof -i :8080

# Use a different port
naeos serve --port 9090
```

### CORS errors in browser

The API server allows localhost by default. For other origins, configure CORS:

```bash
naeos serve --cors-origins "https://myapp.com,https://staging.myapp.com"
```

### WebSocket connection fails

Ensure your proxy/load balancer supports WebSocket upgrades. The WebSocket endpoint is at `/ws`.

## Database

### `connection refused`

Check that your database is running and accessible:

```bash
# PostgreSQL
psql -h localhost -U postgres -c "SELECT 1"

# MySQL
mysql -h localhost -u root -e "SELECT 1"

# SQLite (file-based)
ls -la ./naeos.db
```

### Migration errors

If database migrations fail:

```bash
# Reset the database (WARNING: destroys data)
naeos db reset --confirm

# Or run migrations manually
naeos db migrate --verbose
```

## Performance

### Pipeline is slow

For large specs, try these optimizations:

```bash
# Use caching (skips unchanged stages)
naeos run --input-file spec.yaml --cache

# Generate only specific languages
naeos run --input-file spec.yaml --languages go,typescript

# Parallel generation (if supported)
naeos run --input-file spec.yaml --parallel
```

### High memory usage

Large specs with 100+ modules may use significant memory:

```bash
# Monitor memory usage
naeos run --input-file spec.yaml --profile memory

# Process modules in batches
naeos run --input-file spec.yaml --batch-size 20
```

## Getting Help

If your issue isn't covered here:

1. Check the [GitHub Issues](https://github.com/NAEOS-foundation/naeos/issues) for similar problems
2. Search the [GitHub Discussions](https://github.com/NAEOS-foundation/naeos/discussions)
3. Ask in the [Discord community](https://discord.gg/naeos)
4. Open a new issue with:
   - NAEOS version (`naeos version`)
   - Operating system and architecture
   - Full error message
   - Relevant spec snippet (redact sensitive data)

See also: [Getting Started](/docs/getting-started/), [Installation](/docs/installation/), [CLI Reference](/docs/cli-reference/)
