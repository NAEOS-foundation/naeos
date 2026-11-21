# NAEOS Master Development Plan v1.0
**Audit date:** 2026-11-21  
**Scope:** NAEOS Core, CRM, Intelligence, Control Plane, staging, security and pilot readiness.  
**Status:** Phase 0 inventory established; live CI run conclusions are intentionally limited to verified metadata and PR descriptions.

## Executive decision
Follow the gated sequence **Stabilize → Secure → Integrate → Verify → Pilot → Early Access**. Avoid new feature expansion until critical identity, execution, audit, staging and E2E gates are closed. No production deployment is authorized by this plan.

## Phase 0 goals
1. Inventory open PRs across Core, CRM and Intelligence.
2. Categorize by critical-path dependencies, blockers, mergeability and scope.
3. Establish a canonical milestone/task sequence and release gates.
4. Identify CI or validation evidence gaps.
5. Preserve security gates; do not bypass required checks.

## Repository snapshot
### NAEOS Core — `NAEOS-foundation/naeos`
Open PRs in priority focus:
- #830 — production API identity contract (draft; mergeability currently false)
- #831 — preserve JWT subject/client identity separately (draft; mergeable)
- #832 — persisted principal-agent binding resolver (draft; mergeability currently false)
- #841 — align P2 contract with runtime (draft; mergeability currently false)
- #842 — Docker local staging runtime (draft; mergeable)
- #851 — brand governance stabilization (draft; mergeable)
- #852 — asset validation tooling (draft; mergeable)
- #853 — Social Media asset reference audit (draft; mergeable)
Other open PRs include #823, #820, #818, #817, #815, #794, #696, #692, #676, #675, #670 and older work. These require a dedicated stale/dependency review before disposition.

### NAEOS CRM — `NAEOS-foundation/naeos-crm`
- #69 — database-backed integration test isolation (draft; mergeable)
- #68 — atomic audit prototype for Policy Rule updates (draft; mergeable)
- #67 — audit durability findings/remediation plan (draft; mergeable)
- #66 — company deletion safeguard (draft; mergeable)
- #65 — company access and relationship validation (draft; mergeable)
- #64 — Go-Gate requester self-rejection safeguard (draft; mergeable)
- #63 — CRM read-only Control Plane dashboard (draft; mergeable)
- #62 — audit hash metadata v2 (draft; mergeable)
- #60 — legacy audit chain backfill regression reproduction (draft; mergeable)
- #37 — Prisma/outbox lifecycle stabilization (open; older; requires status/dependency review)

### NAEOS Intelligence — `NAEOS-foundation/naeos-intelligence`
- #78 — Supabase ES256 JWT verification (draft; mergeable)
- #77 — isolated partner deployment baseline (non-draft; mergeable)
- #74 — Intelligence DB runtime role matrix (draft; mergeable)
- #73 — unified DB migration plan (draft; mergeable; migration/cutover remains unapproved)
- #72 — separate JWT subject from client ownership (draft; mergeable)
- #71 — read-only Core repository context integration (non-draft; mergeability currently false)
- #69 — Core/GitHub context resolver integration (open; older verification limits noted in PR)
- #67 — read-only API and M1 governance alignment (open; review its exact current state before closing)

## Dependency-aware critical path
### P0 — security and operational safety
1. **CORE-IDENTITY:** #831 preserve `iss/sub/client_id` distinctions; then align/ratify #830 contract.
2. **CORE-BINDING:** #832 exact issuer/principal-kind/subject → active agent binding; fail closed on missing, ambiguous, inactive or mismatch.
3. **CORE-SCOPES:** verify route-specific scopes and authorization tests; ensure ledger/run/evidence queries are agent/principal scoped.
4. **CORE-STAGING:** #842 local staging Compose validation; confirm validation actually runs (PR description records runner setup failure and Compose validation unverified), then build/preflight/readiness.
5. **CRM-TEST-SAFETY:** #69 local-only database integration guard and safe cleanup.
6. **CRM-AUDIT:** #67 findings plan + #68 prototype; expand atomic mutation/audit remediation to all consequential mutation paths. Fail closed on audit persistence failures.
7. **INT-IAM:** #78 ES256 verifier + #72 subject/ownership; reconcile scopes, issuer/audience, resource ownership, deployment/partner binding. #77 accepted isolated-partner model is documentation evidence, not complete IAM closure.
8. **INT-READONLY:** #71 Core context integration, after API identity/scope contract is stable; no external side effects in pilot Intelligence.

