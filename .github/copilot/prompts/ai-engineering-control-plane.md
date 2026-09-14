# NAEOS — Implement the AI Engineering Control Plane

You are working directly inside the NAEOS repository:

https://github.com/NAEOS-foundation/naeos

Your task is to evolve the existing NAEOS implementation into a focused, demonstrable **AI Engineering Control Plane**.

Do NOT redesign NAEOS from scratch.

Do NOT create a parallel architecture.

Do NOT duplicate existing functionality.

First inspect the repository and reuse the existing architecture, interfaces, CLI commands, NEIR model, pipeline, governance, policy, context, compiler, MCP, artifact, audit, and testing systems wherever possible.

---

## 1. Product Direction

The strategic positioning to implement is:

> **NAEOS is the Engineering Control Plane for the AI-Native Software Era.**

Core message:

> **AI coding agents generate code. NAEOS provides the engineering control plane around them.**

NAEOS should sit above AI coding agents and coordinate:

```text
Engineering Specification
        ↓
Parse
        ↓
Normalize
        ↓
Resolve
        ↓
NEIR
        ↓
Validate
        ↓
Policy / Governance
        ↓
AI Context
        ↓
MCP / AI Agent
        ↓
Execution
        ↓
Artifacts / Evidence / Audit
```

The system must remain vendor-neutral.

Do not hard-code NAEOS around one AI provider.

---

# 2. Primary Objective

Build one end-to-end workflow that demonstrates:

```text
Specification
    ↓
NEIR
    ↓
Validation
    ↓
Policy
    ↓
AI Context
    ↓
AI Agent Instructions
    ↓
Execution
    ↓
Artifacts
    ↓
Audit / Evidence
```

The complete workflow should be executable locally and demonstrable in less than approximately three minutes for a small example project.

The objective is NOT to add dozens of new features.

The objective is to make the existing NAEOS architecture behave like a coherent engineering control plane.

---

# 3. Repository-First Rule

Before modifying anything:

1. Inspect README.md.
2. Inspect ARCHITECTURE-OVERVIEW.md.
3. Inspect DEVELOPMENT_PLAN.md.
4. Inspect ROADMAP.md.
5. Inspect WHITEPAPER-EN.md.
6. Inspect DOCUMENTATION-INDEX.md.
7. Inspect AGENTS.md.
8. Inspect existing NEIR implementation.
9. Inspect specification parser/normalizer/resolver.
10. Inspect validator.
11. Inspect policy/governance.
12. Inspect context generation.
13. Inspect AI compiler.
14. Inspect MCP.
15. Inspect audit/evidence/artifact systems.
16. Inspect existing examples and tests.
17. Inspect existing CLI command registration.

Map existing functionality before writing code.

Create a short internal implementation plan based on what already exists.

Do not assume that a capability is missing simply because its name differs.

---

# 4. Preserve Existing Architecture

Reuse existing NAEOS components.

Expected existing areas include:

```text
cmd/naeos
internal/specification
internal/neir
internal/compiler
internal/context
internal/generation
internal/governance
internal/policy
internal/artifacts
internal/mcp
internal/audit
internal/evidence
internal/security
pkg/pipeline
pkg/kernel
```

Also reuse existing:

* interfaces
* services
* pipeline stages
* CLI conventions
* configuration
* error handling
* logging
* telemetry
* artifact formats
* policy mechanisms
* tests
* documentation conventions

If an existing component already solves a requirement, integrate with it rather than implementing another version.

---

# 5. Define the Control Plane Model

Treat NEIR as the persistent engineering model.

The control plane should conceptually manage:

```text
Intent
Architecture
Dependencies
Policies
Security
AI Context
Execution
Artifacts
Evidence
Audit
```

NEIR must remain the central representation.

Do not turn prompts into the system's source of truth.

The engineering specification remains the source of truth.

AI prompts are context/instructions derived from the engineering model.

---

# 6. Implement an End-to-End "AI Engineering Run"

Create or improve an existing workflow/CLI capability that can demonstrate:

