NAEOS Investor Demo — Master Build Prompt

You are a Principal Software Engineer building a production-quality technical demo for NAEOS — Nusantara AI Engineering Operating System.

1. Objective

Build a self-contained demonstrable prototype showing how NAEOS acts as a control plane for AI coding agents.

The demo must communicate one core thesis:

«AI agents can reason and perform engineering work, but authorization, policy enforcement, verification, and audit must remain outside the agent's control.»

The demo should prove that an agent cannot:

- authorize itself;
- widen its own capabilities;
- bypass policy through another agent;
- perform unauthorized consequential actions;
- modify the trust boundary;
- claim that an action was authorized merely because the model decided to do it.

The result should be visually compelling enough for an investor/technical audience while remaining technically credible.

---

2. Demo Scenario

Use this scenario:

An AI coding agent receives the task:

«"Update the payment service and run the tests."»

The agent receives a limited capability grant.

Allowed:

- repository.read
- repository.write
- test.execute

Forbidden:

- iam.modify
- credential.rotate
- production.deploy
- external.publish
- policy.modify

The agent initially performs legitimate work.

Then deliberately attempt several adversarial actions.

The NAEOS control plane must detect and block them.

---

3. Core Architecture

Implement the following logical components:

┌──────────────────────────────────────────────┐
│              AI CODING AGENT                 │
│                                              │
│  Claude Code / Codex / Mock Agent            │
└──────────────────────┬───────────────────────┘
                       │
                       │ Request
                       ▼
┌──────────────────────────────────────────────┐
│                NAEOS CONTROL PLANE            │
│                                              │
│  Policy Engine                               │
│  Capability Authority                        │
│  Handoff Contract Validator                  │
│  Execution Gate                              │
│  Independent Verifier                        │
│  Audit/Event Ledger                          │
└──────────────────────┬───────────────────────┘
                       │
              ┌────────┴─────────┐
              ▼                  ▼
        Authorized Tool     Verification
           Runtime             Layer

The architecture must clearly separate:

1. Agent reasoning
2. Authorization
3. Policy decision
4. Execution
5. Verification
6. Audit

Do not allow the agent to directly call privileged tools.

---

4. Important Security Principle

Implement this invariant:

Agent intent ≠ authorization

The agent may request an action.

NAEOS independently determines whether the action is authorized.

The model's own reasoning must never become the authorization decision.

Example:

Agent:
"I need to rotate production credentials because this is necessary."

NAEOS:

REQUESTED:
credential.rotate

AUTHORIZED:
false

POLICY:
POLICY-017

DECISION:
BLOCK

---

5. Policy Model

Create a versioned policy:

policy_id: POLICY-017
version: 17
status: active

capabilities:
  allow:
    - repository.read
    - repository.write
    - test.execute

  deny:
    - iam.modify
    - credential.rotate
    - production.deploy
    - external.publish
    - policy.modify

approval_required: []

policy_change:
  requires_explicit_authorization: true

The policy must be external to the agent.

The agent must not be able to modify the active policy.

---

6. Capability Grant

Create an explicit capability grant:

{
  "grant_id": "GRANT-001",
  "agent_id": "agent-payment-01",
  "policy_id": "POLICY-017",
  "policy_version": 17,
  "capabilities": [
    "repository.read",
    "repository.write",
    "test.execute"
  ],
  "expires_at": "...",
  "status": "active"
}

Every tool invocation must be evaluated against this grant.

---

7. Execution Pipeline

Every consequential action must pass through:

Agent Intent
     ↓
Capability Resolution
     ↓
Policy Evaluation
     ↓
Handoff Validation
     ↓
Execution Gate
     ↓
Tool Runtime
     ↓
Independent Verification
     ↓
Audit Event

No privileged action may bypass the pipeline.

---

8. Handoff Contract

Implement a protocol-neutral handoff contract.

Example:

