# AIP Execution Evidence Interoperability Experiment

## Goal

Test a narrow interoperability boundary between an AIP-style authorization decision and NAEOS execution evidence:

`identity/policy decision → authorized tool call → observation → evidence → independent verification`

The experiment does **not** claim a native AIP integration. It uses the concepts exposed by the current AIP v1alpha2 specification as an input shape: agent identity/session, policy version, tool authorization, and audit decision. AIP remains the authorization/enforcement boundary; NAEOS supplies the evidence and independent-verification model after the authorization decision.

AIP v1alpha2 defines policy-based authorization for MCP tool calls, optional agent identity/session management, policy hashing/signing, and an audit log for authorization decisions.

## Boundary

| Concern | AIP | NAEOS experiment |
|---|---|---|
| Agent identity / session | Owner | Carries stable reference |
| Tool authorization | Owner | Consumes decision |
| Allow / block before tool call | Owner | Records decision binding |
| Tool execution | Runtime / MCP server | Observation boundary |
| Execution outcome | Runtime | Evidence payload |
| Evidence integrity | External verifier | Verification assertions |
| Policy semantics | AIP | Not reimplemented |

## Scenario A — ALLOW

1. An AIP-style request identifies an agent session and MCP tool invocation.
2. The authorization layer returns `ALLOW` under policy version `v1`.
3. The tool is executed.
4. The runtime returns an outcome.
5. NAEOS records an evidence envelope binding the invocation to the authorization decision.
6. An independent verifier checks the binding and evidence integrity.

Expected invariant:

`ALLOW + matching invocation/context + observed outcome → verifiable execution evidence`

## Scenario B — BLOCK

1. An AIP-style request identifies an agent session and MCP tool invocation.
2. The authorization layer returns `BLOCK`.
3. The tool is **not** invoked.
4. NAEOS records denial evidence containing the authorization decision and invocation identity.
5. An independent verifier confirms that the evidence contains no execution outcome.

Expected invariant:

`BLOCK → no tool execution outcome`

This distinction is important: a denial record is evidence of an authorization boundary, not evidence that the tool executed and failed.

## Minimal evidence envelope

The experiment uses protocol-neutral fields:

- `invocation_id`: stable identifier for the tool invocation.
- `authorization_id`: identifier binding the invocation to the authorization decision.
- `agent_id`: stable agent identity reference.
- `tool`: normalized tool name.
- `policy_version`: policy version used for the decision.
- `decision`: `ALLOW` or `BLOCK`.
- `context_digest`: digest binding authorization context to the invocation.
- `outcome`: present only when execution actually occurred.
- `evidence_ref`: reference to the durable evidence record.

No AIP private key, credential, token, or secret is copied into the evidence fixture.

## Acceptance criteria

- ALLOW binds to exactly one invocation and policy version.
- ALLOW evidence contains an observed execution outcome.
- BLOCK evidence proves the authorization boundary was reached without asserting tool execution.
- A mismatched policy version or context digest fails independent verification.
- The experiment remains protocol-neutral and does not require changes to AIP or an AIP SDK.
- Fixtures are deterministic and safe to publish.

## Status

Design/fixture experiment. The repository fixtures define expected evidence only; they are not external execution results.

Next step: run the two scenarios against an AIP-compatible runtime or proxy and replace the expected fixtures with independently captured evidence, keeping implementation evidence separate from external validation evidence.
