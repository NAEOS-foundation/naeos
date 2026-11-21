# Production Control Plane Docker deployment

`Dockerfile.control-plane` builds the dedicated production Control Plane and
runs `naeos control-plane serve`; it does not run `naeos api` or the investor
demo. The service listens on port `8080`, runs as a non-root user, and requires
its PostgreSQL persistence and JWKS identity configuration before it starts.

## Build

Build from the repository root and tag the image with the source commit:

```sh
docker build -f Dockerfile.control-plane -t naeos-control-plane:$(git rev-parse --short HEAD) .
```

## Required configuration

Inject secrets from the deployment platform's secret manager. Do not bake them
into the image or commit them to the repository.

- `NAEOS_CONTROLPLANE_DATABASE_URL`: dedicated Control Plane PostgreSQL URL with TLS.
- `NAEOS_CONTROLPLANE_ISSUER`: exact issuer from the access token.
- `NAEOS_CONTROLPLANE_AUDIENCE`: exact expected token audience.
- `NAEOS_CONTROLPLANE_JWKS_URL`: HTTPS endpoint publishing the signing keys.
- `NAEOS_CONTROLPLANE_ADDR`: optional listen address; defaults to `:8080`.

For Supabase, verify the issuer and audience against actual project access
tokens before configuring the service. The runtime validates asymmetric JWKS
signatures and fails startup if the database or JWKS trust configuration is
unavailable. It does not fall back to `DATABASE_URL`, local SQLite, or an
in-memory store. The service uses a dedicated Control Plane database role and
private schema; do not reuse the CRM application's database credentials.

See [Production Control Plane JWKS Authentication](./control-plane/production-jwks-authentication.md)
for token validation, route allowlist, and identity-to-grant behavior. The
runtime does not apply database migrations or perform a production cutover;
those actions require separate review and approval.

## Health and network

The image healthcheck probes `GET /health/live`. Configure an orchestrator
readiness probe for `GET /health/ready`; readiness requires PostgreSQL, schema,
and initialized JWKS trust to be available. The only application route is
`POST /api/control-plane/decision`, which requires a bearer token.

Terminate TLS at a trusted ingress or reverse proxy and keep the container on
a private network. Do not expose port `8080` directly to the public internet.
Use separate database credentials and configuration per environment; never
use production credentials in staging or preview.

## Generic REST API CLI

The standalone `naeos api` command is a separate REST API and supports Supabase
JWT verification, but it is not the entrypoint built by this Dockerfile. See
[naeos api](./cli/naeos_api.md) for its authentication flags, tenant claim
handling, and endpoint limitations.
