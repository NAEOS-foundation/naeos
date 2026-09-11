## naeos observability slo

Show a reference SLO with burn-rate alert rules

### Synopsis

Show a reference SLO (99% availability, 30d window) with its error
budget and Prometheus burn-rate alerting rules.

Example:
  naeos observability slo --service orders-api

```
naeos observability slo [flags]
```

### Options

```
  -h, --help             help for slo
      --service string   service name for the SLO (default "naeos")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos observability](naeos_observability.md)	 - Observability and telemetry management

