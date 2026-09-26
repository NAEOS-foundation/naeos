# AI Engineering Governance Assessment

Summary: A structured technical assessment for teams evaluating governance and control boundaries around AI coding agents.
Last updated: 2026-09-26

## Overview

The NAEOS AI Engineering Governance Assessment is a technical discovery and architecture exercise for organizations already experimenting with AI coding agents.

It is designed to establish the current control boundary before discussing a pilot.

The assessment is not a compliance certification, security certification, production-readiness determination, or vendor scorecard.

## Who it is for

Primary participants:

- CTO / CIO / VP Engineering
- Head of Platform Engineering
- AI Engineering or Developer Productivity lead
- Principal / Staff Engineer
- Security Architect / DevSecOps lead
- Engineering governance or compliance engineering lead

The most useful entry point is usually the platform or AI engineering team that owns the agent workflow, with security and risk stakeholders joining where appropriate.

## Assessment scope

### 1. Agent inventory

Document:

- coding agents in use or under evaluation;
- repositories and environments they can access;
- tools and external systems they can invoke;
- identities and credentials available to them;
- human approval points;
- autonomous execution paths.

### 2. Intent and authorization

Determine:

- how engineering intent is represented;
- how requests are authenticated;
- how agent identity is established;
- where authorization decisions occur;
- whether authorization is evaluated per action or only at session boundaries;
- how approvals are represented and audited.

### 3. Policy enforcement

Map:

- policy sources;
- policy ownership;
- policy evaluation point;
- policy/version binding;
- deny and escalation behavior;
- unsupported or malformed input handling;
- exception paths.

### 4. Runtime boundary

Identify:

- the component that actually executes actions;
- command and tool boundaries;
- repository and environment scope;
- credential boundaries;
- final authorization checks;
- isolation and resource controls;
- rollback or recovery controls.

### 5. Verification

Document which gates exist before and after execution:

- static analysis;
- tests;
- security checks;
- dependency checks;
- benchmark or quality gates;
- human review;
- deployment health checks;
- post-execution verification.

### 6. Evidence and audit

Assess whether the organization can reconstruct:

- request and intent;
- agent identity;
- policy version;
- authorization decision;
- execution metadata;
- verification result;
- resulting artifact or commit;
- external deployment receipt;
- relevant timestamps and correlation identifiers.

The assessment explicitly distinguishes agent-session memory from durable audit evidence.

## Output

The assessment produces a concise engineering control map containing:

| Area | Current state | Evidence | Gap / question | Candidate control |
|---|---|---|---|---|
| Agent inventory |  |  |  |  |
| Identity |  |  |  |  |
| Authorization |  |  |  |  |
| Policy enforcement |  |  |  |  |
| Runtime boundary |  |  |  |  |
| Verification |  |  |  |  |
| Evidence / audit |  |  |  |  |
| Handoff |  |  |  |  |

The output should separate observed facts from proposed controls.

## Assessment workflow

```text
Discovery
   |
   v
Agent + Tool Inventory
   |
   v
Control Boundary Mapping
   |
   v
Policy / Authorization Review
   |
   v
Runtime + Verification Review
   |
   v
Evidence / Audit Review
   |
   v
Gap Map
   |
   v
Pilot Candidate
```

## Questions to answer

1. Where can an AI coding agent take consequential action today?
2. Which actions require explicit authorization?
3. Where is the final controllable authorization boundary?
4. Can policy be evaluated independently of the model?
5. Can execution be distinguished from proposal?
6. Can verification be distinguished from execution?
7. What evidence survives after the agent session ends?
8. Can an auditor reconstruct why an action was allowed?
9. Can a changed or unsupported policy/schema version fail closed?
10. Which single workflow would provide the clearest pilot signal?

## Pilot handoff

A pilot candidate should have:

- a defined engineering workflow;
- a measurable control boundary;
- a limited set of agents/tools;
- explicit success criteria;
- observable evidence artifacts;
- a named technical owner;
- a defined evaluation period.

Recommended first use case: specification-to-service generation with policy and evidence inspection, using the NAEOS Golden Path and Reference Demo.

## Boundaries

This assessment does not establish:

- regulatory compliance;
- complete security assurance;
- production readiness;
- scalability under arbitrary workloads;
- correctness of every model-generated change;
- safety of every consequential external action;
- customer adoption or commercial suitability.

## Related documents

- docs/NAEOS-ENTERPRISE-OVERVIEW.md
- docs/PILOT-READINESS.md
- docs/GOLDEN-PATH.md
- docs/REFERENCE-DEMO.md
- docs/DESIGN-PARTNER-PROGRAM.md
