// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
	"github.com/spf13/cobra"
)

func newControlPlaneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Inspect the production Control Plane persistence dependencies",
	}
	cmd.AddCommand(newControlPlanePreflightCommand())
	return cmd
}

func newControlPlanePreflightCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "preflight",
		Short: "Validate DATABASE_URL, TLS, PostgreSQL connectivity, and private schema",
		Long:  "Open the dedicated Control Plane PostgreSQL repositories and verify that the expected private schema and ledger functions are installed. This command performs read-only connectivity and schema checks; it does not apply migrations or modify production data.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			databaseURL := os.Getenv("DATABASE_URL")
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			runtime, err := controlplane.OpenProductionRuntime(ctx, databaseURL)
			if err != nil {
				// Avoid echoing a connection URL, which may contain credentials.
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return fmt.Errorf("Control Plane preflight timed out or was canceled")
				}
				return fmt.Errorf("Control Plane preflight failed: %w", err)
			}
			defer runtime.Close()
			fmt.Fprintln(cmd.OutOrStdout(), "Control Plane PostgreSQL connectivity and schema checks passed.")
			return nil
		},
	}
}
