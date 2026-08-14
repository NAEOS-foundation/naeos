---
title: Pipeline Engine
description: The 11-stage DAG pipeline — from parsing to artifact output.
---

## Overview

The NAEOS pipeline engine is an 11-stage directed acyclic graph (DAG) that transforms raw YAML/JSON specifications into validated, multi-language outputs. Stages are independently observable and extensible via plugins.

## Pipeline Stages

```text
┌────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐ ┌─────────┐
│ Parse  │→│Normalize │→│Resolve │→│Build NEIR│→│Validate │
└────────┘ └──────────┘ └────────┘ └──────────┘ └────┬────┘
                                                     │
┌──────────┐ ┌────────┐ ┌───────┐ ┌──────────┐      │
│ Write    │←│ Review │←│Generate│←│Schedule  │←─────┘
│ Artifacts│ └────────┘ └───────┘ └──────────┘
└──────────┘
```

The execution pipeline also includes two coordination stages between validation and scheduling:

```text
Validate → Build Graph → Policy Evaluation → Schedule → Generate → Review → Write Artifacts
```

Governance policies (`pipeline.Config.Policies`) are evaluated after the graph is built — a violation aborts the run before any code is generated.

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

### 4. Build NEIR

Builds the NEIR (NAEOS Engineering Intermediate Representation):
- Assembles the canonical model from normalized data
- Applies architecture patterns and templates
- Constructs the module and service graph
- Generates internal identifiers and metadata

### 5. Validate

Comprehensive validation including:
- Semantic model validation
- Optional schema conformance checking (via schema registry)
- Cross-module reference validation
- Business rule validation via plugins

### 6. Build Graph

Builds the execution DAG from the NEIR model:
- Module and service graph construction
- Dependency ordering for downstream scheduling

### 7. Policy Evaluation

Evaluates governance policy rules against the generated model:
- Multi-policy rule evaluation
- Fails the pipeline on policy violations

### 8. Schedule

DAG-based task scheduling:
- Parallel execution group identification
- Topological sort of dependent tasks
- Resource-aware scheduling
- Incremental build support

### 9. Generate

Multi-language code generation:
- Template-driven output per language
- Per-language adapters (Go, TypeScript, Python, Java, Rust, FastAPI, Actix Web)
- Concurrent generation across modules
- Artifact manifest creation
- Optional result caching

### 10. Review

Reviews generated artifacts:
- Automated artifact review and linting
- Review results attached to the pipeline result

### 11. Write Artifacts

Writes all artifacts to disk:
- Generated source code
- Documentation bundles
- AI context files
- Deployment manifests
- Build reports and summaries

## Pipeline Configuration

The pipeline is configured via `naeos.yaml` (or `naeos.json`):

```yaml
pipeline:
  name: demo            # pipeline name
  output_dir: ./output  # where artifacts are written
  verbose: false        # verbose stage logging
  language: [go]        # generation targets
```

Caching is enabled per-run with `--cache-dir` (see below); it is not a
config-file key.

## Running the Pipeline

```bash
# Full pipeline
naeos run --input-file spec.yaml

# Restricted pipeline with preview (write nothing to disk)
naeos run --input-file spec.yaml --dry-run

# Target specific languages only
naeos run --input-file spec.yaml --language go --language typescript

# Enable caching between runs
naeos run --input-file spec.yaml --cache-dir .naeos-cache

# Profile the pipeline
naeos run --input-file spec.yaml --profile --profile-out profile.json

# Watch mode with hot-reload
naeos watch --input-file spec.yaml
```

## Related

- [CLI Reference](/docs/cli-reference/) — full flag reference for `naeos run` and `naeos watch`
- [Architecture](/docs/architecture/) — the pipeline within the broader platform
- [Plugin SDK](/docs/plugin-sdk/) — extending the pipeline with hooks and plugins