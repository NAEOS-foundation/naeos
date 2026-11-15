# Docker

Copy `.env.example` to `.env`. Generate two different random values (for
example, with `openssl rand -hex 32`) and set `NAEOS_API_JWT_SECRET` and
`NAEOS_ENCRYPTION_KEY`. Compose refuses to start if either value is empty.
Keep `.env` private; users with access to the Docker daemon can inspect
container environment variables. Configure the RS256 JWKS issuer, audience,
and URL, and set `NAEOS_GITHUB_ALLOWED_REPOSITORIES` to the exact repositories
Intelligence may read. Set `GITHUB_TOKEN` only when GitHub access needs private
repository permissions or a higher rate limit.

Build and start the local Core API:

```sh
docker compose build naeos-core
docker compose up -d naeos-core
```

The API listens on `http://localhost:8081`. Compose enables user auth and
passes the JWT secret, service verifier, and repository allowlist into the
container. The `naeos api --secret` flag remains available for direct CLI use;
the Docker service reads `NAEOS_API_JWT_SECRET` from its environment instead
of exposing it in the process arguments. The service repository route remains
restricted to the RS256 service scope and the configured allowlist. To build
only the CLI image, run
`docker build -t naeos:local .`. The multi-platform release build is
available through `make docker` and pushes the image, so it is not a local-only
build command.
