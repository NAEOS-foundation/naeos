---
title: Core Principles
description: The 8 engineering principles that guide all NAEOS design decisions.
weight: 2
---

NAEOS is built on 8 core principles that guide all technical and architectural decisions. These principles ensure consistency, reliability, and extensibility across the platform.

## 1. Specification is the Single Source of Truth

The specification defines what exists, what is valid, and what is required. Code, documentation, and engineering decisions must stay synchronized with it.

**Implications:**
- All changes start from the specification
- Code is a product of the specification, not the other way around
- Documentation must be consistent with the specification
- The spec is versioned and immutable once published

## 2. Architecture Precedes Implementation

Architectural decisions must be made and documented before code is written.

**Implications:**
- Create Architecture Decision Records (ADRs) for every significant decision
- Document trade-offs explicitly
- Review architecture before implementation begins
- Use NEIR as the canonical architecture model

## 3. Everything is Declarative

All configuration and behavior are defined declaratively.

**Implications:**
- YAML/JSON specifications are the only way to define the system
- No imperative configuration scripts
- All state can be reconstructed from the specification
- Deterministic output from the same input

## 4. Everything is an Artifact

All pipeline outputs are artifacts that can be versioned, reviewed, and audited.

**Implications:**
- Code, configuration, and documentation are artifacts
- All artifacts are documented and traceable
- Artifacts can be cached and shared
- Content-hash deduplication in the artifact store

## 5. Everything is Versioned

All components use semantic versioning.

**Implications:**
- Specifications have versions (minimum: 0.1.0)
- Artifacts carry version metadata
- Dependencies are locked
- Schema versions enable migration paths

## 6. Everything is Extensible

The system is designed to be extended without modifying the core.

**Implications:**
- Plugin system for custom extensions
- Custom adapters for new output targets
- Custom profiles for industry-specific needs
- WASM runtime for sandboxed plugin execution

## 7. Small Kernel, Powerful Extensions

The kernel is minimal; all features live in extensions.

**Implications:**
- Kernel handles only core services (registry, event bus, telemetry)
- All features are in separate packages
- Extensions can be installed and uninstalled independently
- Plugins are first-class citizens

## 8. Human Review Before Automation

Automation must be preceded by human review.

**Implications:**
- Review artifacts before deployment
- Policy evaluation for governance
- Audit trail for all changes
- Guardrails prevent unintended automated actions

## Applying These Principles

These principles are enforced through:

- **Pipeline validation** — The validator checks spec compliance at every stage
- **Governance policies** — Policy rules block or warn on violations
- **Audit trails** — All changes are logged and traceable
- **Code review** — Human review is required before merging

See also: [Architecture](/docs/architecture/), [Governance](/docs/governance/), [Pipeline Engine](/docs/pipeline-engine/)
