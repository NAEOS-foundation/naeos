# Repository Endpoint Security Test Plan

## Scope

Add deterministic security tests for the Core repository metadata endpoint (`GET /api/v1/repositories/{owner}/{name}`). Keep this work isolated from IAM remediation PR #79. Tests must establish side-effect behavior, not only HTTP status codes.

## Existing implementation observations

- Endpoint is GET-only and requires service-JWKS authentication.
- JWT middleware enforces repository-read scope.
- Repository path is validated as two segments, then normalized to lowercase owner/name.
- Exact allowlist check occurs before provider request; allowlist denial is audited.
- Outbound GitHub API host is fixed to `api.github.com`; redirects are not followed.
- Outbound timeout is 10 seconds; response is read through a 4 MiB `LimitReader`. The current implementation does not explicitly detect a body exceeding the limit; hardening must read at most `4 MiB + 1` and reject if over the cap.
- Successful provider read emits an audit event; audit failure prevents a successful response.

## Required test cases

| Case | Expected HTTP | Provider calls | Additional assertion |
|---|---:|---:|---|
| Not allowlisted | 403 | 0 | Denial audit event emitted when auditor configured |
| Wrong scope | 403 | 0 | Provider never called |
| Wrong method | 405 (handler behavior; confirm through full middleware) | 0 | Provider never called |
| Invalid path segments | 400 | 0 | Provider never called |
| Allowlisted, provider 200 | 200 | 1 | Response metadata; success audit event |
| Allowlisted, provider 404 | 404 | 1 | No success audit event |
| Allowlisted, provider non-200 | 502 | 1 | Safe error response; no provider details leaked |
| Provider transport failure | 502 | 1 | Safe error response |
| Provider redirect | 502 or provider-unsuccessful handling | 1 | Redirect target never requested; credentials not forwarded |
| Audit sink fails on success | 500 | 1 | No repository data in successful response |
| Oversized provider body | 502 (required after strict limit enforcement) | 1 | Explicitly detect >4 MiB; current limit-reader alone is not sufficient to prove strict rejection |


## Testability implementation approach

- Add a private injectable outbound HTTP client or transport to `Server` (default nil uses production client behavior), without exposing user-controlled provider destinations.
- In tests, use a counting `http.RoundTripper` and deterministic mocked responses. Count attempts and assert exactly zero for all denial paths and exactly one for allowlisted provider calls.
- Prefer tests that exercise the full handler/middleware route, signed service JWT, configured allowlist, and audit sink.
- Keep redirect policy enforced by the production client; mock transport must separately assert the redirect target is never contacted.
- Ensure mock state is per test and safe for parallel execution if tests use `t.Parallel`.

## Additional audit findings / implementation gates

- The current source returns 405 from the handler for non-GET methods. Full middleware may alter the response; test the complete `Server.Handler()` and document the observed result instead of assuming 403.
- Current body handling caps bytes read but does not explicitly detect an oversized provider body. The test should fail against current behavior if oversized payloads are accepted, and implementation should read at most `4 MiB + 1` and reject when the extra byte proves the body exceeds the cap.
- PR #867 CI on the docs-only commit reports Sign-off check and DCO failing. Documentation Gate, gofmt-autofix, Secrets scan, Policy Bypass Experiment, Cloudflare Deployment Gate, Slack Notifications, NAEOS Compliance Gate, CodeQL, and CI succeeded. Resolve DCO/sign-off before readiness; this is not code remediation verification.

## Security invariants

1. Authorization failures must happen before provider I/O.
2. Provider destination remains server-controlled and fixed to GitHub API.
3. Redirects must not cause a second request or credential forwarding.
4. Provider response or audit failure must never produce a successful repository-data response.
5. Tests should assert call counts and side-effect ledger, not just status codes.

## Verification and PR gates

- Run focused tests for repository endpoint/JWKS authorization.
- Run full Go test suite and relevant lint/security CI.
- Review diff to confirm only testability seam and tests are introduced; no authorization relaxation, allowlist change, production configuration, or deployment.
- CI must pass on the final PR head. Until these are executed, tests are unverified and the PR remains draft.

## Current execution note

This branch initially contains only the plan. No source-code changes or tests have been made in this PR yet because the remote-only editing workflow cannot execute the Go test suite locally. The plan PR is a tracking artifact, not a security remediation completion claim.
