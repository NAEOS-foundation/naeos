# NAEOS Proof Pack v1

**Baseline:** `main` @ `54113c3bda1971d1a503d779eab38b6675fd83e8`  
**Release:** NAEOS 3.7.0  
**Purpose:** A compact, evaluator-facing index of reproducible NAEOS evidence.

## 1. What NAEOS is proving

NAEOS is an engineering control plane for AI coding agents. This Proof Pack focuses on bounded, falsifiable properties rather than blanket security or production-readiness claims.

```text
Intent → Authorization → Execution Boundary → Observation → Durable Evidence → Independent Verification
```

Core boundaries:
- intent is not authorization;
- policy decisions are distinct from execution;
- observations are distinct from claims;
- evidence persists beyond agent memory;
- verification is independently inspectable;
- untrusted handoff input does not silently become authority.

## 2. Reproducible evidence set

| Evidence | Location | Reproduction | Boundary proved |
|---|---|---|---|
| Golden Path | `docs/GOLDEN-PATH.md` | `go build -o naeos ./cmd/naeos && ./examples/demo-cli/run-demo.sh` | specification → validation → policy → generation → traceable evidence |
| Verification Pack v1 | `docs/VERIFICATION-PACK.md` | `gofmt -l . && go test ./... && go vet ./...` plus listed trust/bridge checks | trusted evidence, exit semantics, runtime bridge, observation, oracle classification |
| External Validation #001 | `docs/EXTERNAL-VALIDATION-REPORT-001.md` | follow `docs/EXTERNAL-VALIDATION.md` and retained workflow artifacts | real external runtime authorization, denial, mutation/freshness, crash recovery, independent verification |
| Governance Boundary Benchmark v1 | `benchmarks/governance-boundary-v1/README.md` | `bash ./scripts/governance-boundary-benchmark.sh` | ALLOW, DENY, REQUIRE_APPROVAL, and detected DIRECT_BYPASS behavior |
| Evidence Tampering Benchmark v1 | `experiments/evidence-tampering-v1/README.md` | `go run ./experiments/evidence-tampering-v1` | post-append evidence mutation is rejected by the existing integrity verifier |

## 3. Golden Path

The local Golden Path is the smallest first-run proof. It validates the specification, derives NEIR, rejects an invalid policy configuration, generates agent context, produces artifacts, and writes traceable run metadata.

It proves a reproducible local control-plane path. It does not establish production readiness, enterprise compliance, or safety of arbitrary external integrations.

## 4. Verification Pack

The Verification Pack v1 records the hardening delivered through PR #542:
- operator-supplied Ed25519 trust anchor;
- explicit CLI exit semantics;
- agent-facing JSONL runtime bridge;
- filesystem observation;
- test-oracle taxonomy.

Blocking acceptance findings include unauthorized protected side effects, gateway bypass, untrusted signer acceptance, unverifiable evidence, stale-authorization bypass, or required governance/security CI regression.

Production private-key custody remains an operator responsibility.

## 5. External Validation #001

External Validation #001 is **PASS** for its documented scenarios. It combines:
- CrewAI + Attenu consequential authorization, denial, revocation, ledger integrity, and independent verification;
- Temporal worker crash/recovery with stable logical invocation identity;
- CrewAI mutation-after-authorization and policy-version freshness enforcement.

It explicitly does not claim exactly-once arbitrary external side effects, production readiness, enterprise compliance, or blanket security.

## 6. Governance Boundary Benchmark v1

The four-scenario benchmark measures:
- **ALLOW** — authorized side effect is observed and verified;
- **DENY** — denied action produces no side effect;
- **REQUIRE_APPROVAL** — approval-gated action produces no side effect without approval;
- **DIRECT_BYPASS** — an out-of-band side effect may occur, but independent verification rejects it as authorized evidence.

v1 score: **4/4 scenario assertions passed**. Any deviation is a regression.

## 7. Evidence Tampering Benchmark v1

Three independent mutations are tested:
1. execution output;
2. authorization decision;
3. stored integrity hash.

All three must fail verification while the authoritative store remains intact.

This proves a bounded post-append integrity property. It does not prove signatures/key management, process-compromise resistance, artifact tamper detection, replay protection, capability-escalation resistance, or distributed durability.

## 8. Evaluator sequence

1. Run the **Golden Path**.
2. Read and execute the **Verification Pack** checks.
3. Run the **Governance Boundary Benchmark**.
4. Run the **Evidence Tampering Benchmark**.
5. Review **External Validation #001** and retained artifacts.
6. Re-run selected negative scenarios and record expected vs observed behavior.

## 9. Proof boundaries

This pack intentionally does not claim:
- universal security against arbitrary AI agents or providers;
- exactly-once semantics for arbitrary external side effects;
- production scalability or latency targets;
- enterprise compliance or certification;
- correctness of arbitrary policies;
- protection against every possible bypass;
- distributed durability beyond tested evidence stores;
- an independent audit of the whole NAEOS project.

Each additional claim requires its own executable evidence and retained reproduction context.

## 10. Promotion rule

Evidence should be promoted only when it is:
1. executable or directly reproducible;
2. tied to a specific repository baseline;
3. explicit about expected behavior;
4. explicit about failure/non-claims;
5. retained in a reviewable artifact;
6. independently inspectable where the claim requires independence.

**Principle:** Evidence should be reproducible before it is persuasive.
