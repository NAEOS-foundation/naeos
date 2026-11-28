# Evidence Verifier Listener Security Boundary

Status: enforced default; external deployment remains unverified
Related: Issue #870, PR #862, PR #871, PR #889

## Default listener

The standalone Evidence Verifier defaults to `127.0.0.1:8081`. This prevents a
configuration-free launch from binding the bearer-token HTTP endpoint to all
network interfaces. The health, config, and readiness routes accept GET only.
`/healthz` indicates process liveness, `/configz` indicates static
configuration was initialized, and `/readyz` remains HTTP 503 until a bounded,
read-only source readiness probe exists.

## Explicit address override

`NAEOS_EVIDENCE_VERIFIER_ADDR` can override the listener address. Any
non-loopback bind must be treated as an unsafe plaintext listener unless the
deployment architecture places it behind authenticated TLS termination and
restricts network access to that proxy and authorized callers. The environment
variable itself does not prove TLS or network isolation.

Before an external/private-network deployment is approved, attach evidence for:

- TLS termination and certificate lifecycle at the ingress/proxy.
- Private network exposure and allow-list/network policy restricting callers.
- Authentication at both ingress and verifier service boundaries as appropriate.
- No provider execution credentials; source credentials must be read-only.
- Health probes use GET and do not equate liveness/configuration with dependency readiness.
- Shutdown and request timeouts remain bounded.
- Deployment configuration and effective listener address have been inspected.

No deployment or production configuration is changed by this document or code change.

## Conformance coverage added by this change

- Default listener is loopback-only.
- `/healthz`, `/configz`, and `/readyz` reject non-GET with 405 and
  `Allow: GET`.
- Nil verifier service returns 503 for the verification route.
