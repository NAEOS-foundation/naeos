# P1.7 Policy Change Mid-Run

This example proves that an authorization issued under policy v1 cannot be reused after the active policy advances to v2.

Flow: T0 request -> policy v1 -> ALLOW; T1 policy v2 active -> old authorization -> EXECUTION_BLOCKED -> no side effect -> evidence -> verify.

The invariant is authorization freshness: a decision is valid only while its bound policy version remains the active policy at the execution boundary.

## Run

./examples/control-plane-policy-change/run-demo.sh

Or: go run ./examples/control-plane-policy-change

No network, credentials, LLM API key, or external service is required.

## Acceptance criteria
- T0 authorization returns ALLOW under policy v1.
- Policy v2 becomes active before execution.
- Reusing the v1 authorization returns DENY with stale_policy.
- Execution evidence is EXECUTION_BLOCKED.
- No side-effect file is created.
- Evidence contains the original authorization and blocked execution.
- Independent session verification returns PASS.
- The command exits non-zero if any assertion fails.