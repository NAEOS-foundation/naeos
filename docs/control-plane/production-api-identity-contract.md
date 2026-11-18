# Production Control Plane API and Identity Contract

**Status:** Proposed — implementation gate  
**Scope:** Production Control Plane API compatibility for `/decision`, `/runs`, and `/evidence`  
**Baseline:** `feat/control-plane-db-target-guard-20261118` at `35c2ddab6c3893cfcc562b3f63b422fb84c86cbc`  
**Environment:** Internal NAEOS use; no pilot-client or multi-tenant authorization is implied.

## 1. Context and observed implementation

The dedicated production runtime is composed separately from the investor demo and uses PostgreSQL repositories, a persistent ledger, and JWKS-backed bearer-token validation.

At this baseline:

- Production registers `GET /health/live`, `GET /health/ready`, and `POST /api/control-plane/decision`.
- The demo API also registers `GET /api/control-plane/runs` and `GET /api/control-plane/evidence`.
- The website has server-side proxy routes for all three API paths. The `runs` and `evidence` proxies currently forward `NAEOS_CONTROLPLANE_API_TOKEN` as a bearer token.
- The production JWKS validator currently chooses `client_id` when present and otherwise `sub`, then exposes that selected value as `JWTClaims.Sub`.
- The production decision handler binds request `agent_id` to that selected value. It does not currently enforce the token's `scope` or `scopes` claims.
- The current Control Plane persistence model has grants keyed by `agent_id`, but no explicit issuer/principal-to-agent binding repository is wired into the production runtime.

These differences mean the demo handlers cannot safely be copied into production as-is. A successful database connection does not establish API or identity compatibility.

## 2. Security invariants

1. A verified JWT is evidence of an issuer-authenticated principal, not by itself proof that the principal may act as any requested agent.
2. Preserve `iss`, `sub`, `client_id`, `aud`, `scope`, and `scopes` as distinct verified claims. Never overwrite `sub` with `client_id`.
3. Resolve a principal to an NAEOS `agent_id` through an explicit, persisted, enabled binding. Never trust a caller-supplied `agent_id` as the binding.
4. Enforce a route-specific scope after JWT signature and registered-claim validation. A valid token without the required scope is not authorized.
5. Apply the resolved agent boundary in the database query itself. Filtering a globally loaded ledger in application memory is not an acceptable production isolation boundary.
6. Caller-supplied filters may narrow a query; they must never widen the resolved principal's permitted data scope.
7. Evidence is verified only when a configured trusted verification key is available and every returned evidence bundle passes verification. Missing trust material must fail closed or explicitly report `UNTRUSTED`.
8. A static shared bearer token is not an end-user identity and must not be treated as a substitute for JWKS identity, scoped authorization, revocation, or actor attribution.
9. The production runtime must not mount demo reset, scenario, in-memory execution, or other demo-only routes.

## 3. Principal and agent identity model

The authentication layer must return a typed principal context rather than a single overloaded string:

- `issuer`: exact validated JWT issuer.
- `subject`: exact validated JWT `sub`.
- `client_id`: validated `client_id`, when present; kept separate from subject.
- `principal_kind`: selected by trusted issuer/audience configuration or a separately validated, issuer-controlled claim. It must not be guessed merely from the presence of `client_id`.
- `scopes`: normalized set derived from validated `scope` and/or `scopes` claims.

The authorization layer resolves this context through an explicit binding, conceptually keyed by `(issuer, principal_kind, principal_id)`, to one or more NAEOS agent identities. A binding must be enabled and unambiguous for the requested operation. The binding store and its migration are not part of this documentation-only change.

For Supabase user tokens, `sub` is the user identity. For machine/service tokens, the service identity must be selected by an explicit trusted token profile. Do not silently treat a machine `client_id` as a human `sub`, or vice versa.

If a server-side website proxy calls the API on behalf of a signed-in user, the request must preserve verifiable end-user attribution. Acceptable future designs include:

- securely forwarding the user's validated access token to the Control Plane, with an explicit user principal binding; or
- minting a short-lived, asymmetric service assertion that carries a verifiable actor/delegation claim and is validated against a configured trust profile.

