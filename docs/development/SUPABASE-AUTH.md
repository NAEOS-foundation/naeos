# Supabase Auth for the NAEOS API

The NAEOS API can validate Supabase Auth access tokens when started with
`naeos api --auth` and `SUPABASE_URL` is configured.

## Signing key requirement

This integration currently accepts **RS256 tokens signed with an RSA key of at
least 2048 bits**. Configure the Supabase project's Auth signing key accordingly.
It does not accept ES256 or HS256 tokens. Supabase documents ES256 as a supported
and recommended signing algorithm, so projects using that default must choose an
RSA signing key before using this integration. See the
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
expiration, not-before time, and RS256 signature are validated. Tokens with
`role` other than `authenticated` are rejected. Repository routes continue
to require the separate service-to-service JWT configuration and repository
allowlist.

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
