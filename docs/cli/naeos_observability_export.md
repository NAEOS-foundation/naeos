## naeos observability export

Export collected spans to an OTLP collector

### Synopsis

Export collected spans to an OpenTelemetry collector via OTLP/HTTP.

Example:
  naeos observability export --endpoint http://localhost:4318 --count 3

```
naeos observability export [flags]
```

### Options

```
      --count int         number of sample spans to export (default 1)
      --endpoint string   OTLP collector base URL (required)
  -h, --help              help for export
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos observability](naeos_observability.md)	 - Observability and telemetry management

