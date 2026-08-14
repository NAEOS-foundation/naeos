## naeos profile run

Run pipeline with performance and memory profiling

### Synopsis

Run the NAEOS pipeline with profiling enabled.

Profiles execution stages (validate, build_graph, policy_eval, schedule,
generate, review, write_artifacts) and captures heap snapshots at each
stage boundary to detect memory leaks and performance bottlenecks.

Examples:
  naeos profile run --input "project: my-app\nmodules:\n  - name: core\n    path: ./core"
  naeos profile run --input-file spec.yaml --config config.yaml
  naeos profile run --input-file spec.yaml --profile profile.json --memprofile mem.json

```
naeos profile run [flags]
```

### Options

```
      --config string       path to pipeline config file
  -h, --help                help for run
      --input string        specification input
      --input-file string   path to specification file
      --memprofile string   write memory profile to JSON file
      --profile string      write pipeline profile to JSON file
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos profile](naeos_profile.md)	 - Manage industry-specific project profiles

