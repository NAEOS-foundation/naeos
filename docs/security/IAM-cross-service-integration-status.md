# IAM cross-service integration status

**Review date:** 2026-11-17  
**Scope:** NAEOS Core PR #817, CRM PR #45, Intelligence PR #72  
**Disposition:** Contract and identity changes are validated by repository CI; tenant enforcement and production rollout remain blocked.

## Evidence snapshot

| Repository / PR | Head commit | PR state | CI evidence | Review state |
|---|---|---|---|---|
| `NAEOS-foundation/naeos#817` | See PR head | Open, Draft | Cross-service CI not asserted by this document | No approval inferred |
| `NAEOS-foundation/naeos-crm#45` | `93c17e2bd00b249a58e82e62466071b211eb6ea7` | Open, Draft | CI run `43754623270`, completed/success; build, dependency audit, Prisma generation, test DB migrations, seed, workspace build and tests succeeded | No submitted reviews or review threads found |
| `NAEOS-foundation/naeos-intelligence#72` | `93d38d17e2ad11fac5018964e4d6043716f0380c` | Open, Draft | CI run `43739318238`, completed/success; module checks, govulncheck, secret scan, vet, tests and build succeeded | No submitted reviews or review threads found |

CI success is evidence for the tested branch state only. It is not a production deployment or proof of cross-tenant isolation.

## CRM PR #45 — delivered vs outstanding

Delivered:
- Canonical issuer/subject principal and exact organization/workspace scope types.
- Fail-closed authorization orchestration with membership, tenant lifecycle and capability checks.
- Unit/API tests for absent/mismatched/inactive membership, tenant lifecycle, denied capability, dependency outage and malformed input.
- Proposed ownership map and rollout gates.

Outstanding:
- No concrete trusted `TenantMembershipResolver` adapter is connected to the private `naeos_identity` schema.
- No API route wiring is enabled; PR description explicitly confirms contract-only scope.
- CRM tenant ownership columns, scoped repository predicates, composite constraints, and two-tenant CRUD/nested-route tests are not delivered in this PR.
- The CRM ownership map records unresolved product decisions: whether CRM is internal NAEOS tenancy, customer-facing multi-tenancy, or both; initial tenant mapping; global-vs-tenant catalog semantics.
- The proposed runtime login role still needs an approved, deployable execution path with least-privilege credentials.

## Intelligence PR #72 — delivered vs outstanding

Delivered:
- Verified issuer, subject and client_id are represented separately.
- Human ownership is issuer-qualified subject; client-only tokens retain legacy client ownership.
- Missing issuer for a human subject fails closed.
- Request storage ownership tests prove human subject does not collapse into client_id.
- Go baseline raised to 1.26.11 and CI security checks passed.

Outstanding:
- This change partitions Intelligence request ownership; it does not resolve NAEOS Organization/Workspace membership.
- No tenant capability resolver or organization/workspace resource authorization is established by this PR.
- Existing client-owned records are not reassigned; a separately approved ownership migration is required if migration is desired.
- The identity issuer/audience/JWKS runtime trust configuration still requires staging evidence under IAM-03.

## Cross-service decision

Do not mark IAM-04/05 PASS merely because CRM orchestration and Intelligence ownership tests pass. The enforcement chain is incomplete until:
1. CRM's product tenancy and initial tenant mapping are approved.
2. A trusted membership resolver uses verified `(iss, sub)` and the dedicated least-privilege runtime DB path.
3. CRM routes and repository/data ownership are tenant-scoped end to end.
4. Core, CRM and Intelligence enforce a consistent tenant/capability contract for each resource and operation.
5. Staging tests demonstrate two-tenant isolation, membership revocation, inactive-tenant denial, resource ownership and dependency-outage fail-closed behavior.
6. IAM-03 live staging trust evidence, rollback rehearsal and security review are recorded.

No Production schema, Auth configuration, data backfill, or route cutover is authorized by this status record.