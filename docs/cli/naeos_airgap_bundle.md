## naeos airgap bundle

Build an air-gapped distribution bundle

```
naeos airgap bundle [flags]
```

### Options

```
      --charts-dir string       directory containing *.tgz chart archives
  -h, --help                    help for bundle
      --images-file string      file listing image name:tag (one per line)
      --name string             bundle name
  -o, --output string           output bundle path
      --sbom string             CycloneDX SBOM JSON file to include
      --signatures-dir string   directory containing *.sig.json signature bundles
      --version string          bundle version
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos airgap](naeos_airgap.md)	 - Air-gapped distribution bundles

