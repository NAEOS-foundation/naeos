---
title: "From One Specification to a Generated Go and TypeScript Project"
description: "A reproducible NAEOS CLI demo showing validation, AI context generation, and multi-language artifact generation."
date: 2026-09-08
author: "NAEOS Foundation"
categories: ["tutorial", "case-study"]
---

This case study uses the checked-in [NAEOS CLI demo](https://github.com/NAEOS-foundation/naeos/tree/main/examples/demo-cli). It is an executable example, not a claim about a production deployment.

## Watch the demo

<video controls poster="/downloads/naeos-marketing-poster.png" width="100%">
  <source src="/downloads/naeos-terminal-demo-en.mp4" type="video/mp4" />
  Your browser does not support the video tag.
</video>

## The example

The specification describes a small `demo-app` with:

- `auth` and `api` modules,
- an `api` dependency on `auth`,
- an HTTP `gateway` service on port 8080,
- a hexagonal architecture pattern,
- Go and TypeScript generation targets.

The engineering intent is kept in [`spec.yaml`](https://github.com/NAEOS-foundation/naeos/blob/main/examples/demo-cli/spec.yaml), while pipeline settings live in [`naeos.yaml`](https://github.com/NAEOS-foundation/naeos/blob/main/examples/demo-cli/naeos.yaml).

## Run it locally

From the repository root:

```bash
go build -o naeos ./cmd/naeos
./examples/demo-cli/run-demo.sh
```

The demo runs three stages:

```text
specification
    ├── validate
    ├── context.md
    └── run
          └── generated/
```

The current run reports 61 generated artifacts. The script then verifies that the context bundle, project README, Go module, and TypeScript package were created and writes a `summary.md` file.

## What the output represents

The generated directory contains module folders, Go source files and tests, TypeScript package metadata, a Dockerfile, CI configuration, and architecture documentation. These artifacts are derived from the same specification and can be inspected before being adopted by a project.

The context bundle is a compact Markdown summary containing the project, modules, service, and dependency graph. It can be reviewed by a developer or used as input to an AI tool.

## Why this workflow matters

The value of this example is the traceable path:

```text
spec.yaml
→ validation
→ context bundle
→ generated artifacts
→ inspectable summary
```

This is deliberately smaller than a real system. Its purpose is to make the workflow easy to reproduce and provide a concrete starting point for experiments.

## Try a different AI target

The default demo does not call an LLM. With the required provider configuration, compile the same specification for an AI tool:

```bash
naeos ai compile \
  --input-file examples/demo-cli/spec.yaml \
  --target opencode
```

Available adapters should be checked against the current [AI compiler documentation](/docs/ai-compiler/).

## Evidence and boundaries

- The commands and files referenced above are present in the NAEOS repository.
- The artifact count is produced by the CLI during the demo run and may change as generators evolve.
- This example does not demonstrate a deployed production service, customer adoption, performance benchmarks, or enterprise compliance.

Start with the [CLI demo source](https://github.com/NAEOS-foundation/naeos/tree/main/examples/demo-cli) and the [Getting Started guide](/docs/getting-started/).
