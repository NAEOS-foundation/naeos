## naeos agent append-action

Append an action to a stored session

```
naeos agent append-action [flags]
```

### Options

```
      --agent-id string        agent identifier for the action
      --decision string        decision outcome for the action
  -h, --help                   help for append-action
  -o, --output string          output format: text, json, yaml (default "text")
      --parameters string      JSON object of additional parameters
      --policy-id string       policy identifier that authorized the action
      --reason string          reason for the action
      --session-id string      session identifier (required)
      --spec-ref stringArray   specification reference to attach to the action
      --store-path string      path to the agent session store JSON file
      --target string          target resource or tool (required)
      --type string            action type (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos agent](naeos_agent.md)	 - Manage autonomous agent sessions
