# P1.11 — Independent Verifier CLI

## Purpose

P1.11 provides a read-only verifier for serialized control-plane EvidenceBundle records. It does not contact the control plane, evaluate policy, or execute actions.

Trust boundary:

EvidenceBundle JSON + operator trust anchor -> verifier -> PASS / FAIL

## Command

Supply the public key through an independent trusted channel; do not trust the key merely because it is embedded in the bundle.

```bash
naeos evidence verify-bundle --input-file evidence.json --trusted-public-key /secure/trust/naeos-evidence.pub
naeos evidence verify-bundle --input-file evidence.json --trusted-public-key /secure/trust/naeos-evidence.pub --output json
```

The trusted public key accepts raw, base64, or hexadecimal Ed25519 key material. Missing or malformed trust configuration fails closed. The verifier compares the bundle's embedded public key with the operator-supplied trust anchor and verifies the signature.

## Verification semantics

The verifier uses `controlplane.VerifyEvidenceWithTrustedKey`. It reports three separate properties:

- **Digest integrity:** recomputed canonical SHA-256 digest matches the bundle.
- **Signature validity:** the embedded Ed25519 signature verifies over the digest.
- **Signer trusted:** the embedded public key exactly matches the independently supplied trust anchor.

Decision and execution bindings are checked independently. A valid self-contained signature is not sufficient for PASS if the signer is not trusted by the operator.

## Scope

P1.11 is a local verifier for exported canonical evidence. It does not provide remote evidence retrieval, trust-store distribution or rotation, production KMS, or enterprise compliance attestation.
