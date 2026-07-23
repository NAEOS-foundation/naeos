## naeos config decrypt

Decrypt an encrypted config file

### Synopsis

Decrypt a base64-encoded encrypted config file back to plaintext.

Example:
  naeos config decrypt --input config.enc --output config.yaml
  naeos config decrypt --input config.enc --passphrase "my-secret" --output config.yaml

```
naeos config decrypt [flags]
```

### Options

```
  -h, --help                help for decrypt
  -i, --input string        path to encrypted config file (required)
  -o, --output string       path to write decrypted output
  -p, --passphrase string   decryption passphrase (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos config](naeos_config.md)	 - Configuration management commands

