package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/sbom"
)

func newSBomCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sbom",
		Short: "Software Bill of Materials generation and verification",
		Long:  `Generate CycloneDX SBOM documents from project artifacts and verify their integrity.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newSBomGenerateCommand())
	cmd.AddCommand(newSBomVerifyCommand())
	cmd.AddCommand(newSBomInspectCommand())
	return cmd
}

func newSBomGenerateCommand() *cobra.Command {
	var (
		project   string
		version   string
		output    string
		dir       string
		outputFmt string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a CycloneDX SBOM from a project directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			gen := sbom.NewGenerator(sbom.GeneratorConfig{
				Project:     project,
				Version:     version,
				ToolVersion: version,
			})

			var bom *sbom.BOM
			var err error
			if dir != "" {
				bom, err = gen.FromDir(dir)
			} else {
				bom, err = gen.Generate(nil)
			}
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(bom, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			if output != "" {
				if err := sbom.Write(bom, output); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "SBOM written to %s (%d components)\n", output, bom.ComponentCount())
				return nil
			}

			data, _ := json.MarshalIndent(bom, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "project name (used as BOM root component)")
	cmd.Flags().StringVar(&version, "version", "", "project version")
	cmd.Flags().StringVar(&output, "output", "", "write SBOM to file (instead of stdout)")
	cmd.Flags().StringVar(&dir, "dir", "", "scan directory for file-level components")
	cmd.Flags().StringVar(&outputFmt, "output-format", "json", "output format: json or table")
	return cmd
}

func newSBomVerifyCommand() *cobra.Command {
	var outputFmt string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify the integrity of a CycloneDX SBOM document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			bom, err := sbom.Load(path)
			if err != nil {
				return err
			}

			checks := verifyBOM(bom)

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(checks, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "BOM: %s\n", path)
			fmt.Fprintf(out, "Format: %s %s\n", bom.BOMFormat, bom.SpecVersion)
			fmt.Fprintf(out, "Components: %d\n\n", bom.ComponentCount())

			passed := 0
			failed := 0
			for _, c := range checks {
				mark := "PASS"
				if !c.Passed {
					mark = "FAIL"
					failed++
				} else {
					passed++
				}
				fmt.Fprintf(out, "  [%s] %s — %s\n", mark, c.Name, c.Detail)
			}
			fmt.Fprintf(out, "\nPassed: %d  Failed: %d\n", passed, failed)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}

func newSBomInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Display a summary of a CycloneDX SBOM document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			bom, err := sbom.Load(path)
			if err != nil {
				return err
			}

			data, _ := json.MarshalIndent(bom, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}

	return cmd
}

type sbomCheck struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

func verifyBOM(bom *sbom.BOM) []sbomCheck {
	var checks []sbomCheck

	checks = append(checks, sbomCheck{
		Name:   "format",
		Passed: bom.BOMFormat == "CycloneDX",
		Detail: fmt.Sprintf("bomFormat=%s", bom.BOMFormat),
	})

	checks = append(checks, sbomCheck{
		Name:   "spec-version",
		Passed: bom.SpecVersion == sbom.SpecVersion,
		Detail: fmt.Sprintf("specVersion=%s", bom.SpecVersion),
	})

	checks = append(checks, sbomCheck{
		Name:   "serial-number",
		Passed: bom.SerialNumber != "",
		Detail: "serialNumber present",
	})

	componentsOk := true
	for i, c := range bom.Components {
		if c.Name == "" {
			componentsOk = false
			checks = append(checks, sbomCheck{
				Name:   fmt.Sprintf("component-%d-name", i),
				Passed: false,
				Detail: "component name is empty",
			})
		}
	}
	checks = append(checks, sbomCheck{
		Name:   "components",
		Passed: componentsOk,
		Detail: fmt.Sprintf("%d components, all named", bom.ComponentCount()),
	})

	hashesOk := true
	missingHash := 0
	for _, c := range bom.Components {
		if len(c.Hashes) == 0 {
			hashesOk = false
			missingHash++
		}
	}
	checks = append(checks, sbomCheck{
		Name:   "hashes",
		Passed: hashesOk,
		Detail: fmt.Sprintf("%d/%d components with hashes", len(bom.Components)-missingHash, len(bom.Components)),
	})

	return checks
}

func loadSBomStorePath() string {
	path, err := os.UserHomeDir()
	if err != nil {
		return ".sbom.json"
	}
	return path + "/.config/naeos/sbom.json"
}
