---
title: Architecture
description: Deep dive into NAEOS architecture and design principles.
---

## System Architecture

NAEOS is built on a layered architecture with five main layers:

```text
┌─────────────────────────────────────────────────────────┐
│                      Input Layer                          │
│            Spec YAML/JSON · CLI Commands                  │
├─────────────────────────────────────────────────────────┤
│                      Core Runtime                         │
│   Parser · Normalizer · Resolver · Validator · Scheduler │
├─────────────────────────────────────────────────────────┤
│                    Reasoning Layer                         │
│            Reasoning Graph · Knowledge Graph               │
├─────────────────────────────────────────────────────────┤
│                    Generation Layer                        │
│   Generator · Adapters · Template Engine · Compiler       │
├─────────────────────────────────────────────────────────┤
│                      Output Layer                          │
│   NEIR Model · Code · Docs · AI Context · Manifests       │
└─────────────────────────────────────────────────────────┘
```

### 1. Input Layer
The entry point where specifications (YAML/JSON) and CLI commands enter the system.

### 2. Core Runtime
Handles parsing, normalization, cross-reference resolution, validation, and DAG-based scheduling.

### 3. Reasoning Layer
The decision-making layer with a reasoning graph for traceability and a knowledge graph for domain understanding.

### 4. Generation Layer
Multi-language code generation with per-language adapters and AI instruction compilation.

### 5. Output Layer
Produces the NEIR model, generated code, documentation, AI context bundles, and deployment manifests.

## Design Principles

- **Human-readable specifications** — YAML/JSON as the single source of truth
- **Machine-readable NEIR** — Canonical intermediate representation for all downstream processing
- **Vendor neutral** — Multi-language, multi-cloud, multi-AI-platform
- **Extensible** — Adapters, plugins, and profiles for customization
- **Deterministic** — Same input always produces the same output

## Key Components

### NEIR Model
The NAEOS Engineering Intermediate Representation describes the entire system: project, architecture, modules, services, APIs, storage, infrastructure, security, AI, documentation, deployment, testing, and metadata.

### Pipeline Engine
A 9-stage DAG-based pipeline: Parse → Normalize → Resolve → Build → Validate → Schedule → Generate → Compile → Export.

### AI Compiler
Transforms NEIR into AI instruction sets for 6 target platforms: GitHub Copilot, Claude Code, Cursor, Gemini CLI, Codex, and OpenCode.

### Governance
Policy evaluation, RBAC, audit trails, and artifact review workflows.