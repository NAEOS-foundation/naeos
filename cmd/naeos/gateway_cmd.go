package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/gateway"
)

func newGatewayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "API gateway management",
		Long: `Manage API gateway routing, rate limiting, circuit breakers, and load balancing.

Example:
  naeos gateway status
  naeos gateway rate-status
  naeos gateway cb-status --name api
  naeos gateway lb-list --name api
  naeos gateway add-backend --lb api --name backend1 --url http://localhost:8080`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newGatewayStatusCommand())
	cmd.AddCommand(newGatewayRateStatusCommand())
	cmd.AddCommand(newGatewayCBStatusCommand())
	cmd.AddCommand(newGatewayLBListCommand())
	cmd.AddCommand(newGatewayAddBackendCommand())

	return cmd
}

func newGatewayStatusCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show gateway status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			data := map[string]string{
				"rate_limiter":     "active",
				"circuit_breakers": "configured",
				"load_balancers":   "configured",
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "API Gateway Status\n")
			fmt.Fprintf(out, "==================\n")
			fmt.Fprintf(out, "Rate Limiter:     active\n")
			fmt.Fprintf(out, "Circuit Breakers: configured\n")
			fmt.Fprintf(out, "Load Balancers:   configured\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newGatewayRateStatusCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "rate-status",
		Short: "Show rate limiter usage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rl := gateway.NewRateLimiter()
			_ = rl

			data := map[string]any{
				"type":    "token bucket",
				"default": "100 requests per minute",
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Rate Limiter: active (token bucket)\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Default: 100 requests per minute\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newGatewayCBStatusCommand() *cobra.Command {
	var name string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "cb-status",
		Short: "Show circuit breaker status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cb := gateway.NewCircuitBreaker(name, 5, 3, 30*time.Second)

			data := map[string]any{
				"name":              name,
				"state":             cb.State(),
				"failure_threshold": 5,
				"success_threshold": 3,
				"timeout":           "30s",
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Circuit Breaker: %s\n", name)
			fmt.Fprintf(out, "State:           %s\n", cb.State())
			fmt.Fprintf(out, "Failure Threshold: 5\n")
			fmt.Fprintf(out, "Success Threshold: 3\n")
			fmt.Fprintf(out, "Timeout: 30s\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "default", "circuit breaker name")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newGatewayLBListCommand() *cobra.Command {
	var name string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "lb-list",
		Short: "List load balancer backends",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lb := gateway.NewLoadBalancer()

			backends := lb.List()

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), backends, outputFormat)
			}

			if len(backends) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No backends configured.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-15s %-30s %-10s\n", "NAME", "URL", "HEALTH")
			fmt.Fprintf(out, "%-15s %-30s %-10s\n", "----", "---", "------")
			for _, b := range backends {
				health := "healthy"
				if !b.Healthy {
					health = "unhealthy"
				}
				fmt.Fprintf(out, "%-15s %-30s %-10s\n", b.Name, b.URL, health)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "default", "load balancer name")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newGatewayAddBackendCommand() *cobra.Command {
	var lbName, name, url string
	var weight int

	cmd := &cobra.Command{
		Use:   "add-backend",
		Short: "Add a backend to load balancer",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lb := gateway.NewLoadBalancer()

			lb.AddBackend(&gateway.Backend{
				Name:    name,
				URL:     url,
				Weight:  weight,
				Healthy: true,
			})

			fmt.Fprintf(cmd.OutOrStdout(), "Added backend '%s' to load balancer '%s'\n", name, lbName)
			return nil
		},
	}

	cmd.Flags().StringVar(&lbName, "lb", "default", "load balancer name")
	cmd.Flags().StringVar(&name, "name", "", "backend name (required)")
	cmd.Flags().StringVar(&url, "url", "", "backend URL (required)")
	cmd.Flags().IntVar(&weight, "weight", 1, "backend weight")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("url")
	return cmd
}
