## naeos evidence log

Log an evidence record from a gateway execution

### Synopsis

Execute a tool request through the governance gateway and record
the evidence of the decision and execution outcome.

Example:
  naeos evidence log --tool shell --action run --actor ci-bot \
    --environment production --policy-file policy.json

```
naeos evidence log [flags]
```

### Options

```
      --action string        action to perform (required)
      --actor string         actor identity
      --environment string   deployment environment
  -h, --help                 help for log
      --output string        output format: table or json (default "table")
      --policy-file string   path to policy JSON file
      --resource string      target resource
      --tool string          tool name (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos evidence](naeos_evidence.md)	 - Governance evidence store — immutable audit trail

