## naeos marketplace publish

Publish a template, profile, or plugin to the marketplace

### Synopsis

Publish a local package to the NAEOS marketplace registry.

The package directory must contain a naeos.yaml manifest with name, version, and type fields.

Example:
  naeos marketplace publish ./my-template
  naeos marketplace publish ./my-plugin --registry https://registry.naeos.dev

```
naeos marketplace publish [path] [flags]
```

### Options

```
  -h, --help   help for publish
```

### Options inherited from parent commands

```
      --cache-dir string       cache directory (default ".naeos/cache")
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos marketplace](naeos_marketplace.md)	 - Browse and install templates, profiles, and plugins

