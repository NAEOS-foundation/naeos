## naeos observability

Observability and telemetry management

### Synopsis

Manage tracing, logging, and metrics collection.

Example:
  naeos observability trace --name "http-request"
  naeos observability log --level info --message "Server started"
  naeos observability metrics
  naeos observability status

```
naeos observability [flags]
```

### Options

```
  -h, --help   help for observability
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos observability dashboard](naeos_observability_dashboard.md)	 - Start the observability dashboard
* [naeos observability log](naeos_observability_log.md)	 - Write a log entry
* [naeos observability metrics](naeos_observability_metrics.md)	 - Show collected metrics
* [naeos observability status](naeos_observability_status.md)	 - Show observability stack status
* [naeos observability trace](naeos_observability_trace.md)	 - Create a new trace span