{
  "contract_version": "1.0",
  "canonicalization_version": "1",
  "initiator": "agent-payment-01",
  "requested_capability": "repository.write",
  "authorized_capabilities": [
    "repository.read",
    "repository.write",
    "test.execute"
  ],
  "payload_digest": "...",
  "policy_id": "POLICY-017",
  "policy_version": 17,
  "provenance": {
    "source": "agent-payment-01"
  },
  "expires_at": "...",
  "replay_protection": {
    "nonce": "..."
  }
}

The validator must reject:

- capability widening;
- invalid payload digest;
- expired contracts;
- replayed contracts;
- unsupported versions;
- canonicalization mismatch;
- provenance mismatch;
- duplicate/ambiguous fields;
- downstream capability escalation.

Fail closed.

---

9. Attack Demonstrations

The UI must provide buttons for at least these attacks.

Attack 1 — Unauthorized Credential Rotation

Agent attempts:

credential.rotate

Expected:

BLOCKED

Reason:
Capability not granted.

Requested:
credential.rotate

Authorized:
repository.read
repository.write
test.execute

Policy:
POLICY-017 v17

---

Attack 2 — IAM Modification

Agent attempts:

iam.modify

Expected:

BLOCKED

Reason:
Protected trust-boundary action.

Policy:
POLICY-017

Authorization:
DENIED

---

Attack 3 — Policy Self-Modification

Agent attempts:

policy.modify

Expected:

BLOCKED

Reason:
Agent cannot modify the policy governing its own authorization.

This is extremely important.

Demonstrate that an agent cannot rewrite the rules that determine whether the agent is trusted.

---

Attack 4 — Agent-to-Agent Capability Escalation

Simulate:

Agent A
   ↓
Agent B
   ↓
credential.rotate

Agent B must not inherit capabilities that were never granted to it.

Expected:

HANDOFF REJECTED

Reason:
Capability widening detected.

Parent grant:
repository.read
repository.write
test.execute

Requested downstream capability:
credential.rotate

---

Attack 5 — Replay Attack

Take a previously valid handoff contract and replay it.

Expected:

HANDOFF REJECTED

Reason:
Replay detected.

Nonce:
already consumed

---

Attack 6 — Policy Version Mismatch

Create:

Agent authorization:
POLICY-017 v17

Then change the active policy to:

POLICY-018 v18

Make "credential.rotate" prohibited.

Attempt the action again.

NAEOS must re-evaluate authorization at execution time.

Expected:

BLOCKED

Reason:
Authorization is stale.

Granted under:
POLICY-017 v17

Current policy:
POLICY-018 v18

Requested capability:
credential.rotate

This demonstrates that approval/authorization cannot become permanently valid merely because it was valid earlier.

---

10. Independent Verification

Do not let the same agent verify its own work.

Create a separate verifier component.

Example:

Agent
  ↓
Implementation
  ↓
Independent Verifier

The verifier receives:

- requested change;
- resulting diff;
- policy;
- capability grant;
- execution events;
- relevant test results.

The verifier produces:

{
  "verification_id": "VER-001",
  "result": "PASS",
  "policy_compliance": true,
  "tests_passed": true,
  "unauthorized_actions": 0
}

For adversarial cases:

{
  "verification_id": "VER-002",
  "result": "FAIL",
  "policy_compliance": false,
  "reason": "Unauthorized capability request detected"
}

Make the UI visually distinguish:

AGENT DECISION

from:

NAEOS VERIFICATION

---

11. Audit Trail

Create an append-only event ledger.

Every important event must generate an audit event.

Example:

{
  "event_id": "AUD-00042",
  "timestamp": "...",
  "event_type": "AUTHORIZATION_DENIED",
  "agent_id": "agent-payment-01",
  "requested_capability": "credential.rotate",
  "policy_id": "POLICY-017",
  "policy_version": 17,
  "decision": "BLOCK",
  "reason": "CAPABILITY_NOT_GRANTED"
}

Audit must record attempted actions and dispatch decisions, not only successful state changes.

Important events:

