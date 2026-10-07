# NAEOS AI Integration Contract v1

Status: Proposed for the agent gateway vertical slice.

## Purpose

This contract defines the boundary between an external AI coding agent and the NAEOS control plane.

NAEOS does not grant authority because an agent is trusted. The agent submits intent; NAEOS normalizes, evaluates, executes, and records the resulting decision.

## Canonical flow

```text
AI Agent
   |
   | tool request
   v
JSON / Agent Adapter
   |
   v
NAEOS Execution Gateway
   |
   +--> Policy Evaluation
   |
   +--> Authorization Decision
   |
   +--> Runtime Sandbox
   |
   +--> Execution Result
   |
   +--> Agent Session / Evidence
```

## Request contract

The protocol-neutral JSON adapter accepts:

```json
{
  "capability": "repository.write",
  "tool": "file-edit",
  "action": "write",
  "resource": "src/app.go",
  "environment": "development",
  "actor": "codex",
  "payload": {},
  "context": {}
}
```

Required fields:

- `tool`
- `action`

Recommended identity fields:

- `actor`
- `capability`
- `resource`
- `environment`

## Enforcement invariants

1. Adapter normalization never grants authorization.
2. Every normalized request crosses the execution gateway.
3. Policy evaluation happens before sandbox execution.
4. DENY and REQUIRE_APPROVAL do not execute the request.
5. The gateway is fail-closed.
6. A policy revalidation boundary is used when the control plane supports it.
7. An agent session may record the decision, but session memory is not the authoritative audit record.
8. Execution results expose policy identity, decision, status, hash, and timestamp.
9. Future adapters must normalize into the same `ToolRequest` contract rather than bypassing the gateway.

## CLI reference implementation

The first protocol-neutral integration is available through:

```bash
naeos agent request --request-file request.json --output json
```

An agent can therefore integrate with NAEOS without embedding NAEOS internals.

## First vertical-slice acceptance criteria

- An external agent can submit a JSON tool request.
- NAEOS evaluates the request against governance policy.
- An allowed request reaches the runtime sandbox.
- A denied request is blocked before execution.
- The decision can be persisted against an agent session.
- The result is machine-readable as JSON.
- No provider-specific AI SDK is required at the enforcement boundary.

## Next contracts

The next integration stages should add:

1. signed or integrity-bound capability grants;
2. policy-version binding;
3. capability expiry and revocation;
4. richer evidence bundles;
5. native adapters for the first production AI coding agent;
6. independent verification of the resulting evidence.
