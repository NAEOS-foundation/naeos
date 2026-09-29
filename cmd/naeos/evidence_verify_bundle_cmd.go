// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

func newEvidenceVerifyBundleCommand() *cobra.Command {
	var inputFile, outputFmt string

	cmd := &cobra.Command{
		Use:   "verify-bundle",
		Short: "Verify an exported control-plane evidence bundle",
		Long: "Verify a canonical EvidenceBundle without contacting the control plane.\n\n" +
			"The verifier reads a serialized EvidenceBundle, validates its decision and\n" +
			"execution bindings, and recomputes the evidence digest. It is read-only and\n" +
			"does not evaluate policy, execute actions, or require a live control-plane\n" +
			"service.\n\n" +
			"Examples:\n" +
			"  naeos evidence verify-bundle --input-file evidence.json\n" +
			"  naeos evidence verify-bundle --input-file evidence.json --output json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" {
				return fmt.Errorf("--input-file is required")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("read evidence bundle: %w", err)
			}

			var bundle controlplane.EvidenceBundle
			if err := json.Unmarshal(data, &bundle); err != nil {
				return fmt.Errorf("parse evidence bundle: %w", err)
			}

			result := controlplane.VerifyEvidence(bundle)
			if outputFmt == "json" {
				payload := map[string]interface{}{
					"verification": result,
					"decision_id": bundle.DecisionID,
				}
				if bundle.ExecutionID != "" {
					payload["execution_id"] = bundle.ExecutionID
				}
				encoded, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("encode verification result: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Evidence verification: %s\n", result.Result)
				fmt.Fprintf(cmd.OutOrStdout(), "Decision ID:           %s\n", bundle.DecisionID)
				if bundle.ExecutionID != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Execution ID:          %s\n", bundle.ExecutionID)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Decision consistent:   %v\n", result.DecisionConsistent)
				fmt.Fprintf(cmd.OutOrStdout(), "Execution consistent:  %v\n", result.ExecutionConsistent)
				fmt.Fprintf(cmd.OutOrStdout(), "Digest consistent:     %v\n", result.LedgerIntegrity)
				for _, issue := range result.Issues {
					fmt.Fprintf(cmd.OutOrStdout(), "Issue:                 %s\n", issue)
				}
			}

			if result.Result != "PASS" {
				return fmt.Errorf("evidence verification failed")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to exported EvidenceBundle JSON")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}
