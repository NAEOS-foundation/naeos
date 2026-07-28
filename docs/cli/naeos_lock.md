## naeos lock

Manage lock files for reproducible builds

### Synopsis

Generate and verify lock files for reproducible builds using SHA-256 checksums.

Example:
  naeos lock generate file1.go file2.go
  naeos lock verify file1.go file2.go

### Options

```
  -h, --help            help for lock
  -o, --output string   path for the lock file (default "naeos.lock")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos lock generate](naeos_lock_generate.md)	 - Generate a lock file from current artifacts
* [naeos lock verify](naeos_lock_verify.md)	 - Verify current artifacts against lock file

