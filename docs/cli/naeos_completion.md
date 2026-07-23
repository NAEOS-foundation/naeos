## naeos completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for NAEOS.

To load completions:

Bash:
  $ source <(naeos completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ naeos completion bash > /etc/bash_completion.d/naeos
  # macOS:
  $ naeos completion bash > $(brew --prefix)/etc/bash_completion.d/naeos

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ naeos completion zsh > "${fpath[1]}/_naeos"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ naeos completion fish | source
  # To load completions for each session, execute once:
  $ naeos completion fish > ~/.config/fish/completions/naeos.fish

PowerShell:
  PS> naeos completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> naeos completion powershell > naeos.ps1
  # and source this file from your PowerShell profile.


```
naeos completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime

