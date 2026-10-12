# NAEOS Environment Separation

This document defines the intended deployment boundary for production, staging, and development. The environments must not share credentials, writable data stores, or production side effects.

## Environment map

| Environment | Git branch | Railway project / services | Cloudflare Worker | Purpose |
|---|---|---|---|---|
| Production | `main` | Existing `pleasing-blessing` project: `naeos-control-plane`, `naeos-crm-api`, and production Postgres | `naeos`, serving the canonical `naeos.dev` domains | Live traffic and approved releases |
| Staging | `staging` | `naeos-sandbase-staging`: `naeos-control-plane-staging` and `naeos-sandbase-authz-staging` | `naeos-staging`, workers.dev only | Release-candidate validation and external integration tests |
| Development | `develop` | `naeos-development`: `naeos-control-plane-development` | `naeos-development`, workers.dev only | Active development and disposable test data |

The Railway Control Plane services use separate per-environment API tokens and JWT signing secrets. Never copy production secrets or production database credentials into staging or development. SandBase authorization remains disabled until an approved policy/grant config and the shared hook token are deliberately configured for that environment.

## Promotion flow

1. Feature branches open pull requests into `develop`; CI must pass.
2. Merge to `develop` deploys the development Worker and development Control Plane.
3. Promote a tested candidate by pull request from `develop` to `staging`; verify the deployed staging commit and run integration acceptance tests.
4. Promote the exact accepted candidate from `staging` to `main`; production deploys only from `main`.
5. Do not make environment-specific hotfixes directly in `main` without back-merging them into `staging` and `develop`.

The long-lived `staging` and `develop` branches were initialized from `main`. Before the first release using this flow, reconcile them with current `main` and confirm branch protections in GitHub. Branch-protection status could not be verified through the current GitHub connection.

## Isolation rules

- Production custom domains are attached only to the top-level `naeos` Worker. Named Workers use `workers.dev` and have no production routes.
- Each website build must set `NEXT_PUBLIC_CONTROL_PLANE_API_URL` to its matching Railway API endpoint. PR previews use the development endpoint; never silently fall back to production.
- Website builds select these isolated API endpoints by branch: staging uses `https://naeos-control-plane-staging-staging.up.railway.app`; development and PR previews use `https://naeos-control-plane-development-development.up.railway.app`. Verify both endpoints and update the URLs if Railway regenerates them.
- The newsletter KV namespace is currently provisioned only by the production workflow. Staging/development intentionally do not bind the production KV namespace; provision separate KV namespaces before testing newsletter writes there.
- Keep SandBase tokens, JWT secrets, API tokens, and database credentials unique per environment. Do not place secrets in `wrangler.jsonc`, workflow YAML, issues, or logs.
- Do not mark an environment healthy or an integration E2E PASS without a successful deployment, live request evidence, and an independent check.

## Current limitations

- Production Railway services remain on `main`.
- Staging and development Control Plane deployments are reporting SUCCESS in Railway. DNS/HTTP reachability and browser CORS still need independent verification before calling the complete website-to-API path healthy.
- This configuration creates persistent staging/development Workers; pull-request previews remain ephemeral.
