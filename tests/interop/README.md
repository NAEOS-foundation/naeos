# Handoff interoperability contract harness

This package is a deterministic mock harness for NAEOS-Handoff-Interop-Test-001.

## Run locally

From the repository root:

```sh
go test -race -count=1 -v ./tests/interop
```

## Coverage

- T-01: fresh approval executes once and correlates provenance.
- T-02: mutated candidate digest is rejected.
- T-03: policy version changes invalidate stale approval.
- T-04: authorization provider failure fails closed.
- T-05: simulated committed side effect with lost response yields UNKNOWN and suppresses retry.
- T-06: reconciliation resolves the mock UNKNOWN using action ID without re-execution.
- T-07: subject/scope mismatch is rejected.
- Additional check: map insertion order does not change the test harness digest.

## Important limits

- This is a mock contract harness, not an integration with NAEOS production runtime or a third-party agent framework.
- The canonicalization profile is currently the test's Go JSON representation (struct fields plus encoding/json map-key ordering). It is **not yet a cross-language canonicalization standard**. Do not compare these digests across implementations until a versioned canonicalization specification is agreed.
- T-05 and T-06 use an in-memory side-effect ledger. They do not prove real-world side-effect semantics.
- A passing test means only that the mock behavior satisfies its assertion. It does not establish product conformance.
- Do not publish PASS for any external adapter until the adapter itself is exercised and artifacts are reviewed.
