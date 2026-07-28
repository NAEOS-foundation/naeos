## naeos distributed

Run pipeline tasks in distributed mode across multiple workers

### Synopsis

Execute pipeline tasks using multiple workers for parallel processing.
Tasks are distributed across workers using round-robin load balancing.

Example:
  naeos distributed --input spec.yaml --workers 4
  naeos distributed --config config.yaml --input-file spec.yaml --workers 8

```
naeos distributed [flags]
```

### Options

```
      --config string   path to config file
  -h, --help            help for distributed
  -w, --workers int     number of parallel workers (default 4)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

