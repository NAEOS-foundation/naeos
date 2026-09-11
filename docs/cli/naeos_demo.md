## naeos demo

Run the investor demo control-plane server

### Synopsis

Start the investor demo HTTP server: authorization, policy
enforcement, verification, and audit. Security decisions run outside
agent control.

Observability: pass --siem-endpoint to forward every audit event to a SIEM
collector (CEF or NDJSON framing), and --otlp-endpoint to export request
traces to an OTLP/HTTP collector.

Example:
  naeos demo --addr :9091
  naeos demo --addr :9091 --siem-endpoint http://localhost:9000 --otlp-endpoint http://localhost:4318

```
naeos demo [flags]
```

### Options

```
      --addr string            listen address (e.g. :9091) (default ":9091")
      --bind string            alias for --addr
  -h, --help                   help for demo
      --otlp-endpoint string   OTLP/HTTP collector base URL (e.g. http://localhost:4318)
      --siem-endpoint string   SIEM collector URL to forward audit events (e.g. http://localhost:9000)
      --siem-format string     SIEM framing format: cef or json (default "cef")
      --tenant-id string       tenant identifier attached to exported events
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

