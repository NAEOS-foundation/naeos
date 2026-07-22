## naeos schema validate

Validate a NEIR spec against the schema registry

### Synopsis

Validate a specification file against the latest NEIR JSON Schema
from the schema registry. Supports YAML and JSON spec files.

The command fetches the canonical schema from the registry and checks
that the spec conforms to it, including required fields and enum values.

Examples:
  naeos schema validate spec.yaml
  naeos schema validate spec.json --output json
  naeos schema validate spec.naeos.yaml --registry https://naeos.dev/schemaregistry/v1/neir.json

```
naeos schema validate [file] [flags]
```

### Options

```
  -h, --help              help for validate
  -o, --output string     output format: text, json, yaml (default "text")
      --registry string   schema registry URL (default "https://naeos.dev/schemaregistry/latest.json")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos schema](naeos_schema.md)	 - NEIR schema registry operations

