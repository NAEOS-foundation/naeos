package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/performance"
)

func newPerfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perf",
		Short: "Performance optimization tools",
		Long: `Manage connection pools, batch processing, and caching.

Example:
  naeos perf pool-create --name db --min 2 --max 10
  naeos perf pool-acquire --name db
  naeos perf pool-stats --name db
  naeos perf cache-set --key mykey --value myvalue --ttl 60s
  naeos perf cache-get --key mykey
  naeos perf cache-stats`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newPerfPoolCreateCommand())
	cmd.AddCommand(newPerfPoolAcquireCommand())
	cmd.AddCommand(newPerfPoolStatsCommand())
	cmd.AddCommand(newPerfCacheSetCommand())
	cmd.AddCommand(newPerfCacheGetCommand())
	cmd.AddCommand(newPerfCacheStatsCommand())

	return cmd
}

func newPerfPoolCreateCommand() *cobra.Command {
	var name string
	var min, max int

	cmd := &cobra.Command{
		Use:   "pool-create",
		Short: "Create a connection pool",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pool := performance.NewConnectionPool(name, min, max)

			total, avail, inUse := pool.Stats()
			fmt.Fprintf(cmd.OutOrStdout(), "Created pool '%s' (min=%d, max=%d)\n", name, min, max)
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %d total, %d available, %d in use\n", total, avail, inUse)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "pool name (required)")
	cmd.Flags().IntVar(&min, "min", 2, "minimum connections")
	cmd.Flags().IntVar(&max, "max", 10, "maximum connections")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newPerfPoolAcquireCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "pool-acquire",
		Short: "Acquire a connection from pool",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pool := performance.NewConnectionPool(name, 2, 10)

			conn, ok := pool.Acquire()
			if !ok {
				return fmt.Errorf("pool '%s' is full", name)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Acquired connection: %s\n", conn.ID)
			pool.Release(conn)
			fmt.Fprintf(cmd.OutOrStdout(), "Released connection: %s\n", conn.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "pool name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newPerfPoolStatsCommand() *cobra.Command {
	var name string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "pool-stats",
		Short: "Show connection pool statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pool := performance.NewConnectionPool(name, 2, 10)

			total, avail, inUse := pool.Stats()

			data := map[string]int{
				"total":     total,
				"available": avail,
				"in_use":    inUse,
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Pool: %s\n", name)
			fmt.Fprintf(out, "Total:     %d\n", total)
			fmt.Fprintf(out, "Available: %d\n", avail)
			fmt.Fprintf(out, "In Use:    %d\n", inUse)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "pool name (required)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newPerfCacheSetCommand() *cobra.Command {
	var key, value string
	var ttl time.Duration

	cmd := &cobra.Command{
		Use:   "cache-set",
		Short: "Set a cache value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cache := performance.NewCache("naeos")

			cache.Set(key, value, ttl)
			fmt.Fprintf(cmd.OutOrStdout(), "Cached '%s' (TTL: %s)\n", key, ttl)
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "cache key (required)")
	cmd.Flags().StringVar(&value, "value", "", "cache value (required)")
	cmd.Flags().DurationVar(&ttl, "ttl", 5*time.Minute, "time to live")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("value")
	return cmd
}

func newPerfCacheGetCommand() *cobra.Command {
	var key string

	cmd := &cobra.Command{
		Use:   "cache-get",
		Short: "Get a cache value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cache := performance.NewCache("naeos")

			val, ok := cache.Get(key)
			if !ok {
				fmt.Fprintf(cmd.OutOrStdout(), "Cache miss: '%s'\n", key)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%v\n", val)
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "cache key (required)")
	cmd.MarkFlagRequired("key")
	return cmd
}

func newPerfCacheStatsCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "cache-stats",
		Short: "Show cache statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cache := performance.NewCache("naeos")

			data := map[string]int{
				"entries": cache.Size(),
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Cache: naeos\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Entries: %d\n", cache.Size())
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}
