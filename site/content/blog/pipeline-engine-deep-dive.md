---
title: "Pipeline Engine Deep Dive: How NAEOS Transforms Specs Into Systems"
description: "A technical walkthrough of the 9-stage DAG pipeline — from YAML parsing to multi-language code generation."
date: 2026-07-20
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Every time you run `naeos run --input-file spec.yaml`, a 9-stage directed acyclic graph (DAG) pipeline fires up. Each stage is independently observable, extensible via plugins, and designed to handle specs of any scale.

In this post, we'll walk through every stage with real code and outputs.

## The Pipeline at a Glance

```text
┌────────┐ ┌──────────┐ ┌────────┐ ┌───────┐ ┌─────────┐
│ Parse  │→│Normalize │→│Resolve │→│ Build │→│Validate │
└────────┘ └──────────┘ └────────┘ └───────┘ └─────┬───┘
                                                    │
┌────────┐ ┌──────────┐ ┌─────────┐ ┌────────┐    │
│ Export │←│ Compile  │←│Generate │←│Schedule│←───┘
└────────┘ └──────────┘ └─────────┘ └────────┘
```

## Stage 1: Parse

The pipeline reads your YAML or JSON spec and converts it into an AST. It handles variable interpolation (`${var}`), environment resolution (`$env{VAR}`), multi-file composition via `$include`, and schema version validation.

```yaml
# spec.yaml
project: blog-platform
$include:
  - ./shared/data-model.yaml
variables:
  region: us-east-1
```

Under the hood, the parser uses Go's `yaml.v3` with custom node hooks for reference resolution. If a `$include` path is missing, the pipeline fails fast with a clear error.

## Stage 2: Normalize

Raw parsed data is rarely consistent. The normalizer converts shorthand notations to canonical form, applies defaults, validates type constraints, and merges included files into a single tree.

For example, this shorthand:
```yaml
services:
  - name: api
    kind: http
    port: 8080
```

Gets expanded with defaults:
```yaml
services:
  - name: api
    kind: http
    protocol: http
    port: 8080
    host: "0.0.0.0"
    timeout: 30s
    retry:
      attempts: 3
      backoff: exponential
```

## Stage 3: Resolve

Cross-references are resolved here. The resolver walks the entire spec tree and resolves `$ref{path}` references, external references, and dependency graphs.

```yaml
modules:
  - name: api-gateway
    dependencies:
      - user-service
      - $ref{./shared/definitions.yaml#/modules/post-service}
```

The resolver detects circular references and reports them with the full chain — a critical safety net for complex specs.

## Stage 4: Build NEIR

The NEIR (NAEOS Engineering Intermediate Representation) is the canonical model. It's a fully typed, self-contained representation of your system that includes:

- Module and service graphs
- Architecture patterns and templates
- Infrastructure requirements
- Governance metadata

The NEIR model is what makes NAEOS language-agnostic. Whether you generate Go, TypeScript, Python, Java, or Rust, the source is the same NEIR tree.

```go
// Internal representation (simplified)
type NEIRModel struct {
    Project      string
    Modules      []Module
    Services     []Service
    Architecture Architecture
    Policies     []Policy
    Generation   GenerationConfig
}
```

## Stage 5: Validate

Seven validation layers run in sequence:

| Layer | What it Checks |
|-------|----------------|
| Schema conformance | Spec matches the NEIR JSON Schema |
| Circular deps | No module depends on itself transitively |
| Cross-module refs | All `$ref` targets exist |
| Policy rules | Custom policy conditions evaluated |
| Business rules | Plugin-supplied validators |
| Naming conventions | Module/service names follow pattern |
| Port conflicts | No two services bind the same port |

Validation errors include the exact path and context so you can fix them without guessing.

## Stage 6: Schedule

The DAG scheduler identifies parallel execution groups, performs topological sort of dependent tasks, and supports incremental builds.

```text
Level 0: [parse, normalize, resolve]
Level 1: [build, validate]
Level 2: [schedule]  # itself
Level 3: [generate]  
Level 4: [compile, export]
```

Modules with no interdependencies generate in parallel. In a 50-module project, this cuts generation time by 70% compared to sequential execution.

## Stage 7: Generate

This is where code hits the disk. Template-driven generators for each target language create project files, module scaffolds, service stubs, tests, Dockerfiles, and CI configs.

```bash
naeos run --input-file spec.yaml --language go,typescript
```

Each language adapter implements a common interface:

```go
type Generator interface {
    GenerateModule(module Module) ([]Artifact, error)
    GenerateService(service Service) ([]Artifact, error)
    GenerateTests(module Module) ([]Artifact, error)
}
```

## Stage 8: Compile

The AI compiler transforms NEIR into instruction sets for six AI platforms: GitHub Copilot, Claude Code, Cursor, Gemini CLI, OpenAI Codex, and OpenCode.

Each platform gets its own format — `copilot-instructions.md`, `CLAUDE.md`, `.cursorrules`, and so on. They all carry the same architectural knowledge but in the dialect each assistant understands.

```bash
naeos compile --all --input-file spec.yaml
```

## Stage 9: Export

Everything lands in the output directory:

```
./out/
├── src/              # Generated source code
├── docs/             # Documentation bundles
├── ai/               # AI context files
├── deploy/           # Kubernetes manifests, Dockerfiles
└── report.json       # Build report with artifacts manifest
```

The export stage also writes a machine-readable manifest so CI systems can inspect what was produced.

## Running the Pipeline

```bash
# Full pipeline
naeos run --input-file spec.yaml

# Skip AI compilation
naeos run --input-file spec.yaml --skip compile

# Watch mode with hot-reload
naeos watch --input-file spec.yaml
```

## What's Next

The 9-stage pipeline is the engine room of NAEOS. We're working on:

- **Incremental stage caching** — Skip stages whose inputs haven't changed
- **Parallel generation** — Concurrent multi-adapter execution via `errgroup`
- **Custom stage hooks** — Inject your own logic between any two stages

For a complete reference of every stage configuration option, see the [Pipeline Engine documentation](/docs/pipeline-engine/).
