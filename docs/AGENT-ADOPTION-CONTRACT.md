# NAEOS External Agent Adoption Contract v1

Status: Proposed

This is the protocol-neutral adoption boundary for external AI coding agents.

## Canonical boundary

External Agent -> Provider Adapter -> Normalized Request -> NAEOS Execution Gateway -> Policy -> Authorization -> Replay Check -> Runtime -> Observation/Evidence -> Result

An adapter translates provider input. It MUST NOT authorize, execute, widen capability, or bypass the gateway.

## Version

A conforming integration declares:

```json
{"contract":"naeos.agent-adoption","version":"1.0"}
```

Unknown major versions MUST fail closed. Minor versions MUST preserve the meaning of existing security fields.

## Normalized request

```json
{
  "contract":"naeos.agent-adoption",
  "version":"1.0",
  "request_id":"req_01J...",
  "invocation_id":"inv_01J...",
  "actor":{"type":"agent","id":"codex"},
  "capability":"repository.write",
  "tool":"file-edit",
  "action":"write",
  "resource":"src/app.go",
  "environment":"development",
  "payload":{},
  "context":{"repository":"example/repo","commit":"abc123"}
}
```

Required: `request_id`, `invocation_id`, `tool`, and `action`. Integrations SHOULD provide actor, capability, resource, environment, repository, and commit identity when available.

Missing context MUST NOT become implicit authority.

## Invocation

`invocation_id` identifies one consequential execution attempt.

When replay protection is enabled:

1. policy and authorization are evaluated;
2. the invocation is claimed immediately before the sandbox side effect;
3. a consumed invocation cannot execute again;
4. durable state survives process restart and shared replicas;
5. replay-state failure fails closed.

An adapter MUST NOT silently reuse an invocation ID for a new consequential action.

## Capability and policy

A request is not authorization.

A capability grant, when present, is validated against requested capability, policy identity/version, current state, expiry, revocation, and scope.

A grant is bounded authority, not a permanent permission ticket.

Policy is authoritative at the control plane. If governing policy changes before the consequential boundary, authorization is revalidated according to gateway semantics. A stale authorization cannot be treated as current authority merely because an agent still holds the old response.

## Decisions

| Decision | Meaning | Execution |
|---|---|---|
| `ALLOW` | Currently authorized | May execute after gateway checks |
| `DENY` | Not authorized | Must not execute |
| `REQUIRE_APPROVAL` | Human approval required | Must not execute until approved |
| `REQUIRE_VERIFICATION` | Verification required | Must not execute until satisfied |

Adapters MUST preserve these semantics and MUST NOT turn denial or pending approval into success.

## Result

A conforming integration exposes structured authorization and execution state:

```json
{
  "contract":"naeos.agent-adoption",
  "version":"1.0",
  "request_id":"req_01J...",
  "invocation_id":"inv_01J...",
  "decision":"ALLOW",
  "status":"completed",
  "policy":{"id":"repo-policy","version":"42","digest":"sha256:..."},
  "grant":{"id":"grant_01J...","digest":"sha256:..."},
  "execution":{"result":"ok"},
  "evidence":{"id":"evidence_01J...","digest":"sha256:..."}
}
```

Transport is intentionally unspecified: CLI, HTTP, RPC, or another transport may conform if semantics are preserved.

## Adapter trust boundary

The adapter contract is a protocol and integration boundary, not a containment mechanism for arbitrary code. An adapter MUST NOT perform consequential side effects outside the NAEOS gateway, but an interface alone cannot technically prevent arbitrary provider code from making filesystem, network, process, or credential calls if that code has the operating-system authority to do so.

Deployments that treat adapters as untrusted code SHOULD run them without credentials or direct side-effect capability. Consequential credentials and capability handles SHOULD remain owned by the NAEOS-controlled runtime. A conforming adapter implementation MUST therefore be evaluated together with its deployment privilege boundary; passing protocol conformance does not by itself prove adapter containment.

## Evidence

Agent transcript, memory, and self-reported completion are not authoritative evidence.

Integrations SHOULD preserve request/invocation identity, policy and grant identity, execution status, observed side effect, evidence reference, and independent verification result where configured.

Integration-generated evidence must remain distinguishable from independently observed evidence.

## Fail-closed semantics

The integration MUST fail closed for unknown major versions, malformed security fields, missing invocation identity when replay protection is enabled, invalid/expired/revoked grants, policy mismatch, stale authorization, replay, unavailable durable replay state, missing approval, or unsatisfied verification.

Transport or execution failure MUST NOT be represented as successful completion.

After an ambiguous consequential execution, integrations MUST NOT automatically retry the same invocation unless the governing execution semantics explicitly permit it.

## Conformance

A provider adapter conforms when it:

1. translates provider input into the normalized request;
2. preserves request and invocation identity;
3. never grants authority;
4. never executes outside the NAEOS gateway;
5. preserves decision semantics;
6. returns execution/evidence state without inventing success;
7. does not silently retry consequential invocations.

The Codex adapter is the first provider-specific implementation. Future adapters SHOULD conform to this contract instead of adding provider-specific authorization paths.

## Minimal adoption test

1. Submit a unique invocation.
2. Observe policy decision.
3. Execute an allowed request through the gateway.
4. Record invocation, policy, grant, execution, and evidence identifiers.
5. Replay the same invocation and confirm denial.
6. Change or invalidate authorization and confirm stale authority cannot cross the consequential boundary.
7. Independently verify evidence where configured.

Negative paths are part of conformance; an ALLOW path alone is insufficient.

## Relationship to existing contracts

- `docs/AI-INTEGRATION-CONTRACT.md` documents the initial gateway/Codex vertical slice.
- The runtime gateway remains the single authorization/execution boundary.
- Capability grant, expiry/revocation, replay protection, and durable replay are existing runtime controls.
- `docs/GOLDEN-PATH.md` and `docs/EXTERNAL-VALIDATION.md` remain the canonical local evaluation path.

This document does not create a second Golden Path or authorization model.

> An agent submits intent. NAEOS determines authority. The runtime enforces the boundary. Evidence records what happened.
