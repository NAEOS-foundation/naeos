## naeos agent

Manage autonomous agent sessions

### Synopsis

Create and inspect agent sessions that capture project context, permissions, and action history.

```
naeos agent [flags]
```

### Options

```
  -h, --help   help for agent
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos agent actions](naeos_agent_actions.md)	 - List stored agent actions
* [naeos agent append-action](naeos_agent_append-action.md)	 - Append an action to a stored session
* [naeos agent create](naeos_agent_create.md)	 - Create a new agent session
* [naeos agent delete](naeos_agent_delete.md)	 - Delete a stored agent session
* [naeos agent get](naeos_agent_get.md)	 - Get a stored agent session
* [naeos agent list](naeos_agent_list.md)	 - List stored agent sessions
* [naeos agent request](naeos_agent_request.md)	 - Authorize and execute an agent tool request through NAEOS
