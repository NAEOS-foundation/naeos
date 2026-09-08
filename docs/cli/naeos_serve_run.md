## naeos serve run

Start the NAEOS daemon in the foreground

```
naeos serve run [flags]
```

### Options

```
      --auth                enable JWT authentication
      --config string       path to server config file (YAML)
  -h, --help                help for run
      --jwt-secret string   JWT secret for API auth
  -p, --port string         API listener port (default "8080")
      --tls-cert string     path to TLS certificate
      --tls-key string      path to TLS private key
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos serve](naeos_serve.md)	 - Run NAEOS as a production daemon

