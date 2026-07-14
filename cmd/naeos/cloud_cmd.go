package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/cloud"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	cloudProvider string
	cloudRegion   string
	cloudProject  string
	cloudEnv      string
	cloudInput    string
)

type cloudSpec struct {
	Cloud struct {
		Provider    string `yaml:"provider"`
		Region      string `yaml:"region"`
		Project     string `yaml:"project"`
		Environment string `yaml:"environment"`
		Resources   []struct {
			Name string            `yaml:"name"`
			Kind string            `yaml:"kind"`
			Type string            `yaml:"type"`
			Spec map[string]string `yaml:"spec"`
		} `yaml:"resources"`
	} `yaml:"cloud"`
}

func loadCloudConfigFromSpec(path string) (*cloud.DeployConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}

	var spec cloudSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse spec file: %w", err)
	}

	if spec.Cloud.Provider == "" {
		return nil, fmt.Errorf("cloud.provider is required in spec")
	}

	config := &cloud.DeployConfig{
		Provider:    cloud.CloudProvider(spec.Cloud.Provider),
		Region:      spec.Cloud.Region,
		Project:     spec.Cloud.Project,
		Environment: spec.Cloud.Environment,
	}

	for _, r := range spec.Cloud.Resources {
		resType := r.Type
		if resType == "" {
			resType = r.Kind
		}
		specMap := make(map[string]any)
		for k, v := range r.Spec {
			specMap[k] = v
		}
		config.Resources = append(config.Resources, cloud.Resource{
			Name: r.Name,
			Type: resType,
			Spec: specMap,
		})
	}

	return config, nil
}

func resolveCloudConfig() (*cloud.DeployConfig, error) {
	if cloudInput != "" {
		return loadCloudConfigFromSpec(cloudInput)
	}

	config := &cloud.DeployConfig{
		Provider:    cloud.CloudProvider(cloudProvider),
		Region:      cloudRegion,
		Project:     cloudProject,
		Environment: cloudEnv,
	}

	return config, nil
}

func newCloudCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cloud",
		Short: "Cloud deployment commands",
		Long:  `Deploy NAEOS projects to AWS, GCP, or Azure.`,
	}

	cmd.AddCommand(newCloudDeployCommand())
	cmd.AddCommand(newCloudPlanCommand())
	cmd.AddCommand(newCloudStatusCommand())
	cmd.AddCommand(newCloudExportCommand())
	cmd.AddCommand(newCloudTypesCommand())

	cmd.PersistentFlags().StringVarP(&cloudProvider, "provider", "p", "aws", "Cloud provider (aws, gcp, azure)")
	cmd.PersistentFlags().StringVarP(&cloudRegion, "region", "r", "", "Cloud region")
	cmd.PersistentFlags().StringVarP(&cloudProject, "project", "j", "", "Cloud project name")
	cmd.PersistentFlags().StringVarP(&cloudEnv, "env", "e", "dev", "Environment (dev, staging, prod)")
	cmd.PersistentFlags().StringVarP(&cloudInput, "input-file", "i", "", "Spec file with cloud configuration (overrides flags)")

	return cmd
}

func newCloudDeployCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy to cloud provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := resolveCloudConfig()
			if err != nil {
				return err
			}

			adapter, err := cloud.GetAdapter(config.Provider)
			if err != nil {
				return err
			}

			if err := adapter.Validate(config); err != nil {
				return err
			}

			result, err := adapter.Deploy(config)
			if err != nil {
				return err
			}

			fmt.Printf("Deployed to %s: %d resources\n", result.Provider, len(result.Resources))
			for _, r := range result.Resources {
				fmt.Printf("  - %s (%s) -> %s\n", r.Name, r.Type, r.ID)
			}
			return nil
		},
	}
}

func newCloudPlanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Plan cloud deployment",
		Long:  `Generate HCL and show resources that would be created without applying.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := resolveCloudConfig()
			if err != nil {
				return err
			}

			adapter, err := cloud.GetAdapter(config.Provider)
			if err != nil {
				return err
			}

			if err := adapter.Validate(config); err != nil {
				return err
			}

			planResult, err := adapter.Plan(config)
			if err != nil {
				return err
			}

			tf, err := adapter.ExportTerraform(config)
			if err != nil {
				return fmt.Errorf("generate HCL: %w", err)
			}

			fmt.Printf("Plan: %d resources to deploy (%s/%s)\n", len(planResult.Resources), config.Provider, config.Region)
			for _, res := range planResult.Resources {
				fmt.Printf("  - %s (%s)\n", res.Name, res.Type)
			}

			fmt.Println("\n--- Cost Estimate ---")
			fmt.Print(planResult.CostEstimate.FormatCost())

			fmt.Println("\n--- Generated HCL ---")
			fmt.Println(tf)
			return nil
		},
	}
}

func newCloudStatusCommand() *cobra.Command {
	var statusProject string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show deployed resource status",
		Long:  `List deployed resources from the cloud state store (~/.naeos/cloud/).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve home dir: %w", err)
			}
			stateDir := filepath.Join(home, ".naeos", "cloud")

			entries, err := os.ReadDir(stateDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No deployments found. State store is empty.")
					return nil
				}
				return fmt.Errorf("read state store: %w", err)
			}

			filterProvider := cloud.CloudProvider(cloudProvider)
			if cloudProvider == "" {
				filterProvider = ""
			}

			found := false
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				if !strings.HasSuffix(entry.Name(), ".json") {
					continue
				}

				data, err := os.ReadFile(filepath.Join(stateDir, entry.Name()))
				if err != nil {
					continue
				}

				var state struct {
					Provider   string                   `json:"provider"`
					Project    string                   `json:"project"`
					Region     string                   `json:"region"`
					Status     string                   `json:"status"`
					Timestamp  string                   `json:"timestamp"`
					Resources  []cloud.DeployedResource  `json:"resources"`
				}
				if err := json.Unmarshal(data, &state); err != nil {
					continue
				}

				if filterProvider != "" && state.Provider != string(filterProvider) {
					continue
				}
				if statusProject != "" && state.Project != statusProject {
					continue
				}

				if !found {
					fmt.Println("Deployed resources:")
					fmt.Println("──────────────────────────────────────────────────────")
					found = true
				}

				fmt.Printf("  Provider:   %s\n", state.Provider)
				fmt.Printf("  Project:    %s\n", state.Project)
				fmt.Printf("  Region:     %s\n", state.Region)
				fmt.Printf("  Status:     %s\n", state.Status)
				fmt.Printf("  Deployed:   %s\n", state.Timestamp)
				if len(state.Resources) > 0 {
					fmt.Printf("  Resources:\n")
					for _, r := range state.Resources {
						fmt.Printf("    - %s (%s) -> %s\n", r.Name, r.Type, r.ID)
					}
				}
				fmt.Println("──────────────────────────────────────────────────────")
			}

			if !found {
				fmt.Println("No deployments found matching the given filters.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&statusProject, "project", "", "Filter by project name")
	return cmd
}

func newCloudExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export Terraform configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := resolveCloudConfig()
			if err != nil {
				return err
			}

			adapter, err := cloud.GetAdapter(config.Provider)
			if err != nil {
				return err
			}

			if err := adapter.Validate(config); err != nil {
				return err
			}

			tf, err := adapter.ExportTerraform(config)
			if err != nil {
				return err
			}

			fmt.Println(tf)
			return nil
		},
	}
}

func newCloudTypesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "List supported resource types",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch cliOutputFormat {
			case "json":
				return FormatOutput(cmd.OutOrStdout(), cloud.SupportedResourceTypes, "json")
			case "yaml":
				return FormatOutput(cmd.OutOrStdout(), cloud.SupportedResourceTypes, "yaml")
			default:
				fmt.Println("Supported resource types:")
				for _, t := range cloud.SupportedResourceTypes {
					fmt.Printf("  - %s\n", t)
				}
				return nil
			}
		},
	}
}
