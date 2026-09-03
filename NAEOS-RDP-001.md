NAEOS Roadmap & Development Plan

From Architecture to Executable AI Engineering Foundation

Document ID: NAEOS-RDP-001
Version: 1.0.0
Status: Proposed
Project: NAEOS — Nusantara AI Engineering Operating System
Maintainer: NAEOS Foundation
Last Updated: September 2026

---

1. Purpose

This document defines the development roadmap for NAEOS from its current architectural foundation toward an executable, production-oriented AI Engineering Operating System.

The roadmap translates the principles established in:

- "NAEOS-PER-001 — Project Evolution Record"
- "NAEOS-NRA-001 — NAEOS Reference Architecture"

into concrete engineering phases, milestones, deliverables, GitHub work items, and acceptance criteria.

The objective is to move NAEOS from:

«Architecture → Specification → Implementation → Validation → Adoption»

---

2. Strategic Objective

NAEOS aims to become a vendor-neutral foundation for governing and executing AI-assisted software engineering.

The core system must eventually provide:

Agent
  │
  ▼
Intent
  │
  ▼
Policy Evaluation
  │
  ▼
Authorization
  │
  ▼
Execution Control
  │
  ▼
Evidence
  │
  ▼
Verification

The development roadmap therefore prioritizes control and governance infrastructure over additional AI capabilities.

---

3. Development Philosophy

NAEOS development follows seven principles.

3.1 Policy Before Automation

Do not automate a behavior before defining the policy governing it.

3.2 Deterministic Enforcement

Consequential authorization should be enforced by deterministic infrastructure rather than model behavior.

3.3 Evidence by Default

Consequential actions should generate durable evidence automatically.

3.4 Modular Architecture

Core components must remain independently replaceable.

3.5 Vendor Neutrality

No single AI provider should become a mandatory dependency.

3.6 Security by Design

Security requirements must be architectural constraints, not post-development additions.

3.7 Incremental Execution

Each phase must produce independently testable value.

---

4. Current Baseline

NAEOS currently has a strong conceptual foundation consisting of:

- project identity
- reference architecture
- constitution model
- governance principles
- policy model
- Control Plane concept
- Policy Registry concept
- Decision Record concept
- Audit Evidence concept
- artifact-bound approval concept
- Runtime concept
- Plugin Registry direction
- independent verification concept
- vendor-neutral positioning

The principal development gap is:

«Turning these concepts into executable, testable infrastructure.»

---

5. Development Phases

The roadmap consists of six major phases.

PHASE 1
Foundation
     │
     ▼
PHASE 2
Policy & Governance
     │
     ▼
PHASE 3
Control Plane
     │
     ▼
PHASE 4
Runtime & Agent Integration
     │
     ▼
PHASE 5
Evidence & Verification
     │
     ▼
PHASE 6
Ecosystem & Production

---

6. Phase 1 — Foundation

Objective

Establish the executable core and repository structure.

Key Deliverables

- repository architecture
- package/module structure
- configuration system
- core data models
- CLI/API foundation
- event model
- logging foundation
- testing framework
- development conventions

Proposed Structure

naeos/
├── docs/
├── constitution/
├── policies/
├── kernel/
├── control-plane/
├── runtime/
├── compiler/
├── plugins/
├── integrations/
├── audit/
├── tests/
└── examples/

GitHub Issues

NAEOS-001  Define repository architecture
NAEOS-002  Define core domain models
NAEOS-003  Define event model
NAEOS-004  Define configuration model
NAEOS-005  Establish test infrastructure
NAEOS-006  Establish development conventions
NAEOS-007  Create local development environment

Acceptance Criteria

Phase 1 is complete when:

- core modules compile/build successfully
- automated tests execute
- domain objects have defined schemas
- event model is documented
- development environment is reproducible
- CI can validate the repository

---

7. Phase 2 — Policy & Governance

Objective

Implement the authoritative policy layer.

This phase converts the Policy Registry concept into an actual system.

Core Components

Policy
Policy Version
Policy Scope
Policy Rule
Policy Decision
Authorization Requirement

Example Policy

id: production-deployment

version: 1.0.0

scope:
  environment: production

rules:
  deployment:
    decision: require_approval

GitHub Issues

NAEOS-010  Define policy schema
NAEOS-011  Implement Policy Registry
NAEOS-012  Implement policy versioning
NAEOS-013  Implement policy validation
NAEOS-014  Implement policy lifecycle
NAEOS-015  Implement policy scope
NAEOS-016  Define authorization semantics
NAEOS-017  Add policy test suite

Acceptance Criteria

The system must be able to:

1. register a policy
2. version a policy
3. validate a policy
4. retrieve the active policy
5. determine policy scope
6. reject invalid policies
7. provide deterministic policy evaluation input

