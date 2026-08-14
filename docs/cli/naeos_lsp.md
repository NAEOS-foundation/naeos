## naeos lsp

Start NEIR Language Server Protocol server

### Synopsis

Start a Language Server Protocol (LSP) server for NEIR specification YAML files.

The LSP server communicates over stdio using the LSP protocol.
Supported features:
  - Autocomplete for NEIR YAML fields and enum values
  - Hover documentation for fields
  - Go-to-definition for references
  - Real-time diagnostics (validation errors and warnings)
  - Document symbols (outline)

Example:
  naeos lsp

```
naeos lsp [flags]
```

### Options

```
  -h, --help   help for lsp
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

