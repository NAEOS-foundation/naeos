---
title: Roadmap
description: The NAEOS execution roadmap from public consistency to Golden Path adoption, ecosystem development, and later commercialization.
---

## Current strategy

NAEOS is sequencing execution around **proof first, adoption second, ecosystem third, commercialization last**.

The current canonical product direction is:

> **The Engineering Control Plane for AI Coding Agents**

The canonical control model is:

```text
Specification
→ NEIR
→ Validation + Policy
→ Agent Intent
→ Authorized Execution
→ Observation
→ Evidence
→ Independent Verification
```

## P0 — Public consistency

**Status: ACTIVE**

Synchronize the public surfaces with the repository's actual engineering state.

- Website positioning and navigation
- SEO titles, descriptions, canonical messaging, and structured public copy
- Whitepaper
- FAQ
- About
- Public roadmap and documentation references
- Distinction between software release versions and engineering experiment milestones

## P1 — Golden Path

**Status: NEXT**

Make the existing control-plane proof the primary developer demonstration.

```text
Control Plane
→ Golden Path
→ P1.6–P1.10 Reference Demo
→ Evidence
→ Independent Verification (P1.11)
```

The goal is a fresh-checkout path that a technical evaluator can run, inspect, and understand in minutes.

## P2 — External adoption

**Status: GATED BY P1**

Work with the first **5–10 developers or repositories**.

Capture:

- successful evaluation paths;
- integration friction;
- deviations from the documented Golden Path;
- failure modes;
- missing documentation;
- evidence quality;
- requests that recur across independent evaluators.

These observations become the input for the next engineering priorities.

## P3 — Ecosystem

**Status: GATED BY P2**

Prioritize SDKs, integrations, and marketplace capabilities only where external adoption evidence shows a recurring need.

The objective is an ecosystem that extends the control boundary without fragmenting its governance or evidence model.

## P4 — Commercialization

**Status: GATED BY P2**

Cloud and Enterprise packaging follows technical adoption evidence rather than preceding it.

The commercial path should be informed by real deployment boundaries, operational requirements, governance needs, and evidence collected from external use.

## Completed engineering proof

- **P1.6** — Public Control-Plane Golden Path — DONE
- **P1.7** — Policy Change Mid-Run / Stale Authorization — DONE
- **P1.8** — Atomic Execution Commit Boundary — DONE
- **P1.9** — Evidence & Verification Plane — DONE
- **P1.10** — Evidence Query & Audit API — DONE
- **P1.11** — Independent Verifier CLI — DONE

P1.11 is a verification capability supporting the proof path, not a reason to open another feature-expansion cycle.

## Strategic sequence

```text
Public consistency
→ Reproducible Golden Path
→ 5–10 external evaluators
→ Evidence-backed ecosystem priorities
→ Cloud / Enterprise
```

The project is intentionally not prioritizing premature dashboard expansion, provider-specific AI features, broad marketplace growth, or commercial packaging before the technical adoption evidence exists.