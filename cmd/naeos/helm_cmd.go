package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/helm"
)

func newHelmCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm",
		Short: "Helm chart scaffolding and validation",
		Long:  `Create, render, and validate Kubernetes Helm charts.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newHelmChartCommand())
	cmd.AddCommand(newHelmInitCommand())
	cmd.AddCommand(newHelmValidateCommand())
	return cmd
}

func newHelmChartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chart",
		Short: "Create a new Helm chart",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newHelmChartInitCommand())
	cmd.AddCommand(newHelmChartRenderCommand())
	return cmd
}

func newHelmInitCommand() *cobra.Command {
	var (
		version     string
		appVersion  string
		description string
		output      string
	)

	cmd := &cobra.Command{
		Use:   "init <name>",
		Short: "Initialize a new Helm chart",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chart, err := helm.NewChart(helm.ChartConfig{
				Name:        args[0],
				Version:     version,
				Description: description,
				AppVersion:  appVersion,
			})
			if err != nil {
				return err
			}

			dir := output
			if dir == "" {
				dir = args[0]
			}
			if err := chart.WriteToDisk(dir); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Created Helm chart %s v%s at %s\n", chart.Name, chart.Version, dir)
			fmt.Fprintf(out, "  %d values\n  %d templates\n", len(chart.Values), len(chart.Templates))
			fmt.Fprintln(out, "\nTemplates:")
			for _, name := range chart.SortedTemplateNames() {
				fmt.Fprintf(out, "  templates/%s\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "chart version (default 0.1.0)")
	cmd.Flags().StringVar(&appVersion, "app-version", "", "app version (default 1.0.0)")
	cmd.Flags().StringVar(&description, "description", "", "chart description")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output directory (default: <name>)")
	return cmd
}

func newHelmChartInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init <name>",
		Short: "Initialize a new Helm chart (alias of helm init)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chart, err := helm.NewChart(helm.ChartConfig{Name: args[0]})
			if err != nil {
				return err
			}
			if err := chart.WriteToDisk(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created Helm chart %s at %s\n", chart.Name, args[0])
			return nil
		},
	}
}

func newHelmChartRenderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render <dir>",
		Short: "Render chart files from a directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chart, err := helm.LoadFromDisk(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Chart: %s v%s\n", chart.Name, chart.Version)
			fmt.Fprintln(out, "----- Chart.yaml -----")
			fmt.Fprint(out, chart.RenderChartMetadata())
			fmt.Fprintln(out, "\n----- values.yaml -----")
			fmt.Fprint(out, chart.RenderValues())
			fmt.Fprintln(out, "\n----- templates -----")
			for _, t := range chart.Templates {
				fmt.Fprintf(out, "===== templates/%s =====\n", t.Name)
				fmt.Fprintln(out, strings.TrimSuffix(t.Content, "\n"))
				fmt.Fprintln(out)
			}
			return nil
		},
	}
	return cmd
}

func newHelmValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <dir>",
		Short: "Validate a Helm chart directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chart, err := helm.LoadFromDisk(args[0])
			if err != nil {
				return err
			}
			errs := chart.Validate()
			out := cmd.OutOrStdout()
			if len(errs) == 0 {
				fmt.Fprintf(out, "Chart %s v%s is valid\n", chart.Name, chart.Version)
				return nil
			}
			for _, e := range errs {
				fmt.Fprintf(out, "  [FAIL] %s\n", e.Error())
			}
			return fmt.Errorf("chart validation failed with %d error(s)", len(errs))
		},
	}
	return cmd
}