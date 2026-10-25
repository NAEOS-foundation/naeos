# Production and Staging Environment Map

**Status:** repository/configuration audit and fail-closed preview isolation changes; not a live infrastructure attestation  
**Last reviewed:** 2026-10-25

This document separates what the repository configures from what still needs to be verified in the Cloudflare and Railway dashboards. A preview deployment is not considered an isolated staging environment unless its upstream services, credentials, data stores, and side effects are isolated too.

## 1. Current deployment map

```text
GitHub: NAEOS-foundation/naeos
│
├── pull request targeting main
│   └── .github/workflows/website.yml
│       ├── typecheck + policy/authorization/publisher tests
│       ├── build site with NEXT_PUBLIC_CONTROL_PLANE_API_URL empty
│       ├── Wrangler "preview" → Cloudflare Worker Preview URL
│       └── Preview upstream binding explicitly empty
│           └── Control Plane proxy fails closed (503); no production fallback
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
- [Website workflow](../../.github/workflows/website.yml)
- [Cloudflare Worker config](../../site/wrangler.jsonc)
- [Control Plane proxy route](../../site/app/api/control-plane/decision/route.ts)
- [Control Plane live proof](control-plane/live-proof.md)
- [Separate SandBase Railway deployment runbook](sandbase-railway-deployment.md)

## 2. Environment inventory

| Component | Production path/configuration | Preview/staging status | Confidence |
|---|---|---|---|
| Website | Cloudflare Worker named `naeos`; custom domains listed in `site/wrangler.jsonc`; deploy is restricted to `main` push or manual dispatch on `main` | PR workflow runs Wrangler `preview`; preview URLs are enabled in Worker config | Config verified; live URL/health not verified here |
| Control Plane demo API | `https://naeos-control-plane-production.up.railway.app` | PR preview has no upstream by default and proxy fails closed with 503; no staging API is provisioned by this change | Isolation config verified in source; live preview behavior still needs CI/runtime verification |
| Control Plane credentials | Proxy reads `NAEOS_CONTROLPLANE_API_TOKEN` from Cloudflare runtime environment and fails with 503 if missing | Preview does not need a token while its upstream is unset; confirm preview secrets are not configured to bypass intended isolation | Source behavior verified; Cloudflare dashboard secret scopes not verified |
| Newsletter storage | `NEWSLETTER_KV` binding; workflow attempts to resolve/create a namespace only on `main` and writes its ID into `wrangler.jsonc` before production deploy | No separate preview KV namespace is declared in the checked-in Worker config (`kv_namespaces: []`) | Source/config verified; live binding and persistence not verified |
| SandBase authorization daemon | Runbook describes a separate Railway service using `Dockerfile.sandbase-authz`; must not replace `naeos-control-plane` | Runbook says live integration requires real staging config and a successful Evidence Run; that is not proof it is deployed or passing | Runbook verified; live service status not verified |
| CLI cloud/deploy environments | CLI docs expose environment labels such as `dev`, `staging`, and `prod` | These labels are CLI configuration options, not evidence of a deployed website/API staging stack | Docs only |

## 3. Isolation fix included in this PR

The PR changes the environment boundary in three places:

1. The website workflow injects `NEXT_PUBLIC_CONTROL_PLANE_API_URL` only when the ref is `main`; PR builds receive an empty value.
2. `site/wrangler.jsonc` configures the production upstream in top-level `vars` and overrides it with an empty value under `previews.vars`.
3. The server-side proxy reads `NAEOS_CONTROLPLANE_UPSTREAM_URL` from the runtime binding and returns HTTP 503 when it is missing or invalid. It no longer contains a hard-coded production fallback. A pure resolver test rejects missing, non-HTTPS, credential-bearing, and non-origin URLs.

The Control Plane live-proof documentation says the demo endpoint evaluates authorization and records a decision; it does not execute the requested side effect. The isolation change is designed to keep PR preview traffic from reaching that production endpoint by default. This is source-level intent until CI and the actual Cloudflare Preview runtime confirm it.

## 4. Staging readiness gaps

1. **No staging Control Plane service exists in this change.** Preview live API calls are intentionally disabled until a non-production upstream is provisioned and explicitly configured.
2. **Preview runtime behavior still needs confirmation.** Verify Cloudflare applies `previews.vars` and the proxy returns 503 when the upstream is unset.
3. **Preview secrets and policy** are not verifiable from source. Confirm the preview Worker does not have a secret/configuration combination that defeats the intended isolation; never share production credentials with untrusted PRs.
4. **Newsletter persistence is not clearly isolated.** The checked-in config has no KV namespace. Confirm preview cannot read/write production subscribers and that preview responses do not imply persistence when storage is absent.
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

This PR now changes the website build environment, Cloudflare Worker configuration, and server-side proxy behavior in addition to documenting the map. It does not modify deployed Cloudflare/Railway resources, DNS, or secret values directly. No production/staging health check has yet been performed, and no live environment is claimed healthy based on repository configuration alone.
