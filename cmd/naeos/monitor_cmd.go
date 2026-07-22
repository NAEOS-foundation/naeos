package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/monitoring"
)

var (
	monitorPort string
)

func newMonitorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start monitoring server with Prometheus metrics",
		Long:  `Start monitoring server exposing Prometheus metrics, health, and readiness endpoints.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			metrics := monitoring.NewMetrics()
			registry := metrics.Registry()

			mux := http.NewServeMux()
			mux.Handle("/metrics", monitoring.PrometheusHandler(registry))
			mux.Handle("/health", monitoring.HealthHandler())
			mux.Handle("/ready", monitoring.ReadyHandler())

			fmt.Printf("Monitor server starting on http://localhost%s\n", monitorPort)
			fmt.Println("  /metrics  - Prometheus metrics")
			fmt.Println("  /health   - Health check")
			fmt.Println("  /ready    - Readiness check")
			srv := &http.Server{
				Addr:              monitorPort,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       15 * time.Second,
				WriteTimeout:      15 * time.Second,
				IdleTimeout:       60 * time.Second,
			}
			return srv.ListenAndServe()
		},
	}

	cmd.Flags().StringVarP(&monitorPort, "port", "p", ":9090", "Monitor server port")

	return cmd
}