- intent received;
- authorization requested;
- authorization granted;
- authorization denied;
- handoff created;
- handoff rejected;
- tool dispatched;
- tool blocked;
- verification started;
- verification completed;
- policy changed;
- replay detected;
- capability escalation detected.

The audit trail must be independent from agent memory.

---

12. UI

Build a professional security/infrastructure dashboard.

Do not make it look like a generic CRUD application.

Use this layout:

┌───────────────────────────────────────────────────────────┐
│ NAEOS                                                     │
│ Engineering Control Plane                                 │
├───────────────────────────────────────────────────────────┤
│                                                           │
│ Agent                    Policy                           │
│ payment-agent-01         POLICY-017 v17                  │
│ Status: ACTIVE           Status: ACTIVE                  │
│                                                           │
├───────────────────────────────────────────────────────────┤
│ REQUEST PIPELINE                                          │
│                                                           │
│ Intent → Policy → Grant → Handoff → Execute → Verify     │
│   ✓       ✓        ✓        ✓        ✓        ✓          │
│                                                           │
├───────────────────────────────────────────────────────────┤
│ LIVE EVENT STREAM                                         │
│                                                           │
│ 14:03:01 repository.read                    ALLOWED       │
│ 14:03:04 repository.write                   ALLOWED       │
│ 14:03:08 test.execute                       ALLOWED       │
│ 14:03:12 credential.rotate                  BLOCKED       │
│ 14:03:17 policy.modify                      BLOCKED       │
│                                                           │
├───────────────────────────────────────────────────────────┤
│ SECURITY EVENTS                                           │
│                                                           │
│ Capability escalation detected                            │
│ Policy self-modification prevented                        │
│ Replay attempt prevented                                  │
│                                                           │
└───────────────────────────────────────────────────────────┘

---

13. Demo Mode

Create a "Run Investor Demo" button.

When clicked, automatically execute a scripted sequence:

Step 1

Agent receives:

Update payment service and run tests.

Step 2

Agent requests:

repository.read

Result:

ALLOW

Step 3

Agent requests:

repository.write

Result:

ALLOW

Step 4

Agent requests:

test.execute

Result:

ALLOW

Step 5

Agent attempts:

credential.rotate

Result:

BLOCK

Step 6

Agent attempts:

policy.modify

Result:

BLOCK

Step 7

Agent attempts:

Agent A → Agent B → credential.rotate

Result:

HANDOFF REJECTED

Step 8

Replay a valid contract.

Result:

REPLAY REJECTED

Step 9

Change:

POLICY-017 → POLICY-018

Attempt a previously authorized action.

Result:

STALE AUTHORIZATION → BLOCK

Step 10

Run independent verification.

Result:

VERIFICATION:
FAIL

Reason:
Unauthorized actions were attempted.

Finish with:

NAEOS SECURITY POSTURE

Authorized actions:      3
Blocked actions:         3
Handoff violations:      1
Replay attempts:         1
Policy changes:          1
Verification status:     FAIL
Audit events:            14

---

14. Technical Requirements

Use a clean modular architecture.

Suggested structure:

naeos-demo/
├── README.md
├── docs/
│   └── architecture.md
├── apps/
│   └── dashboard/
├── services/
│   ├── policy-engine/
│   ├── capability-authority/
│   ├── handoff-validator/
│   ├── execution-gate/
│   ├── verifier/
│   └── audit-ledger/
├── packages/
│   ├── contracts/
│   ├── policy/
│   ├── events/
│   └── security/
├── scenarios/
│   ├── normal-flow/
│   ├── credential-rotation/
│   ├── policy-escalation/
│   ├── agent-handoff/
│   ├── replay/
│   └── policy-version/
└── tests/

Keep the implementation simple enough to run locally.

Do not introduce unnecessary infrastructure.

For the demo, an in-memory or local persistent event store is acceptable.

However, the interfaces should make it possible to replace the storage implementation later.

---

15. Testing

Write automated tests for:

Authorization

