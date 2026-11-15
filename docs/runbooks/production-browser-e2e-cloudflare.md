# Production Browser E2E: Cloudflare 403 runbook

The production browser test must exercise the real Control Plane and its JavaScript hydration. Do not weaken Cloudflare protection globally or make the test pass by skipping hydration assertions.

## Evidence

The production run fails when requests for `https://naeos.dev/_next/static/chunks/*.js` return HTTP 403 with an HTML response. The browser then cannot hydrate the Control Plane. The Playwright test sends `x-naeos-e2e-token` on GET requests for `/en/control-plane`, `/id/control-plane`, and `/_next/static/`; it logs only whether the header was attached, never the token value.

## Cloudflare remediation

1. Open **Security → Events** for the `naeos.dev` zone and filter to the timestamp of the failed GitHub Actions run. Inspect a failed `/_next/static/chunks/*.js` request and record the action, source/product, and rule ID that produced the 403.
2. Confirm the request contains `x-naeos-e2e-token`. The test's failure diagnostics report `e2eTokenHeaderPresent` without revealing the value.
3. Confirm the GitHub repository secret `NAEOS_E2E_CHALLENGE_TOKEN` and the value checked by the Cloudflare rule are identical. Do not put the value in this repository, an issue, or an Actions log.
4. Add or adjust a rule **before the blocking rule** so it matches only:
   - hostname `naeos.dev`;
   - method `GET`;
   - an exact match of `x-naeos-e2e-token` to the secret value; and
   - either the English/Indonesian Control Plane document paths or `/_next/static/` assets.
5. Use the Security Events evidence to skip only the specific blocking rule/product for those matched requests. Do not disable the WAF, Bot protections, or challenges site-wide. If the event is not bypassable with a custom-rule Skip action, use the narrowest supported exception for that identified product.
6. Re-run **Control Plane Browser E2E**. It is fixed only when both locales hydrate and all 12 browser tests pass against production. Recheck Security Events after the run.

## Suggested expression shape

Use this only as a starting point and adapt it to the actual Cloudflare rule/action identified in Security Events. Replace the placeholder in the Cloudflare dashboard with the secret value; never commit the value:

```text
(http.host eq "naeos.dev"
 and http.request.method eq "GET"
 and any(http.request.headers["x-naeos-e2e-token"][*] eq "REPLACE_WITH_SECRET_VALUE")
 and (
   http.request.uri.path in {"/en/control-plane" "/en/control-plane/" "/id/control-plane" "/id/control-plane/"}
   or starts_with(http.request.uri.path, "/_next/static/")
 ))
```

This expression defines the narrow request match; it is not, by itself, a reason to skip all security checks. Scope the action to the specific blocking rule/product confirmed in Security Events.
