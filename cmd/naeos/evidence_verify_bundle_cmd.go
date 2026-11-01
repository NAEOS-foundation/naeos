// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
	"github.com/NAEOS-foundation/naeos/internal/evidence"
)

func newEvidenceVerifyBundleCommand() *cobra.Command {
	var inputFile, outputFmt, trustedPublicKeyFile string

	cmd := &cobra.Command{
		Use:   "verify-bundle",
		Short: "Verify an exported control-plane evidence bundle",
		Long: "Verify a canonical EvidenceBundle without contacting the control plane.\n\n" +
			"The verifier reads a serialized EvidenceBundle, validates its decision and\n" +
			"execution bindings, recomputes the evidence digest, and requires an\n" +
			"operator-supplied trusted Ed25519 public key. It is read-only and does not\n" +
			"evaluate policy, execute actions, or require a live control-plane service.\n\n" +
			"Examples:\n" +
			"  naeos evidence verify-bundle --input-file evidence.json --trusted-public-key trusted.pub\n" +
			"  naeos evidence verify-bundle --input-file evidence.json --trusted-public-key trusted.pub --output json",
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

			var result controlplane.EvidenceVerification
			if trustedPublicKeyFile == "" {
				result = controlplane.VerifyEvidenceWithTrustedKey(bundle, nil)
			} else {
				keyData, err := os.ReadFile(trustedPublicKeyFile)
				if err != nil {
					return fmt.Errorf("read trusted evidence public key: %w", err)
				}
				publicKey, err := evidence.ParseEd25519PublicKey(keyData)
				if err != nil {
					return fmt.Errorf("parse trusted evidence public key: %w", err)
				}
				result = controlplane.VerifyEvidenceWithTrustedKey(bundle, publicKey)
			}

			digestIntegrity := !hasVerificationIssue(result.Issues, "evidence digest mismatch")
			signatureValid := !hasVerificationIssue(result.Issues, "evidence signature invalid or missing")
			signerTrusted := !hasVerificationIssue(result.Issues, "trusted evidence signing key is missing or invalid") &&
				!hasVerificationIssue(result.Issues, "evidence signer is not the configured trust anchor")

			if outputFmt == "json" {
				payload := map[string]interface{}{
					"verification":     result,
					"decision_id":      bundle.DecisionID,
					"digest_integrity": digestIntegrity,
					"signature_valid":  signatureValid,
					"signer_trusted":   signerTrusted,
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
				fmt.Fprintf(cmd.OutOrStdout(), "Digest integrity:      %v\n", digestIntegrity)
				fmt.Fprintf(cmd.OutOrStdout(), "Signature valid:       %v\n", signatureValid)
				fmt.Fprintf(cmd.OutOrStdout(), "Signer trusted:        %v\n", signerTrusted)
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
	cmd.Flags().StringVar(&trustedPublicKeyFile, "trusted-public-key", "", "operator-trusted Ed25519 public key (raw, base64, or hex)")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}

func hasVerificationIssue(issues []string, fragment string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, fragment) {
			return true
		}
	}
	return false
}
