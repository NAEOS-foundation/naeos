# ADR — Production Control Plane Runtime and Supabase Persistence

- **Status:** Accepted for implementation
- **Date:** 2026-11-11
- **Decision owners:** NAEOS maintainers
- **Scope:** Railway service `naeos-control-plane` and its dedicated Supabase PostgreSQL persistence
- **Related:** `docs/architecture/NAEOS-AUTHORIZATION-RUNTIME-CALLER-MAP.md`, `docs/control-plane/live-proof.md`

## 1. Context and observed baseline

The production Railway service currently builds `Dockerfile.control-plane`, whose command is:

```sh
naeos demo --addr :8080
```

That command constructs `investordemo.SetupDemoEnvironment()`. The demo setup seeds a fixed policy and grant in memory, constructs a separate `internal/controlplane.DecisionGateway`, and optionally restores the control-plane ledger from `NAEOS_CONTROLPLANE_LEDGER_PATH`.

The following facts were verified during the 2026-11-11 repository and deployment audit:

- Railway production service `naeos-control-plane` has a `DATABASE_URL` variable name configured.
- The current `naeos demo` entrypoint does not read `DATABASE_URL`.
- `naeos api` reads `NAEOS_DB_DRIVER` and `NAEOS_DB_DATABASE`, and its existing attachment creates only the `pipeline_runs` table. It is not the persistence path for the control-plane policy, grant, approval, or decision ledger.
- `internal/controlplane.Ledger` currently supports an optional JSON snapshot path. Without that path, appends are process-memory-only. Its file snapshot loader validates the event hash chain, but does not provide a multi-replica database transaction boundary.
- The investor-demo setup uses in-memory policy, grant, audit, and approval stores. The demo API also registers reset, scenario, execution, and other demo endpoints.
- The Railway production service has no attached persistent volume. The CRM service is a separate service and is explicitly outside this ADR.
- Supabase reports the `naeos.db` project as `ACTIVE_HEALTHY` on PostgreSQL 17.6 in `ap-northeast-1`. The project currently has only its default branch (no isolated staging branch); the `public` schema contains no application tables. Security, performance, and health advisors returned no lints at audit time.
- A 24-hour sample of Supabase pooler logs showed Supabase internal roles but no identifiable NAEOS Control Plane runtime role. This is absence of positive connection evidence, not proof that no connection has ever occurred. Railway's latest runtime logs show the investor demo starting, not a database-backed production server.
- The demo HTTP router registers demo and Control Plane routes together. Source-level middleware explicitly calls the static-token guard for `POST /api/control-plane/decision`, but does not apply that same guard globally to the other registered routes. Some sampled proxy requests to evidence/runs returned 403; that observed status does not replace a route-by-route source authorization review.
- The Supabase project audit found no application tables in the `public` schema at the time of review. Existing Supabase migration history is not evidence that the repository's runtime schema has been installed.

These observations mean that the presence of `DATABASE_URL` and a successful HTTP deployment do not establish production database usage or durable Control Plane state.

## 2. Decision

Turn `naeos-control-plane` into a dedicated production Control Plane service backed by Supabase PostgreSQL.

The production runtime MUST NOT start the investor-demo composition root. The investor demo remains a separate, explicitly non-production experience.

The production service will use the existing `internal/controlplane` domain contracts for the API decision/evidence path. This does not silently replace or consolidate the separate CLI runtime path documented in `NAEOS-AUTHORIZATION-RUNTIME-CALLER-MAP.md`. Any future unification with `internal/governance/control` or `internal/runtime/gateway` requires a separate ADR and caller-level regression evidence.

## 3. Runtime and persistence requirements

### 3.1 Production composition root

Introduce an explicit production server command/composition root. It must:

- require a valid PostgreSQL connection configuration at startup;
- establish and ping the database before binding the HTTP listener;
- load persisted policy versions, active policy references, grants, approval state, and ledger state;
- fail startup if required schema, signing/trust configuration, or persistence dependencies are unavailable;
- expose only the production API route allowlist;
- keep demo reset, scripted scenario, and demo-only execution endpoints out of the production router;
- implement graceful shutdown and close database resources.

Do not repurpose `naeos api` as the Control Plane merely because it already has a generic database adapter. Its current data model and route surface are different.

### 3.2 PostgreSQL persistence boundary

Use a dedicated repository/storage interface so domain policy evaluation and ledger verification do not depend directly on SQL.

The initial PostgreSQL schema should live in a private application schema (proposed: `naeos_control_plane`) and include, at minimum:

- immutable policy versions plus an explicit active-version reference;
- grants with agent scope, policy/version binding, expiry, revocation status and timestamps;
- approvals with an atomic one-time state transition (pending → approved/rejected/expired/consumed);
- append-only ledger events with monotonically ordered sequence, canonical event payload, previous hash, event hash, and unique event identity;
- a ledger head/lock record to serialize appends and calculate the next hash from the committed predecessor.

Ledger append must be one database transaction: lock the ledger head, read the predecessor hash, construct and hash the canonical event, insert the event, update the head, then commit. On any persistence error, the event must not be reported as committed and no protected side effect may proceed.

The database must reject update/delete operations against committed ledger events for the runtime role. The application role must have only the minimum privileges required for runtime reads and writes. Do not grant `anon` or `authenticated` browser roles direct access to private Control Plane tables.

Policy versions are immutable. Updating a policy creates a new version and changes the active pointer atomically; it must never rewrite the prior version. Grant validation must be re-evaluated against the current active policy at the final controllable boundary.

### 3.3 Connection contract

