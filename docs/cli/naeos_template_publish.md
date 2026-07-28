## naeos template publish

Publish a starter project template to the marketplace

### Synopsis

Publish a starter project template to the NAEOS template marketplace.

The template directory must contain:
  - template.yaml or naeos.yaml — manifest with name, version, description
  - README.md — documentation
  - Project source files

Example:
  naeos template publish ./my-template
  naeos template publish ./my-template --registry https://registry.naeos.dev

To generate a local registry entry without publishing:
  naeos template publish ./my-template --registry file://./local-registry.json

```
naeos template publish [path] [flags]
```

### Options

```
  -h, --help              help for publish
  -j, --json              output template entry as JSON
      --registry string   template registry URL (default: generate entry only, use --registry to publish remotely)
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

