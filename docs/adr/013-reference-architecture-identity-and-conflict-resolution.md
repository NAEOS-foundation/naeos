# ADR-013: Reference Architecture Identity and Authorization Semantics

- ADR ID: ADR-013
- Title: Reference Architecture Identity and Authorization Semantics
- Status: Proposed
- Date: 2026-11-01
- Authors: NAEOS OSS
- Related Documents: [NAEOS-NRA-001](../../Reference%20Architecture/NAEOS-NRA-001.md), [NAEOS-MTS-001](../../NAEOS-MTS-001.md), [NES-016 AI](../NES-016-AI.md), [NES-020 Security](../NES-020-Security.md), [AI Integration Contract](../AI-INTEGRATION-CONTRACT.md), [Agent Adoption Contract](../AGENT-ADOPTION-CONTRACT.md), [Trust Boundary Map](../../TRUST-BOUNDARY-MAP.md)

## Context

The repository contains two Reference Architecture documents with conflicting identity and authority signals. `NAEOS-NRA-001.md` is established as the canonical Reference Architecture by the documentation authority model and repository navigation. `NAEOS-NRA-002.md` repeats the NRA-001 document identifier and Stable status while presenting a divergent architecture draft.

The second draft contains useful agent-control and authorization concepts, but several are already represented in the Master Technical Specification, Engineering Specifications, integration contracts, and trust-boundary documentation. Promoting both documents would create ambiguous authority and duplicate normative material.

The repository also needs an explicit, deterministic policy conflict rule and a clear distinction between an explicit policy outcome and an internal inability to evaluate policy.

## Decision

1. Keep `Reference Architecture/NAEOS-NRA-001.md` as the sole canonical NAEOS Reference Architecture.
2. Archive `Reference Architecture/NAEOS-NRA-002.md` as a non-canonical historical working draft. Preserve its original contents and metadata for traceability, but add a clear archival notice and remove it from active architecture navigation.
3. Map validated technical requirements from the archived draft to the appropriate existing normative or supporting documents. Do not copy the entire draft into another document or imply that every concept is implemented.
4. Resolve conflicting valid policy outcomes using the deterministic precedence `DENY > REQUIRE_APPROVAL > ALLOW`.
5. Keep policy evaluation failures as internal error/indeterminate states, not a fourth externally valid authorization decision. For protected actions, fail closed before consequential side effects; preserve diagnostics that distinguish evaluation failure from explicit DENY.
6. Treat `REQUIRE_APPROVAL` as non-executable until valid approval and the required runtime revalidation have completed.
7. Keep `NAEOS-MTS-001` in its existing Proposed status. This ADR does not promote the MTS or claim implementation conformance.

## Rationale

A single canonical architecture prevents competing sources of truth. Archiving rather than deleting the second draft preserves provenance while removing its misleading authority signals from active navigation.

The policy precedence provides deterministic conflict resolution and ensures that a restrictive applicable outcome cannot be overridden by a less restrictive one. Keeping evaluation failures distinct from DENY preserves semantic accuracy and operational diagnosability without allowing uncertainty to authorize protected actions.

Mapping details into existing specifications and contracts avoids duplicate requirements and keeps each document within its intended scope.

## Alternatives Considered

- **Promote NRA-002 as a second canonical architecture:** rejected because it duplicates the document identifier and conflicts with the established authority hierarchy.
- **Replace NRA-001 with NRA-002:** rejected because NRA-001 is already the canonical approved reference and the second draft is incomplete as an architecture baseline.
- **Merge both documents wholesale:** rejected because this would duplicate existing MTS/NES/contract requirements and blur architecture versus implementation specification.
- **Delete NRA-002:** rejected because archival preservation is preferable for historical traceability.
- **Represent evaluation failure as a fourth authorization decision:** rejected because an evaluator failure is not a policy decision and should remain an internal error state.

## Consequences

- Documentation authority remains unambiguous: NRA-001 is canonical.
- NRA-002 remains available for historical reference but is not normative or implementation authority.
- The MTS and supporting specifications explicitly describe policy precedence and fail-closed evaluation-error handling.
- Implementations and conformance tests must preserve the distinction between explicit DENY and internal evaluation failure.
- This ADR is Proposed until reviewed and accepted through the repository's normal governance process.

## Notes

This decision is documentation and architecture governance only. It does not assert that all described controls are implemented in the current runtime, API, or product. No runtime code, API behavior, CI policy, or MTS status is changed by this ADR.
