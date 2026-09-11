package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/investordemo"
)

const demoLandingPage = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>NAEOS Investor Demo</title></head>
<body>
<h1>NAEOS Investor Demo Control Plane</h1>
<p>Authorization, policy enforcement, verification, and audit run outside agent control.</p>
<ul>
  <li><a href="/api/health">/api/health</a> — server health</li>
  <li><a href="/api/policy">/api/policy</a> — active policy (POLICY-017)</li>
  <li><a href="/api/grants">/api/grants</a> — active grants (GRANT-001)</li>
  <li><a href="/api/audit">/api/audit</a> — audit events (tamper-evident chain)</li>
  <li><a href="/api/verification">/api/verification</a> — independent session verification</li>
  <li><a href="/api/investor-demo">/api/investor-demo</a> — run the 10-step scripted demo</li>
  <li><a href="/api/scenarios">/api/scenarios</a> — run attack scenario matrix</li>
</ul>
</body>
</html>
`

func newDemoCommand() *cobra.Command {
	var (
		addr string
		bind string
	)

	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Run the investor demo control-plane server",
		Long: `Start the investor demo HTTP server: authorization, policy
enforcement, verification, and audit. Security decisions run outside
agent control.

Example:
  naeos demo --addr :9091`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if bind != "" {
				addr = bind
			}

			setup := investordemo.SetupDemoEnvironment()
			apiServer := investordemo.NewAPIServer(setup)

			mux := http.NewServeMux()
			mux.Handle("/api/", apiServer)
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(demoLandingPage))
			})

			srv := &http.Server{
				Addr:              addr,
				Handler:           mux,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       30 * time.Second,
				WriteTimeout:      30 * time.Second,
				IdleTimeout:       60 * time.Second,
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			fmt.Fprintf(cmd.OutOrStdout(), "NAEOS investor demo listening on http://localhost%s (CTRL+C to stop)\n", addr)

			errCh := make(chan error, 1)
			go func() {
				errCh <- srv.ListenAndServe()
			}()

			select {
			case err := <-errCh:
				return fmt.Errorf("demo server error: %w", err)
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
				fmt.Fprintln(cmd.OutOrStdout(), "Demo server stopped.")
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":9091", "listen address (e.g. :9091)")
	cmd.Flags().StringVar(&bind, "bind", "", "alias for --addr")
	return cmd
}
