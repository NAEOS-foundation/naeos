## naeos agent create

Create a new agent session

```
naeos agent create [flags]
```

### Options

```
      --agent-id string                agent identifier (required)
  -h, --help                           help for create
  -o, --output string                  output format: text, json, yaml (default "text")
      --project string                 project name for the session
      --specification string           specification reference for the session
      --specification-version string   specification version for the session
      --store-path string              path to the agent session store JSON file
      --workspace string               workspace path for the session
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos agent](naeos_agent.md)	 - Manage autonomous agent sessions
