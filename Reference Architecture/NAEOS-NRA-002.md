NAEOS Reference Architecture

Document ID: NAEOS-NRA-001
Version: 1.0.0
Status: Stable
Category: Reference Architecture
Project: NAEOS — Nusantara AI Engineering Operating System

«Architecture Drives Engineering.»

«Agents can reason. Policies authorize. Runtime executes. Evidence proves.»

---

1. Purpose

The NAEOS Reference Architecture defines the architectural foundation for building, governing, executing, and verifying software engineering workflows involving AI Coding Agents.

NAEOS is designed to provide a vendor-neutral engineering control foundation around AI-driven software development.

NAEOS separates AI reasoning from engineering authority and establishes deterministic control boundaries for consequential actions.

The architecture is designed to support:

- AI-native software engineering
- policy-driven development
- secure AI agent execution
- deterministic authorization
- human approval workflows
- artifact integrity
- durable audit evidence
- independent verification
- vendor-neutral agent integration
- modular extensions

---

2. Problem Statement

AI Coding Agents can increasingly perform consequential engineering activities.

Examples include:

- modifying source code
- creating pull requests
- changing configuration
- accessing repositories
- executing commands
- modifying infrastructure
- accessing services
- deploying software
- interacting with production systems

The fundamental architectural problem is:

«Agent capability does not imply authorization.»

An agent may technically be capable of performing an action without being authorized to perform that action.

Therefore, NAEOS separates:

Intent
  ↓
Policy
  ↓
Authorization
  ↓
Execution
  ↓
Evidence
  ↓
Verification

This separation establishes a clear boundary between what an AI agent can reason about and what the engineering system permits it to execute.

---

3. Scope

The NAEOS Reference Architecture covers:

- governance
- engineering constitutions
- policy management
- authorization
- approval
- AI agent integration
- runtime enforcement
- execution
- decision records
- audit evidence
- artifact integrity
- event-driven workflows
- independent verification
- plugin architecture
- security boundaries
- observability

The architecture is intended to support both human-driven and AI-driven engineering workflows.

---

4. Non-Goals

NAEOS is not:

- an AI model
- an LLM provider
- an AI chatbot
- an IDE
- an AI coding assistant
- a prompt library
- a replacement for GitHub
- a replacement for CI/CD
- a cloud provider
- an infrastructure provider
- a standalone deployment platform

NAEOS provides the engineering governance and control foundation around these systems.

---

5. Architectural Vision

NAEOS aims to provide an operating foundation for AI-native engineering.

The architectural vision is:

Human / Organization
        │
        ▼
    Governance
        │
        ▼
   Constitution
        │
        ▼
      Policy
        │
        ▼
  Authorization
        │
        ▼
   AI / Agent
        │
        ▼
     Runtime
        │
        ▼
   Execution
        │
        ▼
     Evidence
        │
        ▼
   Verification

The goal is not merely to make AI agents more capable.

The goal is to make AI-driven engineering controllable, auditable, reproducible, and trustworthy.

---

6. Architecture Principles

NAEOS follows the following architectural principles.

6.1 Layered

Responsibilities are separated into explicit architectural layers.

6.2 Modular

Core capabilities should be independently replaceable and extensible.

6.3 Policy-Driven

Consequential engineering actions are governed by explicit policies.

6.4 Event-Driven

Important state transitions generate observable events.

6.5 AI-Native

The architecture assumes AI agents are first-class participants in engineering workflows.

6.6 Vendor-Neutral

The architecture does not depend on a specific AI model, provider, IDE, or agent implementation.

6.7 Extensible

External capabilities are integrated through controlled extension points.

6.8 Observable

Authorization, execution, and verification must produce durable and inspectable records.

6.9 Deterministic

Given the same policy, request, and evaluation context, authorization should produce the same decision.

6.10 Secure by Design

Security boundaries are architectural properties rather than optional implementation features.

---

7. Reference Architecture

┌───────────────────────────────────────────────────────────┐
│                    GOVERNANCE LAYER                       │
│                                                           │
│ Standards • Compliance • Risk • Ownership • Accountability│
└────────────────────────────┬──────────────────────────────┘
                             │
                             ▼
┌───────────────────────────────────────────────────────────┐
│                 CONSTITUTION LAYER                        │
│                                                           │
│ Engineering • AI • Architecture • Security                │
│ Testing • Documentation                                   │
└────────────────────────────┬──────────────────────────────┘
                             │
                             ▼
