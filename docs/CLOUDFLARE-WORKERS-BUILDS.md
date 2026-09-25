# Cloudflare Workers Builds Configuration

This document records how the Cloudflare **Workers Builds** integration for the `naeos` Worker is
intended to be configured, why the check previously failed, and the verification procedure.

The integration settings are stored in the Cloudflare Dashboard, not in the repository. Cloudflare
documentation states that Workers Builds does **not** honor the `[build]` section of a Wrangler
configuration file. Repo changes here only pin the toolchain so that dashboard settings are
deterministic.

## Symptom

Every pull request reports a failed check:

```
Workers Builds: naeos
```

## Root cause

The `naeos` Worker (`site/wrangler.jsonc`) declares:

```jsonc
"main": ".open-next/worker.js"
```

`.open-next/` is the build output of `opennextjs-cloudflare build` and is gitignored, so a clean
checkout contains no `worker.js`. When the Workers Builds integration runs without a build command,
`wrangler deploy` fails locally reproducible with:

```
✘ [ERROR] The entry-point file at ".open-next/worker.js" was not found.
```

The OpenNext build must therefore run before any deploy. The build is not run by the integration by
default because the build settings live in the Dashboard.

## Required dashboard configuration

Dashboard path:

```
Cloudflare Dashboard → Workers & Pages → naeos → Settings → Build → Configure
```

| Setting | Value |
| --- | --- |
| Git branch | `main` |
| Root directory | `site` |
| Build command | `npm run build:cf` |
| Deploy command | `npx wrangler deploy --dry-run` |
| Watch paths | (unset) |
| Node version | 22 (see `site/.node-version`) |

`npm run build:cf` is defined in `site/package.json` and produces `.open-next/worker.js` plus the
asset manifest, matching what `.github/workflows/website.yml` runs in CI.

## Recommended deploy architecture

`website.yml` remains the **single production deployer** of the `naeos` Worker. It performs the
OpenNext build, optionally injects the `NEWSLETTER_KV` binding, and runs `wrangler deploy` only on
`main`.

Set the Workers Builds **deploy command to `npx wrangler deploy --dry-run`** so the integration acts
as build verification and never creates a second deployment or uploads worker versions:

- A second live deploy pipeline would race with `website.yml`.
- Workers Builds cannot run the KV namespace injection that `website.yml` performs before deploy;
  making it the live deployer would permanently break the newsletter subscription endpoint.

Git branch `main` also stops the integration from building on feature and docs-only pull requests
(which already fail for unrelated reasons), keeping the instrument as a production-grade build gate.

## Verification

After the dashboard settings are applied:

1. Push any commit to `main` (for example a merge or rebase of an open pull request).
2. Confirm the `Workers Builds: naeos` check completes with the build stage green.
3. Confirm the `Website` workflow (`.github/workflows/website.yml`) still deploys `naeos` and reports
   green.

Stale `Workers Builds: naeos` failure records attached to old pull-request head commits disappear
once a pull request receives a new commit after the configuration change.

## Alternative architecture (not recommended)

If the team later wants Workers Builds to own production deploys instead of `website.yml`:

- Set deploy command to `npx wrangler deploy`.
- Remove the deploy step from `.github/workflows/website.yml`.
- Resolve the `NEWSLETTER_KV` binding before doing so: either record the KV namespace ID in
  `site/wrangler.jsonc` or accept a non-functional newsletter endpoint, since Workers Builds cannot
  run the current binding-injection step.

Keep this decision documented here if it changes.