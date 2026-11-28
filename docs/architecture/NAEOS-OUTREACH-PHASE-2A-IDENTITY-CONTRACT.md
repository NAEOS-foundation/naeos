# NAEOS Outreach — Phase 2A Identity Contract

**Status:** Draft implementation contract / review required  
**Scope:** Identity schema and validation only. No authentication, authorization, database, provider, or deployment changes.

## Decision boundary

- Credentials are verified only by a future trusted identity adapter.
- Canonical identity key is the issuer + subject pair, consistent with NAEOS identity ADRs.
- Request body, query string, user-editable metadata, and caller-supplied role/capability fields are not identity authority.
- Identity is distinct from tenant membership and authorization.
- Tenant, roles, and capabilities are resolved separately by trusted server-side membership/policy services. They are not part of this v1 principal.
- Unsupported contract versions and invalid/missing issuer or subject fail closed.
- Phase 2A does not connect this contract to HTTP middleware; existing fail-closed middleware remains unchanged.

## Contract

Version: v1

Principal fields:
- issuer: non-empty exact issuer string from verified identity source.
- subject: non-empty exact subject string from verified identity source.

Validation rejects missing/blank values and surrounding whitespace. Issuer/subject are not normalized, to avoid identity collisions.

## Test requirements

- valid issuer + subject accepted
- missing issuer rejected
- missing subject rejected
- whitespace-only values rejected
- surrounding whitespace rejected
- issuer-qualified identity retained exactly
- authorization claims are not part of this principal type

## Deferred to Phase 2B/2C

Trusted adapter implementation and credential verification; policy/capability mapping; tenant membership and resource ownership; middleware integration; protected endpoint harness; forged identity and handler-not-executed tests.
