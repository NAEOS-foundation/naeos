# NAEOS Outreach Server (NOS) — Foundation Design

- **Status:** Draft for review
- **Version:** 0.1.0
- **Scope:** Phase 1 architecture, service boundary, API contract, data model, and acceptance criteria
- **Production:** Out of scope; no production systems or credentials are touched.

## 1. Purpose

NAEOS Outreach Server (NOS) is a governed prospect-research, campaign-preparation, and outreach-workflow service for NAEOS. It supports founder-led, evidence-backed outreach preparation first; any later controlled execution requires a separate, approved stage.

NOS is not a bulk-mail sender. AI recommendations and message drafts are proposals, not authorization to perform external side effects.

## 2. Service and repository boundary

**Recommendation:** Keep NOS as an independently deployable service, separate from NAEOS core and CRM business logic. A standalone repository remains the target repository boundary, with this document maintained in the canonical NAEOS repository as the architecture and governance reference.

The Phase 1A Go skeleton currently lives under `services/naeos-outreach` in this repository. Its placement is explicitly provisional: retain it here until repository ownership/administration is confirmed, then move it to the approved standalone repository without changing the service boundary. Phase 1A currently implements only `GET /health` and `GET /ready`; the remaining routes in the OpenAPI file are design proposals, not implemented endpoints.

Integration boundaries:
- **NAEOS Authorization & Execution Gateway:** the only authority for any future external action authorization, idempotency binding, dispatch lifecycle, durable action ledger, UNKNOWN handling, reconciliation, and execution audit. NOS may submit an eligible action intent through an approved contract and retain a reference/read-only status projection; it must not create a competing execution ledger or dispatch/retry/resolve actions itself.
- **NAEOS Intelligence:** provider-neutral research synthesis, qualification explanations, and draft assistance; outputs preserve evidence references and never authorize actions.
- **NAEOS CRM:** pipeline and activity sync through a documented adapter/API; NOS owns outreach-specific research and campaign-preparation state. No shared mutable CRM tables.
- **PostgreSQL:** a dedicated `outreach` schema/database boundary. Runtime credentials are least-privileged; migrations use separate credentials.
- **Redis/queue:** optional asynchronous work only. Queue state is not authoritative business or execution state.

## 3. Phase 1 MVP scope

Included:
- Prospect/company and public-contact research records with verification status and source provenance.
- Evidence/source records for factual enrichment.
- Rubric-versioned, evidence-linked qualification.
- Campaign definitions and campaign membership.
- Message drafts with version history and factual-claim references.
- Human review/approval state for campaign and draft versions.
- CRM sync mapping and activity-event contract; no direct shared-table writes.
- Typed suppression/opt-out records and checks.
- A non-sending preparation API/dashboard with narrowly scoped writes and dry-run planning.

Excluded:
- Any live external send or provider adapter, including email delivery.
- Autonomous outreach or bulk execution.
- LinkedIn scraping or automation.
- Private contact-data inference, guessed emails, or unsupported claims about intent.
- Production deployment, production database changes, provider secrets, or real provider-account connections.

## 4. Core data model (logical)

- `prospects`: prospect_id, company_name, website, country/region, segment, status, owner, created_at, updated_at.
- `contacts`: contact_id, prospect_id, optional name/role/public channel, verification_status, created_at. Every factual field used in qualification, campaign membership, or a draft must be backed by source evidence; unknown fields remain unknown.
- `sources`: source_id, prospect_id and/or contact_id, source_url, source_type, retrieved_at, evidence_excerpt/hash, confidence, attribution. A source is attached to the subject it supports.
- `qualification_scores`: prospect_id, rubric_version, score, priority, rationale, evidence_source_ids, scored_at, and server-derived actor/model provenance. Caller-supplied actor identity is never authoritative.
- `campaigns`: campaign_id, name, goal, ICP_definition, channel, status, policy_profile, created_by, created_at.
- `campaign_members`: campaign_id, prospect/contact subject, status, suppression_check_at, qualification_snapshot.
- `message_drafts`: draft_id, campaign_member_id, channel, subject/body, version, factual_claims with source references, status, generated_at.
- `approvals`: approval_id, object_type, object_id, exact version, decision, authenticated reviewer, rationale, timestamp.
- `outreach_actions`: only a NOS-side business reference such as campaign_member_id and an opaque Gateway invocation/action reference, plus a read-only status projection and timestamps. The Gateway's approved ledger is authoritative; NOS does not store a second idempotency binding or independently transition execution state.
- `events`: event_id, aggregate_type/id, event_type, event_time, bounded/redacted payload, source, and correlation_id. Security-relevant history is append-only.
- `suppression_entries`: suppression_id, identifier_type (`account_domain`, `contact_id`, or `channel_address`), normalized_identifier, applicable channel when relevant, normalization_version, reason/source, effective_at, created_at.
- `audit_records`: audit_id, server-derived actor, action, target, result, correlation_id, occurred_at, redacted metadata.

