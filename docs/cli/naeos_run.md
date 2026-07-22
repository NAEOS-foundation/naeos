## naeos run

Execute the NAEOS pipeline

### Synopsis

Execute the full NAEOS pipeline: parse, normalize, resolve, build NEIR, generate artifacts.

Example:
  naeos run --config config.yaml --input spec.yaml
  naeos run --config config.yaml --input-file spec.yaml --output json
  naeos run --config config.yaml --input spec.yaml --language go --language typescript

```
naeos run [flags]
```

### Options

```
      --config string          path to JSON or YAML config file (auto-detected if omitted)
      --dry-run                preview artifacts without writing to disk
  -h, --help                   help for run
      --input string           specification input to process
      --input-file string      path to a specification file
      --language stringArray   target language for code generation (go, typescript, python, java, rust)
      --output string          output format: text, json, or yaml (default "text")
      --output-file string     optional file path to write the formatted output
```

### Options inherited from parent commands

```
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

