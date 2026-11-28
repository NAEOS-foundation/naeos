# IAM-04/05 — Principal identity and tenant/client contract

**Status:** CONTRACT DEFINED; application-level membership resolver and staging evidence remain BLOCKED  
**Scope:** Supabase Auth human identity and NAEOS Organization/Workspace tenant scope. This is not approval for Production rollout.

## Decision basis (repository evidence)

The proposed identity/tenant ADR defines:
- Canonical human principal as exact pair `(issuer, subject)`; JWT `client_id` is a distinct client identity and must not replace human `sub`.
- Provider user UUID is an external identity link; existing CRM CUID primary keys and local login/session behavior remain unchanged in the compatibility phase.
- Tenant hierarchy is Organization → Workspace → Membership. Membership is organization-scoped when `workspace_id IS NULL`, otherwise scoped to a workspace nested in the same organization.
- Authentication alone does not grant tenant access. Server-side membership, scoped role/capability and resource ownership checks are required. User-editable `user_metadata` is not authorization authority.

The SQL foundation models membership as `(organization_id, workspace_id, issuer, user_id)`, with `user_id UUID`, uniqueness using `NULLS NOT DISTINCT`, composite FK enforcing workspace/organization consistency, and immutable tenant/principal scope. Role-assignment triggers enforce role scope matching membership scope. RLS policies require actor and active membership context; the ADR explicitly cautions that GUC values must be set only by trusted backend code after token verification and membership resolution.

## IAM-04 — Principal definition

- Human principal key: exact issuer-qualified subject pair `(iss, sub)`; issuer normalization or subject rewriting is not permitted.
- Supabase user UUID (`sub`) maps to the schema's `memberships.user_id`; this is appropriate only for the selected Supabase Auth issuer and UUID subject contract.
- `client_id` is retained separately for client/application identification. It is not a human user, membership key, or authorization grant.
- Legacy client-only tokens may retain compatibility in service-token flows, but must not be treated as Supabase human principals or be used to resolve human memberships.
- CRM legacy CUID remains stable and links additively through the nullable unique provider UUID mapping.

## IAM-05 — Tenant and client boundary

- Tenant context is a trusted server-side resolution result, not accepted as authority from URL, request body, headers, JWT user metadata, or client-supplied GUCs.
- Organization and workspace identifiers are selectors only; the backend must confirm active membership for the verified `(iss, sub)` and validate workspace belongs to the selected organization.
- Organization-level membership (`workspace_id IS NULL`) and workspace membership are distinct scopes; role definitions and assignments must match the membership scope.
- Active status of organization/workspace and membership is required. Suspended, invited, or revoked memberships fail closed.
- A client/app identity (`client_id`) may authenticate a service-to-service caller only under a separately approved client principal, capability, and ownership policy. The current human membership schema does not define such policy; do not infer it.
- Resource ownership and nested/bulk/approval operations require independent server-side tenant checks; opaque resource IDs do not confer access.

## Required implementation/verification before marking PASS

- [ ] Implement a trusted membership resolver in the backend that accepts only verified issuer and subject claims, then resolves active organization/workspace membership and scoped roles/capabilities.
- [ ] Ensure all route handlers propagate tenant IDs only from this resolver; reject request-controlled tenant scope absent authorized selection.
- [ ] Negative tests: issuer mismatch; subject not member; inactive membership; suspended organization/workspace; workspace from another organization; organization/workspace role-scope mismatch; client_id attempting human membership resolution; forged tenant IDs; cross-tenant resource/nested/bulk/action access.
- [ ] Positive tests: org membership, workspace membership, properly scoped role and capability.
- [ ] Confirm actual Supabase subject is UUID for the selected issuer using sanitized staging evidence; no raw token or identifying subject is stored in evidence.
- [ ] Validate resolver and database policy against isolated staging; record commit, environment, timestamp, test results and reviewer.
- [ ] Do not mark PASS solely on schema/unit tests. No Production schema, Auth configuration, tenant backfill, or cutover is authorized by this document.

## Current disposition

The ADR and SQL foundation provide an explicit contract for IAM-04/05, but codebase search/review does not establish that a trusted runtime membership resolver is implemented and wired through all Core/CRM/Intelligence routes. Therefore IAM-04/05 are **not PASS**: contract is defined, enforcement and staging verification remain outstanding.