## naeos runtime bridge

Stream JSON tool calls through the governance gateway

### Synopsis

Read one JSON tool request per line from stdin and emit one JSON execution result per line. This is the protocol-neutral live agent interception boundary.

```
naeos runtime bridge [flags]
```

### Options

```
      --evidence-file string          append canonical runtime evidence JSONL to this file
      --evidence-issuer string        trusted issuer identity embedded in runtime evidence signatures
      --evidence-key-id string        trusted signing key identifier embedded in runtime evidence signatures
      --evidence-signing-key string   Ed25519 private key file for authenticated runtime evidence
      --filesystem-root string        filesystem sandbox root for governed writes
  -h, --help                          help for bridge
      --replay-db string              named database connection for durable replay protection
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos runtime](naeos_runtime.md)	 - Runtime gateway for authorized agent execution
