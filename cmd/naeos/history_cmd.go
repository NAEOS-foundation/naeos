// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/eventsourcing"
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
				_, _ = cmd.OutOrStdout().Write([]byte("No pipeline runs found.\n"))
				return nil
			}

			type historyEntry struct {
				ID     string `json:"id" yaml:"id"`
				Name   string `json:"name" yaml:"name"`
				Status string `json:"status" yaml:"status"`
				Events int    `json:"events" yaml:"events"`
				Dur    string `json:"duration,omitempty" yaml:"duration,omitempty"`
				Error  string `json:"error,omitempty" yaml:"error,omitempty"`
			}

			var entries []historyEntry
			for _, id := range ids {
				events, err := store.Load(id)
				if err != nil {
					entries = append(entries, historyEntry{ID: id, Status: "error", Error: err.Error()})
					continue
				}
				snap := eventsourcing.RebuildFromEvents(id, events)

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

				entries = append(entries, historyEntry{
					ID:     id,
					Name:   name,
					Status: snap.Status,
					Events: len(events),
					Dur:    duration,
					Error:  snap.Error,
				})
			}

			switch cliOutputFormat {
			case "json":
				return FormatOutput(cmd.OutOrStdout(), entries, "json")
			case "yaml":
				return FormatOutput(cmd.OutOrStdout(), entries, "yaml")
			default:
				fmt.Fprintf(cmd.OutOrStdout(), "Pipeline Run History (%d runs)\n", len(ids))
				_, _ = cmd.OutOrStdout().Write([]byte(strings.Repeat("─", 60) + "\n"))
				for _, e := range entries {
					icon := "✓"
					switch e.Status {
					case "failed", "error":
						icon = "✗"
					case "running":
						icon = "●"
					}
					durStr := ""
					if e.Dur != "" {
						durStr = fmt.Sprintf(" (%s)", e.Dur)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  %s %s | %s | %d events%s\n", icon, e.ID, e.Name, e.Events, durStr)
					if e.Error != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "    error: %s\n", e.Error)
					}
				}
				return nil
			}
		},
	}

	cmd.Flags().StringVar(&storeDir, "store-dir", "", "path to event store directory (default: .naeos/events)")
	return cmd
}
