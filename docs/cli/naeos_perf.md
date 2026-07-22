## naeos perf

Performance optimization tools

### Synopsis

Manage connection pools, batch processing, and caching.

Example:
  naeos perf pool-create --name db --min 2 --max 10
  naeos perf pool-acquire --name db
  naeos perf pool-stats --name db
  naeos perf cache-set --key mykey --value myvalue --ttl 60s
  naeos perf cache-get --key mykey
  naeos perf cache-stats

```
naeos perf [flags]
```

### Options

```
  -h, --help   help for perf
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos perf cache-get](naeos_perf_cache-get.md)	 - Get a cache value
* [naeos perf cache-set](naeos_perf_cache-set.md)	 - Set a cache value
* [naeos perf cache-stats](naeos_perf_cache-stats.md)	 - Show cache statistics
* [naeos perf pool-acquire](naeos_perf_pool-acquire.md)	 - Acquire a connection from pool
* [naeos perf pool-create](naeos_perf_pool-create.md)	 - Create a connection pool
* [naeos perf pool-stats](naeos_perf_pool-stats.md)	 - Show connection pool statistics

