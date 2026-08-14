## naeos compliance report

Generate a compliance report for a framework (soc2, hipaa, gdpr)

### Synopsis

Generate a compliance report by evaluating audit events against a compliance framework.

Example:
  naeos compliance report --framework soc2 --output soc2-report.json
  naeos compliance report --framework hipaa --output hipaa-report.json

```
naeos compliance report [flags]
```

### Options

```
  -a, --audit-file string   path to audit log file (default: ~/.naeos/audit.log)
      --framework string    compliance framework: soc2, hipaa, gdpr (default "soc2")
  -h, --help                help for report
  -o, --output string       output file path
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos compliance](naeos_compliance.md)	 - Compliance reporting and audit log export

