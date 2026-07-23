## naeos template

Manage generation templates, prompt library, and template marketplace

### Synopsis

Manage NAEOS generation templates, prompt library, and template marketplace.

Examples:
  naeos template list
  naeos template publish ./my-template
  naeos template search microservices
  naeos template init go-http-api
  naeos template show enrich-spec

### Options

```
  -h, --help                   help for template
      --templates-dir string   templates directory (default ".naeos/templates")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos template add](naeos_template_add.md)	 - Add a custom template
* [naeos template init](naeos_template_init.md)	 - Initialize a project from a template in the marketplace
* [naeos template list](naeos_template_list.md)	 - List available templates
* [naeos template prompt-create](naeos_template_prompt-create.md)	 - Create a custom LLM prompt template
* [naeos template prompt-remove](naeos_template_prompt-remove.md)	 - Remove a custom prompt template
* [naeos template publish](naeos_template_publish.md)	 - Publish a starter project template to the marketplace
* [naeos template remove](naeos_template_remove.md)	 - Remove a custom template
* [naeos template search](naeos_template_search.md)	 - Search for starter project templates in the marketplace
* [naeos template show](naeos_template_show.md)	 - Show details of a prompt template

