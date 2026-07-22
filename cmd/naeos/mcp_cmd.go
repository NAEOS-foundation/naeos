package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/compiler"
	contextbundle "github.com/NAEOS-foundation/naeos/internal/context/bundle"
	"github.com/NAEOS-foundation/naeos/internal/mcp"
)

func newMCPCommand() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP (Model Context Protocol) server",
		Long: `Start an MCP server that exposes NAEOS tools to AI agents.

The server implements the Model Context Protocol and provides tools:
  - parse_spec: Parse a NAEOS specification
  - validate_spec: Validate a specification
  - generate_context: Generate AI context bundle
  - compile_spec: Compile spec to AI instruction sets
  - explain_concept: Explain NAEOS concepts

Example:
  naeos mcp --port 8080
  naeos mcp`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			compiler := compiler.New()
			bundle := contextbundle.NewGenerator(nil)
			server := mcp.NewServer(compiler, bundle)

			addr := fmt.Sprintf(":%d", port)
			fmt.Fprintf(cmd.OutOrStdout(), "NAEOS MCP server starting on %s\n", addr)
			fmt.Fprintf(cmd.OutOrStdout(), "Tools: parse_spec, validate_spec, generate_context, compile_spec, explain_concept\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Health check: http://localhost:%d/health\n", port)

			srv := &http.Server{
				Addr:              addr,
				Handler:           server.Handler(),
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       15 * time.Second,
				WriteTimeout:      15 * time.Second,
				IdleTimeout:       60 * time.Second,
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
			}()

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("MCP server error: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "MCP server stopped.")
			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", 3000, "port for MCP server")

	return cmd
}