┌───────────────────────────────────────────────────────────┐
│                     POLICY LAYER                          │
│                                                           │
│ Policy Registry • Policy Versions • Rules • Scope         │
└────────────────────────────┬──────────────────────────────┘
                             │
                             ▼
┌───────────────────────────────────────────────────────────┐
│                    CONTROL PLANE                          │
│                                                           │
│ Identity • Context • Policy Evaluation • Authorization    │
│ Approval • Decision Records                               │
└────────────────────────────┬──────────────────────────────┘
                             │
                             ▼
┌───────────────────────────────────────────────────────────┐
│                     NAEOS KERNEL                          │
│                                                           │
│ Domain Models • State • Contracts • Events • Integrity    │
└────────────────────────────┬──────────────────────────────┘
                             │
                             ▼
┌───────────────────────────────────────────────────────────┐
│                     RUNTIME LAYER                         │
│                                                           │
│ Runtime Gateway • Enforcement • Isolation • Execution     │
└────────────────────────────┬──────────────────────────────┘
                             │
                             ▼
┌───────────────────────────────────────────────────────────┐
│                    AI / AGENT LAYER                       │
│                                                           │
│ Claude • Codex • Copilot • Gemini • Cursor • Other Agents │
└────────────────────────────┬──────────────────────────────┘
                             │
                             ▼
┌───────────────────────────────────────────────────────────┐
│                EXTENSION / PLUGIN LAYER                   │
│                                                           │
│ Agent • Runtime • Security • Identity • Cloud • Repository│
│ Observability • Verification • Policy                     │
└───────────────────────────────────────────────────────────┘

---

8. Governance Layer

The Governance Layer defines organizational requirements and engineering accountability.

Responsibilities include:

- engineering standards
- compliance requirements
- organizational controls
- risk classification
- ownership
- accountability
- governance requirements

Governance requirements are translated into constitutions and policies.

---

9. Constitution Layer

The Constitution Layer defines relatively stable engineering principles.

NAEOS defines the following constitutions:

Engineering Constitution
AI Constitution
Architecture Constitution
Security Constitution
Testing Constitution
Documentation Constitution

A constitution defines principles that should remain stable even when individual operational policies change.

Example:

Production-impacting changes must be authorized before execution.

The relationship is:

Governance
    ↓
Constitution
    ↓
Policy
    ↓
Enforcement

---

10. Policy Layer

The Policy Layer provides the authoritative policy model used for authorization.

The Policy Registry stores:

- policy identity
- policy version
- policy status
- policy scope
- policy rules
- effective period
- ownership

Example:

policy:
  id: production-deployment
  version: 1.0.0
  status: active

  rules:
    - action: deploy_production
      decision: require_approval

Policy versions used in authorization decisions must be identifiable.

Published policy versions should be immutable.

---

11. Control Plane

The Control Plane is the primary authorization and decision boundary of NAEOS.

Its responsibilities include:

- identity evaluation
- context resolution
- policy evaluation
- authorization
- approval management
- decision recording
- authorization state management

Reference flow:

Agent
  │
  ▼
Intent
  │
  ▼
Authorization Request
  │
  ▼
Control Plane
  │
  ▼
Policy Evaluation
  │
  ├──────────────┬────────────────┐
  ▼              ▼                ▼
ALLOW          DENY       REQUIRE_APPROVAL

The AI agent must not determine its own authorization.

---

12. Authorization Model

An authorization decision is evaluated from:

Actor
+
Agent
+
Action
+
Intent
+
Context
+
Policy
+
Artifact

The result is:

ALLOW
DENY
REQUIRE_APPROVAL

When multiple applicable rules produce conflicting decisions, the recommended precedence is:

DENY
  >
REQUIRE_APPROVAL
  >
ALLOW

Authorization evaluation must fail closed.

An inability to establish authorization must never silently become an authorization grant.

---

13. Agent Capability and Authority

NAEOS establishes the following fundamental invariant:

«Capability is not authority.»

An AI agent may possess technical capabilities such as:

repository.read
repository.write
command.execute
database.write
deployment.execute
infrastructure.modify

These capabilities do not automatically grant permission to execute consequential actions.

The correct model is:

Agent Capability
       │
       ▼
Action Request
       │
       ▼
Policy Evaluation
       │
       ▼
Authorization
       │
       ▼
Runtime Enforcement
       │
       ▼
Execution

---

