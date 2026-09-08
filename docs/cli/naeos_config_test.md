## naeos config test

Resolve a single config reference (e.g. env:FOO)

```
naeos config test <reference> [flags]
```

### Options

```
  -h, --help             help for test
  -s, --secret strings   inject secret ns/name/key=value (repeatable)
  -v, --vault strings    inject vault path#key=value (repeatable)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos config](naeos_config.md)	 - Resolve configuration from environment, files, secrets, and Vault

