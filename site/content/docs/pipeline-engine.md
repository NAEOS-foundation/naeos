---
title: Pipeline Engine
description: The 9-stage DAG pipeline — from parsing to export.
---

## Overview

The NAEOS pipeline engine is a 9-stage directed acyclic graph (DAG) that transforms raw YAML/JSON specifications into validated, multi-language outputs. Each stage is independently observable and can be extended via plugins.

## Pipeline Stages

```text
┌────────┐ ┌──────────┐ ┌────────┐ ┌───────┐ ┌─────────┐
│ Parse  │→│Normalize │→│Resolve │→│ Build │→│Validate │
└────────┘ └──────────┘ └────────┘ └───────┘ └─────┬───┘
                                                   │
┌────────┐ ┌──────────┐ ┌─────────┐ ┌────────┐    │
│ Export │←│ Compile  │←│Generate │←│Schedule│←───┘
└────────┘ └──────────┘ └─────────┘ └────────┘
```

### 1. Parse

Reads and parses YAML/JSON specification files. Supports:
- Variable interpolation (`${var}`)
- Environment variable resolution (`$env{VAR}`)
- Multi-file composition via `$include`
- Schema version validation

### 2. Normalize

Normalizes data structures for consistent downstream processing:
- Converts shorthand notations to canonical form
- Applies default values
- Validates type constraints
- Merges included files

### 3. Resolve

Resolves cross-references and dependencies:
- `$ref{path}` resolution across the spec tree
- External reference resolution
- Circular reference detection
- Dependency graph construction

### 4. Build

Builds the NEIR (NAEOS Engineering Intermediate Representation):
- Assembles the canonical model from normalized data
- Applies architecture patterns and templates
- Constructs the module and service graph
- Generates internal identifiers and metadata

### 5. Validate

Comprehensive validation including:
- Schema conformance checking
- Circular dependency detection
- Cross-module reference validation
- Policy rule evaluation
- Business rule validation via plugins

### 6. Schedule

DAG-based task scheduling:
- Parallel execution group identification
- Topological sort of dependent tasks
- Resource-aware scheduling
- Incremental build support

### 7. Generate

Multi-language code generation:
- Template-driven output per language
- Per-language adapters (Go, TypeScript, Python, Java, Rust)
- Concurrent generation across modules
- Artifact manifest creation

### 8. Compile

Compiles NEIR into AI instruction sets:
- Platform-specific context file generation
- Architecture-aware prompt construction
- Dependency and convention encoding
- 6 target platforms supported

### 9. Export

Exports all artifacts:
- Generated source code
- Documentation bundles
- AI context files
- Deployment manifests
- Build reports and summaries

## Pipeline Configuration

The pipeline can be configured via `naeos.yaml`:

```yaml
pipeline:
  stages:
    - parse
    - normalize
    - resolve
    - build
    - validate
    - schedule
    - generate
    - compile
    - export
  parallel: true
  cache: true
  output_dir: ./generated
```

## Running the Pipeline

```bash
# Full pipeline
naeos run --input-file spec.yaml

# Skip specific stages
naeos run --input-file spec.yaml --skip compile,export

# Run specific stages only
naeos run --input-file spec.yaml --only validate,generate

# Watch mode with hot-reload
naeos watch --input-file spec.yaml
```
