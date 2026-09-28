# P1.6 Public Control-Plane Golden Path

This is the smallest public proof of the NAEOS control-plane boundary.

```
ALLOW → execute → evidence → verify
DENY  → no side effect → evidence → verify
```

The demo uses the real `internal/controlplane` evaluator, decision gateway, append-only evidence ledger, and session verifier. The side effect is deliberately local and disposable: an ALLOW request creates one file in an isolated temporary directory; a DENY request must not create its file.

## Run

```bash
./examples/control-plane-golden-path/run-demo.sh
```

Or:

```bash
NAEOS_P16_OUTPUT_DIR=/tmp/naeos-p1-6 go run ./examples/control-plane-golden-path
```

No network, credentials, LLM API key, or external service is required.

## Acceptance criteria

- ALLOW is returned for `repository.write`.
- ALLOW creates `allow-side-effect.json`.
- ALLOW evidence contains `EXECUTION_ALLOWED` and `SIDE_EFFECT_OBSERVED`.
- DENY is returned for `production.delete`.
- DENY does not create `deny-side-effect.json`.
- DENY evidence contains `EXECUTION_BLOCKED` and `SIDE_EFFECT_OBSERVED`.
- Independent session verification returns PASS.
- The command exits non-zero if any assertion fails.

The local file is only a disposable stand-in for an external side effect. The proof is the invariant that DENY never crosses the side-effect boundary.
