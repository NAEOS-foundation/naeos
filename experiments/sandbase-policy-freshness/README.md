# SandBase Policy Freshness Interoperability Experiment

## Goal

Verify that a long-running agent session cannot reuse an authorization produced under policy v1 after the governing policy changes to v2.

This experiment is intentionally protocol-neutral. SandBase Harness remains responsible for session, sandbox, MCP, permission/approval, audit, and replay behavior; NAEOS supplies policy lifecycle and authorization freshness.

## Golden path

`Agent → Policy → Authorization → Sandboxed Execution → Verification → Evidence → Audit`

### Scenario

1. Session starts under policy v1.
2. A consequential tool action is authorized.
3. Policy changes from v1 to v2 while the session remains active.
4. The agent attempts the same consequential action.
5. NAEOS detects the authorization is stale.
6. Execution is blocked pending re-authorization.
7. The decision and stale-authorization event are recorded as evidence/audit data.

## Expected decision

```text
ALLOW(v1)
POLICY_CHANGED(v2)
STALE_AUTHORIZATION
REAUTHORIZE
DENY_UNTIL_REAUTHORIZED
```

## Minimal authorization contract

```json
{
  "decision": "ALLOW",
  "policy_version": "v1",
  "capability": "tool.execute",
  "target": "repo.write",
  "expires_at": "2026-10-04T18:00:00Z",
  "context_digest": "<digest>"
}
```

Freshness rule:

```text
identity valid
AND capability valid
AND policy_version == current_policy_version
AND context_digest == current_context_digest
→ authorization is fresh

otherwise
→ REAUTHORIZE
```

## Interoperability boundary

| Concern | SandBase Harness | NAEOS |
|---|---|---|
| Session lifecycle | owner | observes/binds |
| Sandbox execution | owner | policy constraint |
| MCP/tool permissions | owner | external policy decision |
| Policy lifecycle | integration point | owner |
| Authorization freshness | integration point | owner |
| Verification | runtime evidence | policy/evidence binding |
| Audit/replay | owner | consumes/emits decision evidence |

## Acceptance criteria

- Existing Harness behavior is unchanged when no external authorization hook is configured.
- A policy-version change invalidates the prior authorization.
- A consequential action is not executed with stale authorization.
- Re-authorization can restore execution under policy v2.
- The stale decision and re-authorization are auditable without secrets.

## Status

Design experiment. Next step: run against a SandBase Harness session and capture the resulting evidence/audit record.
