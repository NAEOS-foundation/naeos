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
go run ./examples/control-plane-golden-path
```

No network, credentials, LLM API key, or external service is required.

## Isolated output for a partner demo

The script creates a fresh temporary output directory by default and prints its path:

```bash
./examples/control-plane-golden-path/run-demo.sh
```

To choose an output path, provide a directory that does not already exist:

```bash
NAEOS_CONTROL_PLANE_DEMO_OUTPUT_DIR=/tmp/naeos-partner-run-001 \
  ./examples/control-plane-golden-path/run-demo.sh
```

The runner refuses to reuse or delete an existing directory. Review and remove old demo output explicitly before reusing a path. The result is written to `result.json`; the ALLOW side effect is written to `allow-side-effect.json`, while the DENY side effect must remain absent.

## One workflow to present

Use the **controlled repository write** scenario:

1. An agent requests the `repository.write` capability.
2. NAEOS evaluates the request and returns ALLOW.
3. The demo performs one disposable local file write.
4. The ledger records authorization, execution, and observed side-effect events.
5. The session verifier checks the evidence and must return PASS.

Then contrast it with the denied `production.delete` request: NAEOS returns DENY, no file is created, and the verification must still return PASS.

This is a local control-plane demonstration, not a live GitHub repository mutation or proof of arbitrary agent integrations.

## Acceptance criteria

- ALLOW is returned for `repository.write`.
- ALLOW creates `allow-side-effect.json`.
- ALLOW evidence contains `AUTHORIZATION_DECISION`, `EXECUTION_ALLOWED`, and `SIDE_EFFECT_OBSERVED`.
- DENY is returned for `production.delete` by the explicit policy deny rule.
- DENY does not create `deny-side-effect.json`.
- DENY evidence contains `AUTHORIZATION_DECISION`, `EXECUTION_BLOCKED`, and `SIDE_EFFECT_OBSERVED`.
- The session verifier returns PASS for the recorded ledger.
- The command exits non-zero if any assertion fails.

The local file is only a disposable stand-in for an external side effect. The proof is the invariant that an explicit policy DENY never crosses the side-effect boundary.
