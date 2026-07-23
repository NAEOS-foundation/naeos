## naeos lock verify

Verify current artifacts against lock file

### Synopsis

Verify that current files match the lock file checksums.

Example:
  naeos lock verify file1.go file2.go
  naeos lock verify --lock-file naeos.lock *.go

```
naeos lock verify [flags]
```

### Options

```
  -h, --help               help for verify
  -l, --lock-file string   path to lock file to verify against (default "naeos.lock")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos lock](naeos_lock.md)	 - Manage lock files for reproducible builds

