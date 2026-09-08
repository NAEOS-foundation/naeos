## naeos sbom generate

Generate a CycloneDX SBOM from a project directory

```
naeos sbom generate [flags]
```

### Options

```
      --dir string             scan directory for file-level components
  -h, --help                   help for generate
      --output string          write SBOM to file (instead of stdout)
      --output-format string   output format: json or table (default "json")
      --project string         project name (used as BOM root component)
      --version string         project version
```

### Options inherited from parent commands

```
      --dry-run   global dry-run mode: preview without writing to disk
      --verbose   enable verbose logging
```

### SEE ALSO

* [naeos sbom](naeos_sbom.md)	 - Software Bill of Materials generation and verification

