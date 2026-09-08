## naeos config resolve

Resolve a configuration file into concrete values

```
naeos config resolve <config.json> [flags]
```

### Options

```
  -h, --help             help for resolve
      --output string    output format: table or json (default "table")
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