```text
naeos run
    ↓
parse specification
    ↓
normalize
    ↓
resolve
    ↓
build NEIR
    ↓
validate
    ↓
evaluate policies
    ↓
compile AI context
    ↓
produce AI instructions
    ↓
execute/generate
    ↓
store artifacts
    ↓
produce audit/evidence
```

Prefer extending `naeos run` or existing commands instead of creating an unnecessary new command.

If a new command is genuinely required, follow the repository's existing CLI architecture and naming conventions.

---

# 7. Introduce an Explicit Engineering Run Context

If the current architecture does not already provide an equivalent concept, introduce a lightweight run context.

Conceptually:

```text
EngineeringRun
├── run_id
├── specification
├── specification_hash
├── NEIR
├── validation_result
├── policy_result
├── context_bundle
├── execution_plan
├── artifacts
├── evidence
├── audit_events
└── status
```

The run context should allow the system to answer:

* What engineering intent was executed?
* Which specification version was used?
* Which NEIR was produced?
* Which validations passed?
* Which policies were evaluated?
* What AI context was generated?
* What execution occurred?
* Which artifacts were produced?
* What evidence exists?
* What failed?

Reuse existing event sourcing, audit, artifact, and evidence systems where possible.

---

# 8. Deterministic Policy Boundary

One of the most important architectural principles:

> Prompts are not security boundaries.

AI agents may propose actions.

Deterministic NAEOS components must decide whether consequential actions are allowed.

The architecture should therefore conceptually separate:

```text
AI Agent
   │
   │ proposes action
   ▼
NAEOS Control Plane
   │
   ├── Validate
   ├── Policy evaluation
   ├── Authorization
   ├── Audit
   └── Execute / Reject
```

Never trust the model to enforce policy by instruction alone.

Reuse existing policy and governance implementations.

Add missing enforcement only where necessary.

---

# 9. AI Context Layer

Strengthen the existing context generation so that AI agents receive structured engineering context rather than only raw project files.

The generated context should include, where available:

```text
Project
Architecture
Modules
Dependencies
APIs
Services
Infrastructure
Security constraints
Policies
Testing requirements
Documentation
Relevant artifacts
Current execution state
```

Context should be derived from NEIR.

Avoid dumping the entire repository into the model.

Prefer deterministic, structured, relevant context.

---

# 10. AI Adapter Strategy

Preserve the existing multi-provider architecture.

NAEOS currently supports multiple AI instruction targets.

Ensure the control-plane workflow can produce appropriate instructions/context for existing targets such as:

* GitHub Copilot
* Claude Code
* Cursor
* Gemini CLI
* Codex
* OpenCode
* Windsurf

Do not create provider-specific business logic inside the core engineering model.

Provider adapters should remain at the integration boundary.

---

# 11. MCP Integration

Inspect the current MCP implementation.

The MCP layer should expose useful engineering-system capabilities rather than simply exposing raw filesystem access.

Where appropriate, expose operations conceptually equivalent to:

```text
get_project_model
get_neir
validate_specification
evaluate_policy
get_engineering_context
get_execution_plan
get_artifacts
get_audit
```

Only implement missing operations.

All consequential operations must pass through NAEOS validation/policy boundaries.

---

# 12. Demonstration Project

Create or improve a canonical example specifically for demonstrating the control plane.

Prefer an existing example directory if appropriate.

The example should be small enough to understand but complex enough to demonstrate:

* modules
* dependencies
* service/API
* architecture
* policy
* AI context
* generation
* validation
* artifacts
* audit/evidence

Example conceptual project:

```text
project: ai-task-service

modules:
  - auth
  - api
  - worker
  - storage

services:
  - api
  - worker

dependencies:
  api → auth
  api → storage
  worker → storage
```

Do not blindly use this structure if the repository already has a better canonical example.

---

# 13. Demonstrate Failure, Not Only Success

The demo must show that NAEOS is a control plane.

Create at least one intentionally invalid scenario.

For example:

```text
Module A
    ↓
Module B

Policy:
Module B cannot depend on Module A
```

or:

