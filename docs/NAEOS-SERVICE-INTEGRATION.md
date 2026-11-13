# NAEOS service integration

## Read-only repository metadata

NAEOS Core serves `GET /api/v1/repositories/{owner}/{name}` for authenticated clients that present an RS256 service JWT with `naeos:repository:read`. The endpoint reads metadata from GitHub, requires an exact owner/repository entry in `NAEOS_GITHUB_ALLOWED_REPOSITORIES`, and records successful reads in the Core audit log. It does not write to GitHub.

The API keeps the existing HS256 user-token path. Configure `NAEOS_API_JWKS_ISSUER`, `NAEOS_API_JWKS_AUDIENCE`, and `NAEOS_API_JWKS_URL` to enable the additional service JWT verifier. This enables authentication for the repository route only; use `naeos api --auth --secret ...` to enable the existing user authentication on other API routes. A service token is restricted to GET on the repository route and must carry the repository-read scope. Set `GITHUB_TOKEN` to a least-privilege GitHub token if private repositories or higher API rate limits are needed.

## CRM → Intelligence → Core

CRM's `POST /api/v1/intelligence/engineering-analysis` accepts an allowlisted repository reference from a CRM admin. CRM calls the protected Intelligence API using a service token scoped to `intelligence:write` and `intelligence:read`. Intelligence calls Core with a separate service JWT scoped only to `naeos:repository:read`. Both CRM and Core maintain their own repository allowlists. No CRM contact or prospect record is sent to Intelligence or Core.

### Minimum permission matrix

| Caller | Audience | Required scope | Allowed operation |
|---|---|---|---|
| CRM service | `naeos-intelligence` | `intelligence:write`, `intelligence:read` | Submit an analysis request and analyze that request |
| Intelligence service | `naeos-core` | `naeos:repository:read` | `GET /api/v1/repositories/{owner}/{name}` only |
| CRM human user | CRM-local policy | `intelligence:analyze` (admin only) | Request analysis for a CRM-allowlisted repository |

Do not grant repository writes, Core admin scopes, Intelligence cancellation, CRM contact access, or GO-Gate execution to either service identity. The audience must be distinct per API, and each service must validate the same issuer and JWKS published by the selected identity provider.

For local development, run Core on port 8081 with `naeos api --port 8081`; run Intelligence on port 8080. Set `NAEOS_INTELLIGENCE_CORE_API_BASE_URL=http://127.0.0.1:8081/api`. Configure the issuer, audience, and JWKS URL for both services with the appropriate audience-specific service tokens. Remote deployments must use HTTPS and a private service network.

The services remain fail-closed when JWKS configuration, service scopes, or either repository allowlist is absent or invalid.
