## naeos import

Import specifications from HCL format to NAEOS YAML/JSON

### Synopsis

Import a specification written in HCL format and convert it to
the NAEOS YAML or JSON format for use with the pipeline.

Supported HCL blocks:
  project "name" { version = "1.0.0" }
  service "name" { port = 8080; type = "backend" }
  infra "name" { engine = "docker" }

Example:
  naeos import --input spec.hcl
  naeos import --input spec.hcl --output spec.yaml --format yaml
  naeos import --input spec.hcl --format json --output-file result.json

```
naeos import [flags]
```

### Options

```
  -f, --format string   output format: yaml, json (default "yaml")
  -h, --help            help for import
  -i, --input string    path to HCL input file (required)
  -o, --output string   path to output file
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

