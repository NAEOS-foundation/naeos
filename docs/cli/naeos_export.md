## naeos export

Export generated artifacts to a directory

### Synopsis

Export all generated artifacts to a directory.

Example:
  naeos export --config config.yaml --input spec.yaml
  naeos export --config config.yaml --input spec.yaml --output-dir ./generated
  naeos export --config config.yaml --input spec.yaml --dry-run

```
naeos export [flags]
```

### Options

```
      --config string          path to JSON or YAML config file (auto-detected if omitted)
      --dry-run                preview artifacts without writing to disk
  -h, --help                   help for export
      --input string           specification input or file path to process
      --language stringArray   target language for code generation (go, typescript, python, java, rust)
      --output-dir string      directory to write exported artifacts
```

### Options inherited from parent commands

```
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos export compose](naeos_export_compose.md)	 - Generate docker-compose.yaml and Dockerfile from spec

