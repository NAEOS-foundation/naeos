## naeos compliance verify

Verify audit log hash chain integrity

### Synopsis

Verify the hash chain integrity of the audit log to detect tampering.

Example:
  naeos compliance verify
  naeos compliance verify --audit-file /custom/path/audit.log

```
naeos compliance verify [flags]
```

### Options

```
  -a, --audit-file string   path to audit log file (default: ~/.naeos/audit.log)
  -h, --help                help for verify
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos compliance](naeos_compliance.md)	 - Compliance reporting and audit log export

