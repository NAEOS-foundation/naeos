## naeos config

Resolve configuration from environment, files, secrets, and Vault

### Synopsis

Use references (env:VAR, file:/path, secret:ns/name/key, vault:path#key) to resolve configuration values.

```
naeos config [flags]
```

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos config resolve](naeos_config_resolve.md)	 - Resolve a configuration file into concrete values
* [naeos config sources](naeos_config_sources.md)	 - List available config sources and their reference syntax
* [naeos config test](naeos_config_test.md)	 - Resolve a single config reference (e.g. env:FOO)

