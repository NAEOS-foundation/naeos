// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/eventsourcing"
)

func newEventsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Event sourcing commands for pipeline audit trail and replay",
	}

	cmd.AddCommand(newEventsReplayCommand())
	cmd.AddCommand(newEventsListCommand())

	return cmd
}

func newEventsReplayCommand() *cobra.Command {
	var inputFile string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "replay",
		Short: "Replay events to reconstruct pipeline state",
		Long: `Replay a series of events to reconstruct the state of a pipeline run.
Events can be loaded from a JSON file or standard input.

Example:
  naeos events replay --input events.json
  cat events.json | naeos events replay`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("read events file: %w", err)
			}

			var events []eventsourcing.Event
			if err := json.Unmarshal(data, &events); err != nil {
				return fmt.Errorf("parse events: %w", err)
			}

			if len(events) == 0 {
				_, _ = cmd.OutOrStdout().Write([]byte("No events to replay.\n"))
				return nil
			}

			streamID := "replayed"
			if len(events) > 0 && events[0].StreamID != "" {
				streamID = events[0].StreamID
			}

			snap := eventsourcing.RebuildFromEvents(streamID, events)

			output, err := json.MarshalIndent(snap, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal snapshot: %w", err)
			}

			if outputPath != "" {
				if err := os.WriteFile(outputPath, output, 0o600); err != nil {
					return fmt.Errorf("write output: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Replayed %d events → %s\n", len(events), outputPath)
			} else {
				_, _ = cmd.OutOrStdout().Write(output)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "path to events JSON file (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "optional output file path")
	_ = cmd.MarkFlagRequired("input")

	return cmd
}

func newEventsListCommand() *cobra.Command {
	var inputFile string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events from a stored event file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("read events file: %w", err)
			}

			var events []eventsourcing.Event
			if err := json.Unmarshal(data, &events); err != nil {
				return fmt.Errorf("parse events: %w", err)
			}

			for _, e := range events {
				_, _ = cmd.OutOrStdout().Write([]byte(eventsourcing.FormatEvent(e) + "\n"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "path to events JSON file (required)")
	_ = cmd.MarkFlagRequired("input")

	return cmd
}

func init() {
	_ = filepath.Clean
}
