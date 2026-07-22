---
title: Spec Language
description: The NAEOS Specification Language v2 — syntax, features, and capabilities.
---

## Overview

The NAEOS Specification Language (NSL) v2 is a declarative YAML-based language for describing engineering systems. A single spec file captures project structure, architecture, modules, services, dependencies, and deployment targets.

## Syntax Features

### Variable Interpolation

Use `${var}` syntax to reference values within the same specification:

```yaml
project: my-service
version: "1.0"
modules:
  - name: api
    path: ./services/${project}
```

### Environment Variables

Resolve environment variables at pipeline runtime with `$env{VAR}`:

```yaml
services:
  - name: api
    port: $env{API_PORT}
    host: $env{HOST}
```

### Cross-References

Reference other parts of the spec with `$ref{path}`:

```yaml
modules:
  - name: auth
    path: ./auth
  - name: api
    path: ./api
    dependencies: [$ref{modules[0].name}]
```

### Multi-File Composition

Compose specs from multiple files with `$include{file}`:

```yaml
project: my-app
$include{./shared/modules.yaml}
$include{./shared/services.yaml}
```

### Custom Functions

Transform values with built-in functions using `$fn{name(args)}`:

| Function | Description | Example |
|----------|-------------|---------|
| `upper` | Uppercase string | `$fn{upper(name)}` |
| `lower` | Lowercase string | `$fn{lower(name)}` |
| `slug` | URL-safe slug | `$fn{slug(project)}` |
| `default` | Default value | `$fn{default(port, 8080)}` |
| `len` | Length of list/string | `$fn{len(modules)}` |
| `coalesce` | First non-null | `$fn{coalesce(host, localhost)}` |

### Conditional Sections

Use `$if{condition}` and `$endif` for conditional blocks:

```yaml
services:
  - name: api
$if{env == production}
    replicas: 3
    resources:
      cpu: "2"
      memory: 4Gi
$endif
```

### Schema Versioning

Every spec includes a version field with auto-validation:

```yaml
# schema: v1
project: my-app
version: "1.0"
```

The pipeline validates that the spec version meets the minimum required version (v0.1.0).

## Full Example

```yaml
project: ecommerce-platform
version: "2.0"
$include{./profiles/saas.yaml}

modules:
  - name: api-gateway
    path: ./gateway
    dependencies: [user-service, product-service]
  - name: user-service
    path: ./services/users
    dependencies: [database]
  - name: product-service
    path: ./services/products
    dependencies: [database, search-engine]
  - name: database
    path: ./infra/db
  - name: search-engine
    path: ./infra/search

services:
  - name: gateway
    kind: reverse-proxy
    port: $fn{default($env{GATEWAY_PORT}, 8080)}
  - name: user-api
    kind: rest
    port: 9001
  - name: product-api
    kind: rest
    port: 9002

architecture:
  pattern: microservices

generation:
  languages: [go, typescript, python]
  output_dir: ./generated
```

## Best Practices

- Use `$include` to split large specs into maintainable modules
- Prefer `$env{VAR}` for environment-specific values
- Use `$ref` instead of duplicating values
- Add schema versioning to all specs
- Keep specs readable — YAML comments are preserved
