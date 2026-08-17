---
title: "Pipeline Engine Deep Dive: How NAEOS Transforms Specs Into Systems"
description: "A technical walkthrough of the 7-stage DAG pipeline — from YAML parsing to multi-language code generation."
date: 2026-07-20
author: "NAEOS Foundation"
categories: ["tutorial"]
---

Every time you run `naeos run --input-file spec.yaml`, a 7-stage directed acyclic graph (DAG) pipeline fires up. Each stage is independently observable, extensible via plugins, and designed to handle specs of any scale.

In this post, we'll walk through every stage with real code and outputs.

## The Pipeline at a Glance

```text
┌──────────┐   ┌─────────────┐   ┌──────────────┐   ┌──────────┐
│ Validate │ → │ Build Graph │ → │ Policy Eval  │ → │ Schedule │
└──────────┘   └─────────────┘   └──────────────┘   └────┬─────┘
                                                         │
┌───────────────┐   ┌──────────┐   ┌─────────────────┐   │
│ Write Artifacts│ ← │ Review   │ ← │    Generate     │ ←─┘
└───────────────┘   └──────────┘   └─────────────────┘
```

## Stage 1: Validate

The pipeline reads your YAML or JSON spec and validates it through a comprehensive multi-step process:

1. **Parse** — Converts the spec into an AST using Go's `yaml.v3` with custom node hooks for reference resolution. Handles variable interpolation (`${var}`), environment resolution (`$env{VAR}`), multi-file composition via `$include`, and schema version validation.

2. **Normalize** — Converts shorthand notations to canonical form, applies defaults, validates type constraints, and merges included files into a single tree.

3. **Resolve** — Resolves cross-references (`$ref{path}`), external references, and dependency graphs. Detects circular references and reports them with the full chain.

4. **Build NEIR** — Constructs the NAEOS Engineering Intermediate Representation — a fully typed, self-contained model of your system that includes modules, services, architecture patterns, infrastructure requirements, and governance metadata.

5. **Validate** — Checks the NEIR model for structural correctness:

| Check | What it Enforces |
|-------|------------------|
| Project | `project:` section present with a non-empty name |
| Modules | At least one module, each with `name` and `path`; unique names |
| Services | Valid `name`; port in range 0–65535; duplicate ports flagged as warnings |
| Architecture | Pattern is one of `layered`, `clean`, `hexagonal`, `microkernel`, `event-driven`, `cqrs`, `monolith`, `monolithic`, `microservices`, `serverless` |

```yaml
# spec.yaml
project: blog-platform
$include:
  - ./shared/data-model.yaml
variables:
  region: us-east-1
```

Under the hood, the parser uses Go's `yaml.v3` with custom node hooks for reference resolution. If a `$include` path is missing, the pipeline fails fast with a clear error.

## Stage 2: Build Graph

The pipeline constructs the execution DAG from the NEIR model — every module and service becomes a node, and dependencies become edges. This graph drives what runs next and enables parallel execution of independent modules.

## Stage 3: Policy Evaluation

Governance policies are evaluated against the NEIR model. A violation aborts the run before any code is generated — governance is enforced at build time, not after the fact.

```go
// Example policy rule
rules := []policy.Rule{
    {
        RuleID:    "require-testing",
        Condition: "exists:testing",
        Priority:  1,
        Action:    "block",
    },
}
```

## Stage 4: Schedule

The DAG scheduler identifies parallel execution groups, performs topological sort of dependent tasks, and supports incremental builds.

```text
Level 0: [validate]
Level 1: [build_graph]
Level 2: [policy_eval]
Level 3: [schedule]
Level 4: [generate]
Level 5: [review]
Level 6: [write_artifacts]
```

Modules with no interdependencies generate in parallel. Stage caching (`--cache-dir`) can skip this stage if inputs haven't changed.

## Stage 5: Generate

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

The NEIR model is what makes NAEOS language-agnostic. Whether you generate Go, TypeScript, Python, Java, or Rust, the source is the same NEIR tree.

## Stage 6: Review

Generated artifacts are passed through the review engine, which produces review results attached to the pipeline output — a lightweight linting pass before anything is written. The reviewer checks for rules like `no-todo`, `no-placeholder`, and code quality constraints.

## Stage 7: Write Artifacts

Everything lands in the output directory:

```
./out/
├── src/              # Generated source code
├── docs/             # Documentation bundles
├── ai/               # AI context files
├── deploy/           # Kubernetes manifests, Dockerfiles
└── report.json       # Build report with artifacts manifest
```

The write stage also emits a machine-readable manifest so CI systems can inspect what was produced. In dry-run mode, this stage is skipped.

## Running the Pipeline

```bash
# Full pipeline
naeos run --input-file spec.yaml

# With caching for faster rebuilds
naeos run --input-file spec.yaml --cache-dir .naeos-cache

# With profiling for performance analysis
naeos run --input-file spec.yaml --profile --profile-out profile.json

# AI instruction compilation (separate command)
naeos ai compile --input-file spec.yaml --target claude

# Watch mode with hot-reload
naeos watch --input-file spec.yaml
```

## What's Next

The 7-stage pipeline is the engine room of NAEOS. We're working on:

- **Richer stage hooks** — Today hooks attach at `BeforeParse`/`AfterParse`, `BeforeRun`/`AfterRun`, and `BeforeGenerate`/`AfterGenerate`; we're working toward finer-grained injection points between any two stages
- **Wider incremental caching** — `--cache-dir` already skips stages whose inputs haven't changed; we're expanding it to more stages

For a complete reference of every stage configuration option, see the [Pipeline Engine documentation](/docs/pipeline-engine/).
