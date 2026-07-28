## naeos ai compile

Compile a specification for a target AI agent using AI

### Synopsis

Compile a NAEOS specification into configuration files for a target AI agent
using AI-powered generation.

Example:
  naeos ai compile --input-file spec.yaml --target claude
  naeos ai compile --input-file spec.yaml --target opencode --provider anthropic

```
naeos ai compile [flags]
```

### Options

```
  -h, --help                help for compile
      --input-file string   path to specification file
      --provider string     LLM provider (openai, anthropic, ollama)
      --target string       target AI agent (claude, copilot, cursor, gemini, codex, opencode, windsurf) (default "opencode")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos ai](naeos_ai.md)	 - AI-powered assistance commands

