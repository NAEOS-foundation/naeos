package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/demoobs"
	"github.com/NAEOS-foundation/naeos/internal/investordemo"
	"github.com/NAEOS-foundation/naeos/internal/observability"
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
		addr         string
		bind         string
		otlpEndpoint string
		siemEndpoint string
		siemFormat   string
		tenantID     string
	)

	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Run the investor demo control-plane server",
		Long: `Start the investor demo HTTP server: authorization, policy
enforcement, verification, and audit. Security decisions run outside
agent control.

Observability: pass --siem-endpoint to forward every audit event to a SIEM
collector (CEF or NDJSON framing), and --otlp-endpoint to export request
traces to an OTLP/HTTP collector.

Example:
  naeos demo --addr :9091
  naeos demo --addr :9091 --siem-endpoint http://localhost:9000 --otlp-endpoint http://localhost:4318`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if bind != "" {
				addr = bind
			}

			logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), nil))

			obsCfg := demoobs.Config{
				OTLPEndpoint: otlpEndpoint,
				SIEMEndpoint: siemEndpoint,
				SIEMFormat:   siemFormat,
				TenantID:     tenantID,
			}

			var siem *demoobs.SIEMForwarder
			if obsCfg.HasSIEM() {
				siem = demoobs.NewSIEMForwarder(
					obsCfg.SIEMEndpoint,
					obsCfg.SIEMFormatValue(),
					demoobs.WithTenant(obsCfg.TenantID),
					demoobs.WithLogger(logger),
				)
				defer siem.Close()
			}

			var tracer *observability.Tracer
			var otlpExp observability.Exporter
			if obsCfg.HasOTLP() {
				tracer = demoobs.NewDemoTracer()
				otlpExp = observability.NewOTLPHTTPExporter(obsCfg.OTLPEndpoint)
			}

			setup := investordemo.SetupDemoEnvironment()
			if siem != nil {
				setup.AuditLedger.SetObserver(siem)
			}
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

			var handler http.Handler = mux
			if tracer != nil && otlpExp != nil {
				handler = demoobs.TracingMiddleware(tracer, otlpExp, logger)(handler)
			}

			srv := &http.Server{
				Addr:              addr,
				Handler:           handler,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       30 * time.Second,
				WriteTimeout:      30 * time.Second,
				IdleTimeout:       60 * time.Second,
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			fmt.Fprintf(cmd.OutOrStdout(), "NAEOS investor demo listening on http://localhost%s (CTRL+C to stop)\n", addr)
			if siem != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  SIEM forwarding:  %s (%s)\n", obsCfg.SIEMEndpoint, obsCfg.SIEMFormatName())
				if obsCfg.TenantID != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  SIEM tenant:      %s\n", obsCfg.TenantID)
				}
			}
			if otlpExp != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  OTLP tracing:     %s/v1/traces\n", obsCfg.OTLPEndpoint)
			}

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
	cmd.Flags().StringVar(&otlpEndpoint, "otlp-endpoint", "", "OTLP/HTTP collector base URL (e.g. http://localhost:4318)")
	cmd.Flags().StringVar(&siemEndpoint, "siem-endpoint", "", "SIEM collector URL to forward audit events (e.g. http://localhost:9000)")
	cmd.Flags().StringVar(&siemFormat, "siem-format", "cef", "SIEM framing format: cef or json")
	cmd.Flags().StringVar(&tenantID, "tenant-id", "", "tenant identifier attached to exported events")
	return cmd
}
