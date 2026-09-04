package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/airgap"
)

func newAirgapCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airgap",
		Short: "Air-gapped distribution bundles",
		Long:  `Bundle charts, images, SBOMs, and signatures for offline deployment.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAirgapBundleCommand())
	cmd.AddCommand(newAirgapImportCommand())
	cmd.AddCommand(newAirgapInspectCommand())
	return cmd
}

func newAirgapBundleCommand() *cobra.Command {
	var (
		name     string
		version  string
		charts   string
		images   string
		sbom     string
		sigs     string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "bundle",
		Short: "Build an air-gapped distribution bundle",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			outputPath := output
			if outputPath == "" {
				outputPath = name + "-" + version + ".airgap.tar.gz"
			}

			b := airgap.NewBuilder(airgap.BundleConfig{
				Name:          name,
				Version:       version,
				ChartsDir:     charts,
				ImagesFile:    images,
				SBOMManifest:  sbom,
				SignaturesDir: sigs,
			})

			bundle, err := b.Build(outputPath)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, bundle.Summary())
			fmt.Fprintf(out, "  Output:     %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "bundle name")
	cmd.Flags().StringVar(&version, "version", "", "bundle version")
	cmd.Flags().StringVar(&charts, "charts-dir", "", "directory containing *.tgz chart archives")
	cmd.Flags().StringVar(&images, "images-file", "", "file listing image name:tag (one per line)")
	cmd.Flags().StringVar(&sbom, "sbom", "", "CycloneDX SBOM JSON file to include")
	cmd.Flags().StringVar(&sigs, "signatures-dir", "", "directory containing *.sig.json signature bundles")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output bundle path")
	return cmd
}

func newAirgapImportCommand() *cobra.Command {
	var (
		dest      string
		verify    bool
		outputFmt string
	)

	cmd := &cobra.Command{
		Use:   "import <bundle.tar.gz>",
		Short: "Extract an air-gapped bundle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := airgap.Extract(args[0], airgap.ExtractOptions{
				Destination:  dest,
				VerifyHashes: verify,
			})
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(bundle, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, bundle.Summary())
			fmt.Fprintf(out, "  Extracted to: %s\n", dest)
			return nil
		},
	}

	cmd.Flags().StringVarP(&dest, "dest", "d", ".", "destination directory")
	cmd.Flags().BoolVar(&verify, "verify-hashes", true, "verify file hashes against manifest")
	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}

func newAirgapInspectCommand() *cobra.Command {
	var outputFmt string

	cmd := &cobra.Command{
		Use:   "inspect <bundle.tar.gz>",
		Short: "Inspect an air-gapped bundle manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundle, err := airgap.ReadBundle(args[0])
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := bundle.Manifest()
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, bundle.Summary())
			valid := bundle.VerifyChecksum()
			mark := "OK"
			if !valid {
				mark = "CORRUPTED"
			}
			fmt.Fprintf(out, "  Manifest hash: %s [%s]\n", bundle.ManifestHash, mark)
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}