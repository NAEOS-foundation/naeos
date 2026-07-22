## naeos events replay

Replay events to reconstruct pipeline state

### Synopsis

Replay a series of events to reconstruct the state of a pipeline run.
Events can be loaded from a JSON file or standard input.

Example:
  naeos events replay --input events.json
  cat events.json | naeos events replay

```
naeos events replay [flags]
```

### Options

```
  -h, --help            help for replay
  -i, --input string    path to events JSON file (required)
  -o, --output string   optional output file path
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos events](naeos_events.md)	 - Event sourcing commands for pipeline audit trail and replay

