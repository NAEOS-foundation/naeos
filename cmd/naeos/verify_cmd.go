package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/verification"
)

func newVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Independently verify governance evidence",
		Long:  `Run the verification chain (evidence integrity, artifact hash, approval binding) over the evidence store.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newVerifyEvidenceCommand())
	cmd.AddCommand(newVerifyReportCommand())
	return cmd
}

func newVerifyEvidenceCommand() *cobra.Command {
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

func newVerifyReportCommand() *cobra.Command {
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
