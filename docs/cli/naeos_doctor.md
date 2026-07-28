## naeos doctor

Run diagnostics on the NAEOS environment and configuration

### Synopsis

Run comprehensive diagnostics to check the health of your NAEOS setup.

Checks include:
  - Go toolchain and version
  - Language runtimes (Node, Python, Java, Rust)
  - Docker and container tools
  - Git version
  - NAEOS configuration
  - Spec validation (if spec provided)
  - Network connectivity
  - Go module status
  - Output directory writability
  - Workspace detection

```
naeos doctor [flags]
```

### Options

```
      --config string   path to config file
  -h, --help            help for doctor
      --quick           skip language runtime and network checks
      --spec string     path to spec file for validation
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

