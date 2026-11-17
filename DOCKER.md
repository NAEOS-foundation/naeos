# Docker

Copy `.env.example` to `.env`. Set `NAEOS_API_JWT_SECRET` to a strong
random value; the service refuses to start with `--auth` when neither this
secret nor Supabase Auth is configured. Configure the RS256 JWKS issuer,
audience, and URL, and set `NAEOS_GITHUB_ALLOWED_REPOSITORIES` to the exact
repositories Intelligence may read. Set `GITHUB_TOKEN` only when GitHub access
needs private repository permissions or a higher rate limit.

Build and start the local Core API:

```sh
docker compose build naeos-core
docker compose up -d naeos-core
```

The API listens on `http://127.0.0.1:8081` (loopback only). Compose enables user auth and
passes the service verifier, repository allowlist, and Supabase Auth settings
(`SUPABASE_URL`, `SUPABASE_AUTH_AUDIENCE`, `SUPABASE_JWKS_URL`) into the
container. When targeting a local Supabase stack over http, also set
`NAEOS_SUPABASE_ALLOW_INSECURE_LOOPBACK=true`; leave it disabled for any
deployed service. The service repository route remains restricted to the RS256
service scope and the configured allowlist. To build only the CLI image, run
`docker build -t naeos:local .`. The multi-platform release build is
available through `make docker` and pushes the image, so it is not a local-only
build command.
