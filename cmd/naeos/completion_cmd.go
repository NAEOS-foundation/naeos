package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for NAEOS.

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
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return cmd
}
