## naeos workflow

Workflow and approval management

### Synopsis

Create, execute, and manage workflows and approval processes.

Example:
  naeos workflow list
  naeos workflow create --name deploy-prod --steps build,test,deploy
  naeos workflow execute --name deploy-prod
  naeos workflow approve --id req-123 --approver admin --comment "LGTM"
  naeos workflow requests

```
naeos workflow [flags]
```

### Options

```
  -h, --help   help for workflow
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos workflow approve](naeos_workflow_approve.md)	 - Approve a pending request
* [naeos workflow create](naeos_workflow_create.md)	 - Create a new workflow
* [naeos workflow execute](naeos_workflow_execute.md)	 - Execute a workflow
* [naeos workflow list](naeos_workflow_list.md)	 - List all workflows
* [naeos workflow reject](naeos_workflow_reject.md)	 - Reject a pending request
* [naeos workflow requests](naeos_workflow_requests.md)	 - List approval requests

