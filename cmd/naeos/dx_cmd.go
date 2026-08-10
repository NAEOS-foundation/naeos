package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/devexperience"
)

func newDXCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dx",
		Short: "Developer experience tools",
		Long: `Generate VS Code extensions, CLI completions, and code snippets.

Example:
  naeos dx vscode-gen
  naeos dx completion-bash
  naeos dx completion-zsh
  naeos dx completion-powershell
  naeos dx snippet-list
  naeos dx snippet-get --name project`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newDXVSCodeGenCommand())
	cmd.AddCommand(newDXCompletionBashCommand())
	cmd.AddCommand(newDXCompletionZshCommand())
	cmd.AddCommand(newDXCompletionPSCommand())
	cmd.AddCommand(newDXSnippetListCommand())
	cmd.AddCommand(newDXSnippetGetCommand())

	return cmd
}

func newDXVSCodeGenCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "vscode-gen",
		Short: "Generate VS Code extension package",
		Long: `Generate a complete VS Code extension with syntax highlighting and LSP integration.

The extension is written to the specified output directory (default: ./naeos-vscode).
It includes:
  - package.json with commands, keybindings, menus, and LSP configuration
  - TextMate grammar for .naeos.yaml syntax highlighting
  - extension.js with LSP client, compile/validate/dashboard commands
  - README.md and launch.json

Example:
  naeos dx vscode-gen
  naeos dx vscode-gen --output ./extensions/naeos-vscode`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ext := devexperience.NewVSCodeExtension(
				"naeos", "1.0.0", "NAEOS project support", "NAEOS",
				[]string{"yaml", "json"},
			)

			if outputDir == "" {
				outputDir = "naeos-vscode"
			}

			if err := ext.GenerateExtension(outputDir); err != nil {
				return fmt.Errorf("generate extension: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "VS Code extension generated in %s/\n", outputDir)
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "To install:")
			fmt.Fprintf(cmd.OutOrStdout(), "  cd %s && npm install -g vsce && vsce package && code --install-extension naeos-*.vsix\n", outputDir)
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintln(cmd.OutOrStdout(), "Or copy to your extensions directory:")
			fmt.Fprintf(cmd.OutOrStdout(), "  cp -r %s ~/.vscode/extensions/naeos\n", outputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "", "output directory for the extension (default: ./naeos-vscode)")
	return cmd
}

func newDXCompletionBashCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion-bash",
		Short: "Generate bash completion script",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := devexperience.NewCompletionEngine()
			fmt.Fprintln(cmd.OutOrStdout(), engine.GenerateBashCompletion())
			return nil
		},
	}
}

func newDXCompletionZshCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion-zsh",
		Short: "Generate zsh completion script",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := devexperience.NewCompletionEngine()
			fmt.Fprintln(cmd.OutOrStdout(), engine.GenerateZshCompletion())
			return nil
		},
	}
}

func newDXCompletionPSCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion-powershell",
		Short: "Generate PowerShell completion script",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := devexperience.NewCompletionEngine()
			fmt.Fprintln(cmd.OutOrStdout(), engine.GeneratePowerShellCompletion())
			return nil
		},
	}
}

func newDXSnippetListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "snippet-list",
		Short: "List available code snippets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sm := devexperience.NewSnippetManager()

			names := sm.List()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-20s\n", "SNIPPET")
			fmt.Fprintf(out, "%-20s\n", "--------")
			for _, name := range names {
				fmt.Fprintf(out, "%-20s\n", name)
			}
			return nil
		},
	}
}

func newDXSnippetGetCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "snippet-get",
		Short: "Get a code snippet",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			sm := devexperience.NewSnippetManager()

			snippet, ok := sm.Get(name)
			if !ok {
				return fmt.Errorf("snippet '%s' not found", name)
			}

			fmt.Fprintln(cmd.OutOrStdout(), snippet)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "snippet name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
