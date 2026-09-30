---
title: About NAEOS
description: NAEOS is an open-source engineering control plane for AI coding agents, built around governance, authorized execution, evidence, and verification.
---

## What NAEOS is

NAEOS (Nusantara AI Engineering Operating System) is an open-source, vendor-neutral engineering control plane for AI coding agents.

The project focuses on a specific engineering boundary: **an AI agent may propose work, but the engineering system must determine what is authorized, what executes, and what evidence remains.**

The current canonical model is:

```text
Specification
    ↓
NEIR
    ↓
Validation + Policy
    ↓
Agent Intent
    ↓
Authorized Execution
    ↓
Observation
    ↓
Evidence
    ↓
Independent Verification
```

## Why we are building it

AI-assisted software development changes the speed and shape of engineering work. The challenge is no longer only how to generate code; teams also need explicit authorization boundaries, reproducible execution, traceability, and evidence that can be inspected after the agent has finished.

NAEOS is being built to provide that layer without requiring a single AI vendor, agent runtime, or cloud provider.

## Current proof

The repository has moved beyond a purely conceptual architecture. The public proof path combines the live Control Plane, the Golden Path, the Reference Demo, evidence records, and the independent verifier.

P1.6–P1.10 are the primary Golden Path proof sequence. P1.11 provides independent verification of serialized evidence.

## Project principles

- **Governance before execution** — policy and authorization define the execution boundary.
- **Evidence over claims** — important behavior should leave inspectable evidence.
- **Independent verification** — verification should not depend on the same agent that performed the work.
- **Vendor neutrality** — the control boundary should not depend on one AI provider.
- **Reproducibility** — technical evaluators should be able to reproduce the documented proof.
- **Human accountability** — AI assists engineering; humans remain accountable for consequential changes.

## Where the project is going

The execution strategy is deliberately staged:

1. **P0 — Public consistency:** synchronize website, SEO, whitepaper, FAQ, About, and public documentation with the repository truth.
2. **P1 — Golden Path:** make P1.6–P1.10 the primary reproducible developer demonstration.
3. **P2 — External adoption:** work with the first 5–10 developers or repositories and capture evidence and failure modes.
4. **P3 — Ecosystem:** prioritize SDKs, integrations, and marketplace capabilities based on adoption evidence.
5. **P4 — Commercialization:** evaluate Cloud/Enterprise packaging after technical adoption evidence exists.

This sequence keeps product expansion downstream of reproducible technical proof and external use.