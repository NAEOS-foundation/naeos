## naeos demo

Run the investor demo control-plane server

### Synopsis

Start the investor demo HTTP server: authorization, policy
enforcement, verification, and audit. Security decisions run outside
agent control.

Example:
  naeos demo --addr :9091

```
naeos demo [flags]
```

### Options

```
      --addr string   listen address (e.g. :9091) (default ":9091")
      --bind string   alias for --addr
  -h, --help          help for demo
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

