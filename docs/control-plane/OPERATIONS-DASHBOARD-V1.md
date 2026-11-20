# NAEOS Control Plane Operations Dashboard v1

- Status: implementation blueprint
- Date: 2026-11-19
- Scope: internal-first operational console
- Repository authority: NAEOS core
- Related architecture: [Production Control Plane Runtime and Supabase Persistence](../architecture/ADR-PRODUCTION-CONTROL-PLANE-SUPABASE-PERSISTENCE.md), [Control Plane Persistence Contracts](../architecture/CONTROL-PLANE-PERSISTENCE-CONTRACTS.md)

## 1. Purpose

Build an authenticated internal operations console for NAEOS systems, environments, Control Plane persistence, policy decisions, evidence, and runtime health. This is distinct from the public `/control-plane` product/investor demonstration page.

The first release is read-only operational visibility. Administrative mutations, deployment orchestration, rollback, policy editing, and database operations are explicitly out of scope for v1.

## 2. Repository reconnaissance (2026-11-19)

The following are confirmed from the current default branches:

### NAEOS core (`NAEOS-foundation/naeos`)

- `internal/controlplane` contains the reusable decision gateway and ledger domain.
- PostgreSQL persistence adapters exist for policy, grants, approvals, and ledger.
- `supabase/migrations/20261111000100_production_control_plane.sql` defines the private `naeos_control_plane` schema, restricted runtime role, immutable policy snapshots, approval state, and append-only ledger.
- `internal/controlplane/production_runtime.go` validates the environment-specific database target, TLS URL, runtime role, schema compatibility, and persisted ledger before serving.
- `cmd/naeos/controlplane_cmd.go` exposes `control-plane preflight` and `control-plane serve`.
- `internal/controlplane/production_auth.go` currently mounts liveness, readiness, and one authenticated `POST /api/control-plane/decision` route. It does not expose the read APIs required by an operations dashboard.
- `site/components/control-plane/ControlPlaneDashboard.tsx` is a public demonstration component. It contains sample decision rows and public-facing product messaging; it is not an authenticated operations console and must not be repurposed to display private operational data.
- The current production runtime is deliberately separate from the investor-demo composition root.

### NAEOS CRM (`NAEOS-foundation/naeos-crm`)

- `apps/web` is an authenticated React/Vite application with a login flow, sidebar navigation, and existing pages for GO-Gate, Policy, Audit, and Dashboard.
- `apps/api` enforces authentication and server-side RBAC. Existing roles are `admin`, `manager`, and `member`.
- The CRM UI currently calls its own API using cookie credentials. The Supabase access token is held privately inside `AuthProvider`; it is not exposed as a general-purpose Control Plane credential.
- CRM's existing `policy` and `audit` permissions govern CRM resources and must not be assumed to grant access to the independent Control Plane persistence boundary.

### Environment and data boundary

The accepted ADR and its 2026-11-18 amendment define:
- Production Control Plane: CRM PostgreSQL database, private `naeos_control_plane` schema, dedicated Control Plane LOGIN role.
- Staging/preview Control Plane: Supabase project/database `naeos.db`, separate runtime role and credentials.
- Local: Docker/local PostgreSQL with explicit local environment configuration.
- No fallback between environments; no CRM application-role reuse; no production migration or cutover from this dashboard work.

## 3. Architecture decision

Use the existing authenticated NAEOS CRM web shell as the internal operator UI, but keep the Control Plane as an independent backend and authorization domain.

Request path:

```
Operator browser
  -> CRM web application (authenticated session)
  -> CRM API (server-side session validation + dedicated cp:read permission)
  -> Control Plane API (short-lived service identity / verified operator context)
  -> Control Plane repositories
  -> private naeos_control_plane schema
```

Requirements:

1. Browser code must never connect directly to PostgreSQL/Supabase or receive a database URL, service-role key, Control Plane runtime credential, or static API token.
2. The CRM API acts as a backend-for-frontend (BFF). It validates the current operator session and the dedicated Control Plane permission before proxying read requests.
3. The Control Plane API independently authenticates the BFF/service identity and authorizes the requested read scope. CRM UI visibility alone is not authorization.
4. Do not forward an arbitrary browser-supplied Authorization header or allow the browser to select an upstream URL.
5. Keep Control Plane database repositories and schema isolated from CRM domain repositories and ORM models.
6. Do not use the public website's `NEXT_PUBLIC_*` variables for secrets or privileged endpoints.
7. Environment selection is server-side and explicit. The browser may select a view only among environments the authenticated principal is authorized to inspect; it cannot choose database URLs or credentials.

If the current identity integration cannot securely represent an operator and a service-to-service caller, implement a dedicated short-lived service assertion with audience, issuer, expiry, subject, actor, environment scope, and read-only capability. Do not solve this with a shared long-lived browser token.

## 4. v1 navigation and page scope

Add a separate **Control Plane** navigation group in the authenticated CRM shell, visible only to principals with `control-plane:read`.

### Overview

- Environment selector: Local / Staging / Production, limited by server authorization.
- System health summary: Healthy / Degraded / Unhealthy / Unknown.
- Dependency status: API, PostgreSQL connectivity, schema compatibility, identity/JWKS readiness.
- Latest verified ledger sequence and chain verification timestamp.
- Recent policy decisions: ALLOW / DENY / PENDING / INDETERMINATE.
- Recent execution lifecycle events, when present.
- Recent security/governance events.
- Last successful backup evidence, or Unknown if no trusted evidence source is connected.
- Data freshness timestamp and request correlation ID.

### Systems & Environments

