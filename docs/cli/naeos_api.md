## naeos api

Start NAEOS REST API server

### Synopsis

Start the NAEOS REST API server for external integrations and web dashboard.

```
naeos api [flags]
```

### Options

```
  -a, --auth            Enable JWT authentication
  -h, --help            help for api
  -p, --port string     API server port (default "8080")
  -s, --secret string   JWT secret key
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

