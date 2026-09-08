NAEOS Master Technical Specification

Core Domain, Policy, Control Plane, Runtime & Evidence

Document ID: NAEOS-MTS-001
Version: 1.0.0
Status: Proposed
Project: NAEOS — Nusantara AI Engineering Operating System
Maintainer: NAEOS Foundation
Last Updated: September 2026

---

1. Purpose

This document defines the technical specification for the core NAEOS platform.

It translates the architectural principles defined by the NAEOS Reference Architecture into implementable technical contracts.

This specification defines:

- core domain objects
- policy model
- authorization model
- Control Plane
- Decision Records
- Runtime
- execution requests
- evidence
- artifact integrity
- event model
- plugin interfaces
- agent integration boundaries
- security requirements

This document is intended to be used by NAEOS engineers and contributors as the primary implementation reference.

---

2. Technical Thesis

NAEOS is based on one fundamental distinction:

«Agent capability is not authority.»

An AI agent may propose an action.

NAEOS determines whether that action is authorized.

The system therefore follows:

Agent
  │
  │ Intent
  ▼
NAEOS Control Plane
  │
  │ Policy Evaluation
  ▼
Authorization Decision
  │
  ▼
Runtime
  │
  │ Controlled Execution
  ▼
Evidence
  │
  ▼
Verification

---

3. System Boundaries

NAEOS is responsible for:

- policy management
- authorization
- execution control
- governance
- decision recording
- evidence capture
- artifact integrity
- agent integration
- plugin lifecycle

NAEOS is not primarily responsible for:

- generating AI models
- replacing AI Coding Agents
- acting as an IDE
- replacing source-control systems
- replacing cloud infrastructure
- replacing observability platforms

---

4. Core Components

The minimum NAEOS platform consists of:

┌─────────────────────────────────────────────┐
│                  NAEOS                      │
│                                             │
│  ┌────────────┐       ┌───────────────┐    │
│  │   Policy   │──────▶│ Control Plane │    │
│  │  Registry  │       └───────┬───────┘    │
│  └────────────┘               │             │
│                              ▼             │
│                       ┌──────────────┐      │
│                       │  Decision    │      │
│                       │   Engine     │      │
│                       └──────┬───────┘      │
│                              │             │
│                              ▼             │
│                       ┌──────────────┐      │
│                       │   Runtime    │      │
│                       └──────┬───────┘      │
│                              │             │
│                              ▼             │
│                       ┌──────────────┐      │
│                       │   Evidence   │      │
│                       │    Store     │      │
│                       └──────────────┘      │
└─────────────────────────────────────────────┘

---

5. Domain Model

The initial domain model contains:

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

---

6. Agent

An Agent represents an external AI system interacting with NAEOS.

Example:

agent:
  id: agent-claude
  provider: anthropic
  type: coding-agent
  version: "x.y.z"

The Agent object identifies the actor but does not automatically grant authority.

---

7. Actor

An Actor represents the entity responsible for initiating an action.

Possible actor types:

human
agent
service
system
plugin

Example:

actor:
  id: agent-claude
  type: agent

---

8. Action

An Action represents an operation that may have engineering consequences.

Examples:

read_repository
write_file
create_commit
merge_pull_request
modify_configuration
execute_command
deploy_staging
deploy_production
modify_infrastructure
rotate_secret
delete_resource

Actions should have stable identifiers.

Example:

action:
  type: deploy_production

---

9. Intent

Intent represents what the agent is attempting to accomplish.

Example:

intent:
  id: intent-001

  actor:
    id: agent-claude
    type: agent

  action:
    type: deploy_production

  target:
    environment: production
    service: api

  reason: release approved application version

Intent is not authorization.

Intent
  ≠
Authorization

---

10. Policy

A Policy defines authoritative rules governing actions.

Example:

policy:
  id: production-deployment
  version: 1.0.0
  status: active

  rules:
    - action: deploy_production
      decision: require_approval

---

11. Policy Version

Policies must be immutable by version.

Example:

production-deployment
    │
    ├── v1.0.0
    ├── v1.1.0
    └── v2.0.0

A historical decision must reference the exact policy version used.

---

12. Policy Rule

A Policy Rule defines the conditions under which an action is allowed.

Example:

rule:
  action: deploy_production

  conditions:
    environment: production

  decision:
    type: require_approval

Supported decisions:

ALLOW
DENY
REQUIRE_APPROVAL

---

13. Authorization Request

An Authorization Request asks the Control Plane whether a proposed action may proceed.

Conceptual structure:

authorization_request:

  request_id: req-001

  actor:
    id: agent-claude
    type: agent

  action:
    type: deploy_production

  target:
    environment: production

  artifact:
    id: build-123
    hash: sha256:...

  context:
    repository: naeos
    branch: main

