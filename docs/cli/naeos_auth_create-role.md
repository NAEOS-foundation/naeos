## naeos auth create-role

Create a custom RBAC role

```
naeos auth create-role <name> [flags]
```

### Options

```
      --deny stringArray         denied permissions (resource:action)
  -h, --help                     help for create-role
      --parent stringArray       parent roles to inherit from
      --permission stringArray   permissions (resource:action, e.g., spec:read)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos auth](naeos_auth.md)	 - Authentication and authorization management

