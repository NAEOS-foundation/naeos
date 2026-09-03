package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

func newEvidenceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evidence",
		Short: "Governance evidence store — immutable audit trail",
		Long:  `Query, verify, and manage the tamper-evident governance evidence store.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newEvidenceLogCommand())
	cmd.AddCommand(newEvidenceQueryCommand())
	cmd.AddCommand(newEvidenceVerifyCommand())
	cmd.AddCommand(newEvidenceSummaryCommand())
	cmd.AddCommand(newEvidenceHashCommand())
	return cmd
}

func evidenceStorePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", naeoserr.Wrapf(err, naeoserr.ErrInternal, "cannot determine home directory")
	}
	return filepath.Join(home, ".config", "naeos", "evidence.json"), nil
}

func loadEvidenceStore() (*evidence.EvidenceStore, error) {
	store := evidence.NewStore()
	path, err := evidenceStorePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "reading evidence store")
	}
	var records []evidence.EvidenceRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrValidation, "parsing evidence store")
	}
	for _, r := range records {
		if _, err := store.Append(r); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func saveEvidenceStore(store *evidence.EvidenceStore) error {
	path, err := evidenceStorePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "creating evidence directory")
	}
	data, err := json.MarshalIndent(store.Records(), "", "  ")
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshaling evidence store")
	}
	return os.WriteFile(path, data, 0o600)
}

func newEvidenceLogCommand() *cobra.Command {
	var tool, action, resource, environment, actor, policyFile, outputFmt string

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Log an evidence record from a gateway execution",
		Long: `Execute a tool request through the governance gateway and record
the evidence of the decision and execution outcome.

Example:
  naeos evidence log --tool shell --action run --actor ci-bot \
    --environment production --policy-file policy.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load or create a policy registry for the control plane.
			var reg *policy.Registry
			if policyFile != "" {
				data, err := os.ReadFile(policyFile)
				if err != nil {
					return naeoserr.Wrapf(err, naeoserr.ErrInternal, "reading policy file")
				}
				reg = policy.NewRegistry()
				var p policy.Policy
				if err := json.Unmarshal(data, &p); err != nil {
					return naeoserr.Wrapf(err, naeoserr.ErrValidation, "parsing policy")
				}
				if err := reg.Register(&p); err != nil {
					return err
				}
			} else {
				reg = policy.NewRegistry()
			}

			cp := control.New(reg)
			sb := gateway.NewDefaultSandbox(gateway.SandboxConfig{})
			gw := gateway.New(cp, sb)

			result, err := gw.Authorize(gateway.ToolRequest{
				Tool:        tool,
				Action:      action,
				Resource:    resource,
				Environment: environment,
				Actor:       actor,
			})
			if err != nil {
				return err
			}

			// Build the evidence record.
			store, err := loadEvidenceStore()
			if err != nil {
				return err
			}

			rec := evidence.EvidenceRecord{
				Actor:             actor,
				Resource:          result.Request.Resource,
				Action:            result.Request.Action,
				Environment:       environment,
				PolicyID:          result.PolicyID,
				RuleID:            result.RuleID,
				Decision:          result.Decision,
				DecisionReasons:   result.Reasons,
				ExecutionStatus:   result.Status,
				ExecutionOutput:   result.Output,
				ExecutionDurationMs: result.Duration.Milliseconds(),
				ArtifactHash:      result.Hash,
			}

			saved, err := store.Append(rec)
			if err != nil {
				return err
			}
			if err := saveEvidenceStore(store); err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(saved, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Evidence: %s\n", saved.ID)
			fmt.Fprintf(out, "Decision: %s\n", saved.Decision)
			fmt.Fprintf(out, "Status:   %s\n", saved.ExecutionStatus)
			fmt.Fprintf(out, "Policy:   %s\n", saved.PolicyID)
			fmt.Fprintf(out, "Hash:     %s\n", saved.Hash[:16]+"...")
			return nil
		},
	}

	cmd.Flags().StringVar(&tool, "tool", "", "tool name (required)")
	cmd.Flags().StringVar(&action, "action", "", "action to perform (required)")
	cmd.Flags().StringVar(&resource, "resource", "", "target resource")
	cmd.Flags().StringVar(&environment, "environment", "", "deployment environment")
	cmd.Flags().StringVar(&actor, "actor", "", "actor identity")
	cmd.Flags().StringVar(&policyFile, "policy-file", "", "path to policy JSON file")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	_ = cmd.MarkFlagRequired("tool")
	_ = cmd.MarkFlagRequired("action")
	return cmd
}