---

8. Phase 3 — Control Plane

Objective

Build the central authorization and enforcement layer.

This is the architectural heart of NAEOS.

             CONTROL PLANE
                   │
       ┌───────────┼───────────┐
       ▼           ▼           ▼
    Policy      Risk        Approval
 Evaluation   Evaluation     State
       │           │           │
       └───────────┼───────────┘
                   ▼
             Decision
                   │
         ┌─────────┼─────────┐
         ▼         ▼         ▼
       ALLOW      DENY    APPROVAL

Core Functions

- evaluate policy
- classify action
- determine authorization
- enforce approval requirements
- issue decisions
- record decisions

GitHub Issues

NAEOS-020  Define Control Plane API
NAEOS-021  Implement policy evaluator
NAEOS-022  Implement authorization engine
NAEOS-023  Implement action classification
NAEOS-024  Implement approval state
NAEOS-025  Implement decision engine
NAEOS-026  Implement enforcement middleware
NAEOS-027  Add deterministic decision tests

Acceptance Criteria

For identical:

Policy
+
Action
+
Context

the Control Plane must produce the same decision.

Example:

Input
  ↓
Policy Evaluation
  ↓
ALLOW / DENY / REQUIRE_APPROVAL

No LLM inference should be required to determine the final authorization result.

---

9. Phase 4 — Runtime & Agent Integration

Objective

Connect NAEOS controls to real AI Coding Agent execution.

The Runtime becomes the bridge between authorization and actual execution.

Agent
  │
  ▼
Intent
  │
  ▼
Control Plane
  │
  ▼
Authorization
  │
  ▼
Runtime
  │
  ▼
Tool / Command

Initial Integrations

Priority should be given to protocol/interface compatibility rather than provider-specific features.

Potential integrations:

- Claude
- Codex
- GitHub Copilot
- Gemini
- OpenCode
- other agent systems

Runtime Responsibilities

- receive authorized actions
- enforce constraints
- execute tools
- isolate execution
- capture execution metadata
- prevent unauthorized execution

GitHub Issues

NAEOS-030  Define Runtime API
NAEOS-031  Define agent adapter interface
NAEOS-032  Implement execution gateway
NAEOS-033  Implement tool authorization
NAEOS-034  Implement command restrictions
NAEOS-035  Implement execution sandbox
NAEOS-036  Implement agent adapter
NAEOS-037  Add runtime integration tests

Acceptance Criteria

An agent must not be able to bypass the Control Plane and directly execute protected operations.

---

10. Phase 5 — Evidence & Verification

Objective

Create durable evidence for consequential engineering actions.

This phase implements the principle:

«The audit trail must outlive the agent's memory.»

Evidence Model

Action
 │
 ▼
Decision Record
 │
 ▼
Execution
 │
 ▼
Evidence
 │
 ├── timestamp
 ├── actor
 ├── policy version
 ├── decision
 ├── artifact hash
 ├── execution result
 └── verification state

GitHub Issues

NAEOS-040  Define Decision Record schema
NAEOS-041  Define Audit Evidence schema
NAEOS-042  Implement immutable evidence store
NAEOS-043  Implement execution evidence capture
NAEOS-044  Implement artifact hashing
NAEOS-045  Implement approval-artifact binding
NAEOS-046  Implement audit queries
NAEOS-047  Add evidence integrity tests

Acceptance Criteria

The system must be able to answer:

Who?
What?
When?
Why?
Under which policy?
Which decision?
Which artifact?
What actually executed?
What was the result?

---

11. Phase 6 — Independent Verification & Ecosystem

Objective

Move NAEOS from an engineering prototype toward production ecosystem infrastructure.

Components

- independent verification
- external verification
- compliance workflows
- plugin ecosystem
- integrations
- enterprise deployment model
- community contribution model

Verification Architecture

NAEOS
  │
  │ Policy Contract
  ▼
Runtime
  │
  │ Execution
  ▼
System
  │
  │ Evidence
  ▼
Independent Verifier
  │
  │ Independent Verification
  ▼
Verification Result

GitHub Issues

NAEOS-050  Define verification contract
NAEOS-051  Define independent verification
NAEOS-052  Implement verification events
NAEOS-053  Implement evidence verification
NAEOS-054  Implement plugin lifecycle
NAEOS-055  Implement plugin registry
NAEOS-056  Create integration SDK
NAEOS-057  Create reference deployment

---

12. Cross-Phase Security Track

Security is not a separate final phase.

It runs across every phase.

Identity
Authorization
Least Privilege
Isolation
Secrets
Audit
Integrity
Policy Enforcement
Supply Chain

Security Issues

