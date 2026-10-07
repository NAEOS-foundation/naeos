// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

func newAgentRequestCommand() *cobra.Command {
	var requestJSON, requestFile, adapterName, outputFmt string
	var sessionID, storePath, invocationID string

	cmd := &cobra.Command{
		Use:   "request",
		Short: "Authorize and execute an agent tool request through NAEOS",
		Long: `Submit a normalized agent tool request to the NAEOS execution gateway.

The request is first evaluated by the active governance policy. Only an allowed request reaches the runtime sandbox. The resulting decision and execution metadata can optionally be persisted to an agent session.

Example:
  naeos agent request --request '{"tool":"file-edit","action":"write","resource":"src/app.go","actor":"codex"}'
  naeos agent request --request-file request.json --session-id sess-123 --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := loadAgentRequestInput(requestJSON, requestFile)
			if err != nil {
				return err
			}
			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("invalid agent request JSON: %w", err)
			}
			reg, err := loadPolicies()
			if err != nil {
				return err
			}
			cp := control.New(reg)
			sb := gateway.NewDefaultSandbox(gateway.SandboxConfig{})
			gw := gateway.New(cp, sb, gateway.WithReplayProtection(true))

			if adapterName == "" {
				adapterName = "json"
			}
			var adapter gateway.AgentAdapter
			switch adapterName {
			case "json":
				adapter = gateway.JSONToolAdapter{}
			case "codex":
				adapter = gateway.CodexToolAdapter{}
			default:
				return fmt.Errorf("unsupported adapter %q", adapterName)
			}
			req, err := adapter.NormalizeTool(raw)
			if err != nil {
				return fmt.Errorf("normalize agent request: %w", err)
			}
			if invocationID != "" {
				req.InvocationID = invocationID
			}
			result, err := gw.Authorize(req)
			if err != nil {
				return err
			}
			if sessionID != "" {
				if err := persistSessionAction(sessionID, storePath, result); err != nil {
					return err
				}
			}
			if outputFmt == "json" {
				encoded, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(encoded)); err != nil {
					return err
				}
				return runtimeResultExit(result)
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Agent:    %s\n", req.Actor); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Tool:     %s/%s\n", req.Tool, req.Action); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Decision: %s\n", result.Decision); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Status:   %s\n", result.Status); err != nil {
				return err
			}
			if result.PolicyID != "" {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Policy:   %s\n", result.PolicyID); err != nil {
					return err
				}
			}
			if result.Hash != "" {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Hash:     %s\n", result.Hash); err != nil {
					return err
				}
			}
			return runtimeResultExit(result)
		},
	}

	cmd.Flags().StringVar(&requestJSON, "request", "", "inline JSON agent tool request")
	cmd.Flags().StringVar(&requestFile, "request-file", "", "path to JSON agent tool request")
	cmd.Flags().StringVar(&adapterName, "adapter", "json", "agent adapter: json or codex")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "persist the decision and execution in an agent session")
	cmd.Flags().StringVar(&storePath, "store-path", "", "path to the agent session store JSON file")
	cmd.Flags().StringVar(&invocationID, "invocation-id", "", "unique invocation identity used for replay protection")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}

func loadAgentRequestInput(inline, file string) ([]byte, error) {
	if inline == "" && file == "" {
		return nil, fmt.Errorf("one of --request or --request-file is required")
	}
	if inline != "" && file != "" {
		return nil, fmt.Errorf("--request and --request-file are mutually exclusive")
	}
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read agent request: %w", err)
		}
		return data, nil
	}
	return []byte(inline), nil
}