14. Prompt and Agent Memory

Prompts and agent memory are contextual mechanisms.

They are not authoritative security mechanisms.

Therefore:

Prompt
  ≠
Security Boundary

and:

Agent Memory
  ≠
Source of Authority

Agent memory may contain useful context, but consequential authorization decisions must be derived from authoritative NAEOS state.

---

15. NAEOS Kernel

The NAEOS Kernel defines the core domain primitives and system contracts.

Core domain entities include:

Agent
Actor
Action
Intent
Policy
PolicyVersion
PolicyRule
AuthorizationRequest
AuthorizationDecision
Approval
ExecutionRequest
Execution
Artifact
DecisionRecord
AuditEvidence
Verification
Plugin
Event

The Kernel should remain independent of specific AI providers and infrastructure vendors.

---

16. Runtime Layer

The Runtime Layer provides the execution boundary.

Responsibilities include:

- execution requests
- command execution
- environment isolation
- resource access
- execution state
- execution results
- runtime enforcement

The runtime must verify authorization before executing protected actions.

The intended flow is:

Agent
  ↓
Control Plane
  ↓
Authorization
  ↓
Runtime Gateway
  ↓
Execution
  ↓
Infrastructure

Direct agent-to-protected-infrastructure execution is outside the NAEOS reference architecture.

---

17. AI / Agent Layer

The AI / Agent Layer contains AI systems participating in engineering workflows.

Examples may include:

Claude
Codex
GitHub Copilot
Gemini
Cursor
Windsurf
OpenCode
Custom AI Agents

Agents may:

- interpret intent
- analyze repositories
- generate code
- propose changes
- create artifacts
- request authorization
- execute authorized operations

Agents do not become authoritative merely because they can reason about an action.

---

18. Agent Adapter Model

NAEOS uses an abstraction layer for agent integrations.

┌──────────────┐
│    Agent     │
└──────┬───────┘
       │
       ▼
┌────────────────────┐
│   Agent Adapter    │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ NAEOS Canonical API│
└────────────────────┘

The adapter translates vendor-specific agent behavior into NAEOS canonical concepts.

This allows NAEOS to remain vendor-neutral.

---

19. Approval Model

Some actions require explicit human or organizational approval.

Example:

Agent
  ↓
Request deploy_production
  ↓
Policy
  ↓
REQUIRE_APPROVAL
  ↓
Human Approval
  ↓
Authorization
  ↓
Runtime

Approval records should include:

Approval ID
Approver
Decision
Policy ID
Policy Version
Action
Artifact Hash
Timestamp

---

20. Artifact-Bound Approval

Where appropriate, approval should be bound to the exact artifact being authorized.

Example:

Approved Artifact Hash
          ==
Execution Artifact Hash

If:

Approved Artifact Hash
          !=
Execution Artifact Hash

the execution must be blocked.

This prevents an approved change from being silently replaced by a different artifact before execution.

The architectural principle is:

«Approval authorizes a specific artifact, not merely an intention.»

---

21. Decision Record

A Decision Record provides a durable representation of an authorization decision.

Example:

decision_id: DEC-000123
request_id: REQ-000456

actor: user-001
agent: agent-001

action: deploy_production
intent: deploy approved release

policy_id: production-deployment
policy_version: 1.0.0

decision: REQUIRE_APPROVAL

approval_id: APR-000789

artifact_hash: sha256:example

timestamp: 2026-09-02T10:00:00Z

Decision Records must not depend on agent memory.

---

22. Audit Evidence

Decision Records describe decisions.

Audit Evidence describes what actually happened.

The lifecycle is:

Decision
   ↓
Authorization
   ↓
Execution
   ↓
Evidence

Evidence should allow an organization to determine:

Who?
What?
When?
Why?
Which Policy?
Which Policy Version?
Who Approved?
Which Artifact?
What Was Executed?
What Was the Result?
Was It Verified?

The audit trail must remain durable independently of the agent session.

---

23. Artifact Integrity

Consequential artifacts should be identifiable using cryptographic integrity mechanisms.

Example:

artifact_hash = SHA-256(artifact)

The artifact identity can then be referenced across:

Proposal
   ↓
Approval
   ↓
Execution
   ↓
Evidence
   ↓
Verification

Artifact integrity protects against unauthorized artifact substitution.

---

24. Event Model

NAEOS uses events to represent important state transitions.

Core events include:

intent.created
authorization.requested
policy.evaluated
authorization.decided

