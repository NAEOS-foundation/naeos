## naeos template init

Initialize a project from a template in the marketplace

### Synopsis

Initialize a new project from a starter template in the marketplace.

Templates include complete project structures with:
  - Source code boilerplate
  - Build configuration (Makefile, Dockerfile)
  - CI/CD workflows
  - NAEOS specification file

Examples:
  naeos template init microservices-go
  naeos template init microservices-go --output ./my-project

```
naeos template init [name] [flags]
```

### Options

```
  -h, --help              help for init
  -o, --output string     output directory (defaults to template name)
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

