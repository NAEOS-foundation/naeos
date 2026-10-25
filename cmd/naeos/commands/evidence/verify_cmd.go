// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	clihelpers "github.com/NAEOS-foundation/naeos/cmd/naeos/internal/clihelpers"

	"github.com/NAEOS-foundation/naeos/internal/verification"
	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func NewVerifyCommand() *cobra.Command {
	var configPath, input, inputFile, outputFile string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a specification using the existing NAEOS pipeline",
		Long: `Verify a specification by running the existing NAEOS pipeline: parse, normalize, resolve, build NEIR, and validate.

This is the first vertical slice of the autonomous control-plane verification foundation.

Examples:
  naeos verify --input-file spec.yaml
  naeos verify --input-file spec.yaml --output-format json
  naeos verify --config config.yaml --input-file spec.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" && inputFile == "" {
				return cmd.Help()
			}

			inputValue, err := clihelpers.LoadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, err := clihelpers.LoadPipelineConfig(configPath, clihelpers.CLIVerbose, nil, clihelpers.CLIDryRun, "")
			if err != nil {
				return err
			}

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			result, err := p.Validate(inputValue)
			if err != nil {
				vr := clihelpers.ValidationResult{
					Valid:  false,
					Status: "invalid",
					Errors: []clihelpers.ValidationError{{
						Code:    "PIPELINE_FAILED",
						Message: err.Error(),
					}},
					Summary: "verification failed",
				}
				return clihelpers.RenderValidation(cmd, vr, clihelpers.CLIOutputFormat, outputFile)
			}

			projectName := ""
			if result.NEIR != nil && result.NEIR.Project != nil {
				projectName = result.NEIR.Project.Name
			}
			vr := clihelpers.ValidationResult{
				Valid:    true,
				Status:   "valid",
				Project:  projectName,
				Modules:  len(result.NEIR.Modules),
				Services: len(result.NEIR.Services),
				Summary:  fmt.Sprintf("valid — project: %s, modules: %d, services: %d", projectName, len(result.NEIR.Modules), len(result.NEIR.Services)),
			}
				return clihelpers.RenderValidation(cmd, vr, clihelpers.CLIOutputFormat, outputFile)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write the verification output")

	cmd.AddCommand(NewVerifyEvidenceCommand())
	cmd.AddCommand(NewVerifyReportCommand())
	return cmd
}

func NewVerifyEvidenceCommand() *cobra.Command {
	var outputFmt string

	cmd := &cobra.Command{
		Use:   "evidence",
		Short: "Verify the entire evidence store",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadEvidenceStore()
			if err != nil {
				return err
			}

			// Build the verification chain.
			chain := verification.NewChain(
				verification.Contract{
					Name:        "naeos-verification",
					Description: "Standard NAEOS evidence verification contract",
					Version:     "1.0.0",
					Requirements: []string{
						"evidence chain integrity",
						"record hash authenticity",
					},
				},
				verification.NewEvidenceChainVerifier(store),
				verification.NewApprovalBindingVerifier(),
			)

			results, err := chain.VerifyEvidence(store)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			verified := 0
			failed := 0
			for _, r := range results {
				fmt.Fprintf(out, "%-26s %-10s %s\n", r.Target, r.Status, r.Contract)
				switch r.Status {
				case verification.StatusVerified:
					verified++
				case verification.StatusFailed:
					failed++
				}
				for _, c := range r.Checks {
					mark := "ok"
					if !c.Passed {
						mark = "!!"
					}
					fmt.Fprintf(out, "    %-4s %s\n", mark, c.Name)
				}
			}
			fmt.Fprintf(out, "\nVerified: %d  Failed: %d\n", verified, failed)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}

func NewVerifyReportCommand() *cobra.Command {
	var outputFmt string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Run independent verification and produce a report",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadEvidenceStore()
			if err != nil {
				return err
			}

			chain := verification.NewChain(
				verification.Contract{
					Name:        "naeos-independent-verification",
					Description: "Standard NAEOS independent verification contract",
					Version:     "1.0.0",
					Requirements: []string{
						"evidence chain integrity",
					},
				},
				verification.NewEvidenceChainVerifier(store),
			)

			results, err := chain.VerifyEvidence(store)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Independent Verification Report")
			fmt.Fprintln(out, "────────────────────────────────")
			fmt.Fprintf(out, "Contract:   %s\n", "naeos-independent-verification")
			fmt.Fprintf(out, "Records:    %d\n", len(results))
			fmt.Fprintln(out)

			verified := 0
			for _, r := range results {
				fmt.Fprintf(out, "  %-26s %s\n", r.Target, r.Status)
				if r.Status == verification.StatusVerified {
					verified++
				}
			}
			fmt.Fprintf(out, "\nVerified:   %d\n", verified)

			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}
