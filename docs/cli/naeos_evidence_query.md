## naeos evidence query

Query evidence records by exact criteria, optional RFC3339 time range, and optional chain verification.

```
naeos evidence query [flags]
```

### Query model

All filters are conjunctive: a record must satisfy every supplied filter.

- `--id` — exact evidence ID
- `--actor` — exact actor identity
- `--resource` — exact resource
- `--action` — exact action
- `--environment` — exact execution environment
- `--policy` — exact policy ID
- `--decision` — exact decision (`ALLOW`, `DENY`, `REQUIRE_APPROVAL`)
- `--from` — inclusive RFC3339 start time
- `--to` — inclusive RFC3339 end time
- `--limit` — maximum number of records; `0` means unlimited
- `--verify-chain` — verify the complete evidence hash chain before returning results
- `--output json` — return full evidence records instead of the human-readable table

Results are returned newest-first.

### Examples

Query production actions by an agent:

```bash
naeos evidence query --actor ci-bot --environment production
```

Find denied actions in a time window:

```bash
naeos evidence query \
  --decision DENY \
  --from 2026-10-01T00:00:00Z \
  --to 2026-10-01T23:59:59Z
```

Retrieve one evidence record as JSON and verify the chain first:

```bash
naeos evidence query --id ev-123 --verify-chain --output json
```

### Output semantics

The table view exposes the primary audit dimensions: ID, timestamp, decision, execution status, actor, action, resource, and environment.

JSON output preserves the complete `EvidenceRecord`, including policy, reasons, artifact integrity fields, approval binding, metadata, and hash-chain fields.

Malformed timestamps and a range where `--from` is after `--to` fail with a non-zero exit status.
