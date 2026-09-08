## naeos control evaluate

Evaluate an authorization request and issue a decision

### Synopsis

Evaluate an authorization request against registered governance policies.

Example:
  naeos control evaluate --resource deploy --action run --environment production \\
    --actor ci-bot --context '{"tls_version":"1.2"}'
  naeos control evaluate --resource deploy --action run --output json

```
naeos control evaluate [flags]
```

### Options

```
      --action string        action being requested (required)
      --actor string         actor issuing the request
      --context string       JSON object of evaluation context
      --environment string   deployment/environment context
      --fail-open            allow requests with no matching policy (default: deny)
  -h, --help                 help for evaluate
      --output string        output format: table or json (default "table")
      --resource string      resource being acted on (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos control](naeos_control.md)	 - Governance control plane evaluation

