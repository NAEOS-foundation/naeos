---
title: Schema Registry
description: Versioned JSON Schemas for validating NAEOS source specifications and serialized NEIR model documents.
---

## Overview

The registry publishes separate [JSON Schemas](https://json-schema.org/) for NAEOS source specification files and serialized NEIR model documents. Use the input-spec schema for `.naeos.yaml` and `.naeos.json` source files. Use the NEIR model schema for JSON emitted by NAEOS.

## Latest Schema

The current stable schemas are:

- **Input specification:** [`/schemaregistry/spec/v1/spec.json`](/schemaregistry/spec/v1/spec.json)
- **Serialized NEIR model:** [`/schemaregistry/v1/neir.json`](/schemaregistry/v1/neir.json)

| Contract | Versioned URL | Latest URL |
|----------|---------------|------------|
| Source input spec | [`/schemaregistry/spec/v1/spec.json`](/schemaregistry/spec/v1/spec.json) | [`/schemaregistry/spec/latest.json`](/schemaregistry/spec/latest.json) |
| Serialized NEIR model | [`/schemaregistry/v1/neir.json`](/schemaregistry/v1/neir.json) | [`/schemaregistry/latest.json`](/schemaregistry/latest.json) |

## Usage

### Editor Integration

Add the input-spec `$schema` URI to your source file:

```yaml
# naeos.yaml
$schema: https://registry.naeos.dev/schemaregistry/spec/latest.json
project: my-project
modules:
  - name: core
    path: ./internal/core
```

This enables IntelliSense and inline validation in editors like VS Code, JetBrains, and others that support JSON Schema.

### CLI Validation

Use the NAEOS CLI to validate a source spec against the input schema:

```bash
naeos schema validate spec.yaml
naeos schema validate spec.json --output json
```

### Programmatic Use

Fetch the input-spec schema programmatically:

```bash
curl -s https://registry.naeos.dev/schemaregistry/spec/latest.json
```

## Versioning

Input-spec schema versions and NEIR model schema versions are separate contracts. A change to one does not automatically version the other. Breaking changes require a new versioned path for the affected contract; the `latest` aliases point to the newest stable version.

| Contract | Schema version | Notes |
|----------|---------------|-------|
| Input spec | v1 | Source YAML/JSON; `project` is a string (or legacy top-level `name`) |
| NEIR model | v1 | Serialized model; `project` is an object |

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
