## naeos dx vscode-gen

Generate VS Code extension package

### Synopsis

Generate a complete VS Code extension with syntax highlighting and LSP integration.

The extension is written to the specified output directory (default: ./naeos-vscode).
It includes:
  - package.json with commands, keybindings, menus, and LSP configuration
  - TextMate grammar for .naeos.yaml syntax highlighting
  - extension.js with LSP client, compile/validate/dashboard commands
  - README.md and launch.json

Example:
  naeos dx vscode-gen
  naeos dx vscode-gen --output ./extensions/naeos-vscode

```
naeos dx vscode-gen [flags]
```

### Options

```
  -h, --help            help for vscode-gen
      --output string   output directory for the extension (default: ./naeos-vscode)
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos dx](naeos_dx.md)	 - Developer experience tools

