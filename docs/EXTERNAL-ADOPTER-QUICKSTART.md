# External Agent Adoption Quickstart

Status: v1 adoption guide

This guide shows how an external AI agent can exercise the public NAEOS adoption boundary without GitHub authorization or private NAEOS context.

## 1. Boundary

```text
External Agent
    ↓
Provider Adapter
    ↓
naeos.agent-adoption envelope
    ↓
NAEOS Execution Gateway
    ↓
Policy → Authorization/Revalidation → Replay → Runtime → Observation/Evidence
    ↓
Result
```

The adapter translates provider input. It does not authorize, execute, widen capabilities, or create a second gateway.

## 2. Build the public CLI

From a public checkout:

```bash
go build -o ./naeos ./cmd/naeos
```

No GitHub token or private repository access is required.

## 3. Use the reference adapter

The CLI exposes the protocol-neutral reference adapter as:

```bash
./naeos agent request \
  --adapter reference-external-agent \
  --request-file request.json \
  --output json
```

Example `request.json`:

```json
{
  "contract": "naeos.agent-adoption",
  "version": "1.0",
  "request_id": "req-adopter-001",
  "invocation_id": "inv-adopter-001",
  "actor": "external-agent",
  "capability": "repository.read",
  "tool": "filesystem",
  "action": "read",
  "resource": "README.md",
  "environment": "development",
  "context": {
    "repository": "example/repo",
    "commit": "abc123"
  }
}
```

The reference adapter only normalizes this envelope. Authorization and execution remain in the existing gateway.

## 4. Replay protection

Replay protection is enabled by the CLI.

### Local/test mode

If `--replay-db` is omitted, the CLI uses an in-memory invocation store. This is intentionally process-local and MUST NOT be interpreted as restart- or replica-durable protection.

### Durable mode

Configure a persistent database connection, for example SQLite:

```bash
./naeos db connect \
  --type sqlite \
  --name adoption-replay \
  --user local \
  --database "$HOME/.naeos/adoption-replay.db"
```

Then use:

```bash
./naeos agent request \
  --adapter reference-external-agent \
  --replay-db adoption-replay \
  --request-file request.json \
  --output json
```

The durable store claims the invocation immediately before the sandbox side effect. A consumed invocation cannot execute again after a process restart when the same durable database is reused.

## 5. Required negative tests

An adopter should verify at least:

1. missing `invocation_id` → fail closed;
2. replaying the same invocation → denied before a second side effect;
3. changing `repository.read` to `repository.write` without valid authority → denied;
4. stale authorization → denied before the consequential boundary;
5. `REQUIRE_APPROVAL` → must not execute;
6. durable replay-store failure → fail closed;
7. unsupported contract major version → fail closed;
8. adapter code cannot be treated as a security containment boundary.

The repository's external-agent conformance suite is the canonical semantic reference.

## 6. External adapter implementation

A provider adapter can be implemented in any language or transport as long as it preserves the contract:

- `request_id` identifies the request;
- `invocation_id` identifies the consequential execution attempt;
- the adapter does not grant authority;
- the adapter does not execute outside NAEOS;
- policy decisions are preserved;
- transport errors and execution failures are not represented as success;
- consequential invocations are not silently retried.

Transport is intentionally unspecified: CLI, HTTP, RPC, or another mechanism may conform.

## 7. Adapter threat model

Protocol conformance does not prove arbitrary-code containment.

An adapter is integration code. If it runs with filesystem, network, process, or credential authority, it can technically make side effects without calling NAEOS. Deployments that treat adapters as untrusted SHOULD run them without consequential credentials or direct side-effect capability. Those capabilities SHOULD remain in the NAEOS-controlled runtime.

## 8. Evidence to record

For an external validation run, record:

- public commit evaluated;
- agent/provider type;
- adapter implementation;
- transport;
- replay-store mode;
- conformance results;
- negative-path results;
- observed side effects;
- evidence/verification references;
- documentation or contract ambiguities.

Do not claim certification, compliance, or production security from this quickstart alone.

> An agent submits intent. NAEOS determines authority. The runtime enforces the boundary. Evidence records what happened.
