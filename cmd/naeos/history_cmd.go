package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/eventsourcing"
	"github.com/spf13/cobra"
)

func newHistoryCommand() *cobra.Command {
	var storeDir string

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show pipeline run history from persisted events",
		Long: `Display the history of past pipeline runs stored as event files.

Example:
  naeos history
  naeos history --store-dir ./events
  naeos history --store-dir ./events --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if storeDir == "" {
				storeDir = ".naeos/events"
			}

			store := eventsourcing.NewFileStore(storeDir)
			ids, err := store.StreamIDs()
			if err != nil {
				return fmt.Errorf("read event store: %w", err)
			}

			if len(ids) == 0 {
				cmd.OutOrStdout().Write([]byte("No pipeline runs found.\n"))
				return nil
			}

			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Pipeline Run History (%d runs)\n", len(ids))))
			cmd.OutOrStdout().Write([]byte(strings.Repeat("─", 60) + "\n"))

			for _, id := range ids {
				events, err := store.Load(id)
				if err != nil {
					cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %s — error loading: %v\n", id, err)))
					continue
				}
				snap := eventsourcing.RebuildFromEvents(id, events)

				status := snap.Status
				icon := "✓"
				if status == "failed" {
					icon = "✗"
				} else if status == "running" {
					icon = "●"
				}

				name := snap.Name
				if name == "" {
					name = "unknown"
				}

				var duration string
				if len(events) >= 2 {
					start := events[0].Timestamp
					end := events[len(events)-1].Timestamp
					if !start.IsZero() && !end.IsZero() {
						duration = end.Sub(start).Round(time.Millisecond).String()
					}
				}

				durStr := ""
				if duration != "" {
					durStr = fmt.Sprintf(" (%s)", duration)
				}

				cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %s %s | %s | %d events%s\n", icon, id, name, len(events), durStr)))
				if snap.Error != "" {
					cmd.OutOrStdout().Write([]byte(fmt.Sprintf("    error: %s\n", snap.Error)))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&storeDir, "store-dir", "", "path to event store directory (default: .naeos/events)")
	return cmd
}