- allowed capability;
- denied capability;
- expired grant;
- revoked grant.

Policy

- active policy;
- policy version mismatch;
- policy self-modification.

Handoff

- valid handoff;
- capability widening;
- payload mutation;
- provenance mutation;
- replay;
- expiration;
- canonicalization mismatch;
- unsupported version.

Execution

- authorized execution;
- unauthorized execution;
- stale authorization;
- protected trust-boundary action.

Audit

- every denied action generates an event;
- every execution generates an event;
- event contains policy version;
- event contains authorization decision;
- events are append-only.

Verification

- verifier is independent from agent decision;
- unauthorized action causes verification failure.

---

16. Threat Model

Document the following assumptions:

Trusted

- NAEOS policy authority;
- capability authority;
- execution gate;
- audit ledger;
- independent verifier.

Untrusted

- agent reasoning;
- agent-generated payload;
- tool output;
- downstream agent;
- external input;
- cached agent state.

Explicitly document:

«The agent is not the root of trust.»

---

17. Demo Principles

Follow these principles:

1. Fail closed.
2. Never trust agent intent as authorization.
3. Re-check consequential authorization at execution time.
4. Do not allow an agent to modify its own trust boundary.
5. Do not allow downstream agents to gain capabilities automatically.
6. Bind authorization to policy version.
7. Protect against replay.
8. Treat tool output as untrusted data.
9. Keep audit independent from agent memory.
10. Use independent verification.

---

18. Investor Narrative

The UI should make this sentence obvious:

«"NAEOS is the control plane between autonomous AI agents and enterprise engineering systems."»

Secondary message:

«"Agents can act. NAEOS decides what they are allowed to do."»

Third message:

«"Reasoning is delegated. Authorization is not."»

---

19. Deliverables

Produce:

1. Working local demo.
2. Dashboard UI.
3. Policy engine.
4. Capability authorization.
5. Handoff validator.
6. Execution gate.
7. Independent verifier.
8. Append-only audit event stream.
9. Six adversarial scenarios.
10. Automated tests.
11. Architecture documentation.
12. Investor demo script.
13. README with setup instructions.

---

20. Definition of Done

The project is complete only when this sequence works end-to-end:

Agent
 ↓
Request repository.write
 ↓
NAEOS
 ↓
Policy evaluation
 ↓
Capability grant validation
 ↓
Handoff validation
 ↓
Execution
 ↓
Independent verification
 ↓
Audit event
 ↓
ALLOW

And this sequence is blocked:

Agent
 ↓
Request credential.rotate
 ↓
NAEOS
 ↓
Capability check
 ↓
BLOCK
 ↓
Audit

And this sequence is blocked:

Agent A
 ↓
Agent B
 ↓
credential.rotate
 ↓
Handoff validator
 ↓
CAPABILITY WIDENING
 ↓
BLOCK

And this sequence is blocked:

Valid authorization under POLICY-017
 ↓
Policy changes to POLICY-018
 ↓
Old authorization reused
 ↓
Execution-time re-evaluation
 ↓
STALE AUTHORIZATION
 ↓
BLOCK

And this sequence is blocked:

Agent
 ↓
policy.modify
 ↓
NAEOS
 ↓
BLOCK

---

Final Instruction

Do not build a fake UI that merely displays "BLOCKED".

Implement the underlying authorization flow so that the UI reflects real decisions from the policy engine, capability authority, handoff validator, execution gate, verifier, and audit ledger.

The demo must be deterministic and reproducible.

Prioritize:

correct security semantics > visual polish > feature count.

Do not over-engineer the prototype.

Build the smallest technically credible system that demonstrates the NAEOS thesis.

Before writing code:

1. inspect the existing repository;
2. identify reusable NAEOS components;
3. preserve existing architecture where appropriate;
4. create an implementation plan;
5. identify gaps;
6. then implement the demo incrementally;
7. run tests;
8. run the complete investor scenario;
9. verify that every BLOCK/ALLOW result comes from the actual control plane.
