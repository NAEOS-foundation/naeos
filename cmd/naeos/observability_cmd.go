package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/audit"
	"github.com/NAEOS-foundation/naeos/internal/monitoring"
	"github.com/NAEOS-foundation/naeos/internal/observability"
)

func newObservabilityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observability",
		Short: "Observability and telemetry management",
		Long: `Manage tracing, logging, and metrics collection.

Example:
  naeos observability trace --name "http-request"
  naeos observability log --level info --message "Server started"
  naeos observability metrics
  naeos observability status`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newObsTraceCommand())
	cmd.AddCommand(newObsLogCommand())
	cmd.AddCommand(newObsMetricsCommand())
	cmd.AddCommand(newObsStatusCommand())
	cmd.AddCommand(newObsDashboardCommand())
	cmd.AddCommand(newObsExportCommand())
	cmd.AddCommand(newObsSIEMCommand())
	cmd.AddCommand(newObsSLOCommand())
	return cmd
}

func newObsTraceCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "trace",
		Short: "Create a new trace span",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stack := observability.NewStack("naeos")

			span := stack.Tracer.StartSpan(name)
			stack.Tracer.SetStatus(span, observability.SpanStatusOK, "completed")
			stack.Tracer.EndSpan(span)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Trace: %s\n", span.TraceID)
			fmt.Fprintf(out, "Span:  %s (%s)\n", span.Name, span.SpanID)
			fmt.Fprintf(out, "Status: OK\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "operation", "span name")
	return cmd
}

func newObsLogCommand() *cobra.Command {
	var level, message, source string

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Write a log entry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := observability.NewLogger(source, observability.LogLevelInfo)

			switch level {
			case "debug":
				logger.Debug(message, nil)
			case "info":
				logger.Info(message, nil)
			case "warn":
				logger.Warn(message, nil)
			case "error":
				logger.Error(message, nil)
			default:
				logger.Info(message, nil)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n", level, source, message)
			return nil
		},
	}

	cmd.Flags().StringVar(&level, "level", "info", "log level (debug, info, warn, error)")
	cmd.Flags().StringVar(&message, "message", "", "log message (required)")
	cmd.Flags().StringVar(&source, "source", "naeos", "log source")
	_ = cmd.MarkFlagRequired("message")
	return cmd
}

func newObsMetricsCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show collected metrics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mc := observability.NewMetricsCollector("naeos")

			mc.Counter("requests_total", map[string]string{"method": "GET"})
			mc.Gauge("active_connections", 42, nil)
			mc.Histogram("request_duration_ms", 125.5, nil)

			metrics := mc.GetMetrics()

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), metrics, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-25s %-12s %-10s %s\n", "NAME", "TYPE", "VALUE", "LABELS")
			fmt.Fprintf(out, "%-25s %-12s %-10s %s\n", "----", "----", "-----", "------")
			for _, m := range metrics {
				labels := ""
				for k, v := range m.Labels {
					labels += k + "=" + v + " "
				}
				fmt.Fprintf(out, "%-25s %-12s %-10.1f %s\n",
					m.Name, m.Type, m.Value, labels)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newObsStatusCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show observability stack status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			data := map[string]string{
				"tracer":  "active",
				"logger":  "active (level: info)",
				"metrics": "active",
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Observability Stack\n")
			fmt.Fprintf(out, "===================\n")
			fmt.Fprintf(out, "Tracer:   active\n")
			fmt.Fprintf(out, "Logger:   active (level: info)\n")
			fmt.Fprintf(out, "Metrics:  active\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newObsDashboardCommand() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Start the observability dashboard",
		Long: `Start a local web dashboard for viewing traces, logs, and metrics.

Example:
  naeos observability dashboard
  naeos observability dashboard --port 9090`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stack := observability.NewStack("naeos-dashboard")

			span := stack.Tracer.StartSpan("pipeline.execute")
			stack.Tracer.SetStatus(span, observability.SpanStatusOK, "completed")
			stack.Tracer.EndSpan(span)

			stack.Logger.Info("pipeline started", map[string]any{"spec": "demo"})
			stack.Logger.Info("pipeline completed", map[string]any{"duration_ms": 125})
			stack.Logger.Warn("slow pipeline detected", map[string]any{"threshold_ms": 100})

			stack.Metrics.Counter("pipelines_total", map[string]string{"status": "success"})
			stack.Metrics.Counter("pipelines_total", map[string]string{"status": "success"})
			stack.Metrics.Gauge("active_workers", 3, nil)
			stack.Metrics.Histogram("pipeline_duration_ms", 125.5, nil)
			stack.Metrics.Histogram("pipeline_duration_ms", 89.2, nil)

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, `<!DOCTYPE html>
<html><head><title>NAEOS Observability Dashboard</title></head>
<body>
<h1>NAEOS Observability Dashboard</h1>
<h2>Available Endpoints</h2>
<ul>
  <li><a href="/traces">GET /traces</a> — View traces</li>
  <li><a href="/logs">GET /logs</a> — View logs</li>
  <li><a href="/metrics">GET /metrics</a> — View metrics</li>
  <li><a href="/status">GET /status</a> — System status</li>
</ul>
</body></html>`)
			})
			mux.HandleFunc("/traces", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				spans := stack.Tracer.GetSpans()
				type spanJSON struct {
					TraceID   string `json:"trace_id"`
					SpanID    string `json:"span_id"`
					Name      string `json:"name"`
					Status    string `json:"status"`
					StartTime string `json:"start_time"`
					EndTime   string `json:"end_time"`
				}
				result := make([]spanJSON, 0, len(spans))
				for _, s := range spans {
					status := "unset"
					switch s.Status.Code {
					case observability.SpanStatusOK:
						status = "ok"
					case observability.SpanStatusError:
						status = "error"
					}
					result = append(result, spanJSON{
						TraceID:   s.TraceID,
						SpanID:    s.SpanID,
						Name:      s.Name,
						Status:    status,
						StartTime: s.StartTime.Format("2006-01-02T15:04:05Z"),
						EndTime:   s.EndTime.Format("2006-01-02T15:04:05Z"),
					})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"traces": result})
			})
			mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				entries := stack.Logger.GetEntries()
				type logJSON struct {
					Timestamp string         `json:"timestamp"`
					Level     string         `json:"level"`
					Message   string         `json:"message"`
					Source    string         `json:"source"`
					Attrs     map[string]any `json:"attributes,omitempty"`
				}
				result := make([]logJSON, 0, len(entries))
				for _, e := range entries {
					level := "info"
					switch e.Level {
					case observability.LogLevelDebug:
						level = "debug"
					case observability.LogLevelWarn:
						level = "warn"
					case observability.LogLevelError:
						level = "error"
					}
					result = append(result, logJSON{
						Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z"),
						Level:     level,
						Message:   e.Message,
						Source:    e.Source,
						Attrs:     e.Attributes,
					})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"logs": result})
			})
			mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				metrics := stack.Metrics.GetMetrics()
				_ = json.NewEncoder(w).Encode(map[string]any{"metrics": metrics})
			})
			mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"tracer":"active","logger":"active","metrics":"active"}`)
			})

			addr := fmt.Sprintf(":%d", port)
			srv := &http.Server{
				Addr:              addr,
				Handler:           mux,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       15 * time.Second,
				WriteTimeout:      15 * time.Second,
				IdleTimeout:       60 * time.Second,
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			fmt.Fprintf(cmd.OutOrStdout(), "Starting observability dashboard on http://localhost:%d\n", port)

			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = srv.Shutdown(shutdownCtx)
			}()

			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("dashboard server error: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Dashboard stopped.")
			return nil
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 9090, "dashboard port")
	return cmd
}

func newObsExportCommand() *cobra.Command {
	var (
		endpoint string
		count    int
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export collected spans to an OTLP collector",
		Long: `Export collected spans to an OpenTelemetry collector via OTLP/HTTP.

