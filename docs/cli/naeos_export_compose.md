## naeos export compose

Generate docker-compose.yaml and Dockerfile from spec

### Synopsis

Generate Docker Compose configuration and Dockerfile from a specification.

Reads the spec, runs the pipeline to produce NEIR, then generates
a docker-compose.yaml and Dockerfile in the output directory.

Example:
  naeos export compose --input spec.yaml
  naeos export compose --input spec.yaml --output-dir ./docker

```
naeos export compose [flags]
```

### Options

```
      --config string          path to config file
  -h, --help                   help for compose
      --input string           specification input or file path
      --language stringArray   target language
      --output-dir string      output directory (default ".")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos export](naeos_export.md)	 - Export generated artifacts to a directory

