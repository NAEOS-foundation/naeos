## naeos lock generate

Generate a lock file from current artifacts

### Synopsis

Generate a SHA-256 based lock file for reproducible builds.

Example:
  naeos lock generate file1.go file2.go
  naeos lock generate -o naeos.lock *.go

```
naeos lock generate [flags]
```

### Options

```
  -h, --help   help for generate
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos lock](naeos_lock.md)	 - Manage lock files for reproducible builds

