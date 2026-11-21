## naeos api

Start NAEOS REST API server

This generic REST API is distinct from the dedicated production Control Plane
container, which runs `naeos control-plane serve`.

### Synopsis

Start the NAEOS REST API server for external integrations and web dashboard.
With `--auth`, the default mode is legacy HS256. Configure a JWT secret using
`NAEOS_API_JWT_SECRET` or `--secret`, and set a distinct
`NAEOS_ENCRYPTION_KEY`; both values must be at least 32 bytes. `--secret` takes
precedence over the environment variable.
For mounted secrets, use `NAEOS_API_JWT_SECRET_FILE` and
`NAEOS_ENCRYPTION_KEY_FILE` instead of their value variables. Database
passwords can likewise be supplied using `NAEOS_DB_PASSWORD_FILE`. Do not set
both the value and file variable for the same secret.

For Supabase access-token verification, set `--auth-mode supabase` (or
`NAEOS_API_AUTH_MODE=supabase`) and `SUPABASE_URL`. The optional
`SUPABASE_JWT_AUDIENCE` defaults to `authenticated`; `SUPABASE_JWKS_URL`
defaults to `<SUPABASE_URL>/auth/v1/.well-known/jwks.json`. Supabase must use
asymmetric ES256 or RS256 signing keys. The API does not accept a Supabase
legacy HS256 project signing secret in this mode.

```sh
SUPABASE_URL=https://project.example naeos api --auth --auth-mode supabase
```

In Supabase mode, NAEOS roles are read from trusted
`app_metadata.naeos_roles`; only `admin`, `developer`, and `viewer` are
recognized, and the default is `viewer`. Routes without explicit RBAC mappings
are denied. Do not put authorization claims in user-editable `user_metadata`.
Tenant-scoped pipeline and schema routes use
`app_metadata.naeos_tenant_id`; conflicting `X-Tenant-ID` headers are rejected.
The artifact endpoint is disabled in Supabase mode until artifact data is
partitioned by tenant.

```
naeos api [flags]
```

### Options

```
  -a, --auth            Enable JWT authentication
      --auth-mode string Authentication mode: hs256 or supabase
  -h, --help            help for api
  -p, --port string     API server port (default "8080")
  -s, --secret string   JWT secret key (defaults to NAEOS_API_JWT_SECRET)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