A static `NAEOS_CONTROLPLANE_API_TOKEN` alone does not carry this attribution. No proxy authentication change is authorized by this ADR.

## 4. Route contract

### `POST /api/control-plane/decision`

- Authentication: valid bearer JWT from the configured issuer and audience.
- Required scope: `naeos:control-plane:decision:write`.
- Identity: resolve the verified principal to an authorized `agent_id`; reject caller identity mismatches.
- Authority: load canonical grant and active policy from PostgreSQL; ignore caller-supplied grant, policy, and authorization context.
- Time: use server time for authorization and expiry checks.
- Durability: persist decision evidence before returning a usable decision.
- Failure: return unauthorized/forbidden for identity or scope failures; return unavailable/indeterminate for persistence or evidence failures.

### `GET /api/control-plane/runs`

- Authentication: valid bearer JWT and required read scope `naeos:control-plane:runs:read`.
- Data source: persistent PostgreSQL ledger/repositories only; never demo in-memory state.
- Output: recent execution/run summaries derived from correlated ledger events, including stable identifiers, agent, capability, state, timestamps, and verification state where available.
- Scope: every query is constrained by the resolved agent binding or an explicitly approved broader read grant.
- Filters: supported correlation filters may narrow results only. Add bounded pagination/limits; do not expose an unbounded ledger scan through the HTTP API.
- Semantics: a run summary is a projection of recorded events, not proof of an external side effect by itself.

### `GET /api/control-plane/evidence`

- Authentication: valid bearer JWT and required read scope `naeos:control-plane:evidence:read`.
- Data source: persistent PostgreSQL ledger and canonical evidence reconstruction.
- Output: correlated ledger events and canonical evidence bundles for the authorized agent scope.
- Scope: enforce the agent boundary in the repository query before returning events or bundles.
- Verification: use the configured trusted evidence verification key. Missing key or failed verification must be explicit; never label untrusted evidence as verified.
- Filters: correlation filters may narrow results only. Add bounded pagination/limits and stable ordering.

## 5. HTTP response semantics

| Condition | Status |
|---|---:|
| Missing or invalid bearer token | 401 |
| Valid token, missing scope, disabled/unmapped principal, or scope violation | 403 |
| Invalid query/body shape or unsupported filter | 400 |
| Requested resource not visible in the resolved principal scope | 404 or an explicitly documented non-disclosing 403 policy |
| Persistence, schema, ledger, or trusted-verification dependency unavailable | 503 |
| Decision allowed | 200 |
| Decision requires approval | 202 |
| Decision denied by policy | 403 |

Do not leak raw database URLs, credentials, JWT contents, or internal authorization mappings in error responses.

## 6. Required implementation and verification gates

Before registering `/runs` or `/evidence` in the production handler:

1. Preserve distinct JWT subject and client identity fields in the validator output.
2. Add typed principal resolution and an explicit persisted principal-to-agent binding model.
3. Enforce per-route scopes for decision, runs, and evidence.
4. Add database-level ledger query methods with agent scoping, deterministic ordering, and bounded pagination.
5. Reconstruct and verify evidence from persisted events; test missing-key and tampering outcomes.
6. Add handler tests for missing/invalid JWT, wrong issuer/audience, missing scope, unmapped/disabled principal, mismatched caller agent, cross-agent query attempts, empty results, pagination bounds, and database failures.
7. Add proxy integration tests proving that the website's upstream credential model supplies the required verifiable principal/actor context.
8. Keep production and staging targets separate; run staging validation only after the implementation and identity contract are reviewed.

## 7. Explicit non-goals and safety boundary

This proposal does not:

- apply or generate Supabase migrations;
- create or modify database roles, grants, RLS policies, or credentials;
- read, rotate, or write secrets;
- change Cloudflare or Railway variables;
- change the website proxy's static-token behavior;
- register new production HTTP routes;
- change the Docker entrypoint or deploy any service;
- claim tenant isolation or pilot-client readiness.

The existing environment database target guard remains a separate concern in PR #829. This contract is a prerequisite for a later implementation PR, not authorization to perform a production cutover.

## 8. Review decision requested

Review this contract for correctness of the identity model, route scopes, principal-to-agent binding requirements, and evidence-read semantics. Implementation should proceed only after these boundaries are accepted.
