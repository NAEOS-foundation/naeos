## naeos benchmark

Run pipeline benchmarks

### Synopsis

Benchmark the pipeline performance by running multiple iterations.
Reports timing statistics including average, min, max, and p95.

Example:
  naeos benchmark --iterations 100
  naeos benchmark --input spec.yaml --iterations 50
  naeos benchmark --output json --iterations 100

```
naeos benchmark [flags]
```

### Options

```
      --config string    path to config file
  -h, --help             help for benchmark
  -n, --iterations int   number of iterations (default 10)
  -o, --output string    output format: text, json, yaml
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

