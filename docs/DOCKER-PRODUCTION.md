# Production Docker deployment

`Dockerfile.control-plane` builds and runs the NAEOS Core REST API. It does not
run the investor demo. The container listens on port `8080`, enables JWT auth,
uses a non-root account, and stores local state under `/home/naeos`.

## Build

Build from the repository root and tag the image with the source commit:

```sh
docker build -f Dockerfile.control-plane -t naeos-core:$(git rev-parse --short HEAD) .
```

## Required configuration

Inject secrets from the deployment platform's secret manager. Do not bake them
into the image or commit them to the repository.

- `SUPABASE_URL`: the Supabase project URL. Use asymmetric JWT signing keys
  (ES256 or RS256); legacy HS256 project signing secrets are not accepted.
- `SUPABASE_JWT_AUDIENCE`: optional; defaults to `authenticated`.
- `SUPABASE_JWKS_URL`: optional override; defaults to
  `<SUPABASE_URL>/auth/v1/.well-known/jwks.json`.
- `NAEOS_ENCRYPTION_KEY` or `NAEOS_ENCRYPTION_KEY_FILE`: set a separate value
  of at least 32 bytes if application data encryption is used.

The image starts with `--auth-mode supabase` and refuses to start when the
Supabase URL or HTTPS JWKS cannot be configured. Access tokens must have the
expected issuer, audience, expiry, issued-at claim, and the `authenticated`
Supabase role. NAEOS roles are read only from trusted `app_metadata.naeos_roles`;
only `admin`, `developer`, and `viewer` are recognized, with `viewer` as the
least-privilege default. Routes without an explicit RBAC mapping are denied.
The API does not seed a local admin account in Supabase mode.
Tenant-scoped pipeline and schema routes use the trusted
`app_metadata.naeos_tenant_id` claim; a conflicting `X-Tenant-ID` header is
rejected. The artifact endpoint remains unavailable in Supabase mode because
the current artifact store does not partition data by tenant.

The CLI still supports the legacy HS256 mode for controlled deployments by
setting `--auth-mode hs256` (or `NAEOS_API_AUTH_MODE=hs256`),
`NAEOS_API_JWT_SECRET`, and a distinct
`NAEOS_ENCRYPTION_KEY`; both secrets must be at least 32 bytes. This mode does
not validate an issuer or audience and is not the Docker image default.

The image defaults to a SQLite database at
`/home/naeos/.naeos/naeos.db`. Keep `/home/naeos` on a persistent volume.
The supported deployment is one API instance. PostgreSQL may be used instead
of SQLite:

- `NAEOS_DB_DRIVER=postgres`
- `NAEOS_DB_HOST`, `NAEOS_DB_PORT` (defaults to `5432`), `NAEOS_DB_USER`
- `NAEOS_DB_PASSWORD`, `NAEOS_DB_DATABASE`
- `NAEOS_DB_SSLMODE=verify-full` (default)

Use a PostgreSQL certificate valid for the configured hostname. The container
also writes artifacts, pipelines, auth data, and audit records under its home
directory, so retain the volume even when using PostgreSQL. PostgreSQL alone
does not make the API stateless or make multiple replicas safe; artifacts and
other runtime state still use local files. Do not scale beyond one replica
until those stores and tenant isolation are made shared and verified.
Supply the database password through `NAEOS_DB_PASSWORD` or the mounted-file
alternative `NAEOS_DB_PASSWORD_FILE`; do not set both.

## Backup, restore, and key rotation

Back up the complete `/home/naeos` volume, not just the SQLite database. Stop
the API before taking a filesystem-level SQLite backup, and test restores into
an isolated environment. When using PostgreSQL, back up the database with the
managed database's supported procedure and separately back up the API volume.

The JWKS cache refreshes every five minutes and refreshes early when it sees an
unknown key ID. Rotate Supabase signing keys using the provider's documented
procedure, retaining old verification keys until tokens signed with them have
expired. Never replace the public-key flow with a project signing secret.

The builder and runtime base images are pinned by digest in
`Dockerfile.control-plane`. Review and update those digests on a regular
security-maintenance cadence, then rebuild and rerun the API, Docker, and
dependency-license checks before publishing a new image.

## Network and runtime

Terminate TLS at a trusted ingress or reverse proxy and keep the container on
a private network. Do not publish port `8080` directly to the public internet.
The image healthcheck probes `GET /healthz` for liveness. Configure a separate
readiness probe for `GET /readyz`; it returns `503` when a configured database
health check fails.

Example for a pre-created private Docker network, a secret env file, and a
persistent named volume:

```sh
docker volume create naeos-core-data
docker run -d \
  --name naeos-core \
  --restart unless-stopped \
  --network naeos-private \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  --volume naeos-core-data:/home/naeos \
  --env-file /etc/naeos/core.env \
  naeos-core:$(git rev-parse --short HEAD)
```

Attach the TLS ingress to `naeos-private` and route it to `naeos-core:8080`.
Restrict `/etc/naeos/core.env` to the deployment account.

## Authentication limitation

Supabase verifies user identity but does not by itself establish NAEOS tenant
membership or grant application roles. Assign NAEOS roles through trusted
Supabase `app_metadata`, and set `naeos_tenant_id` there before relying on
tenant-scoped routes. Never place authorization claims in user-editable
`user_metadata`. Review each route's RBAC mapping before exposing it. Keep the
API behind TLS ingress and rotate Supabase signing keys using the provider's
documented procedure.
