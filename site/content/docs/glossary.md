---
title: Glossary
description: Key terms and definitions used throughout the NAEOS platform.
weight: 22
---

| Term | Definition |
|------|------------|
| **NAEOS** | Nusantara Engineering & Architecture Operating System — an open-source declarative platform engineering system |
| **NEIR** | NAEOS Engineering Intermediate Representation — the unified intermediate model that represents the entire project |
| **Spec** | Specification — a YAML or JSON document that defines the project, modules, services, and architecture |
| **Pipeline** | The processing chain: parse → normalize → resolve → build → validate → schedule → generate → compile → export |
| **Kernel** | The core runtime component that manages the service registry, event bus, telemetry, and lifecycle |
| **Artifact** | Any output produced by the pipeline: code, configuration, documentation, or AI context |
| **Profile** | An industry-specific preset configuration (SaaS, AI Agent, FinTech, Healthcare, Government) |
| **Adapter** | An output generator for a specific target (Copilot, Claude, Cursor, Gemini, Codex, OpenCode) |
| **Compiler** | The component that transforms NEIR into AI instruction sets for coding assistants |
| **Context Bundle** | A project summary in markdown or plain text format, optimized for LLM consumption |
| **Governance** | The system of policies, validation rules, and review workflows that enforce standards |
| **Policy** | A rule evaluated during pipeline execution (operators: exists, not_empty, contains, gt, lt, in) |
| **Schema Version** | The SemVer version of the specification format (minimum: 0.1.0) |
| **Module** | A unit of code within a project, defined by name, path, and dependencies |
| **Service** | A runtime component (http, grpc, worker, cli, job) with endpoints and configuration |
| **Endpoint** | An API entry point defined by method, path, and action |
| **DAG** | Directed Acyclic Graph — the data structure used for dependency resolution and task scheduling |
| **Artifact Store** | Persistent storage for pipeline artifacts with content-hash deduplication |
| **Migration** | The process of upgrading a specification from one schema version to another |
| **LSP** | Language Server Protocol — provides IDE features like autocomplete and diagnostics for spec files |
| **MCP** | Model Context Protocol — enables AI agent integration with the NAEOS runtime |