func newEvidenceQueryCommand() *cobra.Command {
	var actor, resource, policyID, decision, outputFmt string
	var limit int

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query evidence records by criteria",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadEvidenceStore()
			if err != nil {
				return err
			}

			var dec control.Decision
			if decision != "" {
				dec = control.Decision(decision)
			}

			results := store.Query(evidence.EvidenceQuery{
				Actor:    actor,
				Resource: resource,
				PolicyID: policyID,
				Decision: dec,
				Limit:    limit,
			})

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-12s %-8s %-10s %-20s %s\n", "ID", "DECISION", "STATUS", "ACTOR", "RESOURCE")
			fmt.Fprintf(out, "%-12s %-8s %-10s %-20s %s\n",
				strings.Repeat("-", 12), strings.Repeat("-", 8), strings.Repeat("-", 10),
				strings.Repeat("-", 20), strings.Repeat("-", 20))
			for _, r := range results {
				id := r.ID
				if len(id) > 11 {
					id = id[:11] + "…"
				}
				fmt.Fprintf(out, "%-12s %-8s %-10s %-20s %s\n",
					id, r.Decision, r.ExecutionStatus, r.Actor, r.Resource)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&actor, "actor", "", "filter by actor")
	cmd.Flags().StringVar(&resource, "resource", "", "filter by resource")
	cmd.Flags().StringVar(&policyID, "policy", "", "filter by policy ID")
	cmd.Flags().StringVar(&decision, "decision", "", "filter by decision (ALLOW/DENY/REQUIRE_APPROVAL)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max records to return")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}

func newEvidenceVerifyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify the integrity of the evidence chain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadEvidenceStore()
			if err != nil {
				return err
			}

			idx, err := store.Verify()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "CHAIN BROKEN at index %d: %v\n", idx, err)
				return err
			}

			summary := store.Summary()
			fmt.Fprintf(cmd.OutOrStdout(), "Chain integrity: OK\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Total records:   %d\n", summary.TotalRecords)
			fmt.Fprintf(cmd.OutOrStdout(), "Chain intact:    %v\n", summary.ChainIntact)
			return nil
		},
	}
}

func newEvidenceSummaryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "summary",
		Short: "Show aggregate evidence statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadEvidenceStore()
			if err != nil {
				return err
			}

			summary := store.Summary()
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Evidence Summary\n")
			fmt.Fprintf(out, "───────────────\n")
			fmt.Fprintf(out, "Total records:        %d\n", summary.TotalRecords)
			fmt.Fprintf(out, "Allowed:              %d\n", summary.ByDecision[control.DecisionAllow])
			fmt.Fprintf(out, "Denied:               %d\n", summary.DeniedCount)
			fmt.Fprintf(out, "Approval required:    %d\n", summary.ApprovalRequiredCount)
			fmt.Fprintf(out, "Approved (explicit):  %d\n", summary.ApprovedCount)
			fmt.Fprintf(out, "With artifacts:       %d\n", summary.WithArtifacts)
			fmt.Fprintf(out, "Chain intact:         %v\n", summary.ChainIntact)
			fmt.Fprintf(out, "\nBy Policy:\n")
			for pol, count := range summary.ByPolicy {
				fmt.Fprintf(out, "  %-20s %d\n", pol, count)
			}
			fmt.Fprintf(out, "\nBy Actor:\n")
			for actor, count := range summary.ByActor {
				fmt.Fprintf(out, "  %-20s %d\n", actor, count)
			}
			return nil
		},
	}
}

func newEvidenceHashCommand() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "hash",
		Short: "Compute the SHA-256 hash of a file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			data, err := os.ReadFile(file)
			if err != nil {
				return naeoserr.Wrapf(err, naeoserr.ErrInternal, "reading file")
			}
			hash := evidence.ComputeArtifactHash(data)
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n", hash, file)
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "file to hash (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
