package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/rollback"
)

func newRollbackCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to a previous snapshot of generated artifacts",
		Long: `Manage snapshots and rollback generated artifacts.

Example:
  naeos rollback list
  naeos rollback restore <snapshot-id>`,
	}

	rollbackList := &cobra.Command{
		Use:   "list",
		Short: "List available snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := rollback.NewStore(".")
			snaps, err := store.List()
			if err != nil {
				return err
			}
			if len(snaps) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No snapshots found")
				return nil
			}
			for _, s := range snaps {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %s\n", s.ID, s.Timestamp.Format(time.RFC3339))
			}
			return nil
		},
	}

	rollbackRestore := &cobra.Command{
		Use:   "restore [snapshot-id]",
		Short: "Restore artifacts from a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := rollback.NewStore(".")
			if outputDir == "" {
				outputDir = "."
			}
			if err := store.Restore(args[0], outputDir); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Restored snapshot %s to %s\n", args[0], outputDir)
			return nil
		},
	}

	cmd.AddCommand(rollbackList)
	cmd.AddCommand(rollbackRestore)
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "directory to restore artifacts to")
	return cmd
}
