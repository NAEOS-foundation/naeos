## naeos status

Show current pipeline, system and project status

### Synopsis

Display the current status of the NAEOS project and pipeline configuration.

Example:
  naeos status
  naeos status --config config.yaml
  naeos status -o json

```
naeos status [flags]
```

### Options

```
      --config string         path to JSON or YAML config file (auto-detected if omitted)
  -h, --help                  help for status
      --metrics-port string   prometheus metrics endpoint (e.g. :9090)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

