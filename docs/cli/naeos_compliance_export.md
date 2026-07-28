## naeos compliance export

Export audit log for compliance reporting

### Synopsis

Export the audit trail in JSON or CSV format for compliance purposes.

Example:
  naeos compliance export --format json --output audit-export.json
  naeos compliance export --format csv --output audit-export.csv

```
naeos compliance export [flags]
```

### Options

```
  -a, --audit-file string   path to audit log file (default: ~/.naeos/audit.log)
  -f, --format string       export format: json or csv (default "json")
  -h, --help                help for export
  -o, --output string       output file path (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos compliance](naeos_compliance.md)	 - Compliance reporting and audit log export

