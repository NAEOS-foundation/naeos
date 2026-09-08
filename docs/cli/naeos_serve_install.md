## naeos serve install

Install a systemd unit for the NAEOS daemon

```
naeos serve install [flags]
```

### Options

```
      --binary string   absolute path to the naeos binary (default: current executable)
      --config string   path to server config file (required)
      --group string    system group to run the service as
  -h, --help            help for install
      --user string     system user to run the service as
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos serve](naeos_serve.md)	 - Run NAEOS as a production daemon

