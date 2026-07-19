package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/broker"
)

func newBrokerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "broker",
		Short: "Message broker management",
		Long: `Manage message broker connections (Redis, NATS, Memory, etc.).

Example:
  naeos broker connect --type redis --name myredis --host localhost --port 6379
  naeos broker connect --type nats --name mynats --host localhost --port 4222
  naeos broker connect --type memory --name mymem
  naeos broker list
  naeos broker publish --name myredis --channel events --message '{"event":"created"}'
  naeos broker subscribe --name myredis --channel events
  naeos broker disconnect --name myredis`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newBrokerConnectCommand())
	cmd.AddCommand(newBrokerListCommand())
	cmd.AddCommand(newBrokerPublishCommand())
	cmd.AddCommand(newBrokerDisconnectCommand())

	return cmd
}

func newBrokerConnectCommand() *cobra.Command {
	var brokerType, name, host, password string
	var port, db int

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a message broker",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := broker.NewConnectionStore()

			cfg := &broker.Config{
				Host:     host,
				Port:     port,
				Password: password,
				DB:       db,
			}

			b := broker.New(brokerType)
			if b == nil {
				return fmt.Errorf("unsupported broker type: %s", brokerType)
			}

			if err := b.Connect(cfg); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

			if err := store.Add(name, brokerType, cfg); err != nil {
				return fmt.Errorf("failed to persist connection: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Connected to %s broker '%s'\n", brokerType, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&brokerType, "type", "redis", "broker type (redis, nats, memory)")
	cmd.Flags().StringVar(&name, "name", "", "connection name (required)")
	cmd.Flags().StringVar(&host, "host", "localhost", "broker host")
	cmd.Flags().IntVar(&port, "port", 6379, "broker port")
	cmd.Flags().StringVar(&password, "password", "", "broker password")
	cmd.Flags().IntVar(&db, "db", 0, "Redis database number")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newBrokerListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all broker connections",
		Long:  "List all broker connections, including those persisted across CLI invocations.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := broker.NewConnectionStore()

			saved, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list connections: %w", err)
			}

			if len(saved) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No broker connections.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-15s %-10s %-15s\n", "NAME", "TYPE", "HOST")
			fmt.Fprintf(out, "%-15s %-10s %-15s\n", "----", "----", "----")
			for _, s := range saved {
				host := "localhost"
				if s.Config != nil && s.Config.Host != "" {
					host = s.Config.Host
				}
				fmt.Fprintf(out, "%-15s %-10s %-15s\n", s.Name, s.Driver, host)
			}
			return nil
		},
	}
}

func newBrokerPublishCommand() *cobra.Command {
	var name, channel, message string

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a message to a channel",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := broker.NewManager()

			b := broker.New(mustGetDriverFromStore(name))
			if b == nil {
				return fmt.Errorf("broker '%s' not found", name)
			}

			store := broker.NewConnectionStore()
			saved, err := store.Get(name)
			if err != nil {
				return fmt.Errorf("broker '%s' not found: %w", name, err)
			}

			if err := b.Connect(saved.Config); err != nil {
				return fmt.Errorf("failed to reconnect: %w", err)
			}

			mgr.Register(name, b)

			msg := &broker.Message{
				Payload: []byte(message),
			}

			if err := b.Publish(channel, msg); err != nil {
				return fmt.Errorf("failed to publish: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Published to '%s' on '%s'\n", channel, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "broker name (required)")
	cmd.Flags().StringVar(&channel, "channel", "", "channel name (required)")
	cmd.Flags().StringVar(&message, "message", "", "message payload (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("channel")
	_ = cmd.MarkFlagRequired("message")
	return cmd
}

func newBrokerDisconnectCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect from a broker",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := broker.NewConnectionStore()

			if err := store.Remove(name); err != nil {
				return fmt.Errorf("failed to remove connection: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Disconnected from '%s'.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "broker name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func mustGetDriverFromStore(name string) string {
	store := broker.NewConnectionStore()
	saved, err := store.Get(name)
	if err != nil {
		return "redis"
	}
	return saved.Driver
}
