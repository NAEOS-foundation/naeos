## naeos agent request

Authorize and execute an agent tool request through NAEOS

### Synopsis

Submit a normalized agent tool request to the NAEOS execution gateway.

The request is first evaluated by the active governance policy. Only an allowed request reaches the runtime sandbox. Replay protection is enabled by default; the default in-memory store is process-local and suitable only for local/test use. Use --replay-db for durable restart/shared-replica protection. The resulting decision and execution metadata can optionally be persisted to an agent session.

Example:
  naeos agent request --adapter reference-external-agent \
    --allow-tool filesystem --allow-action read \
    --allow-environment development --allow-capability filesystem.read \
    --request-file request.json --output json

```
naeos agent request [flags]
```

### Options

```
      --adapter string              agent adapter: json, codex, or reference-external-agent (default "json")
      --allow-action strings        adapter policy action grant (repeat or comma-separate; explicit '*' for wildcard)
      --allow-capability strings    adapter policy capability grant (optional; empty means unconstrained)
      --allow-environment strings   adapter policy environment grant (optional; empty means unconstrained)
      --allow-tool strings          adapter policy tool grant (repeat or comma-separate; explicit '*' for wildcard)
  -h, --help                        help for request
      --invocation-id string        unique invocation identity used for replay protection
      --output string               output format: table or json (default "table")
      --replay-db string            named database connection for durable replay protection; omit only for local/test in-memory replay
      --request string              inline JSON agent tool request
      --request-file string         path to JSON agent tool request
      --session-id string           persist the decision and execution in an agent session
      --store-path string           path to the agent session store JSON file
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos agent](naeos_agent.md)	 - Manage autonomous agent sessions
