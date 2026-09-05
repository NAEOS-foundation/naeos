## naeos evidence query

Query evidence records by criteria

```
naeos evidence query [flags]
```

### Options

```
      --actor string      filter by actor
      --decision string   filter by decision (ALLOW/DENY/REQUIRE_APPROVAL)
  -h, --help              help for query
      --limit int         max records to return (default 20)
      --output string     output format: table or json (default "table")
      --policy string     filter by policy ID
      --resource string   filter by resource
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos evidence](naeos_evidence.md)	 - Governance evidence store — immutable audit trail

