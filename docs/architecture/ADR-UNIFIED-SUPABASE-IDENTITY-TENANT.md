# ADR — Unified Supabase Identity and NAEOS Tenant Foundation

- **Status:** Proposed
- **Date:** 2026-11-15
- **Decision owners:** NAEOS maintainers
- **Scope:** NAEOS CRM, Core/Control Plane, and Intelligence identity and tenant contracts
- **Related:** Intelligence ADR-005 (Supabase Auth hosted IdP), Core Supabase Auth integration, Control Plane persistence ADR

## Context

NAEOS services currently have different authentication and authorization models:

- CRM production authentication currently issues a local HS256 session from its own password hash and Prisma User record.
- NAEOS Intelligence accepted Supabase Auth as hosted IdP and selected the naeos-crm-production project (oruykkehukoiipldzcvc) as issuer, but exact live JWT issuer/audience/signing-key compatibility and authorization claim mapping remain gated.
- NAEOS Core validates Supabase JWTs with a JWKS verifier and maps trusted app_metadata.naeos_roles; this is not a universal tenant-membership contract.
- CRM, Core/Control Plane, and Intelligence persistence are separate projects. Cross-project JWT authentication does not grant direct PostgREST/database access.
- CRM users use CUID primary keys and local password/session behavior. Changing those identifiers or removing local authentication in-place risks existing login and ownership relations.

A single physical database is not required for a unified platform. Domain ownership and separate data stores should remain until an explicit migration decision and data-integrity proof authorize consolidation.

## Decision

1. Hosted identity provider remains Supabase Auth in naeos-crm-production, following Intelligence ADR-005. Do not activate service issuer/JWKS configuration until real token claims, live JWKS key, audience, and trusted capability mapping are verified.
2. Canonical human principal is the pair (issuer, subject). Do not conflate JWT client_id with a human subject. Store provider user UUID as an external identity link; do not rewrite CRM CUID primary keys in the first phase.
3. CRM identity linking is additive and nullable. A Supabase Auth UUID mapping is unique when present. Existing CRM password hashes, sessions, User IDs, and ownership relations remain intact during migration and rollback window.
4. Tenant hierarchy is Organization → Workspace → Membership. A membership is scoped either to an organization (`workspace_id IS NULL`) or to one workspace within that organization. Tenant/resource ownership must be checked server-side.
5. Authentication and authorization are separate. JWT proves identity; membership, role assignment, policy and resource ownership determine access. User-editable user_metadata is never authorization authority.
6. RBAC is supplemented by policy checks. Platform, organization, and workspace roles are distinct scopes. Roles and capabilities are server-controlled and least-privilege.
7. Domain persistence remains federated. CRM owns CRM records; Core owns Control Plane records; Intelligence owns its request/execution/evidence records. Cross-project references use canonical UUIDs and contract validation, not cross-project foreign keys.
8. Audit events use a shared correlation contract: issuer-qualified actor, organization/workspace IDs where applicable, action, resource, request/trace IDs, outcome, timestamp, and schema version. Each domain remains authoritative for its own audit storage.
9. RLS and API authorization are both required. The proposed private identity schema separates a read-only `naeos_identity_runtime` group from write-capable `naeos_identity_admin`; both are NOLOGIN, INHERIT, and NOBYPASSRLS. Runtime row policies scope organizations, workspaces, memberships, role assignments/capabilities, and audit reads using transaction-local context settings (`naeos.organization_id`, `naeos.workspace_id`, `naeos.issuer`, `naeos.subject`) plus active membership checks. A database role able to execute arbitrary SQL can set custom GUCs, so these values are not an independent authentication boundary: a trusted backend must derive them only after JWT verification and trusted membership resolution, set them transaction-locally (`set_config(..., true)`), and never expose the runtime database credential or a client-controlled SQL/GUC setter. RLS additionally denies access when the organization or selected workspace is not active; the narrow `is_active_tenant` SECURITY DEFINER helper returns only this lifecycle predicate and is executable only by the backend identity roles. Runtime audit inserts must match the context actor and active membership. Role capability mappings are stored in `role_capabilities`. Runtime audit reads require the explicit `naeos.identity.audit.read` capability, granted by default only to the built-in `organization_admin` and `workspace_admin` roles; active membership alone does not authorize audit visibility. Distinct LOGIN roles must be provisioned with secret-managed credentials and granted the appropriate group; Supabase `service_role` and browser roles receive no schema access. Backend routes independently enforce membership and resource ownership.
10. Migration is staged and backward-compatible. No production schema/auth configuration changes are authorized by this ADR. Production cutover requires staging tests, backup proof, rollback rehearsal, and explicit operator approval.

