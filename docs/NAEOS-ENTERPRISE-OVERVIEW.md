# NAEOS Enterprise Overview

Summary: A concise enterprise-facing overview of NAEOS as an engineering control plane for governing AI coding agents.
Last updated: 2026-09-26

## Overview

AI coding agents can move from generating suggestions to reading repositories, changing files, invoking tools, running commands, and participating in delivery workflows.

NAEOS provides an engineering control plane around those workflows. It helps teams make engineering intent explicit, apply policy before execution, verify outcomes, and retain durable evidence of what happened.

**Positioning:** NAEOS is the engineering control plane that governs how AI coding agents turn engineering intent into executable, verifiable software.

> Model proposes. Policy decides. Runtime executes. Evidence proves.

NAEOS is designed to work around heterogeneous coding agents and existing engineering systems rather than replace them.

## Enterprise problem

Teams adopting agentic development need to answer operational questions that ordinary prompts and documentation do not reliably answer:

- Which agent is acting?
- What was the agent asked or authorized to do?
- Which policy applied?
- Was the action allowed, denied, or escalated?
- What actually executed?
- What verification ran?
- What evidence remains after the agent session ends?
- Can an incident or review reconstruct the decision and execution path?

These questions become more important as agent permissions, tool access, repositories, CI/CD systems, credentials, and external services become connected.

## Control model

NAEOS separates four concerns:

1. **Model proposal** — the agent proposes an action or change.
2. **Policy decision** — NAEOS evaluates the request against declared policy and context.
3. **Runtime execution** — an authorized runtime performs the action within the approved boundary.
4. **Observation and evidence** — execution and verification produce durable evidence for later inspection.

This separation is intended to reduce ambiguity between what an agent suggested, what policy authorized, what actually executed, and what was verified.

## Reference architecture

```text
AI Coding Agents
  Copilot | Claude Code | Codex | Cursor | Other Agents
                       |
                       v
              NAEOS Control Plane
        +-----------------------------+
        | Policy | Authorization      |
        | Governance | Verification    |
        | Handoff | Risk Controls     |
        +-----------------------------+
                       |
                       v
                  Runtime
                       |
                       v
             Evidence / Audit Ledger
                       |
                       v
       Git | CI/CD | Cloud | Enterprise Systems
```

## Enterprise use cases

### AI coding governance

Define and enforce controls around agent-driven repository and tool actions.

### Platform engineering

Provide a common governance boundary across multiple coding agents and development workflows.

### Security and risk control

Make authorization, policy decisions, verification requirements, and execution evidence inspectable.

### Regulated engineering

Support evidence-oriented workflows where decisions and execution need to remain reconstructable.

### Agent handoffs

Use protocol-neutral handoff contracts so work can move between agents or runtimes without losing the governing context.

## What NAEOS does not replace

NAEOS is complementary to:

- AI coding agents
- source-control platforms
- CI/CD systems
- IAM and identity providers
- secrets management
- SIEM and observability systems
- cloud platforms
- developer portals and internal platforms

The intended role is the governance and engineering control layer between intent and executable engineering action.

## Evaluation path

Organizations can evaluate NAEOS without committing to a broad platform rollout:

1. Identify one agent-driven engineering workflow.
2. Map the existing control boundary.
3. Run the NAEOS Golden Path / Reference Demo.
4. Inspect policy, execution, verification, and evidence artifacts.
5. Compare the observed workflow with the team's operational requirements.
6. Define a narrow design-partner or pilot scope if additional integration is justified.

## Success criteria

A technical evaluation should establish observable evidence for:

- policy decision recorded;
- authorization outcome recorded;
- execution trace available;
- verification result available;
- durable evidence retained;
- deterministic or reproducible behavior where required;
- fail-closed behavior for unsupported or invalid control inputs.

These are evaluation criteria, not claims that every NAEOS workflow already satisfies every enterprise requirement.

## Related documents

- docs/PILOT-READINESS.md
- docs/GOLDEN-PATH.md
- docs/REFERENCE-DEMO.md
- docs/DESIGN-PARTNER-PROGRAM.md
- docs/policy-decision-record.md
- docs/change-risk-governance.md
