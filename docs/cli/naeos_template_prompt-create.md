## naeos template prompt-create

Create a custom LLM prompt template

### Synopsis

Create a custom LLM prompt template that can be used with 'naeos ai enrich'.

The command opens an interactive editor or accepts --system and --user flags.
Example:
  naeos template prompt-create my-custom-prompt --system "You are an expert" --user "Analyze this: {{.SpecContent}}"

```
naeos template prompt-create [name] [flags]
```

### Options

```
      --description string   description of the prompt
  -h, --help                 help for prompt-create
      --provider string      LLM provider (openai, anthropic, ollama) (default "openai")
      --system string        system prompt content
      --user string          user prompt content (required)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --templates-dir string   templates directory (default ".naeos/templates")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos template](naeos_template.md)	 - Manage generation templates, prompt library, and template marketplace

