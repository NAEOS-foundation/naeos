# Unified Supabase Integration — Implementation Plan

**Status:** Staged implementation plan; Production changes not authorized  
**Date:** 2026-11-15  
**Parent ADR:** ADR-UNIFIED-SUPABASE-IDENTITY-TENANT.md

## Execution status (2026-11-16)

Status is reported by PR artifact and check run, not inferred from file existence.

- **Phase 0 — Baseline and authority verification:** Open. Live JWT claims/signing key, audience, trusted capability mapping, and isolated staging owner are not yet evidenced.
- **Phase 1 — CRM identity mapping:** Implemented in draft [NAEOS CRM PR #44](https://github.com/NAEOS-foundation/naeos-crm/pull/44). The current PR head CI is green; migration and runtime configuration are not applied.
- **Phase 2 — Shared tenant foundation:** Draft [NAEOS Core PR #816](https://github.com/NAEOS-foundation/naeos/pull/816) contains the private-schema foundation migration. Identity read access and identity-admin writes are separated into NOLOGIN/NOBYPASSRLS groups; `role_capabilities` binds scoped roles to server-controlled capability keys. The dedicated ephemeral PostgreSQL 17.6 integration job previously passed migration execution, runtime/admin privilege separation, RLS enablement, cross-org/workspace FK rejection, duplicate membership rejection, role-scope guard, and audit append-only trigger/privilege tests. The current PR update hardens runtime RLS with active membership-bound session context, cross-tenant read-denial assertions, forged audit actor denial, and INHERIT role-attribute validation; its CI verification is pending. Supabase Preview/Integration was skipped, so this is not a Supabase staging run and no migration has been applied to a Supabase project.
- **Tenant authorization contract:** Draft [NAEOS CRM PR #45](https://github.com/NAEOS-foundation/naeos-crm/pull/45) adds fail-closed resolver/policy interfaces, tests, and CRM ownership map. CI is in progress. No API routes are wired to this contract yet.
- **Phase 4 — Intelligence identity:** Draft [NAEOS Intelligence PR #72](https://github.com/NAEOS-foundation/naeos-intelligence/pull/72) separates JWT issuer/subject/client_id and canonical owner logic; CI is green. Existing records are not reassigned.
- **Phase 3 and Phase 5:** Not started; blocked on tenant product/ownership decision, verified membership resolver execution path, staged migrations, and integration evidence.

No Production schema, auth configuration, or data has been changed.


## Implementation sequence

### Phase 0 — Baseline and authority verification
- [ ] Confirm selected Auth project (naeos-crm-production) and project access boundary.
- [ ] Inspect live signing key configuration and JWKS; confirm RS256 key/kid.
- [ ] Obtain a real access token locally and validate iss, aud, sub, exp, nbf; do not store or share token.
- [ ] Decide trusted NAEOS role/capability mapping; authenticated is not a product permission.
- [ ] Confirm staging projects/environments and migration owner.

### Phase 1 — CRM identity mapping (backward-compatible)
- [ ] Add nullable unique Supabase Auth UUID to CRM User.
- [ ] Keep CUID primary key, password hash, local HS256 session and existing role field unchanged.
- [ ] Add mapping/linking flow only after proof of account ownership; email match alone must not silently link accounts.
- [ ] Add tests for duplicate provider UUID, unlinked legacy account, inactive user, account-link takeover and existing login.
- [ ] Run Prisma validation, generated client checks, migration review and CRM CI.

### Phase 2 — Shared tenant foundation
- [ ] Create organizations, workspaces, memberships, role definitions and scoped role assignments in isolated staging.
- [ ] Enforce workspace-to-organization consistency.
- [ ] Design explicit RLS policies and grants per table; avoid permissive policies.
- [ ] Set runtime session context (`naeos.organization_id`, `naeos.workspace_id`, `naeos.issuer`, `naeos.subject`) only from a verified token and trusted membership resolver; never expose the database role or a client-controlled context setter.
- [ ] Define service contract for trusted membership/permission lookup.
- [ ] Test cross-tenant denial and member lifecycle.
- [ ] Verify forged audit actor/outcome fields are denied and runtime audit rows match active membership.
- [ ] Require `naeos.identity.audit.read` for runtime audit visibility; verify ordinary active members cannot read audit events and organization/workspace administrators can read only their authorized scope.

### Phase 3 — CRM tenant-aware persistence and API
- [ ] Agree initial tenant ownership/backfill semantics.
- [ ] Add tenant ownership to CRM resources in small migration groups.
- [ ] Enforce tenant context server-side for every sensitive API route.
- [ ] Add tests covering tenant escape via IDs, nested relations, bulk actions and GO-Gate operations.
- [ ] Verify no existing record is assigned a tenant by inference.

### Phase 4 — Core and Intelligence identity propagation
- [ ] Core: resolve issuer-qualified subject to trusted actor/membership; keep current role registry and authorization gates.
- [ ] Intelligence: separate human sub from OAuth client_id; revise principal and storage ownership model with tests.
- [ ] Propagate organization/workspace IDs only from trusted server-side authorization resolution.
- [ ] Correlate audit events using request/trace ID and canonical actor identifier.
- [ ] Keep Control Plane persistence service role and private schema separate from browser access.

### Phase 5 — Staging rollout and production gate
- [ ] Apply versioned migrations only to isolated staging.
- [ ] Run identity, RLS, API authorization, tenant isolation, session and integration suites.
- [ ] Rehearse migration, backup restoration, and rollback.
- [ ] Review observability, secrets, and runtime connection roles.
- [ ] Produce evidence-based rollout checklist and obtain explicit production approval.

## Definition of done for each phase

- Migration is versioned and repeatable.
- Tests cover positive and negative security cases.
- No auth/session compatibility regression.
- No credentials or user tokens in repository or logs.
- Changes reviewed in PR; required checks pass without bypass.
- Scope and deployment status are reported accurately.

## Known blockers

- Live Supabase JWT claims and active signing key must be verified.
- CRM legacy CUID user IDs must remain stable during initial mapping.
- Tenant backfill requires an explicit ownership decision.
- Intelligence scope claims and canonical principal semantics need implementation and test evidence.
- Production database modifications are explicitly gated.
