# Evidence Verifier durable replay guard

The verifier requires PostgreSQL-backed one-time request claims. In-memory replay protection is not supported.

## Isolated staging/test database

Apply `supabase/migrations/20261127000100_evidence_verifier_replay_guard.sql` only to an isolated verifier staging/test database. Never run it against production. Provision a dedicated login role via a secure administrative session and grant it membership in `naeos_evidence_verifier_runtime`. Do not reuse the Control Plane role or Supabase service-role credentials. The runtime startup checks TLS, schema/table existence, schema USAGE, INSERT-only table permissions, and role membership/attributes.

Example role provisioning (run only in the isolated verifier test/staging database; supply a managed secret, never commit it):

```sql
CREATE ROLE naeos_evidence_verifier_login LOGIN INHERIT PASSWORD '<managed-secret>';
GRANT naeos_evidence_verifier_runtime TO naeos_evidence_verifier_login;
```

## Required configuration

- `NAEOS_EVIDENCE_SOURCE_URL` — authoritative read-only HTTPS observation endpoint
- `NAEOS_EVIDENCE_SOURCE_TOKEN` — source credential
- `NAEOS_EVIDENCE_VERIFIER_TOKEN` — verifier service credential
- `NAEOS_EVIDENCE_VERIFIER_DATABASE_URL` — dedicated PostgreSQL URL with TLS sslmode=require/verify-ca/verify-full
- `NAEOS_EVIDENCE_VERIFIER_SCOPE` — stable verifier scope (1–128 characters)
- `NAEOS_EVIDENCE_POLICY_VERSION` — explicit policy version
- `NAEOS_EVIDENCE_VERIFIER_ADDR` — listener; defaults to loopback `127.0.0.1:8081`

Each request ID is normalized to lowercase and claimed in the isolated PostgreSQL table before source observation. Duplicate claims return HTTP 409. Database errors/timeouts return HTTP 503 and never call the observation source. Claims are intentionally retained; no TTL/release behavior is implemented. A failed observation/policy evaluation still consumes the request ID.

## Tests

Unit tests cover concurrent single-winner claim behavior and service duplicate/storage-failure paths. PostgreSQL integration test is opt-in with `NAEOS_EVIDENCE_VERIFIER_TEST_DATABASE_URL`; it requires the dedicated test migration and runtime role and verifies concurrent single-winner, duplicate persistence across separate pools, and scope isolation. Use an isolated ephemeral test database only.
