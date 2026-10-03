## naeos sbom verify

Verify the integrity of a CycloneDX SBOM document

```
naeos sbom verify [flags] <bom-file>
```

Use `--require-licenses` in release checks to fail unless every listed component
has a CycloneDX license identifier or SPDX license expression.

### Options

```
  -h, --help            help for verify
      --output string   output format: table or json (default "table")
      --require-licenses fail if any component has no CycloneDX license declaration
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos sbom](naeos_sbom.md)	 - Software Bill of Materials generation and verification
