# NAEOS Verification Pack v1

**Baseline:** `main` @ `75fc784ce965b200c92d410216774c6702018f0f`  
**Scope:** External-review hardening merged by PR #542  
**Purpose:** Reproducible verification surface for reviewers, security practitioners, and design partners.

## Verification posture

The baseline includes trusted Ed25519 evidence verification, explicit CLI exit semantics, an agent-facing JSONL runtime bridge, filesystem observation, and a test-oracle taxonomy.

### Verification matrix

| Control | Repository evidence | Expected result |
|---|---|---|
| Trusted evidence signer | `VerifyEvidenceWithTrustedKey` and trust-anchor tests | Unknown signer = FAIL |
| Evidence signing | `Ledger.BuildEvidence` / ledger signer | Ed25519 signature and public key recorded |
| Policy denial | CLI policy/runtime tests | DENY returns exit code 10 |
| Execution failure | `runtimeResultExit` | Failed governed execution returns 11 |
| Verification failure | Evidence verification command | Verification failure returns 12 |
| Agent boundary | `naeos runtime bridge` + JSON adapter | Requests are normalized before gateway authorization |
| Runtime observation | `filesystemObserver` | Protected filesystem writes can produce artifact hash evidence |
| Oracle classification | `docs/TEST-ORACLE-TAXONOMY.md` | Failures are classified before being called bypasses |

## Reproduction

From a clean checkout:

```bash
git checkout 75fc784ce965b200c92d410216774c6702018f0f
gofmt -l .
go test ./...
go vet ./...
```

Relevant trust-anchor tests:

```text
TestVerifyEvidenceWithTrustedKeyRejectsUnknownSigner
TestEvidenceSignerRequiresValidKey
```

Agent boundary:

```bash
naeos runtime bridge --help
```

The bridge consumes one JSON tool request per line and emits one JSON result per line. Normalization does not grant authorization; requests still cross the governance gateway.

## Security acceptance criteria

Blocking findings are:

1. Unauthorized protected side effect.
2. Tool execution bypassing the governance gateway.
3. Evidence accepted with an untrusted signer.
4. Evidence that cannot be independently verified.
5. Silent bypass through stale authorization.
6. Required security/governance CI regression.

A non-zero process exit alone is not evidence of POLICY_BYPASS. Use `docs/TEST-ORACLE-TAXONOMY.md`.

## Deployment boundary

The repository exposes the trust-anchor interface. Production private-key custody remains an operator responsibility; private signing material should remain outside evidence bundles and use an appropriate secret-management/KMS boundary.

This pack does not claim an independent production audit. It defines the controls and evidence expected for one.

## External review handoff

1. Checkout the exact baseline.
2. Run repository tests/security checks.
3. Verify the trusted-key tests.
4. Exercise CLI exit semantics.
5. Exercise the JSONL runtime bridge.
6. Verify filesystem observation and artifact hashing.
7. Run a stale-policy adversarial scenario.
8. Classify failures using the oracle taxonomy.
9. Record expected behavior, observed behavior, evidence, and reproduction commands.

## Acceptance status

**PR #542 remediation:** merged to `main`.

**Independent re-test:** pending. This is intentionally separate from implementation acceptance; the next reviewer should validate the exact baseline above.
