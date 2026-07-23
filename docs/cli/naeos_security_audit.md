## naeos security audit

Run security audit on project files

### Synopsis

Scan project files for common security issues:
  - Hardcoded secrets and API keys
  - SQL injection patterns
  - XSS vulnerabilities
  - Unsafe eval/deserialization
  - Debug mode enabled
  - Missing health check endpoints

Example:
  naeos security audit
  naeos security audit --input ./src
  naeos security audit --output json

```
naeos security audit [flags]
```

### Options

```
  -h, --help            help for audit
  -i, --input string    directory or file to audit (default: current directory)
  -o, --output string   output format: text, json, yaml
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos security](naeos_security.md)	 - Security and secrets management

