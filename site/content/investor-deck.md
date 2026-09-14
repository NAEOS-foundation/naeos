---
title: Investor Deck
description: A concise investor brief for NAEOS, the engineering control plane for AI-native software delivery.
---

# NAEOS Investor Deck

> NAEOS is not another AI coding assistant. It is an engineering control plane that turns specifications into validated, governable, AI-ready software workflows.

## View or download the deck

- [Open the investor pitch deck in your browser](/NAEOS-PITCHDECK.pdf)
- [Download the investor pitch deck PDF](/NAEOS-PITCHDECK.pdf)

## Thesis

NAEOS is building the system layer for AI-native software engineering.

The current repository already demonstrates a coherent technical foundation: declarative specification input, pipeline execution, NEIR construction, policy and governance hooks, AI context generation, artifact review, evidence/audit support, CLI workflows, and demo control-plane capabilities.

## Why this matters now

AI coding tools are improving faster than many engineering systems are adapting.

Software delivery now requires more than code generation:

- consistent engineering intent
- architecture and dependency validation
- policy enforcement
- context management across tools
- traceability from spec to artifact
- governance around agent actions
- reproducible execution

This is the gap NAEOS is designed to address.

## What exists today

The repository already contains meaningful proof of product substance:

- `pkg/pipeline` — the main execution pipeline
- `internal/neir` — the NEIR model and validation components
- `internal/governance/policy` — policy evaluation and enforcement hooks
- `internal/context/bundle` — AI context generation
- `internal/evidence` — evidence and audit support
- `cmd/naeos` — CLI workflows for validate, run, context, policy, and demo flows

For current source references, see:

- [README.md](https://github.com/NAEOS-foundation/naeos/blob/main/README.md)
- [ARCHITECTURE-OVERVIEW.md](https://github.com/NAEOS-foundation/naeos/blob/main/ARCHITECTURE-OVERVIEW.md)
- [WHITEPAPER.md](https://github.com/NAEOS-foundation/naeos/blob/main/WHITEPAPER.md)
- [ROADMAP.md](https://github.com/NAEOS-foundation/naeos/blob/main/ROADMAP.md)

## Strategic moat

### NEIR as the core differentiator

NEIR is the repository's central engineering representation. It gives NAEOS a structured model that can be validated, governed, transformed, documented, and supplied to downstream AI tools.

This offers a strong strategic advantage:

- specification is the source of truth
- the model is machine-readable
- validation and policy happen before generation
- context can be compiled consistently
- evidence can be attached to artifacts and execution

### Product differentiation

| Category | AI coding tools | NAEOS |
|---|---|---|
| Core role | Code generation | Engineering control plane |
| Source of truth | Prompt or file changes | Declarative specification + NEIR |
| Governance | Often external or ad hoc | Built into the pipeline |
| Validation | Variable or partial | Deterministic and structured |
| AI context | Often tool-specific | Compiled systematically |
| Auditability | Limited | Evidence and traceability built in |

## Open source and commercialization

NAEOS is open source under Apache 2.0 and built in public. That is an adoption advantage because it enables technical validation, external contribution, and ecosystem development.

A realistic commercialization path is to keep the open core broadly useful while monetizing the control-plane layer through:

- centralized governance and policy management
- audit and evidence workflows
- enterprise integrations
- managed control-plane services
- hosted or managed deployment surfaces
- team-level collaboration and compliance controls

## Current development stage

The repository is already substantial and demonstrably functional. The current implementation includes runnable CLI workflows, a working pipeline, policy evaluation, context generation, demo control-plane capabilities, and evidence/audit hooks.

The next priority is to turn that technical foundation into a sharper developer wedge and clearer commercial story:

- clearer proof of the control-plane workflow
- stronger end-to-end examples
- stronger governance and audit narratives
- stronger enterprise and partner positioning
- broader design-partner validation

## 12-month targets

These are targets, not current results:

- 10–20 design partners
- 100+ active engineering teams using the project
- 3–5 enterprise pilots
- 8–10 technical integrations across AI and developer tooling
- 50+ external contributors
- a stronger, repeatable control-plane workflow for real teams

## Investor takeaway

NAEOS is attractive to investors who understand developer infrastructure, AI systems, open-source platforms, and the emerging category of AI-native engineering tools.

The strongest argument is not that NAEOS is “another AI tool.” It is that NAEOS is building the system infrastructure that AI-native software delivery requires.

## Bottom line

NAEOS already has real technical substance. The opportunity is to turn that substance into a durable developer platform, a trusted control plane, and a strong long-term infrastructure company.
