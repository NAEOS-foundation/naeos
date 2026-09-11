# NAEOS Investor Demo — Control Plane for AI Agent Authorization

## Overview

The NAEOS Investor Demo is a technical demonstration of NAEOS acting as a control plane between autonomous AI agents and enterprise engineering systems.

**Core thesis:**

> AI agents can reason and perform engineering work, but authorization, policy enforcement, verification, and audit must remain outside the agent's control.

## Architecture

```
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
```

## Key Components

### 1. Policy Engine (`policy_engine.go`)

Evaluates whether capabilities are allowed or denied by policy.

- Stores and manages policies with versioning
- Evaluates capability requests against the active policy
- Identifies protected capabilities (e.g., `iam.modify`, `policy.modify`)

### 2. Capability Authority (`policy_engine.go`)

Checks if an agent is authorized for a specific capability.

- Validates that the agent has a valid grant
- Checks grant expiration and revocation status
- Ensures the capability is in the grant's allowed list
- Verifies the capability is allowed by the policy

### 3. Handoff Contract Validator (`execution_gate.go`)

Validates handoff contracts between components and detects attacks.

- **Replay protection**: Detects if a contract with the same nonce has been seen before
- **Capability widening**: Ensures requested capabilities are in the authorized list
- **Downstream escalation**: Prevents agents from gaining capabilities via handoff
- **Payload tampering**: Verifies payload digest integrity
- **Expiration checks**: Ensures contracts haven't expired

### 4. Execution Gate (`execution_gate.go`)

The main authorization barrier for all consequential actions.

- Re-checks authorization at execution time (prevents stale auth)
- Validates handoff contracts
- Blocks protected capabilities
- Runs independent verification

### 5. Independent Verifier (`verifier.go`)

Independently verifies that authorized actions comply with policy.

- Separate from the agent's own verification
- Checks for unauthorized attempts in audit log
- Validates policy compliance

### 6. Audit Ledger (`verifier.go`)

Append-only event log for all security decisions.

- Records every authorization decision (granted or denied)
- Tracks attempted attacks
- Records execution events
- Cannot be modified or deleted

## Attack Demonstrations

The demo includes six attack scenarios:

### Scenario 1: Authorized Repository Read ✓ ALLOW
Agent requests `repository.read` which is in its grant.

### Scenario 2: Unauthorized Credential Rotation ✗ BLOCK
Agent attempts `credential.rotate` which is:
- Not in its grant
- Protected by policy

### Scenario 3: Policy Self-Modification ✗ BLOCK
Agent attempts `policy.modify` to change the policy governing its own authorization.

**Critical invariant**: Agents cannot widen their own trust boundaries.

### Scenario 4: Capability Escalation via Agent Handoff ✗ BLOCK
Agent A hands off to Agent B requesting a capability that was never granted.

The validator detects capability widening and blocks the handoff.

### Scenario 5: Replay Attack ✗ BLOCK
Agent attempts to replay a previously valid handoff contract using the same nonce.

The validator maintains a nonce ledger and rejects replayed contracts.

### Scenario 6: Stale Authorization ✗ BLOCK
Authorization is initially valid under `POLICY-017`, but when the policy changes to `POLICY-018`, old authorizations are no longer valid.

## Running the Demo

### Prerequisites

- Go 1.21 or later
- NAEOS repository

### Build

```bash
cd /workspaces/naeos
go build -o cmd/naeos-demo/naeos-demo ./cmd/naeos-demo
```

### Run

```bash
./cmd/naeos-demo/naeos-demo -addr :9090
```

### Access

- **Dashboard**: http://localhost:9090/
- **API**: http://localhost:9090/api/
- **WebSocket demo**: ws://localhost:9090/ws

### API Endpoints

#### Health Check
```bash
curl http://localhost:9090/api/health
```

#### Check Authorization
```bash
curl -X POST http://localhost:9090/api/authorize \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-payment-01",
    "capability": "repository.read"
  }'
```

#### Execute Action
```bash
curl -X POST http://localhost:9090/api/execute \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-payment-01",
    "capability": "repository.read",
    "payload": {}
  }'
```

#### Get Audit Events
```bash
curl http://localhost:9090/api/audit
```

#### Run All Scenarios
```bash
curl -X POST http://localhost:9090/api/scenarios
```

#### Verify Agent Session
```bash
curl -X POST http://localhost:9090/api/verification \
  -H "Content-Type: application/json" \
  -d '{"agent_id": "agent-payment-01"}'
```

#### Reset Demo
```bash
curl -X POST http://localhost:9090/api/reset
```

## Testing

Run the comprehensive test suite:

```bash
cd /workspaces/naeos
go test ./internal/investordemo -v
```

Test categories:

- **Policy Engine Tests**: Allowed/denied capabilities, protected capabilities
- **Capability Authority Tests**: Invalid grants, expired grants, revoked grants
- **Handoff Validator Tests**: Valid contracts, expired contracts, replay detection, capability widening, payload mutation, provenance mutation, unsupported contract version, canonicalization mismatch, signed/tampered signature
- **Execution Gate Tests**: Authorized/unauthorized execution, protected capabilities, stale authorization, IAM modification, external publish
- **Audit Ledger Tests**: Event recording, append-only guarantee
- **Verifier Tests**: Authorized actions, unauthorized attempts
- **Scenario Tests**: All six attack scenarios

