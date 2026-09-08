## naeos runtime exec

Execute a tool request through the governance gateway

### Synopsis

Execute an agent tool request after policy evaluation.

Example:
  naeos runtime exec --tool shell --action run --resource scripts/deploy.sh \
    --environment production --actor ci-bot
  naeos runtime exec --tool file-edit --action write --output json

```
naeos runtime exec [flags]
```

### Options

```
      --action string        action to perform (required)
      --actor string         actor identity
      --adapter string       agent adapter to use
      --context string       JSON evaluation context
      --environment string   deployment environment
  -h, --help                 help for exec
      --output string        output format: table or json (default "table")
      --resource string      target resource
      --tool string          tool name (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos runtime](naeos_runtime.md)	 - Runtime gateway for authorized agent execution

