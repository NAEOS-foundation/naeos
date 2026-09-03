package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
)

func newControlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "control",
		Short: "Governance control plane evaluation",
		Long:  `Evaluate authorization requests against registered policies to produce deterministic decisions (ALLOW, DENY, REQUIRE_APPROVAL).`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newControlEvaluateCommand())
	return cmd
}

func newControlEvaluateCommand() *cobra.Command {
	var resource, action, environment, actor, contextJSON string
	var failOpen bool
	var output string

	cmd := &cobra.Command{
		Use:   "evaluate",
		Short: "Evaluate an authorization request and issue a decision",
		Long: `Evaluate an authorization request against registered governance policies.

Example:
  naeos control evaluate --resource deploy --action run --environment production \\
    --actor ci-bot --context '{"tls_version":"1.2"}'
  naeos control evaluate --resource deploy --action run --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadPolicies()
			if err != nil {
				return err
			}
			plane := control.New(reg, control.FailClosed(!failOpen))

			ctx := map[string]any{}
			if contextJSON != "" {
				if err := json.Unmarshal([]byte(contextJSON), &ctx); err != nil {
					return fmt.Errorf("invalid --context JSON: %w", err)
				}
			}

			rec, err := plane.Evaluate(control.Request{
				Resource:    resource,
				Action:      action,
				Environment: environment,
				Actor:       actor,
				Context:     ctx,
			})
			if err != nil {
				return err
			}

			if output == "json" {
				data, err := json.MarshalIndent(rec, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling decision: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Request:  %s/%s environment=%s actor=%s\n",
				orWildcard(rec.Request.Resource), orWildcard(rec.Request.Action),
				orWildcard(rec.Request.Environment), orWildcard(rec.Request.Actor))
			fmt.Fprintf(out, "Decision: %s\n", rec.Decision)
			if rec.PolicyID != "" {
				fmt.Fprintf(out, "Policy:   %s v%s\n", rec.PolicyID, rec.PolicyVersion)
			}
			if rec.RuleID != "" {
				fmt.Fprintf(out, "Rule:     %s\n", rec.RuleID)
			}
			if len(rec.Reasons) > 0 {
				fmt.Fprintf(out, "Reason:   %s\n", strings.Join(rec.Reasons, "; "))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&resource, "resource", "", "resource being acted on (required)")
	cmd.Flags().StringVar(&action, "action", "", "action being requested (required)")
	cmd.Flags().StringVar(&environment, "environment", "", "deployment/environment context")
	cmd.Flags().StringVar(&actor, "actor", "", "actor issuing the request")
	cmd.Flags().StringVar(&contextJSON, "context", "", "JSON object of evaluation context")
	cmd.Flags().StringVar(&output, "output", "table", "output format: table or json")
	cmd.Flags().BoolVar(&failOpen, "fail-open", false, "allow requests with no matching policy (default: deny)")
	_ = cmd.MarkFlagRequired("resource")
	_ = cmd.MarkFlagRequired("action")
	return cmd
}
