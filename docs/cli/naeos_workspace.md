## naeos workspace

Manage multi-module workspaces

### Synopsis

Manage multi-module workspaces for NAEOS projects.

Example:
  naeos workspace init my-workspace
  naeos workspace add my-module ./modules/my-module
  naeos workspace list
  naeos workspace remove my-module

### Options

```
  -h, --help          help for workspace
      --root string   workspace root directory (default ".")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos workspace add](naeos_workspace_add.md)	 - Add a module to the workspace
* [naeos workspace info](naeos_workspace_info.md)	 - Show workspace information and status
* [naeos workspace init](naeos_workspace_init.md)	 - Initialize a workspace
* [naeos workspace list](naeos_workspace_list.md)	 - List workspace modules
* [naeos workspace lock](naeos_workspace_lock.md)	 - Generate or update workspace lockfile
* [naeos workspace remove](naeos_workspace_remove.md)	 - Remove a module from the workspace

