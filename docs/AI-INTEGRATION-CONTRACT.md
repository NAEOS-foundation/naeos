# NAEOS AI Integration Contract v1

Status: Proposed for the agent gateway vertical slice.

## Purpose

This contract defines the boundary between an external AI coding agent and the NAEOS control plane.

NAEOS does not grant authority because an agent is trusted. The agent submits intent; NAEOS normalizes, evaluates, executes, and records the resulting decision.

## Canonical flow

```text
AI Agent
   |
   v
Agent Adapter
   |
   v
NAEOS Execution Gateway
   |
   +--> Policy Evaluation
   +--> Authorization Decision
   +--> Runtime Sandbox
   +--> Execution Result
   +--> Agent Session / Evidence
```

## Request contract

The protocol-neutral adapter accepts a normalized request:

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

Required fields are `tool` and `action`. Identity and scope fields such as `actor`, `capability`, `resource`, and `environment` should be supplied whenever the agent can provide them.

## Codex bridge envelope

The first provider-specific adapter uses a deliberately narrow bridge envelope. It accepts a function-call object and converts it into the same `ToolRequest` used by the gateway:

```json
{
  "type": "function_call",
  "name": "file-edit",
  "arguments": {
    "action": "write",
    "resource": "src/app.go",
    "capability": "repository.write",
    "environment": "development",
    "payload": {
      "operation": "replace"
    }
  }
}
```

The adapter is a boundary translator, not an authorization layer. It does not execute a Codex action and it does not widen capabilities. The normalized request still crosses the NAEOS execution gateway.

## Enforcement invariants

1. Adapter normalization never grants authorization.
2. Every normalized request crosses the execution gateway.
3. Policy evaluation happens before sandbox execution.
4. DENY and REQUIRE_APPROVAL do not execute the request.
5. The gateway is fail-closed.
6. A policy revalidation boundary is used when the control plane supports it.
7. Agent session memory is not the authoritative audit record.
8. Execution results expose policy identity, decision, status, hash, and timestamp.
9. Future adapters must normalize into the same `ToolRequest` contract rather than bypassing the gateway.

## CLI

```bash
naeos agent request --request-file request.json --output json
naeos agent request --adapter codex --request-file codex-request.json --output json
```

## Acceptance criteria

- An external agent can submit a tool request.
- NAEOS evaluates the request against governance policy.
- An allowed request reaches the runtime sandbox.
- A denied request is blocked before execution.
- The decision can be persisted against an agent session.
- The result is machine-readable as JSON.
- Provider-specific normalization remains outside the policy and execution layers.

## Next

The next integration stages should add signed or integrity-bound capability grants, policy-version binding, capability expiry and revocation, richer evidence bundles, and independent verification of resulting evidence.
