## naeos history

Show pipeline run history from persisted events

### Synopsis

Display the history of past pipeline runs stored as event files.

Example:
  naeos history
  naeos history --store-dir ./events
  naeos history --store-dir ./events --output json

```
naeos history [flags]
```

### Options

```
  -h, --help               help for history
      --store-dir string   path to event store directory (default: .naeos/events)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

