# Evidence Verifier HTTP Conformance

The standalone verifier defaults to loopback `127.0.0.1:8081`. Its HTTP route
contract is:

| Route | Method | Meaning |
|---|---|---|
| `/healthz` | GET | Process liveness only |
| `/configz` | GET | Required static configuration initialized |
| `/readyz` | GET | Bounded read-only PostgreSQL replay-guard readiness |
| `/v1/verify` | POST | Authenticated evidence verification |

Health/config/readiness routes reject non-GET requests with 405 and `Allow: GET`.
Readiness returns 503 if the readiness probe is absent or fails. It does not
assert provider-specific execution-effect verification.

Non-loopback listener overrides are plaintext and must remain behind authenticated
TLS termination with restrictive private-network ingress. This PR does not alter
production deployment or secrets.