### P1 — integration and proof
9. Integrate CRM operations dashboard (#63) against scoped read-only Control Plane API.
10. Complete staging migration rehearsal only after explicit review/approval; unified DB migration (#73) is not authorized by this plan.
11. E2E critical scenario: identity → context → plan → policy/authorization → approval boundary → guarded action (only when enabled) → observation → reconciliation → evidence → independent verification.
12. Failure-injection and recovery: unauthorized, wrong scope/binding, revoked principal, duplicate/replay, timeout before/after side effect, executor/observation/evidence failure, reconciliation, concurrent requests.
13. Backup/restore and rollback rehearsal; deployment/monitoring runbook.

### P2 — pilot and Early Access
14. Build a bounded read-only engineering governance pilot (repository analysis, risk, plan, PR review, evidence).
15. Pilot report and KPI baseline. Suggested acceptance: zero unauthorized execution; zero duplicate side effects; critical evidence durable; 100% critical/high finding verification; 100% critical E2E scenario pass.
16. Early Access release package only after gates pass: security review, staging evidence, operational runbook, onboarding, support/escalation, release artifacts, known limitations and rollback.

## PR disposition rules (Phase 0)
- **Merge candidate:** all required CI/checks pass, review and DCO complete, no unresolved blocking comments, dependency/base is merged, change scope is understood.
- **Revise:** security/CI/validation gaps, incomplete tests, stale base, conflict, or unmet acceptance criteria.
- **Hold:** depends on critical contract/binding/runtime work not yet merged.
- **Close:** superseded, duplicate, no longer relevant, or unsafe; only after reviewing issue/PR history and confirming replacement.
- Never bypass required security, dependency-risk, DCO or CI gates.

## CI and verification evidence policy
- PR descriptions and mergeability metadata are not a substitute for current check-run results.
- Record exact head SHA and latest check conclusion before merge decisions.
- PR #842's own description says 37 checks passed, Supabase Preview skipped as expected, but the Compose validation workflow failed during runner setup before validation; Docker Compose configuration remains unverified.
- Where tool/API access does not provide check-run detail, mark **UNVERIFIED** and request/trigger a proper rerun rather than infer success.
- PR head SHA and mergeability should be refreshed immediately before any merge.

## Phase 0 action checklist
- [x] Inventory open PRs in Core, CRM, Intelligence.
- [x] Record initial mergeability/draft metadata for critical PR subset.
- [x] Establish dependency-aware critical path and safety rules.
- [ ] Refresh current checks/head SHA for critical PRs in batches.
- [ ] Review PR comments/reviews and identify stale or conflicting PRs.
- [ ] Confirm base branches and stacked dependencies.
- [ ] Create/maintain issue or task tracker for all P0 items.
- [ ] Publish Phase 0 disposition recommendations (merge/revise/hold/close) with evidence.
- [ ] Reconcile plan after PR merges and update status.

## Gate definitions
- **G0 Backlog controlled:** inventory and dependencies reconciled; no unsafe bypass.
- **G1 Identity secure:** identity contract, exact binding, scopes and ownership tests pass.
- **G2 Execution governed:** durable action ledger, idempotency/replay protection, UNKNOWN outcome and authoritative reconciliation.
- **G3 CRM durable:** consequential mutation + audit durability/atomicity and safe integration tests proven.
- **G4 Staging proven:** deploy/preflight/readiness, migrations, secrets, backup/restore, rollback and observability tested in isolated staging.
- **G5 E2E verified:** critical scenarios and failure injection pass with evidence.
- **G6 Pilot accepted:** bounded scope, KPI and report; no critical/high unresolved blocker for pilot scope.
- **G7 Early Access:** repeatable isolated deployment, onboarding, support, release and rollback package.

## Safety constraints
- This roadmap does not authorize production deployment, production migration, credential changes, public client access, or execution-capability activation.
- Do not place secrets in repository, issue comments, PR bodies or logs.
- Use staging-only credentials and target verification.
- External side effects remain disabled until authorization/execution gateway readiness and explicit scope approval.

## Current next action
Finish Phase 0 check-run/review/dependency refresh for critical PRs, then begin Phase 1 by resolving identity contract and binding dependencies in order.