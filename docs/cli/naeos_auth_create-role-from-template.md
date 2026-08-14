## naeos auth create-role-from-template

Create a role from a compliance template (auditor, soc2_auditor, gdpr_admin, hipaa_admin)

```
naeos auth create-role-from-template <template-name> [flags]
```

### Options

```
  -h, --help                 help for create-role-from-template
      --parent stringArray   parent roles to inherit from
      --role-name string     custom role name (defaults to template name)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos auth](naeos_auth.md)	 - Authentication and authorization management

