package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newKernelCommand() *cobra.Command {
	var configPath, outputFormat, topic, payload string

	cmd := &cobra.Command{
		Use:   "kernel",
		Short: "Inspect the NAEOS kernel and service registry",
		Long: `Inspect and interact with the NAEOS kernel services, metrics, and event bus.

Example:
  naeos kernel services --config config.yaml
  naeos kernel metrics --config config.yaml --output json
  naeos kernel publish --topic my-topic --payload hello
  naeos kernel subscribe --topic my-topic --payload hello`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "services",
		Short: "List registered kernel services",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			cfg, err := pipeline.ConfigFromFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			p, err := pipeline.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			services := p.RegisteredKernelServices()
			rendered, err := renderOutput(services, outputFormat, func() []byte {
				return []byte(strings.Join(services, "\n") + "\n")
			})
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(rendered)
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "metrics",
		Short: "Show kernel telemetry metrics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			cfg, err := pipeline.ConfigFromFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			p, err := pipeline.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			metrics := p.KernelMetrics()
			rendered, err := renderOutput(metrics, outputFormat, func() []byte {
				return []byte(fmt.Sprintf("events=%d\nlast_event=%s\n", metrics.Events, metrics.LastEvent.Name))
			})
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(rendered)
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "events",
		Short: "List active kernel event topics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			cfg, err := pipeline.ConfigFromFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			p, err := pipeline.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			topics := p.KernelTopics()
			rendered, err := renderOutput(topics, outputFormat, func() []byte {
				return []byte(strings.Join(topics, "\n") + "\n")
			})
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(rendered)
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "publish",
		Short: "Publish an event to the kernel event bus",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if topic == "" {
				return fmt.Errorf("missing required --topic")
			}

			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			cfg, err := pipeline.ConfigFromFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			p, err := pipeline.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			if err := p.Publish(topic, payload); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "published topic=%s payload=%v\n", topic, payload)
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to a kernel event topic and optionally publish a sample payload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if topic == "" {
				return fmt.Errorf("missing required --topic")
			}

			resolved, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			cfg, err := pipeline.ConfigFromFile(resolved)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			p, err := pipeline.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			var received any
			if err := p.Subscribe(topic, func(v any) {
				received = v
			}); err != nil {
				return err
			}

			if payload != "" {
				if err := p.Publish(topic, payload); err != nil {
					return err
				}
			}

			rendered, err := renderOutput(map[string]any{"topic": topic, "received": received}, outputFormat, func() []byte {
				return []byte(fmt.Sprintf("topic=%s received=%v\n", topic, received))
			})
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(rendered)
			return err
		},
	})

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.PersistentFlags().StringVar(&outputFormat, "output", "text", "output format: text, json, or yaml")
	cmd.PersistentFlags().StringVar(&topic, "topic", "", "kernel event topic")
	cmd.PersistentFlags().StringVar(&payload, "payload", "", "event payload to publish")
	return cmd
}
