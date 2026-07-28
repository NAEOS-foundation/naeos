## naeos preview

Preview generated artifacts without writing them

### Synopsis

Preview what artifacts would be generated without writing to disk.

Example:
  naeos preview --config config.yaml --input spec.yaml
  naeos preview --config config.yaml --input-file spec.yaml

```
naeos preview [flags]
```

### Options

```
      --config string   path to JSON or YAML config file (auto-detected if omitted)
  -h, --help            help for preview
      --input string    specification input or file path to process
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

