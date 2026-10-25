## naeos verify

Verify a specification using the existing NAEOS pipeline

### Synopsis

Verify a specification by running the existing NAEOS pipeline: parse, normalize, resolve, build NEIR, and validate.

This is the first vertical slice of the autonomous control-plane verification foundation.

Examples:
  naeos verify --input-file spec.yaml
  naeos verify --input-file spec.yaml --output-format json
  naeos verify --config config.yaml --input-file spec.yaml

```
naeos verify [flags]
```

### Options

```
      --config string        path to JSON or YAML config file (auto-detected if omitted)
  -h, --help                 help for verify
      --input string         specification input to process
      --input-file string    path to a specification file
      --output-file string   optional file path to write the verification output
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos verify evidence](naeos_verify_evidence.md)	 - Verify the entire evidence store
* [naeos verify report](naeos_verify_report.md)	 - Run independent verification and produce a report
