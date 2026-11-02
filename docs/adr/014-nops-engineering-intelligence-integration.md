# ADR-014: Integrate NOPS as an External Engineering Intelligence Service

> Status: Proposed
> Date: 2026-11-02
> Decision owners: NAEOS maintainers

## Context

NAEOS is a Go-based engineering framework with explicit policy, execution, verification, and evidence boundaries. NOPS (NAEOS-foundation/Nops) is a separate Python prototype that provides AI-assisted code audit, refactoring proposals, and deployment-configuration analysis through configurable model providers.

The capabilities are complementary, but NOPS is not currently a production-grade NAEOS component. Its current implementation includes a simplified string-replacement patch applier, minimal tests, a Flask starter endpoint, and development artifacts committed to the repository. Directly importing NOPS into the NAEOS core would couple the core to Python and risk confusing advisory AI output with trusted execution or verification.

NAEOS already defines an external-agent adoption boundary. This decision must remain consistent with that boundary and the documentation authority hierarchy.

## Decision

1. Keep NOPS in its own repository and runtime, initially as an **external engineering intelligence service**.
2. Integrate through a versioned, protocol-neutral NOPS contract and a NAEOS-owned adapter. The adapter translates requests and responses; it does not grant permissions.
3. Treat all NOPS/LLM findings, plans, patches, and deployment recommendations as untrusted proposals.
4. Route any consequential repository or deployment operation through the existing NAEOS control plane, policy evaluation, authorized runtime, verification, and evidence mechanisms.
5. Do not allow NOPS to write to repositories, invoke deployment credentials, or perform consequential side effects as a consequence of an AI response.
6. Keep model-provider credentials and provider-specific behavior inside NOPS; do not expose provider keys to NAEOS agents or generated artifacts.
7. Start with read-only code/deployment analysis. Introduce patch proposal and application only after the patch contract, authorization boundary, rollback behavior, and deterministic tests are accepted.
8. Do not merge NOPS source code into the NAEOS core as part of this integration.

## Initial scope

- Read-only code audit requests and structured findings.
- Read-only deployment configuration analysis.
- Refactoring plans represented as proposed patches, not automatically applied changes.
- Explicit provenance for repository, revision, file, provider/model, prompt version, and analysis result.
- NAEOS-side policy and evidence handling for any future action.

## Non-goals

- Replacing the NAEOS policy engine, runtime, verifier, or evidence ledger.
- Treating model output or an NOPS status field as proof of correctness.
- Giving NOPS production deployment credentials.
- Running arbitrary generated code or patches in the NOPS service.
- Making NOPS a trusted policy decision point.

## Security and trust invariants

- **Fail closed:** malformed, unsupported, or unverifiable contract messages are rejected.
- **No authority by suggestion:** a recommendation, patch, or SAFE_TO_DEPLOY label is not authorization.
- **Least privilege:** the first integration is read-only and receives only the source context required for the requested analysis.
- **Secret minimization:** secrets and sensitive environment values are redacted before external model calls where possible; redaction failures block the request.
- **Revision binding:** every result is bound to an immutable repository revision and explicit file paths/digests.
- **Traceability:** NAEOS records request identity, contract version, NOPS version, provider/model metadata when available, result digest, and downstream verification references.
- **No bypass:** all consequential side effects remain subject to NAEOS policy and runtime enforcement.

## Alternatives considered

### Embed NOPS into NAEOS core

Rejected. It introduces a Python runtime and provider dependencies into the Go core and blurs the trust boundary between advisory AI and deterministic enforcement.

### Allow NOPS to execute patches directly

Rejected. The current patch mechanism is simplified string replacement and does not establish authorization, rollback, or behavioral correctness.

### Expose NOPS only as an independent developer CLI

Deferred as a useful standalone mode, but insufficient for a governed NAEOS integration with consistent identity and evidence.

## Consequences

### Positive

- Preserves NAEOS core language and architecture boundaries.
- Reuses NAEOS governance for consequential operations.
- Allows NOPS to evolve and support multiple model providers independently.
- Establishes a narrow path for audit, refactor, and deployment-analysis capabilities.

### Costs and risks

- Requires a maintained adapter and versioned contract.
- Requires careful data minimization for source code sent to external providers.
- NOPS must be hardened and tested before production use.
- Evidence quality depends on recording provenance and independently validating claims.

## Rollout gates

1. **Documentation gate:** this ADR and the NOPS integration contract are reviewed and accepted.
2. **NOPS hygiene gate:** remove committed virtual environments/build artifacts, correct packaging metadata, pin and audit runtime dependencies, and establish meaningful CI tests.
3. **Read-only adapter gate:** implement request/response validation, timeouts, bounded payloads, authentication, and revision binding.
4. **Evidence gate:** persist provenance and result digests; distinguish analysis completion from verification success.
5. **Patch gate:** require structured diff validation, path allowlists, exact-base checks, backup/rollback, explicit authorization, and post-application tests.
6. **Production gate:** security review, operational limits, observability, and documented incident/rollback procedures.

No later gate is implied to be complete by acceptance of this ADR.

## References

- [NAEOS Documentation Authority Model](../../DOCUMENTATION-AUTHORITY.md)
- [NAEOS External Agent Adoption Contract](../AGENT-ADOPTION-CONTRACT.md)
- [NAEOS AI Integration Contract](../AI-INTEGRATION-CONTRACT.md)
- [NOPS repository](https://github.com/NAEOS-foundation/Nops)