Secrets stay in a secret manager. Provider credentials, bearer tokens, and unnecessary personal data must never be stored in prospect, campaign, or message records or written to logs.

## 5. API contract (initial, versioned)

Base path: `/api/v1`. The contract is proposed; aside from `/health` and `/ready`, endpoints are not implemented by the current Phase 1A skeleton.

- `GET /health` and `GET /ready` — process health and readiness.
- `POST /prospects`, `GET /prospects`, `GET /prospects/{prospect_id}` — create, list/search, and retrieve prospects with evidence.
- `POST /prospects/{prospect_id}/contacts`, `GET /prospects/{prospect_id}/contacts`, `GET /contacts/{contact_id}` — create/list/retrieve contact research records; records remain unverified until evidence is attached.
- `POST /prospects/{prospect_id}/sources`, `POST /contacts/{contact_id}/sources` — attach evidence to the exact prospect/contact it supports.
- `POST /prospects/{prospect_id}/qualification` — record an evidence-backed assessment.
- `POST /campaigns`, `GET /campaigns`, `GET /campaigns/{campaign_id}` — create/list/retrieve campaigns.
- `POST /campaigns/{campaign_id}/members` — add an eligible prospect/contact subject.
- `POST /campaigns/{campaign_id}/drafts` — create a non-sending draft proposal.
- `POST /drafts/{draft_id}/approval` — approve/reject a specific version; the server derives reviewer identity from the authenticated principal.
- `POST /campaigns/{campaign_id}/dry-run` — produce an eligibility plan; it never sends.
- `GET /suppression`, `POST /suppression` — read/manage typed suppression records with separate restricted capabilities.
- `GET /events` — paginated activity/event stream.

All protected operations require an authenticated identity and an operation-specific capability. All requests accept an optional `X-Request-ID`; if omitted, the server generates one, returns it in the same response header, and records it with the request. The actor is derived from trusted authentication middleware, never a body field. All mutations are validated and authorized. External side effects are absent in Phase 1.

## 6. Prospect qualification

Scoring is rubric-versioned and explainable. It is a prioritization aid, not a claim about company intent or pain. Each result records:
- rubric version, score, and priority band;
- rationale and source IDs supporting factual inputs;
- explicit hypothesis labels for inferred fit/pain;
- timestamp and server-derived actor/model provenance.

For human-scored records, provenance identifies the authenticated actor. For model-assisted records, provenance identifies the model and version and, where applicable, the responsible authenticated initiator. These fields are server-recorded; callers cannot assert another actor. Unknown information remains unknown. Never invent contact identity, private email, intent, customer status, or a pain point.

## 7. Approval and future action states

Campaign: `draft → review → approved → paused/archived`.

Draft: `draft → review → approved/rejected → superseded`. Approval binds to one exact draft version and is invalidated by material edits.

Phase 1 supports only dry-run plans; no send action is implemented. For a future stage, NOS may retain a proposal/reference and consume a read-only projection of Gateway status. Only the approved NAEOS Gateway may own authorization, idempotency, durable intent, dispatch, UNKNOWN, and reconciliation transitions. NOS must not independently authorize, dispatch, retry, or resolve an UNKNOWN result. If suppression status cannot be checked and bound to a future authorization at the last controllable boundary, the action remains blocked.

## 8. Compliance and safety controls

- Default to manual/human review; approval is not itself permission to send.
- Use legitimate, public, relevant data sources and record URL, attribution, and collection time. URL fetching must be limited to public HTTP(S) resources with redirect and private-network/SSRF protections.
- No guessed personal contact data, restricted-platform scraping, or automated LinkedIn activity.
- Contact facts remain unverified until attached evidence supports them; no contact is campaign-eligible based on an unsupported name, role, or channel.
- Suppression scope is explicit: account domain, internal contact ID, or a channel address. Normalize by a versioned, channel/type-specific rule; use exact matching only (no fuzzy name matching). Store only the minimum needed identifier, restrict access, and redact it from logs. Apply suppression before campaign membership and again immediately before any future execution; if the check is unavailable or ambiguous, fail closed.
- Any future send requires a separate approved integration that binds the latest suppression decision to the Gateway request and is revalidated before dispatch.
- Provide purpose, sender identity, and a clear opt-out path before any future messages. Legal basis, retention, anti-spam, privacy, platform, and local-law review are prerequisites to live outreach.
- Campaign/sender budgets, rate limits, quiet hours, and a kill switch are required before any live-execution phase.
- Secrets are isolated from AI workers. Audit records are append-only for security-relevant events.

