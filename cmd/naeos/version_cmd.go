package main

import (
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show NAEOS version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "naeos %s\n", version.Full())
			return err
		},
	}
}