SEC-001  Threat model
SEC-002  Agent privilege model
SEC-003  Runtime isolation
SEC-004  Secret handling
SEC-005  Policy integrity
SEC-006  Artifact integrity
SEC-007  Audit integrity
SEC-008  Plugin security
SEC-009  Dependency security
SEC-010  Security test suite

---

13. Cross-Phase Testing Strategy

Testing should occur at multiple levels.

Unit Tests

Test individual components.

Contract Tests

Test interfaces between components.

Policy Tests

Test policy behavior.

Authorization Tests

Test ALLOW / DENY / APPROVAL behavior.

Runtime Tests

Test execution restrictions.

Integration Tests

Test complete workflows.

Security Tests

Test bypass attempts.

Adversarial Tests

Test malicious or unexpected agent behavior.

---

14. Reference End-to-End Workflow

A minimum NAEOS workflow should eventually support:

1. Agent creates intent
          │
          ▼
2. NAEOS receives action request
          │
          ▼
3. Policy Registry provides policy
          │
          ▼
4. Control Plane evaluates request
          │
          ▼
5. Decision generated
          │
     ┌────┼────┐
     ▼    ▼    ▼
   ALLOW DENY APPROVAL
     │         │
     │         ▼
     │      Approval
     │         │
     └────┬────┘
          ▼
6. Runtime executes
          │
          ▼
7. Evidence captured
          │
          ▼
8. Artifact verified
          │
          ▼
9. External verification

---

15. GitHub Milestone Structure

Recommended GitHub milestones:

M0 — Architecture

Architecture baseline
Specifications
Schemas
ADRs

M1 — Foundation

Core
Models
Events
Testing
CI

M2 — Policy

Policy Registry
Policy Schema
Policy Versioning
Policy Evaluation

M3 — Control

Authorization
Decision Engine
Approval
Enforcement

M4 — Runtime

Runtime
Agent Adapters
Tool Execution
Sandbox

M5 — Evidence

Decision Record
Audit Evidence
Artifact Hash
Integrity

M6 — Verification

Independent Verification
Verification
Compliance

M7 — Ecosystem

Plugin Registry
SDK
Integrations
Community

---

16. Priority Classification

Each issue should receive a priority.

P0 — Critical

Required for core system operation.

P1 — High

Required for production readiness.

P2 — Medium

Important but not blocking.

P3 — Future

Experimental or ecosystem enhancement.

---

17. Definition of Done

A feature is not considered complete merely because code exists.

A NAEOS feature is Done when:

Specification
     +
Implementation
     +
Tests
     +
Documentation
     +
Security Review
     +
Integration Validation

have been completed.

---

18. Architecture Decision Records

Important architectural decisions should be recorded as ADRs.

Recommended initial ADRs:

ADR-001  NAEOS Core Architecture
ADR-002  Policy Registry Authority Model
ADR-003  Agent Capability vs Authorization
ADR-004  Control Plane Architecture
ADR-005  Decision Record Model
ADR-006  Audit Evidence Model
ADR-007  Artifact-Bound Approval
ADR-008  Runtime Enforcement Boundary
ADR-009  Vendor-Neutral Agent Integration
ADR-010  Independent Verification
ADR-011  Plugin Registry Architecture

---

19. Repository Documentation Map

Recommended documentation structure:

docs/
│
├── README.md
│
├── architecture/
│   ├── reference-architecture.md
│   ├── control-plane.md
│   ├── runtime.md
│   └── plugin-registry.md
│
├── governance/
│   ├── governance-model.md
│   ├── policy-model.md
│   └── authorization.md
│
├── constitution/
│   ├── engineering.md
│   ├── ai.md
│   ├── architecture.md
│   ├── security.md
│   ├── testing.md
│   └── documentation.md
│
├── decisions/
│   └── adr/
│
├── security/
│   ├── threat-model.md
│   └── security-model.md
│
├── runtime/
│   └── execution-model.md
│
├── audit/
│   ├── decision-record.md
│   └── evidence-model.md
│
├── integrations/
│
└── roadmap/
    └── development-plan.md

---

20. First Reference Implementation

The first implementation should deliberately remain small.

Recommended scope:

Agent
  │
  ▼
Action Request
  │
  ▼
Policy Registry
  │
  ▼
Control Plane
  │
  ▼
ALLOW / DENY / APPROVAL
  │
  ▼
Runtime
  │
  ▼
Execution
  │
  ▼
Audit Evidence

Avoid initially attempting:

- full autonomous agents
- large-scale distributed infrastructure
- complex enterprise UI
- every AI provider
- every cloud provider
- broad plugin ecosystem

The first objective is to prove the governance loop.

---

21. Minimum Viable Governance Loop

