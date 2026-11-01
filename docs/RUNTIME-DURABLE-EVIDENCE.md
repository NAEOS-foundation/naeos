# Authenticated Durable Runtime Evidence

The runtime bridge can optionally persist authenticated evidence for a single
execution run. This mode is stricter than the legacy `--evidence-file` mode:
it requires an operator-managed Ed25519 signing key and an independently
configured trusted public key.

## Required inputs

- `--evidence-file`: canonical signed runtime evidence JSONL.
- `--evidence-ledger`: append-only runtime event journal. Each event binds the
  run ID, sequence, event name, and runtime evidence digest.
- `--evidence-receipt`: signed completion receipt, written only after the
  ledger is sealed.
- `--evidence-run-id`: stable identifier for this run. Reuse it only to resume
  the same incomplete run; choose a new ID for a new run.
- `--evidence-signing-key`: Ed25519 private key material (raw seed, raw private
  key, base64, or hexadecimal).
- `--evidence-trusted-public-key`: independently trusted Ed25519 public key
  (raw, base64, or hexadecimal).
- `--evidence-issuer` and `--evidence-key-id`: trust metadata pinned by the
  verifier.

All of these options are mandatory together when any durable-evidence option
is supplied. The bridge refuses incomplete configuration.

## Example

The signing key and trusted public key must be provisioned out of band. Do not
derive trust by accepting a public key embedded in the evidence artifact.

```bash
./naeos runtime bridge \
  --filesystem-root /var/lib/naeos/sandbox \
  --replay-db /var/lib/naeos/replay.db \
  --evidence-file /var/lib/naeos/evidence/runtime-evidence.jsonl \
  --evidence-ledger /var/lib/naeos/evidence/runtime-events.jsonl \
  --evidence-receipt /var/lib/naeos/evidence/runtime-receipt.json \
  --evidence-run-id run-2026-11-02-001 \
  --evidence-signing-key /secure/naeos-runtime-ed25519.key \
  --evidence-trusted-public-key /secure/trust/naeos-runtime-ed25519.pub \
  --evidence-issuer naeos-runtime \
  --evidence-key-id key-2026-11-01
```

The ledger is verified before accepting an existing journal. If a completion
receipt already exists, the bridge verifies its ledger digest, run identity,
issuer, key ID, and Ed25519 signature against the configured trust anchor, then
refuses to append to that completed run. If the process stopped before receipt
creation, the same run ID can resume from the next event sequence, provided the
existing journal verifies and the completion receipt is absent.

Each evidence record is appended and synced before its corresponding durable
ledger event is published. The bridge does not emit a successful response for
an evidence write or sync failure. A failure after the governed side effect
cannot undo that side effect; it means the run has no successful evidence
completion and must be reconciled explicitly.

## Verification boundaries and operational assumptions

- SHA-256 event chaining and receipt digests detect mutation relative to the
  retained journal; they do not independently authenticate the operator.
- Ed25519 receipt signatures authenticate the receipt only when the verifier
  supplies a separately trusted public key and pins the expected issuer and
  key ID. Key rotation requires explicit trust-store updates.
- Private keys must remain outside evidence artifacts and should be held by an
  external secret manager or KMS in production.
- The implementation relies on the host filesystem's append, file-sync, and
  atomic-rename semantics. It does not claim protection against a privileged
  host administrator, storage rollback, malicious filesystem, or loss of the
  storage device.
- This journal is not a distributed consensus protocol, immutable/WORM storage,
  replicated audit service, or production KMS integration.
- A durable receipt does not establish exactly-once execution for arbitrary
  external side effects. Replay protection must use a durable replay database
  when restart or cross-worker replay resistance is claimed.
- The bridge's independent observer currently covers filesystem writes only.
  Other tool/action classes remain denied at the adapter boundary.

Legacy `--evidence-file` operation remains available for compatibility but
does not provide a durable sealed-ledger receipt unless the complete set of
authenticated durable-evidence options is configured.
