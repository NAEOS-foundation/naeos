## naeos mcp

Start MCP (Model Context Protocol) server

### Synopsis

Start an MCP server that exposes NAEOS tools to AI agents.

The server implements the Model Context Protocol and provides tools:
  - parse_spec: Parse a NAEOS specification
  - validate_spec: Validate a specification
  - generate_context: Generate AI context bundle
  - compile_spec: Compile spec to AI instruction sets
  - explain_concept: Explain NAEOS concepts

Example:
  naeos mcp --port 8080
  naeos mcp

```
naeos mcp [flags]
```

### Options

```
  -h, --help       help for mcp
      --port int   port for MCP server (default 3000)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