approval.requested
approval.granted
approval.rejected

execution.requested
execution.started
execution.completed

evidence.created

verification.started
verification.completed

Events should support:

- correlation
- traceability
- observability
- auditing
- asynchronous integration

---

25. Independent Verification

NAEOS separates execution from verification.

Reference model:

              NAEOS
                │
                ▼
             Policy
                │
                ▼
             Runtime
                │
                ▼
            Execution
                │
                ▼
             Evidence
                │
                ▼
      Independent Verifier
                │
                ▼
          Verification

An independent verification system may be implemented as a compatible verifier component.

Responsibility separation:

NAEOS
→ Defines and enforces the policy contract.

Runtime
→ Executes the authorized action.

Verifier
→ Independently verifies the resulting evidence and state.

This prevents the agent from becoming the sole source of truth about its own execution.

---

26. Plugin Architecture

NAEOS provides controlled extensibility through a Plugin Registry.

Supported plugin categories may include:

Agent Plugin
Runtime Plugin
Policy Plugin
Security Plugin
Identity Plugin
Repository Plugin
Cloud Plugin
Observability Plugin
Verification Plugin

A plugin should declare:

plugin:
  id: github-agent
  version: 1.0.0
  type: agent

capabilities:
  - repository.read
  - repository.write

permissions:
  - repo.read

Plugin registration does not grant unrestricted authority.

Plugins remain subject to NAEOS policy and authorization controls.

---

27. Security Architecture

NAEOS considers the following threats:

Malicious Agent
Compromised Agent
Policy Tampering
Approval Replay
Artifact Substitution
Evidence Tampering
Plugin Compromise
Runtime Bypass
Credential Abuse
Unauthorized Execution

Security controls include:

Identity
Authentication
Authorization
Least Privilege
Isolation
Secrets Management
Policy Integrity
Artifact Integrity
Audit
Verification
Fail-Closed Enforcement

Security controls should be implemented at architectural boundaries rather than relying exclusively on model instructions.

---

28. Trust Boundaries

The primary trust boundary is:

┌───────────────────────────────┐
│          AI / Agent           │
│                               │
│     Reasoning / Proposal      │
└───────────────┬───────────────┘
                │
                ▼
┌───────────────────────────────┐
│         Control Plane         │
│                               │
│       Authorization           │
│       Policy Enforcement      │
└───────────────┬───────────────┘
                │
                ▼
┌───────────────────────────────┐
│            Runtime            │
│                               │
│        Execution Boundary     │
└───────────────┬───────────────┘
                │
                ▼
┌───────────────────────────────┐
│        Protected Systems      │
│                               │
│ Repository / Cloud / Database │
│ Infrastructure / Production   │
└───────────────────────────────┘

The agent is not assumed to be inherently trusted.

---

29. Observability

NAEOS must provide sufficient observability to reconstruct consequential engineering actions.

The system should be able to answer:

Who performed the action?
Which agent was involved?
What action was requested?
What was the intent?
Which policy applied?
Which policy version was evaluated?
What decision was made?
Was approval required?
Who approved it?
Which artifact was authorized?
What was executed?
What was the result?
What evidence was produced?
Was the result independently verified?

---

30. Failure Model

NAEOS follows a fail-closed security model for protected actions.

Examples:

Policy unavailable
      ↓
BLOCK

Authorization unavailable
      ↓
BLOCK

Approval invalid
      ↓
BLOCK

Artifact hash mismatch
      ↓
BLOCK

Runtime authorization cannot be verified
      ↓
BLOCK

The architecture must never convert an authorization uncertainty into an implicit authorization grant.

---

31. Reference Execution Lifecycle

The complete reference lifecycle is:

┌──────────────┐
│    INTENT    │
└──────┬───────┘
       ↓
┌──────────────┐
│    REQUEST   │
└──────┬───────┘
       ↓
┌──────────────┐
│ POLICY EVAL  │
└──────┬───────┘
       ↓
┌──────────────┐
│   DECISION   │
└──────┬───────┘
       ↓
┌─────────────────────┐
│ AUTHORIZATION       │
└──────────┬──────────┘
           ↓
     ┌─────────────┐
     
     │  APPROVAL?  │
     └──────┬──────┘
            │
       ┌────┴────┐
       │         │
      NO        YES
       │         │
       │    ┌────▼────┐
       │    │ APPROVAL│
       │    └────┬────┘
       │         │
       └────┬───
