## naeos rollback

Rollback to a previous snapshot of generated artifacts

### Synopsis

Manage snapshots and rollback generated artifacts.

Example:
  naeos rollback list
  naeos rollback restore <snapshot-id>

### Options

```
  -h, --help                help for rollback
  -o, --output-dir string   directory to restore artifacts to (default ".")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos rollback list](naeos_rollback_list.md)	 - List available snapshots
* [naeos rollback restore](naeos_rollback_restore.md)	 - Restore artifacts from a snapshot