- Registered systems and service identifiers.
- Environment-to-database target mapping by safe logical identifier only.
- Runtime version, commit SHA, observed-at timestamp, health state.
- Never render hostname, database name, username, connection string, secret, or internal provider credentials.

### Decisions & Evidence

- Paginated, filterable decisions and ledger events.
- Decision ID, principal/agent, capability, policy ID/version, result, reason code, timestamp.
- Evidence verification state and ledger sequence.
- Link to a read-only evidence detail view.
- Distinguish missing evidence from unavailable persistence and failed verification.

### Security & Audit

- Authentication and authorization events relevant to Control Plane operations.
- Policy/grant/approval changes and actor provenance.
- Read-only audit export metadata; no mutation or deletion actions in v1.

## 5. Backend read API contract

The core production router needs explicit read-only endpoints. Proposed initial surface:

| Endpoint | Purpose | Access |
|---|---|---|
| `GET /api/control-plane/overview` | Aggregated environment health and counts | `control-plane:read` |
| `GET /api/control-plane/systems` | System/service registry | `control-plane:read` |
| `GET /api/control-plane/environments` | Authorized environment metadata | `control-plane:read` |
| `GET /api/control-plane/decisions` | Paginated decision ledger view | `control-plane:read` |
| `GET /api/control-plane/decisions/{id}` | Canonical decision and evidence summary | `control-plane:read` |
| `GET /api/control-plane/audit-events` | Security-relevant Control Plane audit events | `control-plane:audit:read` |

Contract requirements:

- Strict query validation, bounded page size, stable ordering, and cursor pagination.
- Every response includes `observed_at`, `environment`, and `request_id`.
- Return `Unknown`/unavailable when a source is missing; never convert an upstream failure into an empty successful result.
- No generic SQL, arbitrary table names, arbitrary URL proxying, or unrestricted query parameters.
- Read APIs use read-only repository methods and a least-privilege database role where feasible.
- Health/readiness probes remain unauthenticated only if they reveal no sensitive dependency detail.
- Existing `POST /api/control-plane/decision` behavior and trust boundary remain unchanged.

## 6. Data presentation rules

The UI must not imply live health from mock or cached sample data.

- `Healthy`: fresh positive probe with all required checks passing.
- `Degraded`: service responds, but one or more non-critical checks fail.
- `Unhealthy`: a required dependency or integrity check fails.
- `Unknown`: no trustworthy recent observation is available.
- `Stale`: data was previously observed but exceeds its freshness threshold.

Each status includes its source and last-observed time. No hard-coded example counts in the production console. Sample data may appear only in storybook/test fixtures with a visible fixture label.

## 7. Authorization model

Introduce Control Plane-specific permissions, separate from CRM permissions:

- `control-plane:read` — overview, systems, environments, decisions.
- `control-plane:audit:read` — security and audit event details.
- Future, not enabled in v1: `control-plane:policy:write`, `control-plane:approval:write`, `control-plane:deployment:approve`, `control-plane:operations:execute`.

Initial assignment: Owner/admin only, after explicit mapping and server-side enforcement. Do not automatically grant Control Plane access to every CRM admin, manager, or member without a reviewed policy mapping. v1 must not expose mutation controls.

## 8. Delivery phases

### CPD-1 — API read contracts and authorization

- Define typed read models and repository query interfaces.
- Add Control Plane-specific authorization scope.
- Add read-only handlers, pagination, error semantics, and audit of operator reads where required.
- Add tests for unauthorized, wrong environment, malformed cursor, oversized page, unavailable DB, tampered ledger, and empty-but-valid datasets.

### CPD-2 — CRM BFF integration

- Add CRM API proxy endpoints with server-side session validation.
- Bind service identity to operator identity, environment, audience, expiry, and read-only scope.
- Add a dedicated `control-plane:read` permission and admin-only initial policy mapping.
- Never expose upstream credentials to browser code.

### CPD-3 — Internal UI

- Add the Control Plane navigation group and protected routes to `naeos-crm/apps/web`.
- Build Overview, Systems & Environments, Decisions & Evidence, and Security & Audit pages.
- Use explicit loading, error, stale, empty, and unknown states.
- Add responsive layouts and accessible tables/filters.

### CPD-4 — Local and staging verification

- Run the Control Plane API and CRM stack with Docker/local dependencies.
- Use only the staging `naeos.db` project for remote staging persistence.
- Verify role isolation, environment mismatch rejection, no browser secrets, no CRM table access from Control Plane repositories, and audit provenance.
- No production migration, production credential change, or production deployment.

### CPD-5 — Production readiness review (not authorized by this plan)

Requires a separate change approval, backup/restore evidence, dedicated production role verification, security review, deployment receipt, and explicit production cutover authorization.

## 9. Acceptance criteria for v1

- An unauthenticated browser cannot access the console or its data.
- A CRM principal without `control-plane:read` is denied by the CRM API and Control Plane API.
- The browser receives no Control Plane database credential or privileged static token.
- Environment identity is derived and checked server-side; staging cannot read production and vice versa.
- Dashboard statuses are backed by timestamped observations and show Unknown/Stale when appropriate.
- Decision/evidence reads preserve canonical ledger ordering and verification state.
- Database, identity, or ledger integrity errors are surfaced as errors, never empty success.
- All tests pass; no security gate is bypassed.
- No production database migration, CRM data mutation, or production deployment occurs as part of v1 implementation.

## 10. Immediate implementation order

1. Implement CPD-1 in `NAEOS-foundation/naeos` and add contract/security tests.
2. Implement CPD-2 and CPD-3 in `NAEOS-foundation/naeos-crm`.
3. Validate end-to-end in local Docker, then isolated staging.
4. Conduct a security and operational readiness review before considering production access.
