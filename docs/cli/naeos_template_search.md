## naeos template search

Search for starter project templates in the marketplace

### Synopsis

Search the template marketplace for starter project templates.

Examples:
  naeos template search go
  naeos template search "machine learning"
  naeos template search python --output json

```
naeos template search [query] [flags]
```

### Options

```
  -h, --help              help for search
  -o, --output string     output format: text, json (default "text")
      --registry string   template registry URL (default "https://naeos.dev/templates/registry.json")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --templates-dir string   templates directory (default ".naeos/templates")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos template](naeos_template.md)	 - Manage generation templates, prompt library, and template marketplace

