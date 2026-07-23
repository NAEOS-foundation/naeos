## naeos migration status

Show migration status for all configured databases

### Synopsis

Display the current migration status of all configured databases.

Reads saved connections and queries each database's _migrations table.

Example:
  naeos migration status
  naeos migration status --output json

```
naeos migration status [flags]
```

### Options

```
  -h, --help            help for status
  -o, --output string   output format: text, json, yaml
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos migration](naeos_migration.md)	 - Database migration management