---

14. Authorization Decision

The Control Plane produces a deterministic decision.

decision:
  id: decision-001

  result: REQUIRE_APPROVAL

  policy:
    id: production-deployment
    version: 1.0.0

  request:
    id: req-001

Valid results:

ALLOW
DENY
REQUIRE_APPROVAL

---

15. Decision Determinism

For the same:

Policy
+
Request
+
Relevant Context

the authorization engine must produce the same decision.

Formally:

D(P, R, C) = Decision

where:

- "P" = policy
- "R" = request
- "C" = relevant context

The authorization decision must not depend on probabilistic LLM reasoning.

---

16. Approval

Approval represents an authorized human or system decision allowing a previously restricted action.

Example:

approval:
  id: approval-001

  decision_id: decision-001

  approver:
    id: user-123
    type: human

  result: approved

  artifact_hash: sha256:...

  timestamp: ...

---

17. Approval Binding

For artifact-sensitive actions, approval should be bound to:

Decision
+
Artifact Hash
+
Policy Version

Example:

Approval
 ├── decision_id
 ├── policy_version
 └── artifact_hash

Execution is valid only when these values remain consistent.

---

18. Execution Request

Once authorization is satisfied, an Execution Request is created.

execution_request:
  id: exec-001

  authorization:
    decision_id: decision-001

  action:
    type: deploy_production

  artifact:
    hash: sha256:...

  constraints:
    environment: production

---

19. Runtime

The Runtime is the execution enforcement layer.

Its responsibility is to ensure that execution cannot occur outside the authorization boundary.

Execution Request
       │
       ▼
Runtime
       │
       ├── validate authorization
       ├── validate artifact
       ├── enforce constraints
       ├── execute
       └── capture result

---

20. Runtime Security Boundary

A protected action must not be executable simply because an agent has technical access to a tool.

Incorrect:

Agent ───────────────▶ Tool

Correct:

Agent
  │
  ▼
Control Plane
  │
  ▼
Runtime
  │
  ▼
Tool

The Runtime is therefore an enforcement boundary.

---

21. Execution

Execution represents the actual operation performed.

Example:

execution:
  id: execution-001

  request_id: exec-001

  status: succeeded

  started_at: ...
  completed_at: ...

  result:
    exit_code: 0

Possible states:

PENDING
RUNNING
SUCCEEDED
FAILED
BLOCKED
CANCELLED

---

22. Artifact

An Artifact represents a concrete version of something being acted upon.

Examples:

- source tree
- commit
- build
- container
- deployment package
- infrastructure plan

Artifact identity should be cryptographically addressable where practical.

Example:

artifact:
  id: build-123
  digest:
    algorithm: sha256
    value: ...

---

23. Artifact Integrity

The system must detect artifact changes between authorization and execution.

Approved Artifact
       │
       ▼
SHA-256
       │
       ▼
Approval

Execution Artifact
       │
       ▼
SHA-256
       │
       ▼
Compare

If:

approved_hash != execution_hash

the execution must be rejected for artifact-bound approvals.

---

24. Decision Record

Decision Records preserve the governance decision independently from agent memory.

A Decision Record contains:

Request
Policy
Policy Version
Decision
Actor
Context
Approval
Artifact
Timestamp

Example:

decision_record:
  id: dr-001

  request_id: req-001

  decision: approved

  policy:
    id: production-deployment
    version: 1.0.0

  actor:
    id: agent-claude

  artifact_hash: sha256:...

  approval:
    id: approval-001

---

25. Audit Evidence

Audit Evidence represents durable proof of what happened.

Evidence should include, where applicable:

who
what
when
where
why
policy
decision
approval
artifact
execution
result
verification

The fundamental requirement is:

«Evidence must not depend on agent memory.»

---

26. Evidence Integrity

Audit evidence should support integrity verification.

Conceptually:

Evidence
   │
   ▼
Canonical Representation
   │
   ▼
Cryptographic Digest
   │
   ▼
Integrity Verification

Future implementations may additionally support:

- append-only storage
- signed records
- tamper-evident chains
- external timestamping
- WORM storage

---

27. Event Model

NAEOS should use an event-oriented model for important state transitions.

Example events:

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

Events should have stable identifiers and timestamps.

---

28. Event Envelope

Example:

event:
  id: event-001

  type: authorization.decided

  timestamp: ...

  actor:
    id: naeos-control-plane
    type: system

  subject:
    id: decision-001

  data:
    result: require_approval

---

29. Control Plane API

The initial API should expose operations conceptually equivalent to:

