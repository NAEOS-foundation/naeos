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

The current `main` branch already contains a real filesystem execution experiment and a read-only EvidenceBundle verifier. This remediation closes the remaining evidence-authenticity gap and records the remaining runtime work explicitly.

## Evidence integrity remediation

EvidenceBundle verification now requires:

- canonical SHA-256 digest;
- Ed25519 signature over the digest;
- public key carried with the bundle;
- explicit signature algorithm metadata.

Changing a bundle and recomputing its SHA-256 digest is therefore insufficient to pass verification: the original signature no longer validates.

The public key embedded in a bundle establishes cryptographic authenticity of the bundle under that signing key. A future trust-anchor mechanism is required if an evaluator must additionally establish that the key belongs to an approved NAEOS issuer.

## Remaining acceptance work

- [ ] evidence signing key lifecycle/trust anchor suitable for production deployment;
- [ ] evidence-backed completion in the runtime path;
- [ ] runtime observation as a first-class enforcement component;
- [ ] live coding-agent execution through the gateway;
- [ ] explicit CLI exit semantics for DENY / execution failure / verification failure;
- [ ] test-oracle taxonomy separating oracle defects from actual policy bypasses;
- [ ] independent re-test after all items are complete.

## Evaluation posture

Do not request a paid independent re-test until the acceptance list above is green. The goal is a re-testable commit with reproducible commands, exit codes, evidence artifacts, and verification results rather than an evaluation of known open gaps.
