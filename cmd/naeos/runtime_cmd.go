package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

func newRuntimeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtime",
		Short: "Runtime gateway for authorized agent execution",
		Long:  `Evaluate and execute agent tool requests through the governance control plane. Every request is authorized before execution.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newRuntimeExecCommand())
	cmd.AddCommand(newRuntimeRestrictionsCommand())
	return cmd
}

func newRuntimeExecCommand() *cobra.Command {
	var tool, action, resource, environment, actor, contextJSON, adapterName, outputFmt string

	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute a tool request through the governance gateway",
		Long: `Execute an agent tool request after policy evaluation.

Example:
  naeos runtime exec --tool shell --action run --resource scripts/deploy.sh \
    --environment production --actor ci-bot
  naeos runtime exec --tool file-edit --action write --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadPolicies()
			if err != nil {
				return err
			}
			cp := control.New(reg)
			sb := gateway.NewDefaultSandbox(gateway.SandboxConfig{})
			gw := gateway.New(cp, sb)

			ctx := map[string]any{}
			if contextJSON != "" {
				if err := json.Unmarshal([]byte(contextJSON), &ctx); err != nil {
					return fmt.Errorf("invalid --context JSON: %w", err)
				}
			}

			req := gateway.ToolRequest{
				Tool:        tool,
				Action:      action,
				Resource:    resource,
				Environment: environment,
				Actor:       actor,
				Context:     ctx,
			}

			var result gateway.ExecutionResult
			if adapterName != "" {
				result, err = gw.AuthorizeFromAdapter(adapterName, req)
			} else {
				result, err = gw.Authorize(req)
			}
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling result: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Tool:     %s/%s\n", result.Request.Tool, result.Request.Action)
			fmt.Fprintf(out, "Decision: %s\n", result.Decision)
			if result.PolicyID != "" {
				fmt.Fprintf(out, "Policy:   %s\n", result.PolicyID)
			}
			fmt.Fprintf(out, "Status:   %s\n", result.Status)
			if result.Hash != "" {
				fmt.Fprintf(out, "Hash:     %s\n", result.Hash)
			}
			if len(result.Reasons) > 0 {
				fmt.Fprintf(out, "Reason:   %s\n", strings.Join(result.Reasons, "; "))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tool, "tool", "", "tool name (required)")
	cmd.Flags().StringVar(&action, "action", "", "action to perform (required)")
	cmd.Flags().StringVar(&resource, "resource", "", "target resource")
	cmd.Flags().StringVar(&environment, "environment", "", "deployment environment")
	cmd.Flags().StringVar(&actor, "actor", "", "actor identity")
	cmd.Flags().StringVar(&contextJSON, "context", "", "JSON evaluation context")
	cmd.Flags().StringVar(&adapterName, "adapter", "", "agent adapter to use")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	_ = cmd.MarkFlagRequired("tool")
	_ = cmd.MarkFlagRequired("action")
	return cmd
}

func newRuntimeRestrictionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restrictions",
		Short: "Show current runtime command restrictions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Runtime restrictions are configured via governance policies.")
			fmt.Fprintln(cmd.OutOrStdout(), "Use 'naeos policy register' to define command restrictions.")
			return nil
		},
	}
	return cmd
}