```text
Service A → port 8080
Service B → port 8080
```

The system should reject the invalid state deterministically.

The demo should clearly show:

```text
Specification
     ↓
Validation
     ↓
REJECTED
     ↓
Reason
     ↓
Policy / Dependency violation
```

Then demonstrate the corrected specification.

---

# 14. Observability

Make the end-to-end workflow observable.

The user should be able to see the major stages:

```text
[1] Specification
[2] Parse
[3] Normalize
[4] Resolve
[5] NEIR
[6] Validation
[7] Policy
[8] AI Context
[9] Execution
[10] Artifacts
[11] Evidence
```

Use existing logging, telemetry, profiling, or pipeline output mechanisms.

Do not introduce a completely separate logging framework.

---

# 15. Output Format

Where existing CLI conventions permit, provide machine-readable output.

Prefer:

```text
--output json
--output yaml
```

or the repository's existing equivalent.

A successful engineering run should expose enough information for automation.

Conceptually:

```json
{
  "run_id": "...",
  "status": "success",
  "specification": "...",
  "neir": "...",
  "validation": "passed",
  "policy": "passed",
  "context": "...",
  "artifacts": [],
  "evidence": [],
  "audit": []
}
```

Do not invent a conflicting schema if an existing result schema already exists.

Extend existing schemas where possible.

---

# 16. Testing Requirements

Every implementation must include tests.

At minimum:

### Unit tests

Test:

* engineering run orchestration
* NEIR integration
* validation
* policy enforcement
* context generation
* artifact registration
* audit/evidence creation

### Integration test

Create one complete:

```text
spec → NEIR → validate → policy → context → execution → artifact → audit
```

test.

### Failure test

Verify that an invalid policy/dependency state is rejected before execution.

### Regression tests

Run the existing test suite and ensure existing functionality remains intact.

---

# 17. Performance

Use existing pipeline caching and profiling capabilities.

Do not unnecessarily recompute:

```text
parse
normalize
resolve
NEIR
validation
context
```

when inputs have not changed.

Where practical, the canonical demo should feel fast enough for an interactive developer workflow.

Do not sacrifice correctness for speed.

---

# 18. Documentation

Update documentation only after implementation works.

At minimum update the most appropriate existing documents.

Add:

## AI Engineering Control Plane

Explain:

```text
Specification
      ↓
NEIR
      ↓
Validation
      ↓
Policy
      ↓
AI Context
      ↓
Agent
      ↓
Execution
      ↓
Evidence
```

Explain why NAEOS is different from:

* AI coding assistants
* project generators
* prompt libraries
* generic orchestration frameworks

Use technically precise language.

Avoid unsupported marketing claims.

---

# 19. Demo Documentation

Create a canonical demo guide.

It should allow a developer to execute something approximately like:

```bash
go build -o naeos ./cmd/naeos

naeos init

naeos run --input-file examples/.../spec.yaml

naeos context --input-file examples/.../spec.yaml

naeos ai compile --input-file examples/.../spec.yaml --target opencode
```

Use the actual repository commands after inspecting the current implementation.

Do not invent commands that do not exist.

The demo should explain:

1. specification
2. NEIR
3. validation
4. policy
5. AI context
6. AI adapter
7. execution
8. artifacts
9. evidence/audit

---

# 20. Product Boundary

Do NOT attempt to implement all future NAEOS ambitions in this task.

Do NOT build:

* a complete cloud platform
* billing
* enterprise SaaS
* a massive marketplace redesign
* a new AI model
* a new coding agent
* unnecessary UI
* unnecessary database architecture
* duplicate governance systems

The immediate objective is to make the existing NAEOS technical foundation demonstrate a coherent and compelling control-plane workflow.

---

# 21. Commercial Architecture Awareness

Keep future commercialization in mind without prematurely implementing it.

Potential future commercial capabilities:

```text
Open Source Core
       ↓
NAEOS Control Plane
       ↓
Team Collaboration
       ↓
Centralized Governance
       ↓
Audit / Compliance
       ↓
Managed Infrastructure
       ↓
Enterprise Integrations
```

