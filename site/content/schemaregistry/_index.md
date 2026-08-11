---
title: Schema Registry
description: JSON Schema registry for the NEIR specification format — versioned schemas for validating NAEOS Engineering Intelligence Representation documents.
---

## Overview

The NEIR Schema Registry hosts versioned [JSON Schema](https://json-schema.org/) definitions for the NAEOS Engineering Intelligence Representation (NEIR) specification format. These schemas enable IDE autocompletion, validation, and tooling support for `.naeos.yaml` and `.naeos.json` files.

## Latest Schema

The current stable schema is **[v1](/schemaregistry/v1/neir.json)**.

| Version | Schema URL | Status |
|---------|-----------|--------|
| v1 | [`/schemaregistry/v1/neir.json`](/schemaregistry/v1/neir.json) | Stable |
| latest | [`/schemaregistry/latest.json`](/schemaregistry/latest.json) | Latest stable |

## Usage

### Editor Integration

Add a `$schema` field to your NEIR specification file:

```yaml
# naeos.yaml
$schema: https://registry.naeos.dev/schemaregistry/latest.json
project: my-project
modules:
  - name: core
    path: ./internal/core
```

This enables IntelliSense and inline validation in editors like VS Code, JetBrains, and others that support JSON Schema.

### CLI Validation

Use the NAEOS CLI to validate a spec against the registry schema:

```bash
naeos schema validate spec.yaml
naeos schema validate spec.json --output json
```

### Programmatic Use

Fetch the schema programmatically:

```bash
curl -s https://registry.naeos.dev/schemaregistry/latest.json
```

## Versioning

Schema versions follow the NEIR specification version. Backward-incompatible changes to the NEIR model produce a new schema version. Minor additions (new optional fields) are additive within a version.

| NEIR Version | Schema Version | Notes |
|-------------|---------------|-------|
| 1.x | v1 | Initial stable schema |

## Schema Contents

The NEIR JSON Schema defines these top-level sections:

- `project` — Project metadata (name, version, description, license, authors)
- `architecture` — Architecture pattern and layers
- `domain` — Domain model with bounded contexts, aggregates, entities
- `modules` — Module definitions with paths and packages
- `components` — Component catalog with kinds and dependencies
- `services` — Service definitions with endpoints and ports
- `apis` — API specifications with protocols and schemas
- `storage` — Storage backends (SQL, NoSQL, cache, queue, blob)
- `infrastructure` — Infrastructure providers and resources
- `security` — Authentication, authorization, encryption, secrets
- `ai` — AI model integrations, prompts, context bundles
- `documentation` — Documentation guides, references, ADRs
- `deployment` — Deployment strategy, environments, scaling
- `testing` — Testing strategy, frameworks, coverage, fixtures
- `metadata` — Spec metadata with version tracking and timestamps
- `generation` — Code generation configuration
