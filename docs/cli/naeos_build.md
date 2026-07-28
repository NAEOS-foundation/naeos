## naeos build

Build artifacts from a specification

### Synopsis

Build artifacts from a specification using the NAEOS pipeline.

By default, build runs locally. Use --distributed to distribute work
across multiple workers for parallel processing.

Example:
  naeos build --config config.yaml --input spec.yaml
  naeos build --config config.yaml --input-file spec.yaml --distributed --workers 8

```
naeos build [flags]
```

### Options

```
      --config string          path to JSON or YAML config file
      --distributed            enable distributed building across workers
      --dry-run                preview artifacts without writing to disk
  -h, --help                   help for build
      --input string           specification input to process
      --input-file string      path to a specification file
      --language stringArray   target language for code generation
      --output string          output format: text, json, or yaml (default "text")
      --output-file string     optional file path to write formatted output
  -w, --workers int            number of parallel workers (used with --distributed) (default 4)
```

### Options inherited from parent commands

```
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

