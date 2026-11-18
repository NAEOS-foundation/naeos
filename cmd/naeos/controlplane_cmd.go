// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

func newControlPlaneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control-plane",
		Short: "Inspect the production Control Plane persistence dependencies",
	}
	cmd.AddCommand(newControlPlanePreflightCommand(), newControlPlaneServeCommand())
	return cmd
}

func newControlPlanePreflightCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "preflight",
		Short: "Validate the environment-scoped database URL, TLS, PostgreSQL connectivity, and private schema",
		Long:  "Open the dedicated Control Plane PostgreSQL repositories and verify that the environment-scoped Supabase project, expected private schema, and ledger functions are installed. This command performs read-only connectivity and schema checks; it does not apply migrations or modify data.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			databaseURL := os.Getenv("NAEOS_CONTROLPLANE_DATABASE_URL")
			environment := os.Getenv("NAEOS_CONTROLPLANE_ENVIRONMENT")
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			runtime, err := controlplane.OpenProductionRuntime(ctx, databaseURL, environment)
			if err != nil {
				// Avoid echoing a connection URL, which may contain credentials.
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return fmt.Errorf("control plane preflight timed out or was canceled")
				}
				return fmt.Errorf("control plane preflight failed: %w", err)
			}
			defer runtime.Close()
			fmt.Fprintln(cmd.OutOrStdout(), "Control Plane PostgreSQL connectivity and schema checks passed.")
			return nil
		},
	}
}

func newControlPlaneServeCommand() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the production Control Plane with JWKS authentication and repository authorization",
		Long:  "Require NAEOS_CONTROLPLANE_DATABASE_URL, NAEOS_CONTROLPLANE_ENVIRONMENT, and NAEOS_CONTROLPLANE_ISSUER, NAEOS_CONTROLPLANE_AUDIENCE, and NAEOS_CONTROLPLANE_JWKS_URL. Verify the database project target, initialize PostgreSQL, and fetch the JWKS keyset before binding the HTTP listener.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			databaseURL := os.Getenv("NAEOS_CONTROLPLANE_DATABASE_URL")
			environment := os.Getenv("NAEOS_CONTROLPLANE_ENVIRONMENT")
			openCtx, cancelOpen := context.WithTimeout(ctx, 10*time.Second)
			runtime, err := controlplane.OpenProductionRuntime(openCtx, databaseURL, environment)
			cancelOpen()
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return errors.New("control plane startup canceled or timed out")
				}
				return fmt.Errorf("control plane startup failed: %w", err)
			}
			defer runtime.Close()

			identity := controlplane.ProductionIdentityConfig{
				Issuer:   os.Getenv("NAEOS_CONTROLPLANE_ISSUER"),
				Audience: os.Getenv("NAEOS_CONTROLPLANE_AUDIENCE"),
				JWKSURL:  os.Getenv("NAEOS_CONTROLPLANE_JWKS_URL"),
			}
			if err := runtime.ConfigureIdentity(identity); err != nil {
				return fmt.Errorf("control plane identity initialization failed: %w", err)
			}

			if addr == "" {
				addr = os.Getenv("NAEOS_CONTROLPLANE_ADDR")
			}
			if addr == "" {
				addr = ":8080"
			}
			server := &http.Server{
				Addr:              addr,
				Handler:           runtime.ProductionHandler(),
				ReadHeaderTimeout: 5 * time.Second,
				ReadTimeout:       15 * time.Second,
				WriteTimeout:      30 * time.Second,
				IdleTimeout:       60 * time.Second,
				MaxHeaderBytes:    1 << 20,
			}
			serveErr := make(chan error, 1)
			go func() {
				serveErr <- server.ListenAndServe()
			}()

			select {
			case <-ctx.Done():
				shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancelShutdown()
				if err := server.Shutdown(shutdownCtx); err != nil {
					return fmt.Errorf("control plane graceful shutdown failed: %w", err)
				}
				return nil
			case err := <-serveErr:
				if errors.Is(err, http.ErrServerClosed) {
					return nil
				}
				return fmt.Errorf("control plane HTTP server stopped: %w", err)
			}
		},
	}
	cmd.Flags().StringVar(&addr, "addr", "", "HTTP listen address (default from NAEOS_CONTROLPLANE_ADDR or :8080)")
	return cmd
}
