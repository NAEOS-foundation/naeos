# Production Control Plane JWKS Authentication

This document describes the authentication-to-authorization binding for the dedicated
production Control Plane. The runtime is not the investor demo and does not mount
demo reset, scenario, or execution routes.

## Required runtime configuration

The `naeos control-plane serve` command requires these environment variables:

- `DATABASE_URL`: dedicated Control Plane PostgreSQL login/connection URL (TLS required).
- `NAEOS_CONTROLPLANE_ISSUER`: exact JWT `iss` claim value.
- `NAEOS_CONTROLPLANE_AUDIENCE`: exact expected JWT `aud` value.
- `NAEOS_CONTROLPLANE_JWKS_URL`: HTTPS URL that publishes the signing key set.
- Optional `NAEOS_CONTROLPLANE_ADDR`: HTTP listen address; defaults to `:8080`. CLI `--addr` overrides it.

For Supabase hosted Auth, the documented JWKS endpoint pattern is:

```text
https://<project-ref>.supabase.co/auth/v1/.well-known/jwks.json
```

The issuer value must match the issuer in the project's actual access tokens. Confirm
both issuer and audience from a real access token before configuring the service; do
not infer the project reference, issuer, audience, or signing algorithm.

## Accepted token trust

- JWT signature is verified against a JWKS key selected by `kid`.
- Accepted asymmetric algorithms are RS256 (RSA-2048 or stronger) and ES256 (P-256).
- Issuer, audience, expiry, not-before, and client/subject identity are validated.
- JWKS redirects are not followed, responses are bounded, and keys are cached for a
  bounded TTL. Unknown `kid` values do not trigger a network refresh while the cache
  is fresh.
- The server warms the JWKS endpoint at startup. If the keyset cannot be loaded, the
  HTTP listener does not start.
- HS256 shared-secret tokens are not accepted by this JWKS validator. Supabase must
  be configured to use an asymmetric signing key and publish its public key.

## Identity-to-grant binding

The verified token identity returned by the validator (client ID when present,
otherwise subject) is bound to both request `agent_id` and `action.agent_id`.
If request fields are omitted, the server fills them from the verified identity.
A mismatch is rejected with HTTP 403.

Authorization does not trust caller-supplied `grant`, `policy`, or authorization
context objects. The server clears those fields and calls
`RepositoryDecisionGateway.Authorize`, which resolves the canonical grant and active
policy from PostgreSQL and commits decision evidence to the ledger before returning
a usable decision. A capability outside the persisted grant is denied. Repository
or evidence failures return INDETERMINATE/HTTP 503 rather than ALLOW.

For Supabase user JWTs, the `sub` is usually the user UUID. The corresponding
Control Plane grant's `agent_id` must be provisioned to match that trusted identity
(or a validated service `client_id` when that is the identity being used). Do not
derive authorization identity from user-editable profile metadata.

## Route allowlist

- `GET /health/live`: liveness only.
- `GET /health/ready`: readiness; PostgreSQL, schema, and initialized JWKS trust must
  be available.
- `POST /api/control-plane/decision`: requires `Authorization: Bearer <JWT>`.

No other routes are registered in the production handler. A decision endpoint only
returns a policy decision; it is not a protected side-effect execution endpoint.

## Operator safety

This implementation adds a server command and source-level authentication/authorization
wiring. It does not apply Supabase migrations, create grants, change Railway variables,
deploy the service, or perform a production cutover. These steps require separate
operator review and approval.

References:
- [Supabase JWT signing keys](https://supabase.com/docs/guides/auth/signing-keys)
- [Supabase JWT verification guide](https://supabase.com/docs/guides/auth/jwts)
