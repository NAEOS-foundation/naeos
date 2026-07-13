package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration",
		Short: "Database migration management",
	}

	cmd.AddCommand(newMigrationStatusCommand())

	return cmd
}

func newMigrationStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long: `Display the current migration status of all configured databases.

Example:
  naeos migration status
  naeos migration status --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var sb strings.Builder
			sb.WriteString("Migration Status\n")
			sb.WriteString(strings.Repeat("─", 50) + "\n")
			sb.WriteString(fmt.Sprintf("  %-20s %-15s %-15s\n", "Database", "Version", "Status"))
			sb.WriteString(strings.Repeat("─", 50) + "\n")
			sb.WriteString(fmt.Sprintf("  %-20s %-15s %-15s\n", "postgresql", "0", "not connected"))
			sb.WriteString(fmt.Sprintf("  %-20s %-15s %-15s\n", "mysql", "0", "not connected"))
			sb.WriteString(fmt.Sprintf("  %-20s %-15s %-15s\n", "sqlite", "0", "not connected"))
			sb.WriteString(strings.Repeat("─", 50) + "\n")
			sb.WriteString("Note: Connect to a database to see actual migration status.\n")

			cmd.OutOrStdout().Write([]byte(sb.String()))
			return nil
		},
	}

	return cmd
}