## Tenant contract

- organization_id: canonical UUID identifying the customer/tenant.
- workspace_id: canonical UUID nested within an organization.
- user_id: Supabase Auth auth.users.id UUID at the identity-provider boundary.
- membership_id: canonical UUID for the user's membership, scoped to either an organization (`workspace_id IS NULL`) or one workspace.
- CRM legacy user key: existing CUID remains stable; it maps to canonical identity through a nullable unique provider UUID.
- Cross-service resource identifiers are opaque and are not permission by themselves.

A workspace's organization relationship must be enforced by a composite constraint or equivalent server-side validation. It must not be possible to create a membership with mismatched organization/workspace IDs.

## Migration stages

1. Add nullable unique external Supabase identity mapping to CRM User, preserving local auth.
2. Verify actual Supabase Auth tokens and implement a feature-flagged verifier/session bridge with negative tests.
3. Define and migrate Organization, Workspace, Membership, and scoped role assignments in isolated staging.
4. Backfill existing CRM data to a deliberately chosen initial organization only after owner and access semantics are agreed and tested; no guessed backfill.
5. Apply tenant ownership columns to CRM entities and enforce tenant isolation in API and RLS, with cross-tenant regression tests.
6. Integrate Core and Intelligence using issuer-qualified principals and trusted server-side capability/membership resolution.
7. Run cutover rehearsal, verify rollback and backup restoration, then request explicit production approval.

## Security and acceptance gates

- Real access token confirms alg=RS256, matching kid in JWKS, exact iss, expected aud, valid sub, and time claims.
- Token verifier rejects HS256, malformed tokens, wrong issuer/audience, expiry, unknown key, JWKS outage, and missing subject.
- Capability/role mapping uses server-controlled claims or trusted lookup; authenticated role and user metadata do not imply NAEOS access.
- Cross-tenant read/write/approve operations are denied and tested across CRM, Core, and Intelligence.
- Runtime database session context is set only from a verified token and trusted membership resolution; client-controlled organization/workspace settings are rejected at the application boundary.
- Existing CRM local login and sessions remain functional throughout compatibility phase.
- No secret key, service-role key, database password, or raw access token is committed or logged.
- Migration is repeatable in staging; backup and rollback are demonstrated before production.

## Alternatives considered

- Merge all domain databases immediately: rejected; it increases coupling and blast radius without being necessary for unified identity.
- Replace CRM CUIDs with Supabase UUIDs immediately: rejected; it risks breaking ownership relations, sessions, and existing references.
- Use JWT role claim as complete tenant authorization: rejected; roles do not prove organization/workspace membership or resource ownership.
- Trust user-editable metadata for roles/tenant: rejected; users can alter it.
- Keep independent identity providers indefinitely: rejected as the target because it creates fragmented user lifecycle and inconsistent authorization semantics; temporary backward compatibility is permitted.

## Consequences

Identity consolidation is incremental and compatible with current CRM data. NAEOS services share a canonical human principal and tenant contract while keeping domain persistence boundaries. Membership and authorization lookup APIs are required before enabling multi-tenant access. The SQL foundation has an ephemeral PostgreSQL integration test; Supabase staging execution and secure service connectivity remain unverified. This ADR is proposed and does not claim tenant isolation or unified login is already implemented.
