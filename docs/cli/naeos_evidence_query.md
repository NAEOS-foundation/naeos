## naeos evidence query

Query evidence records by criteria

### Synopsis

Query the evidence store using exact filters and an optional RFC3339 time range.

```
naeos evidence query [flags]
```

### Options

```
      --action string        filter by action
      --actor string         filter by actor
      --decision string      filter by decision (ALLOW/DENY/REQUIRE_APPROVAL)
      --environment string   filter by environment
      --from string          inclusive start time (RFC3339)
  -h, --help                 help for query
      --id string            filter by exact evidence ID
      --limit int            max records to return; 0 means unlimited (default 20)
      --output string        output format: table or json (default "table")
      --policy string        filter by policy ID
      --request-id string    filter by exact logical request ID
      --resource string      filter by resource
      --to string            inclusive end time (RFC3339)
      --verify-chain         verify the evidence hash chain before returning results
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos evidence](naeos_evidence.md)	 - Governance evidence store — immutable audit trail