The first meaningful NAEOS prototype should demonstrate:

REQUEST
   ↓
POLICY
   ↓
DECISION
   ↓
AUTHORIZATION
   ↓
EXECUTION
   ↓
EVIDENCE

Example:

Agent:
"Deploy application to production."

        ↓

Policy Registry:
Production deployment requires approval.

        ↓

Control Plane:
REQUIRE_APPROVAL

        ↓

Human / authorized process:
APPROVED

        ↓

Runtime:
Execute deployment.

        ↓

Audit:
Record policy, approval, artifact hash,
execution result, and timestamp.

If this workflow works reliably, NAEOS has demonstrated its core architectural thesis.

---

22. Key Engineering Metrics

NAEOS should measure more than lines of code.

Governance

- percentage of consequential actions evaluated by policy
- authorization decision accuracy
- policy coverage

Security

- unauthorized execution attempts blocked
- privilege violations prevented
- policy bypass rate

Reliability

- deterministic decision consistency
- execution failure rate
- policy evaluation latency

Auditability

- percentage of consequential actions with evidence
- evidence integrity failures
- artifact verification coverage

Ecosystem

- integrations
- plugins
- contributors
- active repositories
- production deployments

---

23. Risk Register

Risk| Impact| Mitigation
Over-engineering| High| Start with minimum governance loop
Vendor dependency| High| Adapter architecture
Policy complexity| High| Simple deterministic schema
Agent bypass| Critical| Runtime enforcement
Audit manipulation| Critical| Immutable evidence
Plugin compromise| High| Plugin security model
Architecture drift| Medium| ADR process
Lack of adoption| High| Reference implementation
Excessive scope| High| Phase-based roadmap
Weak testing| Critical| Automated policy/security tests

---

24. Strategic Non-Goals

NAEOS should explicitly avoid becoming:

Another AI Coding Assistant

NAEOS does not compete primarily on code generation.

Another IDE

NAEOS should remain IDE-independent.

Another LLM

NAEOS is not a model provider.

Prompt Library

Prompts may be supported but are not the architectural core.

Generic Workflow Automation

Automation without governance is not the primary objective.

Vendor Lock-In Platform

NAEOS must preserve interoperability.

---

25. Long-Term Vision

The long-term NAEOS ecosystem should look like:

                       NAEOS FOUNDATION
                              │
             ┌────────────────┼────────────────┐
             │                │                │
        GOVERNANCE         CONTROL          RUNTIME
             │                │                │
        Constitutions      Policy          Execution
        Standards          Auth            Sandbox
        Compliance         Approval        Tools
             │                │                │
             └────────────────┼────────────────┘
                              │
                              ▼
                          EVIDENCE
                              │
                              ▼
                         VERIFICATION
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
           AGENTS          PLUGINS        INTEGRATIONS
              │               │               │
              └───────────────┼───────────────┘
                              ▼
                       AI ENGINEERING
                         ECOSYSTEM

---

26. Success Criteria

NAEOS can be considered production-oriented when an organization can deploy an AI Coding Agent and reliably answer:

1. What action did the agent propose?
2. Which policy governed the action?
3. Which policy version was active?
4. Was the action authorized?
5. Was approval required?
6. Who or what approved it?
7. What artifact was approved?
8. What artifact was actually executed?
9. What exactly happened during execution?
10. What evidence proves it?
11. Can the result be independently verified?

The system should answer these questions without depending on the agent's memory.

---

27. North Star

The North Star for NAEOS is:

«Every consequential AI engineering action should be governable, authorized, controlled, and provable.»

This can be represented as:

           GOVERNABLE
                │
                ▼
           AUTHORIZED
                │
                ▼
            CONTROLLED
                │
                ▼
             EXECUTED
                │
                ▼
             PROVABLE

---

28. Final Development Direction

The immediate priority is not to build more AI capability.

The immediate priority is to make the existing AI capability safe and governable.

Therefore the development sequence is:

1. Define
   ↓
2. Specify
   ↓
3. Implement
   ↓
4. Enforce
   ↓
5. Test
   ↓
6. Capture Evidence
   ↓
7. Verify
   ↓
8. Integrate
   ↓
9. Deploy
   ↓
10. Scale

The central engineering objective remains:

«Separate what an agent can do from what an agent is authorized to do.»

And the architectural expression of that principle is:

«Agents can reason. Policies authorize. Runtime executes. Evidence proves.»

---

29. Document Control

Field| Value
Document ID| NAEOS-RDP-001
Version| 1.0.0
Status| Proposed
Project| NAEOS
Maintainer| NAEOS Foundation
Scope| Development Roadmap
Review Cycle| Per major milestone
Last Updated| September 2026

End of Document