- `DATABASE_URL` is the single runtime connection-string contract for this service.
- Parse the URL with a PostgreSQL driver/config parser; do not manually concatenate credentials into a DSN.
- Require TLS. Prefer `verify-full` with the Supabase root certificate mounted/configured when operationally available; `require` is the minimum interim mode only if hostname/CA verification cannot yet be provisioned.
- Use a bounded connection pool, explicit connection lifetime/idle limits, connection timeout, and startup ping.
- Railway's network path must be checked before selecting the endpoint. For an IPv4-only path, use the Supabase shared **session** pooler URL copied from the Supabase Connect dialog. Do not guess the pooler hostname or username from the region. Avoid transaction mode for this persistent backend because it does not preserve all session-level PostgreSQL features.
- Use a separate migration connection/role for schema deployment. Prefer the direct database endpoint for migration and backup operations when the network supports it; otherwise use the supported session-mode connection.
- Never log the connection string, password, or query parameters containing credentials.
- Do not copy or reuse the CRM database URL. The Control Plane must use its own dedicated Supabase database role/credentials scoped to its private schema.

### 3.4 Identity and authorization

The hosted NAEOS Identity decision uses Supabase as the selected provider. Production API authentication must validate issuer, audience, expiry, signature, and required claims using the provider's published signing keys/JWKS and a bounded key-refresh policy.

Do not treat `NAEOS_CONTROLPLANE_API_TOKEN` as the final end-user identity system. A static bearer token may remain temporarily for controlled service-to-service compatibility, but it must not be embedded in browser code or become a substitute for user/agent identity, scoped authorization, or revocation.

Do not assume the provider's signing algorithm or JWKS behavior. Verify the actual project signing-key configuration and token claims before selecting accepted algorithms. Reject `alg=none`, unsupported algorithms, wrong issuer/audience, expired tokens, unknown key IDs, and invalid signatures.

### 3.5 Health and readiness

Separate liveness from readiness:

- liveness reports that the process is alive;
- readiness reports success only when PostgreSQL is reachable, required schema is compatible, and required persistence/key dependencies are initialized.

A health response must not expose secrets, SQL, connection strings, or internal infrastructure details.

## 4. Migration and rollout plan

Implement as independently reviewable phases. No production data migration or cutover is implied by accepting this ADR.

| Phase | Scope | Required evidence / exit gate |
|---|---|---|
| P0 — Architecture baseline | This ADR, runtime inventory, route allowlist, schema and failure contracts | Reviewed source-backed design; no production mutation |
| P1 — Storage contracts | Repository interfaces, PostgreSQL ledger/policy/grant/approval adapters, migrations | Unit tests, PostgreSQL integration tests, hash-chain and concurrency tests, failure-injection tests |
| P2 — Production composition | Dedicated production command/router, startup dependency checks, health/readiness, no demo routes | Route allowlist tests, fail-closed startup tests, no demo endpoint exposure |
| P3 — Identity integration | Hosted identity JWT/JWKS verification and scoped service authorization | Positive/negative token tests, key rotation/cache tests, issuer/audience/algorithm rejection tests |
| P4 — Staging validation | Apply schema to an isolated staging database; seed controlled test policy/grant; exercise persistence across restart | Migration status, restart recovery, concurrent append, RLS/privilege tests, advisor review |
| P5 — Production cutover | Configure dedicated runtime role/URL and deploy the production composition root | Explicit preflight, backup/restore proof, deployment receipt, readiness and smoke tests, rollback path |
| P6 — Hardening | Observability, retention, backup cadence, incident and recovery runbooks | Operational evidence and reviewed runbooks |

## 5. Security and operational gates

The following are release blockers:

- database connection failure, missing schema, unsupported schema version, or ledger integrity failure;
- missing/invalid signing and trust configuration where required by an endpoint;
- policy/grant lookup ambiguity or stale policy/grant state;
- ledger append failure or hash-chain mismatch;
- an approval that is expired, rejected, already consumed, or bound to a different decision/artifact;
- any route outside the approved production route allowlist;
- any production migration without a reviewed migration file and a recoverable backup.

No security gate may be bypassed to obtain a green deployment. Never silently fall back to in-memory stores or JSON files in production.

## 6. Rollback

Rollback is a code/deployment rollback, not a destructive schema rollback. Production schema changes must be forward-compatible during the rollout window. Preserve committed ledger events; never truncate or rewrite the ledger to roll back application code.

Before cutover, record the last known-good Railway deployment and validate the prior image can be redeployed. If the new runtime fails readiness, restore the previous deployment while retaining the database schema and data. Any down migration that drops or rewrites persisted policy, approval, or ledger data is prohibited as an automatic rollback action.

## 7. Explicit non-goals

This ADR does not:

- modify the CRM database, service, schema, or credentials;
- apply migrations to the production Supabase project;
- provision a new Supabase project or paid resource;
- claim that the current production service is already using Supabase;
- make the investor demo production-safe by relabeling it;
- claim that a successful HTTP health check proves durable state;
- merge the distinct CLI runtime authorization packages without a separate design review.

## 8. Verification checklist

Before P5, attach evidence for:

1. migration files reviewed and applied to staging;
2. database role grants and private-schema access checked;
3. ledger event hash chain verified before and after restart;
4. concurrent append test proves unique order and no forked hash chain;
5. failed database writes do not become visible as committed events;
6. policy version change invalidates stale grants at execution boundary;
7. approval consumption is atomic and one-time across concurrent requests;
8. unknown/invalid JWT key, issuer, audience, expiry, and algorithm fail closed;
9. production router rejects demo-only endpoints;
10. readiness becomes unhealthy on database/schema/key dependency failure;
11. backup restore test and Railway rollback procedure are documented.

**Decision summary:** production Control Plane state belongs in a dedicated Supabase PostgreSQL schema behind a least-privilege runtime role. The demo remains separate. Persistence, identity, and authorization failures are fail-closed.
