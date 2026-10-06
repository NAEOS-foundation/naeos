# External Review Remediation

## Source

Reduced independent review by Agentic Thinking Ltd at commit `55ec8ab` / v3.6.0.

## Six-link acceptance chain

NAEOS treats these as separate proof links:

1. Policy decision
2. Authorization
3. Execution
4. Observation
5. Durable evidence
6. Independent verification

The current `main` branch contains real filesystem execution, first-class runtime observation, evidence-backed completion, a signed EvidenceBundle verifier, and stale-policy enforcement. This hardening change closes the remaining reviewer gaps with explicit trust anchors, process exit semantics, an agent-facing JSONL gateway bridge, and a test-oracle taxonomy.

## Evidence integrity remediation

EvidenceBundle verification now requires:

- canonical SHA-256 digest;
- Ed25519 signature over the digest;
- public key carried with the bundle;
- explicit signature algorithm metadata.

Changing a bundle and recomputing its SHA-256 digest is therefore insufficient to pass verification: the original signature no longer validates.

The public key embedded in a bundle establishes cryptographic authenticity of the bundle under that signing key. A future trust-anchor mechanism is required if an evaluator must additionally establish that the key belongs to an approved NAEOS issuer.

## Remaining acceptance work

- [x] evidence signing key trust anchor: `VerifyEvidenceWithTrustedKey` requires an operator-supplied Ed25519 public key; production private keys remain externally managed/KMS-backed;
- [x] evidence-backed completion in the control-plane execution path;
- [x] runtime observation as a first-class enforcement component;
- [x] live agent-facing execution boundary through `naeos runtime bridge` (JSONL protocol; every request is normalized and authorized by the gateway);
- [x] explicit CLI exit semantics for DENY / execution failure / verification failure (`10/11/12`);
- [x] test-oracle taxonomy separating oracle defects from actual policy bypasses (`docs/TEST-ORACLE-TAXONOMY.md`);
- [ ] independent re-test after all items are complete.

## Evaluation posture

Do not request a paid independent re-test until the acceptance list above is green. The goal is a re-testable commit with reproducible commands, exit codes, evidence artifacts, and verification results rather than an evaluation of known open gaps.

