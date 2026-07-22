## naeos audit

Security audit of generated or source files

### Synopsis

Run a security audit on the specified files.

Example:
  naeos audit main.go config.yaml
  naeos audit internal/**/*.go

```
naeos audit [file1] [file2] ... [flags]
```

### Options

```
  -h, --help   help for audit
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

