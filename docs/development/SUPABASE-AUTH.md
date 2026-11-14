# Supabase Auth for the NAEOS API

The NAEOS API can validate Supabase Auth access tokens when started with
`naeos api --auth` and `SUPABASE_URL` is configured.

## Supported signing keys

The integration accepts asymmetric JWT signing keys published by Supabase:
**RS256** with RSA keys of at least 2048 bits, and **ES256** with P-256 keys.
It does not accept HS256 tokens. Supabase recommends ES256, as published by Supabase. See the
[Supabase signing keys guide](https://supabase.com/docs/guides/auth/signing-keys).

## Configuration

Set these variables in the API service environment:

- `SUPABASE_URL`: project URL, such as `https://<project-ref>.supabase.co`.
  The API derives issuer `<SUPABASE_URL>/auth/v1` and JWKS URL
  `<issuer>/.well-known/jwks.json`.
- `SUPABASE_JWKS_URL`: optional override for the JWKS URL.
- `SUPABASE_AUTH_AUDIENCE`: optional audience override; defaults to
  `authenticated`.

The project URL and JWKS endpoint must use HTTPS. The configured URL is a trust
anchor; the validator does not follow redirects.

Start the API with `naeos api --auth`. The configured audience, issuer,
expiration, not-before time, and asymmetric signature are validated. Tokens with
`role` other than `authenticated` are rejected. Repository routes continue
to require the separate RS256 service-to-service JWT configuration and
repository allowlist.

## NAEOS role mapping

NAEOS reads role names only from signed `app_metadata.naeos_roles`. This claim
must be managed through a trusted Supabase administrative path; do not put
authorization roles in user-editable `user_metadata`. Only role names present
in NAEOS's RBAC registry are accepted. When there is no valid mapped role, the
user receives the `viewer` role.

Supabase embeds metadata in access tokens, so a role change takes effect when
the user presents a refreshed token. Supabase documents the distinction between
`app_metadata` and `user_metadata` in its
[JWT claims reference](https://supabase.com/docs/guides/auth/jwt-fields).
