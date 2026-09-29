# P1.8 Atomic Execution Commit Boundary

This example demonstrates the P1.8 execution boundary:

    authorize v1 -> atomic execution lock -> side effect -> EXECUTION_ALLOWED

The side effect is a disposable local file. The real invariant is that the policy store cannot activate v2 in the middle of the atomic execution callback.

## Run

    ./examples/control-plane-atomic-execution/run-demo.sh

Or:

    go run ./examples/control-plane-atomic-execution

No network, credentials, or external service is required.

## Acceptance criteria

- Authorization under policy v1 returns ALLOW.
- The side effect is created only inside ExecuteAtomic.
- Evidence contains AUTHORIZATION_DECISION and EXECUTION_ALLOWED.
- Policy v2 becomes active only after the atomic side effect returns.
- Independent session verification returns PASS.
- The command exits non-zero if an assertion fails.
