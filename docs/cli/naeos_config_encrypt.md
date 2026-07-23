## naeos config encrypt

Encrypt a config file with AES-256-GCM

### Synopsis

Encrypt a configuration file at rest using AES-256-GCM with a passphrase.
Output is written as base64-encoded ciphertext.

Example:
  naeos config encrypt --input config.yaml --output config.enc
  naeos config encrypt --input config.yaml --passphrase "my-secret" --output config.enc

```
naeos config encrypt [flags]
```

### Options

```
  -h, --help                help for encrypt
  -i, --input string        path to config file (required)
  -o, --output string       path to write encrypted output
  -p, --passphrase string   encryption passphrase (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos config](naeos_config.md)	 - Configuration management commands

