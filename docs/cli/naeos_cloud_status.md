## naeos cloud status

Show deployed resource status

### Synopsis

List deployed resources from the cloud state store (~/.naeos/cloud/).

```
naeos cloud status [flags]
```

### Options

```
  -h, --help             help for status
      --project string   Filter by project name
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
  -e, --env string             Environment (dev, staging, prod) (default "dev")
  -i, --input-file string      Spec file with cloud configuration (overrides flags)
      --output-format string   output format: json, yaml, table (default "table")
  -p, --provider string        Cloud provider (aws, gcp, azure) (default "aws")
  -r, --region string          Cloud region
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos cloud](naeos_cloud.md)	 - Cloud deployment commands

