## naeos watch

Watch for specification changes and re-run the pipeline

### Synopsis

Watch for specification file changes and automatically re-run the pipeline.

Only .yaml, .yml, and .json files trigger a re-run.
The watcher watches the directory containing the input spec file.

Example:
  naeos watch --config config.yaml --input spec.yaml
  naeos watch --config config.yaml --input-file spec.yaml --language go

```
naeos watch [flags]
```

### Options

```
      --config string          path to JSON or YAML config file (auto-detected if omitted)
  -h, --help                   help for watch
      --input string           specification input to process
      --input-file string      path to a specification file
      --language stringArray   target language for code generation
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

