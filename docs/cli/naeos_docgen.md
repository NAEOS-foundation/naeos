## naeos docgen

Generate documentation from specification

### Synopsis

Auto-generate API docs, module docs, and architecture docs from specs.

Example:
  naeos docgen --input-file spec.yaml
  naeos docgen --input-file spec.yaml --output api
  naeos docgen --input-file spec.yaml --output modules
  naeos docgen --input-file spec.yaml --output architecture

```
naeos docgen [flags]
```

### Options

```
  -h, --help                 help for docgen
      --input string         specification input
      --input-file string    path to specification file
      --output string        output type: full, api, modules (default "full")
      --output-file string   optional file path to write output
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

