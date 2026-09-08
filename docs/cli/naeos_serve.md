## naeos serve

Run NAEOS as a production daemon

### Synopsis

Run the NAEOS server as a production daemon with TLS, graceful
shutdown, multiple listeners, and systemd integration.

Example:
  naeos serve                          # run with defaults on :8080
  naeos serve --config server.yaml     # run from a config file
  naeos serve run --port 9443 --tls-cert cert.pem --tls-key key.pem
  naeos serve config > server.yaml     # print a starter config
  naeos serve install --config server.yaml   # install a systemd unit

```
naeos serve [flags]
```

### Options

```
      --auth                enable JWT authentication
      --config string       path to server config file (YAML)
  -h, --help                help for serve
      --jwt-secret string   JWT secret for API auth
  -p, --port string         API listener port (default "8080")
      --tls-cert string     path to TLS certificate (enables HTTPS)
      --tls-key string      path to TLS private key
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos serve config](naeos_serve_config.md)	 - Print a starter server configuration (YAML)
* [naeos serve install](naeos_serve_install.md)	 - Install a systemd unit for the NAEOS daemon
* [naeos serve run](naeos_serve_run.md)	 - Start the NAEOS daemon in the foreground
* [naeos serve uninstall](naeos_serve_uninstall.md)	 - Remove the installed systemd unit

