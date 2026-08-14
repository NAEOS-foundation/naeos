---
title: "Pipeline Engine Deep Dive: How NAEOS Transforms Specs Into Systems"
description: "A technical walkthrough of the 11-stage DAG pipeline — from YAML parsing to multi-language code generation."
date: 2026-07-20
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Every time you run `naeos run --input-file spec.yaml`, an 11-stage directed acyclic graph (DAG) pipeline fires up. Each stage is independently observable, extensible via plugins, and designed to handle specs of any scale.

In this post, we'll walk through every stage with real code and outputs.

## The Pipeline at a Glance

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

Two coordination stages run between validation and scheduling: **Build Graph** (constructs the execution DAG) and **Policy Evaluation** (enforces governance rules).

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

The validator checks the NEIR model for structural correctness:

| Check | What it Enforces |
|-------|------------------|
| Project | `project:` section present with a non-empty name |
| Modules | At least one module, each with `name` and `path`; unique names |
| Services | Valid `name`; port in range 0–65535; duplicate ports flagged as warnings |
| Architecture | Pattern is one of `layered`, `clean`, `hexagonal`, `microkernel`, `event-driven`, `cqrs`, `monolith`, `monolithic`, `microservices`, `serverless` |
| Schema conformance | Optional — spec checked against the NEIR JSON Schema when a schema source is configured |

Optionally, policy rules are evaluated against the generated model; violations fail the pipeline with the exact rule that was broken.

## Stage 6: Build Graph

The pipeline constructs the execution DAG from the NEIR model — every module and service becomes a node, and dependencies become edges. This graph drives what runs next.

## Stage 7: Policy Evaluation

Governance policies (from `pipeline.Config.Policies`) are evaluated against the model. A violation aborts the run before any code is generated — governance is enforced at build time, not after the fact.

## Stage 8: Schedule

The DAG scheduler identifies parallel execution groups, performs topological sort of dependent tasks, and supports incremental builds.

```text
Level 0: [parse, normalize, resolve]
Level 1: [build NEIR, validate]
Level 2: [schedule]  # itself
Level 3: [generate]
Level 4: [review, write artifacts]
```

Modules with no interdependencies generate in parallel.

## Stage 9: Generate

This is where code hits the disk. Template-driven generators for each target language create project files, module scaffolds, service stubs, tests, Dockerfiles, and CI configs.

```bash
naeos run --input-file spec.yaml --language go --language typescript
```

Each language adapter implements a common interface:

```go
type Generator interface {
    GenerateModule(module Module) ([]Artifact, error)
    GenerateService(service Service) ([]Artifact, error)
    GenerateTests(module Module) ([]Artifact, error)
}
```

## Stage 10: Review

Generated artifacts are passed through the review engine, which produces review results attached to the pipeline output — a lightweight linting pass before anything is written.

## Stage 11: Write Artifacts

Everything lands in the output directory:

```
./out/
├── src/              # Generated source code
├── docs/             # Documentation bundles
├── ai/               # AI context files
├── deploy/           # Kubernetes manifests, Dockerfiles
└── report.json       # Build report with artifacts manifest
```

The write stage also emits a machine-readable manifest so CI systems can inspect what was produced.

## Running the Pipeline

```bash
# Full pipeline
naeos run --input-file spec.yaml

# AI instruction compilation (separate command)
naeos ai compile --input-file spec.yaml --target claude

# Watch mode with hot-reload
naeos watch --input-file spec.yaml
```

## What's Next

The 11-stage pipeline is the engine room of NAEOS. We're working on:

- **Richer stage hooks** — Today hooks attach at `BeforeParse`/`AfterParse`, `BeforeRun`/`AfterRun`, and `BeforeGenerate`/`AfterGenerate`; we're working toward finer-grained injection points between any two stages
- **Wider incremental caching** — `--cache-dir` already skips stages whose inputs haven't changed; we're expanding it to more stages

For a complete reference of every stage configuration option, see the [Pipeline Engine documentation](/docs/pipeline-engine/).
