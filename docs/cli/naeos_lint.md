## naeos lint

Lint a specification file

### Synopsis

Lint a NAEOS specification file for issues and optionally auto-fix them.

Example:
  naeos lint --input-file spec.yaml
  naeos lint --input-file spec.yaml --fix
  naeos lint --input-file spec.yaml --output json

```
naeos lint [flags]
```

### Options

```
      --fix                 automatically fix issues where possible
  -h, --help                help for lint
      --input-file string   path to a specification file to lint
  -o, --output string       output format: text, json (default "text")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

