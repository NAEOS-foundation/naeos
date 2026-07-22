## naeos docs

Generate project documentation

### Synopsis

Generate API documentation and architecture diagrams.

Example:
  naeos docs api --project my-app
  naeos docs architecture --project my-app -o ./docs

### Options

```
  -h, --help             help for docs
  -o, --output string    output directory
  -p, --project string   project name (default "project")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos docs api](naeos_docs_api.md)	 - Generate API documentation
* [naeos docs architecture](naeos_docs_architecture.md)	 - Generate architecture diagram

