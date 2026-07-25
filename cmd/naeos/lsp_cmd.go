package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/lsp"
)

func newLSPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start NEIR Language Server Protocol server",
		Long: `Start a Language Server Protocol (LSP) server for NEIR specification YAML files.

The LSP server communicates over stdio using the LSP protocol.
Supported features:
  - Autocomplete for NEIR YAML fields and enum values
  - Hover documentation for fields
  - Go-to-definition for references
  - Real-time diagnostics (validation errors and warnings)
  - Document symbols (outline)

Example:
  naeos lsp`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := lsp.NewServer()
			fmt.Fprintln(cmd.ErrOrStderr(), "NEIR LSP server started on stdio")
			if err := server.Run(); err != nil {
				return fmt.Errorf("LSP server error: %w", err)
			}
			return nil
		},
	}
	return cmd
}