Example:
  naeos observability export --endpoint http://localhost:4318 --count 3`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if endpoint == "" {
				return fmt.Errorf("--endpoint is required (e.g. http://localhost:4318)")
			}

			stack := observability.NewStack("naeos")
			for i := 0; i < count; i++ {
				span := stack.Tracer.StartSpan(fmt.Sprintf("pipeline.execute.%d", i))
				span.Attributes["instance"] = i
				stack.Tracer.SetStatus(span, observability.SpanStatusOK, "completed")
				stack.Tracer.EndSpan(span)
			}

			exp := observability.NewOTLPHTTPExporter(endpoint)
			if err := exp.ExportSpans(stack.Tracer.GetSpans()); err != nil {
				return fmt.Errorf("export: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Exported %d spans to %s/v1/traces\n", count, endpoint)
			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "endpoint", "", "OTLP collector base URL (required)")
	cmd.Flags().IntVar(&count, "count", 1, "number of sample spans to export")
	return cmd
}

func newObsSIEMCommand() *cobra.Command {
	var (
		endpoint string
		format   string
	)

	cmd := &cobra.Command{
		Use:   "siem",
		Short: "Export sample audit events to a SIEM collector",
		Long: `Export a sample audit event to a SIEM collector using CEF (default)
or NDJSON framing.

Example:
  naeos observability siem --endpoint http://localhost:9000
  naeos observability siem --endpoint http://localhost:9000 --format json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if endpoint == "" {
				return fmt.Errorf("--endpoint is required (e.g. http://localhost:9000)")
			}

			opts := []observability.SIEMOption{}
			if format == "json" {
				opts = append(opts, observability.WithSIEMFormat(observability.SIEMJson))
			}

			exp := observability.NewSIEMExporter(endpoint, opts...)
			event := audit.AuditEvent{
				ID:       "sample-001",
				UserID:   "agent-payment-01",
				Action:   "execute",
				Resource: "production.deploy",
				Status:   "blocked",
				Details:  "CAPABILITY_NOT_GRANTED",
				Metadata: map[string]string{"decision": "deny"},
			}
			if err := exp.ExportEvent(event); err != nil {
				return fmt.Errorf("siem export: %w", err)
			}

			frame := observability.FormatCEF(event, "naeos")
			fmt.Fprintf(cmd.OutOrStdout(), "Exported audit event to %s\n", endpoint)
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", frame)
			return nil
		},
	}

	cmd.Flags().StringVar(&endpoint, "endpoint", "", "SIEM collector URL (required)")
	cmd.Flags().StringVar(&format, "format", "cef", "framing format: cef or json")
	return cmd
}

func newObsSLOCommand() *cobra.Command {
	var service string

	cmd := &cobra.Command{
		Use:   "slo",
		Short: "Show a reference SLO with burn-rate alert rules",
		Long: `Show a reference SLO (99% availability, 30d window) with its error
budget and Prometheus burn-rate alerting rules.

Example:
  naeos observability slo --service orders-api`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			slo := monitoring.DefaultSLO(service)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "SLO: %s (%s)\n", slo.Name, service)
			fmt.Fprintf(out, "Target:            %.2f%%\n", slo.Target*100)
			fmt.Fprintf(out, "Window:            %s\n", slo.Window)
			fmt.Fprintf(out, "Error budget:      %.0f seconds\n", slo.ErrorBudget().Seconds())
			fmt.Fprintf(out, "Burn-rate alerts:  \n")
			for _, r := range slo.AlertRules() {
				fmt.Fprintf(out, "  - %-8s %-4ss at %gx\n", r.Alert, "for "+r.For, r.BurnRate)
			}
			fmt.Fprintf(out, "\nPrometheus alerting rules:\n%s", slo.RulesYAML())
			return nil
		},
	}

	cmd.Flags().StringVar(&service, "service", "naeos", "service name for the SLO")
	return cmd
}
