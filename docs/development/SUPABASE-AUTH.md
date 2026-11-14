# Supabase Auth with the NAEOS API

NAEOS validates Supabase Auth access tokens on protected API routes. The
integration uses Supabase for sign-in and session issuance, then maps signed
user identity and trusted NAEOS roles into the NAEOS API authorization layer.

## 1. Configure the API

Set `SUPABASE_URL` to the project URL in the API service environment, then
start the API with authentication enabled:

```sh
SUPABASE_URL=https://<project-ref>.supabase.co naeos api --auth
```

The API derives issuer `<SUPABASE_URL>/auth/v1` and JWKS URL
`<issuer>/.well-known/jwks.json`. `SUPABASE_JWKS_URL` and
`SUPABASE_AUTH_AUDIENCE` can override those defaults. The configured URLs
must use HTTPS.

The API validates issuer, audience, expiry, not-before time, token role,
subject, and signature. It accepts Supabase asymmetric **RS256** (RSA keys of
at least 2048 bits) and **ES256** (P-256) signing keys. HS256 is not supported.
Repository routes retain their separate RS256 service-token and repository
allowlist checks.

## 2. Sign in and call the API

For a browser or application client, use the Supabase client SDK with the
project publishable key. The SDK handles the sign-in flow and session refresh.
Pass its access token to NAEOS as a bearer token:

```ts
const { data, error } = await supabase.auth.signInWithPassword({
  email,
  password,
});
if (error) throw error;

const response = await fetch(`${NAEOS_API_URL}/api/v1/profiles`, {
  headers: {
    Authorization: `Bearer ${data.session.access_token}`,
  },
});
```

The same API accepts access tokens issued through any configured Supabase Auth
provider. Do not send the Supabase publishable key as the bearer token.

For the NAEOS CLI, configure the same project and sign in:

```sh
naeos supabase init --project-ref <project-ref> \
  --anon-key "$SUPABASE_PUBLISHABLE_KEY" \
  --service-role-key "$SUPABASE_SERVICE_ROLE_KEY"
naeos supabase auth signin --email <email> --password '<password>'
```

The publishable key is used for sign-in. Add the service-role key only in a
trusted admin environment if you will assign roles with the CLI.

The CLI saves the user session separately under `~/.naeos/supabase/session.json`
with mode `0600`, refreshes it when an auth request finds it near expiry, and
clears it after sign-out. To intentionally print the current bearer token for
another client, run `naeos supabase auth access-token`; treat stdout as a
credential and keep it out of logs and shell history.

## 3. Map Supabase users to NAEOS roles

NAEOS trusts roles only from the signed
`app_metadata.naeos_roles` claim. It ignores `user_metadata` for
authorization, filters role names against the NAEOS RBAC registry, and grants
`viewer` if there is no valid NAEOS role.

Use the CLI admin operation to preserve existing app metadata while setting
the NAEOS roles:

```sh
naeos supabase auth admin set-naeos-roles \
  --user-id <supabase-user-id> \
  --role developer
```

This command requires the Supabase service-role key in the local CLI
configuration. Keep that key in a trusted admin environment only; it is never
needed by the NAEOS API or browser clients. A user must refresh their Supabase
session before the updated app metadata appears in their access token.

## Session and revocation behavior

The NAEOS API verifies JWTs locally using the project's public JWKS. A token
already issued by Supabase can therefore remain usable at NAEOS until its
expiration, even after Supabase sign-out or session revocation. Keep access
token lifetimes appropriate for the application. The validator caches JWKS for
five minutes and does not follow redirects when fetching keys.

Supabase references:
- [Signing keys](https://supabase.com/docs/guides/auth/signing-keys)
- [JWT claims](https://supabase.com/docs/guides/auth/jwt-fields)
- [User sessions](https://supabase.com/docs/guides/auth/sessions)
