## naeos schema

NEIR schema registry operations

### Synopsis

Manage and validate against the NEIR JSON Schema registry.

The schema registry hosts versioned JSON Schema definitions for the
NEIR specification format. Use this command to validate specs against
the canonical schema, or query schema version information.

Examples:
  naeos schema validate spec.yaml
  naeos schema validate spec.yaml --registry https://naeos.dev/schemaregistry/latest.json
  naeos schema validate spec.json --output json
  naeos schema info

### Options

```
  -h, --help   help for schema
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos schema info](naeos_schema_info.md)	 - Show schema registry information
* [naeos schema validate](naeos_schema_validate.md)	 - Validate a NEIR spec against the schema registry

