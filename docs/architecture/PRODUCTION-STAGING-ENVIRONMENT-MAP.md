# Production and Staging Environment Map

**Status:** repository/configuration audit; not a live infrastructure attestation  
**Last reviewed:** 2026-10-25

This document separates what the repository configures from what still needs to be verified in the Cloudflare and Railway dashboards. A preview deployment is not considered an isolated staging environment unless its upstream services, credentials, data stores, and side effects are isolated too.

## 1. Current deployment map

```text
GitHub: NAEOS-foundation/naeos
│
├── pull request targeting main
│   └── .github/workflows/website.yml
│       ├── typecheck + policy/authorization/publisher tests
│       ├── build site with NEXT_PUBLIC_CONTROL_PLANE_API_URL
│       │   = https://naeos-control-plane-production.up.railway.app
│       └── Wrangler "preview" → Cloudflare Worker Preview URL
│           └── Control Plane server route currently forwards to
│               the hard-coded PRODUCTION Railway endpoint
│
└── push to main / manual dispatch on main
    └── .github/workflows/website.yml
        ├── build site with production Control Plane URL
        ├── optionally resolve/create NEWSLETTER_KV namespace
        └── Wrangler "deploy" → Cloudflare Worker "naeos"
            ├── naeos.dev
            ├── www.naeos.dev
            ├── docs.naeos.dev
            └── registry.naeos.dev
                │
                └── POST /api/control-plane/decision
                    └── https://naeos-control-plane-production.up.railway.app
                        (separate Railway service: naeos-control-plane)
```

Sources:
- [Website workflow](../.github/workflows/website.yml)
- [Cloudflare Worker config](../site/wrangler.jsonc)
- [Control Plane proxy route](../site/app/api/control-plane/decision/route.ts)
- [Control Plane live proof](control-plane/live-proof.md)
- [Separate SandBase Railway deployment runbook](sandbase-railway-deployment.md)

## 2. Environment inventory

| Component | Production path/configuration | Preview/staging status | Confidence |
|---|---|---|---|
| Website | Cloudflare Worker named `naeos`; custom domains listed in `site/wrangler.jsonc`; deploy is restricted to `main` push or manual dispatch on `main` | PR workflow runs Wrangler `preview`; preview URLs are enabled in Worker config | Config verified; live URL/health not verified here |
| Control Plane demo API | `https://naeos-control-plane-production.up.railway.app` | No distinct staging API URL is wired in the website workflow or proxy route inspected | Production target verified in source; Railway runtime state not verified |
| Control Plane credentials | Proxy reads `NAEOS_CONTROLPLANE_API_TOKEN` from Cloudflare runtime environment and fails with 503 if missing | Need to confirm whether preview has its own token/secret and whether that token is accepted by the production upstream | Source behavior verified; dashboard secret scopes not verified |
| Newsletter storage | `NEWSLETTER_KV` binding; workflow attempts to resolve/create a namespace only on `main` and writes its ID into `wrangler.jsonc` before production deploy | No separate preview KV namespace is declared in the checked-in Worker config (`kv_namespaces: []`) | Source/config verified; live binding and persistence not verified |
| SandBase authorization daemon | Runbook describes a separate Railway service using `Dockerfile.sandbase-authz`; must not replace `naeos-control-plane` | Runbook says live integration requires real staging config and a successful Evidence Run; that is not proof it is deployed or passing | Runbook verified; live service status not verified |
| CLI cloud/deploy environments | CLI docs expose environment labels such as `dev`, `staging`, and `prod` | These labels are CLI configuration options, not evidence of a deployed website/API staging stack | Docs only |

## 3. Key finding: Preview is not backend-isolated

The website workflow builds both PR previews and production with:

```yaml
NEXT_PUBLIC_CONTROL_PLANE_API_URL: https://naeos-control-plane-production.up.railway.app
```

Additionally, `site/app/api/control-plane/decision/route.ts` hard-codes the same production Railway origin for its server-side upstream request. Therefore, creating a Cloudflare Worker Preview URL does **not** create a staging Control Plane. Requests made from a preview can reach the production Control Plane route, subject to its authentication/configuration and policy behavior.

The live-proof documentation says the demo endpoint evaluates authorization and records a decision; it does not execute the requested side effect. That reduces the intended side-effect scope, but it does not make production data, audit records, rate limits, credentials, or service availability isolated from preview traffic.

## 4. Staging readiness gaps

1. **No explicit staging Control Plane origin** is configured for PR builds.
2. **Server proxy origin is hard-coded**, so changing only a build-time public variable would not isolate the server-side API route.
3. **Preview secrets and policy** are not verifiable from source. Confirm the preview Worker has a deliberately scoped token, and do not reuse or expose a production secret to untrusted PRs.
4. **Newsletter persistence is not clearly isolated.** The checked-in config has no KV namespace. Confirm that preview cannot read/write production subscribers and that preview responses do not imply persistence when storage is absent.
5. **SandBase is a separate deployment concern.** The runbook explicitly says not to repoint the existing `naeos-control-plane` Railway service. It also states that the staging integration is not complete until a real SandBase runtime and Evidence Run pass.
6. **Live deployment state remains unverified.** Repository workflow/configuration is not sufficient to confirm current Cloudflare deployment health, Railway service health, secret values, KV binding IDs, or data-store isolation.

## 5. Recommended target boundary

```text
PRODUCTION
  Cloudflare Worker: naeos + production domains
      └── production Control Plane Railway service
          └── production credentials / ledger / data

PR PREVIEW
  isolated Cloudflare Worker Preview URL
      └── no production upstream by default
          └── either:
              A. explicit non-production Control Plane endpoint,
                 with isolated credentials and test data; or
              B. fail closed / disable live Control Plane actions
                 until staging exists
      └── isolated preview KV, or newsletter persistence disabled
```

A preview must not fall back silently to production. If a staging API is not available, the safer default is to make the preview UI buildable while the live API proxy returns a clear unavailable response, rather than sending preview traffic to production.

## 6. Verification checklist

- [ ] Confirm current Cloudflare Worker production deployment, custom-domain routes, and health.
- [ ] Confirm whether the Cloudflare account has a distinct staging/preview Worker and its actual preview URL.
- [ ] Confirm Railway production Control Plane health and the intended persistence/ledger backing.
- [ ] Decide whether staging will use a separate Railway service or whether PR previews should disable the live API.
- [ ] Move the server-side upstream origin to an environment-specific, server-only binding/configuration; do not use `NEXT_PUBLIC_*` for credentials.
- [ ] Give staging separate credentials, policy/grant identities, ledger, and test data.
- [ ] Verify preview newsletter writes cannot touch production KV; bind a separate namespace or disable persistence.
- [ ] Confirm PR previews do not receive production secrets, particularly for untrusted forks/Dependabot.
- [ ] Run a smoke test against the preview URL and assert the resolved upstream environment in a safe, non-secret way.
- [ ] For SandBase, use the separate service and complete the documented real staging Evidence Run before calling it live.

## 7. Scope and safety

This is a documentation-only map. It does not modify Cloudflare, Railway, DNS, secrets, databases, or deployment workflows. No production/staging health check was performed as part of this source review, and no live environment is claimed to be healthy based on repository configuration alone.
