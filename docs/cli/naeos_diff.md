## naeos diff

Compare generated artifacts with existing output directory

### Synopsis

Compare current pipeline output with existing files in the output directory.

Example:
  naeos diff --config config.yaml --input spec.yaml
  naeos diff --config config.yaml --input spec.yaml --output-dir ./out
  naeos diff --config config.yaml --input spec.yaml --format unified

```
naeos diff [flags]
```

### Options

```
      --config string       path to JSON or YAML config file (auto-detected if omitted)
      --format string       diff format: unified (default "unified")
  -h, --help                help for diff
      --input string        specification input to process
      --input-file string   path to a specification file
      --output-dir string   existing output directory to compare against
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

