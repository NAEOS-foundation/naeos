# Control Plane Docker Local Staging

Railway deployment is deferred for this internal staging validation. This compose
configuration runs the dedicated production-style Control Plane runtime locally,
not the investor demo.

## Safety boundaries

- Host bind is localhost only: `127.0.0.1:8080`.
- Environment must be `staging`; runtime target validation is fail-closed.
- Principal kind is explicitly required as `human` or `service`; never inferred from `client_id`.
- Use only a dedicated staging PostgreSQL login scoped to the private
  `naeos_control_plane` schema in `naeos.db`.
- Never use the CRM production database, CRM credentials, Supabase
  `service_role` API key, or a production token.
- Compose does not apply migrations, seed policy/grants, or deploy anything.
- Container is read-only, non-root (from Dockerfile), drops Linux capabilities,
  and enables no-new-privileges.
- Readiness health check requires database/schema and JWKS initialization.
- Do not commit the local `.env` file.

## Configuration

1. Copy `.env.control-plane.local.example` to `.env` (the root `.env` is gitignored).
2. Populate all six configuration values from verified staging settings and the dedicated staging
   runtime role. Set `NAEOS_CONTROLPLANE_PRINCIPAL_KIND` to exactly `human` or `service`, based on
   the actual token profile. Do not infer it from `client_id`; do not paste secrets into chat or commit them.
3. Confirm issuer, audience, and JWKS URL from actual staging Auth configuration
   and token claims; do not guess.
4. Confirm the database URL is staging-only and TLS is required (sslmode=require,
   verify-ca, or verify-full).

## Build and start

```sh
docker compose -f compose.control-plane.local.yml config
docker compose -f compose.control-plane.local.yml build
docker compose -f compose.control-plane.local.yml up -d
docker compose -f compose.control-plane.local.yml ps
docker compose -f compose.control-plane.local.yml logs --tail=100 control-plane
```

The Compose health check targets `GET /health/ready`. The runtime must remain
unhealthy until PostgreSQL, schema/role checks, and JWKS warmup succeed.

## Read-only preflight before running

```sh
docker compose -f compose.control-plane.local.yml run --rm control-plane \
  /usr/local/bin/naeos control-plane preflight
```

Note: `run` uses the configured entrypoint command; if the image command override
needs adjustment, use the local built image directly with `docker run --rm` and
the same environment, without publishing ports.

## Stop / cleanup

```sh
docker compose -f compose.control-plane.local.yml down
```

This stops/removes the local container and network only. It does not remove or alter
Supabase data. Do not use `down -v` because this compose file defines no database
volume and staging database state must be preserved.
