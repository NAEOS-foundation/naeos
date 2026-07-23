## naeos inspect

Inspect the NAEOS pipeline result

### Synopsis

Inspect the pipeline result showing project details, artifacts, and tasks.

Example:
  naeos inspect --config config.yaml --input spec.yaml
  naeos inspect --config config.yaml --input-file spec.yaml --output json

```
naeos inspect [flags]
```

### Options

```
      --config string        path to JSON or YAML config file (auto-detected if omitted)
  -h, --help                 help for inspect
      --input string         specification input or file path to process
      --input-file string    path to a specification file
      --output string        output format: text, json, or yaml (default "text")
      --output-file string   optional file path to write the formatted output
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

