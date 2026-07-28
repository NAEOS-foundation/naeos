## naeos dx

Developer experience tools

### Synopsis

Generate VS Code extensions, CLI completions, and code snippets.

Example:
  naeos dx vscode-gen
  naeos dx completion-bash
  naeos dx completion-zsh
  naeos dx completion-powershell
  naeos dx snippet-list
  naeos dx snippet-get --name project

```
naeos dx [flags]
```

### Options

```
  -h, --help   help for dx
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos dx completion-bash](naeos_dx_completion-bash.md)	 - Generate bash completion script
* [naeos dx completion-powershell](naeos_dx_completion-powershell.md)	 - Generate PowerShell completion script
* [naeos dx completion-zsh](naeos_dx_completion-zsh.md)	 - Generate zsh completion script
* [naeos dx snippet-get](naeos_dx_snippet-get.md)	 - Get a code snippet
* [naeos dx snippet-list](naeos_dx_snippet-list.md)	 - List available code snippets
* [naeos dx vscode-gen](naeos_dx_vscode-gen.md)	 - Generate VS Code extension package

