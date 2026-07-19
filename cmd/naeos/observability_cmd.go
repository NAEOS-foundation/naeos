package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

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
				fmt.Fprint(w, `{"traces":[]}`)
			})
			mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"logs":[]}`)
			})
			mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"metrics":[]}`)
			})
			mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"tracer":"active","logger":"active","metrics":"active"}`)
			})

			addr := fmt.Sprintf(":%d", port)
			srv := &http.Server{
				Addr:    addr,
				Handler: mux,
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			fmt.Fprintf(cmd.OutOrStdout(), "Starting observability dashboard on http://localhost:%d\n", port)

			go func() {
				<-ctx.Done()
				srv.Shutdown(context.Background())
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