## Dashboard

The web dashboard provides:

- **Agent Status**: Current grant and policy
- **Test Capabilities**: Test individual capabilities
- **Attack Scenarios**: Run predefined attacks
- **Authorization Pipeline**: Visual representation of the authorization flow
- **Live Event Stream**: Real-time audit events
- **Verification Status**: Independent verification results
- **Security Posture**: Metrics on allowed/blocked/attempted actions

## Threat Model (§16)

### Trusted components

- **Policy Engine** — evaluates capabilities against versioned policy.
- **Capability Authority** — validates grants against policy.
- **Execution Gate** — the single enforcement point for all consequential actions.
- **Handoff Contract Validator** — rejects tampered, replayed, escalated, and expired contracts.
- **Independent Verifier** — separate from execution path; reviews the audit log.
- **Audit Ledger** — append-only, cannot be modified by any agent or component.

### Untrusted components

- **Agent reasoning** — the model's justification for an action is never the authorization decision.
- **Agent-generated payloads** — payloads are data, not permission.
- **Tool output** — treated as untrusted data.
- **Downstream agents** — no capability inheritance beyond explicit grant.
- **External inputs** — may be adversarial.
- **Cached agent state** — may be stale or tampered.

### Key invariant

> The agent is not the root of trust.

Agent intent is requests; NAEOS is authorization. The model's reasoning never becomes the authorization decision.

## Key Invariants

The demo enforces these critical security invariants:

1. **Agent intent ≠ authorization**: The agent's request is never the authorization decision.
2. **Fail closed**: Unknown or ambiguous cases default to BLOCK.
3. **Re-check at execution**: Authorization is re-evaluated at execution time.
4. **Agent cannot modify its trust boundary**: No self-modification attacks.
5. **Downstream capability protection**: Agents cannot escalate capabilities via handoff.
6. **Replay protection**: Nonces prevent contract replay attacks.
7. **Audit independence**: Audit log is separate from agent state.
8. **Protected capabilities**: Certain actions are never delegated to agents.

## Policy Model

Default policy (`POLICY-017`):

```json
{
  "policy_id": "POLICY-017",
  "version": 17,
  "status": "active",
  "allowed_capabilities": [
    "repository.read",
    "repository.write",
    "test.execute"
  ],
  "protected_capabilities": [
    "iam.modify",
    "credential.rotate",
    "production.deploy",
    "external.publish",
    "policy.modify"
  ]
}
```

Default grant (`GRANT-001`):

```json
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
  "expires_at": "2024-09-12T...",
  "status": "active"
}
```

## Deployment Notes

For production use of NAEOS as an authorization control plane:

1. **Policy Authority**: Policies must be managed by a trusted authority, not by agents.
2. **Audit Trail**: Maintain audit events in append-only storage (e.g., blockchain, write-once storage).
3. **Independent Verification**: Verification must be performed by a separate, trusted component.
4. **Policy Versioning**: Always bind authorizations to specific policy versions.
5. **Replay Protection**: Maintain nonce ledgers across system restarts.
6. **Capability Grants**: Implement short expiration windows and regular re-authorization.

## Files

```
internal/investordemo/
├── types.go                    # Core data types
├── policy_engine.go            # Policy and capability engines
├── grant_store.go              # Grant storage
├── execution_gate.go           # Handoff validator and execution gate
├── verifier.go                 # Audit ledger and verifier
├── scenarios.go                # Demo scenarios
├── api_server.go               # HTTP API
└── investordemo_test.go        # Comprehensive tests

cmd/naeos-demo/
└── main.go                     # HTTP server with dashboard
```

## Architecture Decisions

1. **Modular design**: Each component is independent and testable.
2. **In-memory storage**: Acceptable for demo; replace with persistent storage in production. The interfaces `GrantRepository`, `AuditEventStore`, and `PolicyRepository` (defined in `storage.go`) document and enforce the replacement seam — concrete types implement them and a compile-time assertion verifies the contract.
3. **HTTP + WebSockets**: RESTful API for scenarios, WebSockets for interactive demo.
4. **Append-only audit**: Events cannot be modified or deleted.
5. **Simple policy model**: Demonstrated capabilities; extend with more complex policies as needed.

## Future Enhancements

- [ ] Persistent storage (database)
- [ ] Multi-tenant policies
- [ ] Policy approval workflows
- [ ] Rate limiting per capability
- [ ] Time-based capability expiration
- [ ] Role-based access control (RBAC)
- [ ] Attribute-based access control (ABAC)
- [ ] Integration with GitHub Copilot
- [ ] Integration with Claude Code
- [ ] Kubernetes admission controller

## License

This demo is part of the NAEOS Foundation project and follows the same license.

## Support

For questions or issues, please open an issue on the NAEOS GitHub repository.
