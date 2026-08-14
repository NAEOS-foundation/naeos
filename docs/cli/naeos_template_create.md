## naeos template create

Scaffold a new starter template directory for publishing

### Synopsis

Create a new starter project template structure in the specified directory.
The scaffold includes a template.yaml manifest, README.md, and source boilerplate.

After creation, use 'naeos template publish <dir>' to publish it to the marketplace.

Examples:
  naeos template create my-starter --author "My Name" --desc "My starter template"
  naeos template create go-api --lang go --tags "api, rest"

```
naeos template create [name] [flags]
```

### Options

```
      --author string   template author name
  -D, --desc string     template description (default "NAEOS starter template")
  -d, --dir string      output directory (defaults to template name)
  -h, --help            help for create
  -l, --lang string     primary programming language (default "go")
      --tags strings    comma-separated tags
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