## 9. Suggested repository and deployment layout

NOS should be an independently deployable Go service. A standalone repository is recommended once ownership and repository administration are confirmed. Keep a separate web dashboard or integrate with the CRM UI only after that UI boundary is decided. Phase 1 is limited to local development/staging; no production rollout is implied.

Suggested standalone layout:

```text
cmd/nos-server/
internal/
  auth/
  prospects/
  contacts/
  sources/
  qualification/
  campaigns/
  drafts/
  approvals/
  crm/
  compliance/
  audit/
  storage/
api/openapi/
migrations/
tests/
deploy/
```

Do not create a repository or paid infrastructure until the service boundary and owner are confirmed. The existing monorepo skeleton is provisional and can be moved after review without changing the service's logical boundary.

## 10. Security baseline

- Authentication: NAEOS identity integration or a secure staging-only auth adapter.
- Authorization: role/capability-based access, least privilege, and explicit per-operation capabilities; suppression administration is separately restricted.
- Database: dedicated role/schema; migrations use separate privileged credentials; runtime role is non-owner and least privilege.
- Tenant/owner isolation is enforced at both API and database layers before multi-tenant use.
- Strict IDs, URL schemes/targets, pagination limits, request sizes, and enums; source retrieval must prevent SSRF and revalidate redirects.
- TLS at the deployment boundary, secure headers, request IDs, bounded timeouts, and correlation across API/audit events.
- Queue messages contain IDs rather than secrets or unnecessary PII; workers revalidate state.
- Structured, redacted logs; no message bodies, suppression identifiers, or credentials by default.
- Tests cover unauthorized access, role restrictions, evidence ownership, suppression normalization/matching, idempotency at the Gateway boundary, migrations, and failure paths.

## 11. Phase 1 acceptance criteria

1. Design/service boundary, provisional repository placement, and data ownership are documented.
2. OpenAPI is machine-readable and matches the proposed Phase 1 routes, request correlation, capabilities, and error responses.
3. Only `/health` and `/ready` are claimed as implemented by the current skeleton; proposed routes are not represented as delivered functionality.
4. Prospect/contact records and source evidence preserve exact subject association; unsupported contact facts remain unverified.
5. Qualification is rubric-versioned, evidence-linked, and records server-derived actor/model provenance.
6. Campaign and draft state machines reject invalid transitions; approval is version-specific and material edits invalidate it.
7. Suppression is typed, deterministically normalized, and visible in dry-run decisions; tests cover account/contact/channel scope.
8. CRM integration is an adapter/contract; there are no direct shared-table writes.
9. No email-send provider is connected; tests prove dry-run has no external side effects.
10. Any future external action uses the single approved Gateway ledger and final authorization boundary; NOS cannot dispatch or resolve UNKNOWN independently.
11. Security tests cover unauthorized access, least privilege, invalid inputs, tenant boundaries, SSRF protections, and fail-closed behavior.
12. CI and self-review evidence are required before the PR is marked ready.

## 12. Implementation sequence

1. Review and ratify this architecture note and approve the service/repository owner.
2. Reconcile the proposed API and error model with the Phase 1A skeleton and deployment topology.
3. Decide the staging database boundary; add migrations and repository interfaces only after review.
4. Implement authenticated prospect/contact CRUD and source provenance with tests.
5. Implement qualification provenance and campaign/draft/version-specific approval states.
6. Add typed suppression and dry-run checks with exact evidence.
7. Add the CRM adapter contract without shared-table writes.
8. Add CI, security review, and a staging-only smoke test.
9. Review separately before authorizing any Gateway integration or live-send capability; no production release is implied.

## 13. Decision record

- NOS is an independently deployable service, separate from NAEOS core and CRM business logic; current monorepo placement is provisional until owner/repository administration is confirmed.
- The NAEOS Gateway remains the sole authority for any future external action lifecycle and execution ledger.
- Factual enrichment and contact claims require attached evidence; unknown information remains unknown.
- Human review is mandatory in the MVP, but approval does not authorize a send.
- Phase 1 is non-sending preparation and dry-run only; live outreach is explicitly disabled.
