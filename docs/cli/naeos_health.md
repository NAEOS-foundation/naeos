## naeos health

Run system health checks and diagnostics

### Synopsis

Perform comprehensive health checks on the NAEOS installation,
configuration, and dependencies.

Example:
  naeos health
  naeos health -o json
  naeos health -o yaml

```
naeos health [flags]
```

### Options

```
  -h, --help   help for health
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