POST /authorization/evaluate
GET  /policies/{id}
GET  /policies/{id}/versions
POST /approvals
GET  /decisions/{id}
GET  /evidence/{id}
POST /verification

The exact transport mechanism may be REST, gRPC, event-driven, or another implementation choice.

The semantic contract is more important than the transport.

---

30. Authorization API

Example request:

{
  "actor": {
    "id": "agent-claude",
    "type": "agent"
  },
  "action": {
    "type": "deploy_production"
  },
  "target": {
    "environment": "production"
  },
  "artifact": {
    "hash": "sha256:..."
  }
}

Example response:

{
  "decision": "REQUIRE_APPROVAL",
  "policy": {
    "id": "production-deployment",
    "version": "1.0.0"
  }
}

---

31. Authorization State Machine

              ┌─────────────┐
              │   REQUEST   │
              └──────┬──────┘
                     │
                     ▼
             ┌───────────────┐
             │   EVALUATING  │
             └──────┬────────┘
                    │
          ┌─────────┼─────────┐
          ▼         ▼         ▼
       ALLOW      DENY     APPROVAL
          │         │         │
          │         │         ▼
          │         │      APPROVED
          │         │         │
          └─────────┴─────────┘
                    │
                    ▼
                EXECUTABLE

---

32. Approval State Machine

PENDING
   │
   ├── APPROVED ──▶ VALID
   │
   └── REJECTED ──▶ INVALID

Artifact-bound approval introduces an additional condition:

VALID
  │
  ▼
Artifact Hash Match?
  │
 ┌┴────┐
YES    NO
 │      │
 ▼      ▼
EXECUTE BLOCK

---

33. Policy Evaluation Pipeline

The Policy Engine should perform:

1. Identify applicable policies
2. Resolve policy version
3. Normalize request
4. Evaluate rules
5. Resolve conflicts
6. Produce decision
7. Record decision metadata

---

34. Policy Conflict Resolution

NAEOS must define deterministic conflict resolution.

Recommended initial rule:

DENY
  >
REQUIRE_APPROVAL
  >
ALLOW

This provides a conservative default.

However, the exact precedence model must eventually be formalized as an ADR before production use.

---

35. Multi-Policy Evaluation

A request may be governed by multiple policies.

Example:

Engineering Policy
Security Policy
Production Policy
Organization Policy

The Control Plane evaluates the applicable policy set.

Conceptually:

Request
  │
  ├── Engineering Policy
  ├── Security Policy
  ├── Environment Policy
  └── Organization Policy
          │
          ▼
    Policy Resolution
          │
          ▼
       Decision

---

36. Agent Adapter

Agents should interact with NAEOS through an adapter interface.

Agent
  │
  ▼
Agent Adapter
  │
  ▼
NAEOS Interface

The adapter normalizes provider-specific behavior into a common NAEOS request model.

---

37. Agent Adapter Contract

Conceptual operations:

submit_intent()
request_authorization()
submit_execution()
receive_decision()
receive_execution_result()

The adapter must not become an authorization layer.

Authorization remains the responsibility of the Control Plane.

---

38. Plugin Model

Plugins extend NAEOS without modifying core components.

Potential plugin categories:

agent
runtime
policy
security
observability
repository
cloud
verification
identity

Example:

Plugin
 ├── manifest
 ├── capabilities
 ├── permissions
 ├── version
 └── interface

---

39. Plugin Security

Plugins should be treated as potentially privileged components.

A future Plugin Registry should support:

- plugin identity
- versioning
- capability declaration
- permission declaration
- integrity verification
- lifecycle management
- compatibility validation

Plugins must not automatically inherit unrestricted system authority.

---

40. Security Model

NAEOS security should follow least privilege.

Core principles:

Least Privilege
Explicit Authorization
Isolation
Deterministic Enforcement
Immutable Evidence
Artifact Integrity
Credential Protection
Supply Chain Security

---

41. Threat Model

Initial threat categories:

T1 — Malicious Agent

Agent intentionally attempts unauthorized execution.

T2 — Compromised Agent

Agent credentials or environment are compromised.

T3 — Policy Tampering

An attacker modifies policy.

T4 — Approval Replay

An old approval is reused.

T5 — Artifact Substitution

Approved artifact is replaced with another artifact.

T6 — Evidence Tampering

Audit records are modified.

T7 — Plugin Compromise

A malicious plugin attempts to escalate privileges.

T8 — Runtime Bypass

An agent attempts to bypass NAEOS Runtime enforcement.

---

42. Security Requirements

At minimum:

