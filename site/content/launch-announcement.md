---
title: NAEOS Launch Announcement
description: Introducing NAEOS, the open-source engineering control plane for AI coding agents, with explicit policy, authorization, execution, evidence, and verification.
---

# NAEOS Launch Announcement

> NAEOS is the Engineering Control Plane for AI Coding Agents.

NAEOS helps engineering teams define intent, enforce policy, authorize execution, capture durable evidence, and independently verify outcomes across AI-assisted engineering workflows.

## Why this matters

AI coding agents can generate and propose changes quickly. Engineering teams still need a system to decide what is authorized, what executes, and what evidence remains after execution.

NAEOS is being built around this control boundary—not simply around prompts or code generation.

## The control model

```
Specification
→ NEIR
→ Validation + Policy
→ Agent Intent
→ Authorized Execution
→ Observation
→ Evidence
→ Independent Verification
```

## Current public proof

The repository's public proof path includes the control-plane workflow, Golden Path, reference demo, evidence records, and independent verifier.

- P1.6–P1.10 form the primary Golden Path proof sequence.
- P1.11 provides independent verification of serialized evidence.
- The proof scope is evidence-based and does not imply blanket production assurance or compliance certification.

NAEOS retains specification-first engineering capabilities: NEIR connects engineering intent to validation, generated artifacts, and AI context.

## Evaluate NAEOS

For a reproducible technical evaluation, start with the public Control Plane and run the Golden Path and Reference Demo from a fresh checkout. Inspect the resulting evidence and verify it independently.

- [Repository](https://github.com/NAEOS-foundation/naeos)
- [Getting Started](https://naeos.dev/docs/getting-started)
- [Architecture](https://naeos.dev/architecture-overview)
- [Whitepaper](https://naeos.dev/whitepaper)
- [Roadmap](https://naeos.dev/roadmap)

We are building in public. Technical feedback, reproducible issues, and external evaluation evidence will guide the next phase.