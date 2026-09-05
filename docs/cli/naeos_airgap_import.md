## naeos airgap import

Extract an air-gapped bundle

```
naeos airgap import <bundle.tar.gz> [flags]
```

### Options

```
  -d, --dest string     destination directory (default ".")
  -h, --help            help for import
      --output string   output format: table or json (default "table")
      --verify-hashes   verify file hashes against manifest (default true)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos airgap](naeos_airgap.md)	 - Air-gapped distribution bundles