SR-001  Protected actions require authorization.
SR-002  Authorization is policy-driven.
SR-003  Policy versions are identifiable.
SR-004  Approval can be bound to artifacts.
SR-005  Runtime validates authorization.
SR-006  Audit evidence is durable.
SR-007  Agent memory is not authoritative evidence.
SR-008  Plugins declare capabilities.
SR-009  Privileged operations require explicit permissions.
SR-010  Security-sensitive events are auditable.

---

43. Observability

NAEOS should expose operational telemetry for:

- authorization latency
- policy evaluation
- execution duration
- blocked actions
- approval latency
- runtime failures
- evidence creation
- verification failures

Observability must not compromise sensitive evidence.

---

44. Error Model

The system should distinguish:

POLICY_ERROR
AUTHORIZATION_ERROR
APPROVAL_ERROR
RUNTIME_ERROR
ARTIFACT_ERROR
INTEGRITY_ERROR
PLUGIN_ERROR
VERIFICATION_ERROR

A policy evaluation failure should not silently degrade into ALLOW.

Recommended behavior:

Evaluation Failure
       │
       ▼
FAIL CLOSED

for protected actions.

---

45. Fail-Safe Principle

For high-risk actions:

«When authorization cannot be established, execution must not proceed.»

Example:

Policy unavailable
      │
      ▼
Authorization unknown
      │
      ▼
BLOCK

This should be configurable by policy for lower-risk operations.

---

46. Auditability Requirements

For consequential operations, NAEOS should preserve:

Request ID
Actor
Action
Target
Policy ID
Policy Version
Decision
Approval
Artifact Hash
Execution ID
Execution Result
Timestamp
Verification Status

---

47. Idempotency

Authorization and execution APIs should support idempotency where appropriate.

Example:

request_id = req-001

Repeated submission should not accidentally create multiple independent consequential actions.

---

48. Reproducibility

NAEOS should preserve enough metadata to reconstruct why a decision occurred.

A decision should reference:

Policy Version
Request
Context
Decision Engine Version
Artifact
Approval

This supports future audit and incident investigation.

---

49. Decision Engine Versioning

The authorization engine itself should be versioned.

Example:

Policy Version: 1.2.0
Decision Engine: 0.4.0

This matters because changes to evaluation semantics can affect authorization outcomes.

---

50. Reference End-to-End Protocol

┌────────────┐
│    Agent   │
└─────┬──────┘
      │
      │ Intent
      ▼
┌────────────┐
│   NAEOS    │
│  Gateway   │
└─────┬──────┘
      │
      │ Authorization Request
      ▼
┌────────────┐
│   Policy   │
│  Registry  │
└─────┬──────┘
      │
      ▼
┌────────────┐
│  Control   │
│   Plane    │
└─────┬──────┘
      │
      ▼
 Decision
      │
 ┌────┼───────────┐
 ▼    ▼           ▼
ALLOW DENY     APPROVAL
 │                │
 │                ▼
 │             Approval
 │                │
 └───────┬────────┘
         ▼
    ┌─────────┐
    │ Runtime │
    └────┬────┘
         │
         ▼
     Execution
         │
         ▼
    Audit Evidence
         │
         ▼
     Verification

---

51. Minimum Viable Technical Implementation

The first executable version should implement only:

Policy Registry
        +
Policy Evaluator
        +
Authorization Decision
        +
Runtime Gateway
        +
Decision Record
        +
Audit Evidence

This creates the minimum governance loop.

---

52. MVP Example

Request

deploy_production

Policy

action: deploy_production
decision: require_approval

Control Plane

REQUIRE_APPROVAL

Approval

APPROVED

Runtime

EXECUTE

Evidence

decision
policy
approval
artifact hash
execution result
timestamp

This should become the first canonical NAEOS demonstration.

---

53. Recommended Implementation Order

1. Domain Models
        ↓
2. Policy Schema
        ↓
3. Policy Registry
        ↓
4. Policy Evaluator
        ↓
5. Authorization API
        ↓
6. Decision Record
        ↓
7. Runtime Gateway
        ↓
8. Evidence Store
        ↓
9. Artifact Integrity
        ↓
10. Agent Adapter
        ↓
11. Plugin Registry
        ↓
12. Independent Verification

---

54. Technical Acceptance Criteria

NAEOS Core v1.0 should satisfy the following.

Policy

- policies can be registered
- policies have immutable versions
- policies can be evaluated deterministically

Authorization

- actions can be classified
- authorization decisions are explicit
- DENY cannot silently become ALLOW
- approval requirements are enforceable

Runtime

- protected operations pass through Runtime
- unauthorized operations are blocked
- authorization is revalidated before execution

Evidence

- consequential actions produce evidence
- evidence references policy version
- evidence references decision
- artifact hash can be recorded

Security

- least privilege is enforced
- plugin permissions are explicit
- authorization failures fail closed for pro
