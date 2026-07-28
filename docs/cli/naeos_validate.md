## naeos validate

Validate a specification using the NAEOS pipeline

### Synopsis

Validate a specification file through the NAEOS pipeline without generating artifacts.

Output formats:
  text  — human-readable text output (default)
  json  — structured JSON with error codes and field locations

Error codes:
  SPEC_EMPTY        — specification is empty
  SPEC_INVALID_YAML — specification contains invalid YAML
  PROJECT_MISSING   — project section is missing
  PROJECT_NAME_MISSING — project name is missing
  SERVICE_DUPLICATE — duplicate service name
  PORT_INVALID      — invalid port number
  PIPELINE_FAILED   — pipeline validation failed

Example:
  naeos validate --input spec.yaml
  naeos validate --input spec.yaml --output json
  naeos v --input-file spec.yaml --output json --output-file result.json

```
naeos validate [flags]
```

### Options

```
      --config string          path to JSON or YAML config file (auto-detected if omitted)
  -h, --help                   help for validate
      --input string           specification input to process
      --input-file string      path to a specification file
      --language stringArray   target language for code generation
      --output string          output format: text, json (default "text")
      --output-file string     optional file path to write the output
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

