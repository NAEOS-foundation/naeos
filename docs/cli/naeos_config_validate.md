## naeos config validate

Validate a NAEOS config file against the schema

### Synopsis

Validate a configuration file (YAML or JSON) against the NAEOS config schema.
Reports missing required fields and type mismatches.

Example:
  naeos config validate --input naeos.yaml
  naeos config validate --input config.json --output json

```
naeos config validate [flags]
```

### Options

```
  -h, --help           help for validate
  -i, --input string   path to config file (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos config](naeos_config.md)	 - Configuration management commands

