## naeos security

Security and secrets management

### Synopsis

Manage encrypted secrets, sanitize input, and validate data.

Example:
  naeos security set-secret --name db-pass --value secret123
  naeos security get-secret --name db-pass
  naeos security list-secrets
  naeos security sanitize --input '<script>alert("xss")</script>'
  naeos security hash-password --password mypass
  naeos security validate --name email --value test@example.com

```
naeos security [flags]
```

### Options

```
  -h, --help   help for security
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos security audit](naeos_security_audit.md)	 - Run security audit on project files
* [naeos security get-secret](naeos_security_get-secret.md)	 - Retrieve a secret value
* [naeos security hash-password](naeos_security_hash-password.md)	 - Hash a password with bcrypt
* [naeos security list-secrets](naeos_security_list-secrets.md)	 - List all stored secrets
* [naeos security sanitize](naeos_security_sanitize.md)	 - Sanitize input string
* [naeos security set-secret](naeos_security_set-secret.md)	 - Store an encrypted secret
* [naeos security validate](naeos_security_validate.md)	 - Validate a value against rules

