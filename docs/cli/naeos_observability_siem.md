## naeos observability siem

Export sample audit events to a SIEM collector

### Synopsis

Export a sample audit event to a SIEM collector using CEF (default)
or NDJSON framing.

Example:
  naeos observability siem --endpoint http://localhost:9000
  naeos observability siem --endpoint http://localhost:9000 --format json

```
naeos observability siem [flags]
```

### Options

```
      --endpoint string   SIEM collector URL (required)
      --format string     framing format: cef or json (default "cef")
  -h, --help              help for siem
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos observability](naeos_observability.md)	 - Observability and telemetry management