Do not implement paid functionality unless the existing repository architecture already requires it.

Design interfaces so these capabilities can be added later.

---

# 22. Engineering Principles

Follow these principles strictly:

### Principle 1 — Specification is source of truth

Never make an AI prompt the authoritative representation of engineering intent.

### Principle 2 — NEIR is the engineering model

Use NEIR as the central representation across pipeline stages.

### Principle 3 — Deterministic enforcement

Policies and authorization must be enforced outside the model.

### Principle 4 — Vendor neutrality

AI providers are replaceable execution adapters.

### Principle 5 — Auditability

Important engineering decisions and consequential actions should be traceable.

### Principle 6 — Reproducibility

The same specification and configuration should produce deterministic or explainably equivalent results.

### Principle 7 — Composability

Prefer existing NAEOS interfaces and pipeline stages.

### Principle 8 — Backward compatibility

Do not break existing CLI behavior or APIs without a documented migration path.

### Principle 9 — Minimal new architecture

Extend before replacing.

### Principle 10 — Evidence over claims

Every major capability should have a working example and automated test.

---

# 23. Implementation Process

Execute the work in phases.

## Phase 1 — Repository Analysis

Inspect the repository and produce:

```text
Current architecture
Existing reusable components
Missing control-plane capabilities
Recommended implementation points
Potential conflicts
```

Do not modify code yet.

## Phase 2 — Architecture Plan

Create a concise implementation plan.

Map:

```text
Requirement → Existing Component → Required Change → Test
```

## Phase 3 — Core Implementation

Implement the smallest coherent control-plane workflow.

## Phase 4 — Integration

Connect:

```text
NEIR
Validation
Policy
Context
AI compiler
MCP
Artifacts
Audit
Evidence
```

## Phase 5 — Tests

Add unit, integration, failure, and regression tests.

## Phase 6 — Demo

Build the canonical example.

## Phase 7 — Documentation

Document the actual working behavior.

## Phase 8 — Verification

Run:

```bash
go test ./...
go vet ./...
go build ./cmd/naeos/
```

Also run the canonical demo end-to-end.

If the repository has lint, fuzz, benchmark, or CI-specific commands, inspect and run the appropriate existing commands.

---

# 24. Definition of Done

This task is complete only when:

* [ ] Existing architecture has been inspected.
* [ ] Existing components are reused.
* [ ] Specification remains source of truth.
* [ ] NEIR remains central engineering model.
* [ ] Validation is part of the control-plane flow.
* [ ] Policy evaluation is part of the flow.
* [ ] AI context is derived from engineering state.
* [ ] AI execution remains vendor-neutral.
* [ ] Consequential actions have deterministic enforcement.
* [ ] Artifacts are traceable.
* [ ] Audit/evidence is produced where supported.
* [ ] Successful workflow works end-to-end.
* [ ] Invalid workflow is rejected deterministically.
* [ ] Automated tests exist.
* [ ] Existing tests continue to pass.
* [ ] Canonical demo works.
* [ ] Documentation explains the workflow.
* [ ] No duplicate architecture was introduced.
* [ ] No unnecessary large feature expansion was introduced.

---

# 25. Final Deliverable

When finished, report:

## Architecture changes

What was changed and why.

## Files changed

List every important file.

## New capabilities

Describe what NAEOS can now do.

## Demo

Provide exact commands to reproduce the end-to-end workflow.

## Tests

Report:

```text
unit tests
integration tests
failure tests
regression tests
build
lint
```

## Before / After

Explain:

```text
Before:
existing NAEOS capability

After:
AI Engineering Control Plane workflow
```

## Remaining gaps

Clearly identify anything that should NOT be implemented yet.

---

# Final strategic constraint

Do not optimize for number of features.

Optimize for one undeniable product experience:

> **A developer specifies engineering intent once. NAEOS turns that intent into a validated engineering model, enforces policies, prepares AI context, coordinates AI execution, and leaves behind traceable artifacts and evidence.**

That workflow is the product wedge.

Everything else is secondary.
