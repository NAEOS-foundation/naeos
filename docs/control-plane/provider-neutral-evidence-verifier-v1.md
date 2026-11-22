# Provider-Neutral Evidence Verification Contract v1

**Status:** Implementation contract  
**Version:** 1.0.0  
**Scope:** The production Control Plane's remote evidence-verifier boundary

## Trust boundary

The Control Plane does not infer an external effect from an agent, executor, operator narrative, local process log, or caller-supplied evidence reference. The HTTP verifier is a separately operated trust service. It MUST independently query or validate an authority capable of observing the effect (for example, the relevant provider API) before returning `VERIFIED`.

This repository implements the authenticated client and strict receipt validation. It does **not** implement a provider adapter or claim that a verifier service/provider is already deployed. A service that merely echoes the submitted reference is non-conforming.

## Request

The Control Plane sends one HTTPS POST to the configured endpoint:

```json
{
  "request_id": "random-128-bit-hex",
  "execution_id": "naeos-execution-id",
  "expected_outcome": "COMMITTED",
  "evidence_ref": "untrusted optional lookup hint"
}
```

`evidence_ref` is untrusted input. The verifier MUST bind its independent observation to the exact NAEOS execution ID and requested outcome. It MUST NOT treat the hint or the operator's assertion as proof.

## Successful response

Only HTTP 200 with a strict JSON object is eligible:

```json
{
  "request_id": "same request_id",
  "execution_id": "same execution_id",
  "outcome": "COMMITTED",
  "status": "VERIFIED",
  "evidence_ref": "provider-specific immutable receipt reference",
  "authority": "provider authority identifier",
  "evidence_sha256": "64 hexadecimal characters",
  "observed_at": "2026-11-22T00:00:00Z",
  "valid_until": "2026-11-22T00:05:00Z",
  "verification_policy_version": "provider-policy-v1"
}
```

The client rejects mismatched request/execution/outcome bindings, non-VERIFIED status, missing authority or policy version, malformed digest/reference, future observations, expired receipts, unknown fields, oversized/multiple JSON values, non-200 responses, and redirects. Network calls are bounded by a five-second timeout. The verified reference returned to the ledger is the canonical reference in this receipt, never the caller's submitted hint.

## Verifier service MUST

- authenticate the NAEOS client over TLS and authorize only the evidence-verification operation;
- use least-privilege, read-only provider credentials where possible;
- independently query the configured provider authority and bind observations to the NAEOS execution ID, target, operation, and expected outcome;
- enforce effect-specific verification policy and freshness before returning `VERIFIED`;
- preserve contradictory observations as non-confirming/UNKNOWN;
- distinguish observed evidence from estimates, cached state, and executor claims;
- return immutable evidence references and a SHA-256 digest of the evidence material;
- prevent replay across request IDs and execution IDs;
- retain auditable provenance, provider identity, observation time, policy version, and verification result;
- never possess NAEOS execution credentials or side-effect authority merely by being a verifier.

The service core returns `VERIFIED` only after the source returns a bound observation and the injected policy accepts it. The policy MUST expose a non-empty immutable version identifier, which is retained in the receipt. The service MUST return a non-200 response when evidence is missing, stale, contradictory, insufficient, or cannot be independently observed. It MUST NOT manufacture a terminal outcome to satisfy a caller.

## Client configuration

The production CLI accepts these optional environment variables:

- `NAEOS_CONTROLPLANE_EVIDENCE_VERIFIER_URL`: full HTTPS endpoint for the v1 verification operation.
- `NAEOS_CONTROLPLANE_EVIDENCE_VERIFIER_TOKEN`: dedicated service credential; store in the deployment secret manager and never commit it.

Both must be set together. If neither is set, the Control Plane starts with reconciliation unavailable (HTTP 503), preserving fail-closed behavior. If only one is set or the URL is invalid, startup fails before the listener binds. Setting these values does not prove the remote service is conforming; deployment approval requires a separate service security review and provider-specific conformance evidence.

## Security limitations

TLS plus a dedicated bearer credential authenticates the configured verifier service; it does not independently prove the provider's statement to the Control Plane. The verifier service implementation, its provider trust anchors, credential scope, evidence retention, freshness policy, and operational isolation remain explicit deployment trust dependencies.

This contract intentionally does not select GitHub, a cloud provider, or another provider as NAEOS's universal source of truth. Provider-specific adapters must be implemented and tested separately. No production provider is configured by this change.

## Standalone verifier service

A standalone executable is available at `cmd/naeos-evidence-verifier` and can be built with `Dockerfile.evidence-verifier`. It exposes `POST /v1/verify`, `GET /healthz`, and `GET /readyz`.

Required environment: `NAEOS_EVIDENCE_SOURCE_URL` (HTTPS read-only observation API), `NAEOS_EVIDENCE_SOURCE_TOKEN`, `NAEOS_EVIDENCE_VERIFIER_TOKEN`, `NAEOS_EVIDENCE_POLICY_VERSION`; optional `NAEOS_EVIDENCE_VERIFIER_ADDR` defaults to `:8081`.

The configured source API must implement a read-only GET using query parameters `execution_id`, `expected_outcome`, optional `evidence_ref_hint`, and return the observation object described in the contract. HTTP 404/409/422 are non-confirming; other non-200 responses fail as unavailable. The generic policy only binds the expected outcome and is not provider-specific effect verification.

**Deployment status:** runnable service code, no provider source or service deployment is configured. Controlled staging/integration use only until the source API, effect-specific policy, TLS/network isolation, credential scope, and conformance evidence are reviewed. Health/readiness do not validate provider reachability. Never expose the service publicly.
