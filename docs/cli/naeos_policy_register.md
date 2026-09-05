## naeos policy register

Register a governance policy from a JSON file

### Synopsis

Register a governance policy, persisting it for use by the control plane.

Example:
  naeos policy register --file policy.json

```
naeos policy register [flags]
```

### Options

```
  -f, --file string   path to policy JSON file (required)
  -h, --help          help for register
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos policy](naeos_policy.md)	 - Manage governance policies

