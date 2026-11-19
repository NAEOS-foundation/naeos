# P2 Runtime Integration Contract — Supabase Control Plane

- Status: implementation contract; source baseline updated 2026-11-19
- Scope: production Control Plane runtime integration, originally tracked under Issue #786
- Related ADR: [Production Control Plane Runtime and Supabase Persistence](../architecture/ADR-PRODUCTION-CONTROL-PLANE-SUPABASE-PERSISTENCE.md)

## Verified repository baseline (main)

The production Control Plane runtime implementation is present on `main`:

- `naeos control-plane preflight` performs read-only database target, TLS, connectivity, runtime-role, private-schema, and required-function checks.
- `naeos control-plane serve` opens the environment-scoped PostgreSQL runtime, verifies the ledger, configures JWT/JWKS identity, and then starts the dedicated HTTP server.
- The runtime uses `NAEOS_CONTROLPLANE_DATABASE_URL` and `NAEOS_CONTROLPLANE_ENVIRONMENT`; it does not silently fall back to generic `DATABASE_URL`, in-memory state, or file snapshots.
- Database target validation binds production to the approved CRM PostgreSQL Supabase project and staging/preview to the approved `naeos.db` Supabase project.
- Runtime startup requires TLS, a dedicated login role with constrained membership, the private `naeos_control_plane` schema and required tables/functions, and a valid ledger integrity check.
- Production health routes separate liveness and readiness. The dedicated production handler does not mount investor-demo reset, scenario, or execution routes.
- Identity configuration requires issuer, audience, and JWKS URL; the identity validator is initialized before the listener is bound.

These are source-level implementation findings. They do **not** prove that production or staging credentials have been configured, that the preflight has run against either live database, that migrations have been applied, or that a deployment/cutover has occurred.

## Compatibility and safety requirements

- Keep the investor-demo composition root and its route surface separate from the production runtime.
- Runtime startup and authorization must fail closed on target mismatch, missing schema/trust configuration, database failures, invalid identity, or ledger integrity failure.
- Do not use a generic database variable, in-memory store, or JSON file as a production fallback.
- Keep production CRM database access behind the dedicated least-privilege Control Plane role and private schema; staging/preview uses separate credentials and the `naeos.db` target.
- Do not apply migrations, edit environment secrets, deploy, or perform production cutover without a separately reviewed and authorized operational step.

## Implementation and verification status

| Workstream | Source status | Remaining evidence / gate |
|---|---|---|
| Environment-scoped DB contract and target validation | Implemented | Live preflight per environment |
| PostgreSQL repositories and ledger adapter | Present | Staging migration, integration and recovery evidence |
| Dedicated production composition root and route allowlist | Implemented | Live deployment verification |
| Startup DB/schema/role/ledger checks | Implemented | Live preflight evidence |
| JWT/JWKS identity configuration | Implemented | Confirm actual issuer/audience/key configuration and live token tests |
| Liveness/readiness | Implemented | Runtime monitoring and dependency failure observation |
| Staging validation | Not evidenced as complete | Confirm target, review/apply migration, persistence/restart/concurrency/privilege tests |
| Production cutover | Not authorized / not evidenced | Backup and restore proof, schema/role review, explicit approval, deployment and rollback receipts |

## Required tests and evidence

- Environment mismatch, invalid database URL, and TLS requirement fail closed.
- Runtime role is dedicated, non-superuser, non-BYPASSRLS, and has no additional role memberships.
- Required private schema tables/functions are present before startup.
- Ledger integrity is verified before binding the listener; persistence failure cannot become an ALLOW.
- JWT negative cases reject wrong issuer/audience, expired/not-before tokens, invalid signatures, unsupported algorithms, and unknown keys.
- Production router allowlist excludes demo reset, scenario, execution, and other demo-only routes.
- Readiness fails when database, schema, role, or identity dependencies are unavailable; liveness remains independent.
- Staging integration tests cover migration, restart recovery, concurrent ledger append, least-privilege grants, and fail-closed behavior.

## Rollout boundaries

This document records repository implementation progress only. It does not authorize or claim:
- a staging migration or live staging preflight;
- a production database migration, role/credential change, or cutover;
- a Railway environment-variable change or deployment;
- live production health or durable-state proof.

Maintain the ADR release gates and never bypass security or governance checks to obtain a green status.

Operational evidence must be recorded with environment identity, timestamp, command, outcome, and sanitized output.
