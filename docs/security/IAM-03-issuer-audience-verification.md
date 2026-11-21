# IAM-03 — Supabase issuer and audience verification

**Status:** BLOCKED — runtime configuration and live-token evidence required
**Scope:** Staging only. This document does not authorize Production changes.

## Trust contract

NAEOS must validate Supabase access tokens against explicit trust anchors:
- Exact issuer (iss) configured for the selected Supabase Auth project.
- Expected audience (aud); default Supabase access-token audience is commonly authenticated, but must be confirmed against the selected project's live token configuration.
- Project JWKS endpoint belonging to the same Auth project.
- Allowed signing algorithm and key identifier (kid) from the trusted JWKS.

The validator must reject issuer mismatch, audience mismatch, unknown/untrusted kid, invalid signature, expired token, not-yet-valid token, and malformed claims. A human principal is issuer-qualified (iss, sub); client_id remains a distinct claim.

## Required evidence (do not paste credentials or tokens)

- [ ] Selected Supabase Auth project is approved and its project ref recorded in protected staging configuration.
- [ ] Issuer value is read from the project's trusted OIDC discovery metadata and recorded (non-secret).
- [ ] JWKS URL is verified to belong to that same project; fetched JWKS contains acceptable signing keys and expected kid.
- [ ] Audience expectation is confirmed from a locally obtained staging access token; record only the claim value, never the token.
- [ ] Local staging token validation succeeds for the correct issuer/audience/key.
- [ ] Negative tests reject a token with wrong issuer.
- [ ] Negative tests reject a token with wrong audience.
- [ ] Negative tests reject unknown kid, expired token, and invalid signature.
- [ ] Evidence links, environment name, timestamp, verifier, and commit SHA are recorded.

## Verification procedure

1. Use the protected staging environment only. Confirm SUPABASE_STAGING_PROJECT_REF matches the hostname in SUPABASE_URL.
2. Inspect the project's OIDC discovery document and JWKS endpoint over HTTPS. Do not accept redirects to a different host.
3. Obtain a staging access token locally via the approved Supabase client flow. Do not commit, log, share, or paste the token into an issue/chat.
4. Decode claims locally for inspection (decoding alone is not verification). Record only iss, aud, sub presence, exp/nbf validity, algorithm, and kid; redact the subject value if it can identify a person.
5. Run the validator and negative-case tests using controlled fixtures.
6. Attach sanitized CI/test evidence and mark the control complete only after an authorized reviewer confirms configuration/evidence.

## Current implementation review

- internal/api/jwks.go checks exact issuer, configured audience (string or array), signature, expiry, not-before, issued-at skew, and trusted kid.
- internal/api/supabase_auth.go additionally requires Supabase role authenticated and binds parsed issuer/subject to verified claims.
- internal/controlplane/production_auth.go accepts explicit issuer, audience, and JWKS URL; readiness remains unavailable until JWKS warmup succeeds.

## Outstanding decision / evidence

No live staging issuer, audience, JWKS keyset, token-validation run, or staging owner evidence has been supplied in this PR context. Therefore IAM-03 remains open. Do not infer runtime trust values from code defaults or mark the control passed based on unit tests alone.

## Production gate

Production configuration changes, migration, or rollout require separate explicit approval and documented staging verification, rollback rehearsal, and security review.