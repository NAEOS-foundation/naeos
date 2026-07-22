package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/schemaregistry"
)

func newSchemaCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "NEIR schema registry operations",
		Long: `Manage and validate against the NEIR JSON Schema registry.

The schema registry hosts versioned JSON Schema definitions for the
NEIR specification format. Use this command to validate specs against
the canonical schema, or query schema version information.

Examples:
  naeos schema validate spec.yaml
  naeos schema validate spec.yaml --registry https://naeos.dev/schemaregistry/latest.json
  naeos schema validate spec.json --output json
  naeos schema info`,
	}

	cmd.AddCommand(newSchemaValidateCommand())
	cmd.AddCommand(newSchemaInfoCommand())

	return cmd
}

func newSchemaValidateCommand() *cobra.Command {
	var registryURL string
	var outputFormat string

	valCmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a NEIR spec against the schema registry",
		Long: `Validate a specification file against the latest NEIR JSON Schema
from the schema registry. Supports YAML and JSON spec files.

The command fetches the canonical schema from the registry and checks
that the spec conforms to it, including required fields and enum values.

Examples:
  naeos schema validate spec.yaml
  naeos schema validate spec.json --output json
  naeos schema validate spec.naeos.yaml --registry https://naeos.dev/schemaregistry/v1/neir.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath := args[0]

			client := schemaregistry.NewNEIRClient(registryURL)
			schema, err := client.FetchSchema()
			if err != nil {
				return fmt.Errorf("fetch schema: %w", err)
			}

			result, err := schemaregistry.ValidateNEIRSpec(specPath, schema)
			if err != nil {
				return err
			}

			switch outputFormat {
			case "json":
				data, _ := json.MarshalIndent(result, "", "  ")
				_, _ = cmd.OutOrStdout().Write(data)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			case "yaml":
				fmt.Fprintf(cmd.OutOrStdout(), "valid: %t\n", result.Valid)
				fmt.Fprintf(cmd.OutOrStdout(), "version: %s\n", result.Version)
				if !result.Valid {
					fmt.Fprintln(cmd.OutOrStdout(), "errors:")
					for _, e := range result.Errors {
						fmt.Fprintf(cmd.OutOrStdout(), "  - field: %s\n    message: %s\n", e.Field, e.Message)
					}
				}
			default:
				if result.Valid {
					fmt.Fprintf(cmd.OutOrStdout(), "✓ Valid — conforms to schema %s\n", result.Version)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "✗ Invalid — does not conform to schema %s\n", result.Version)
					for _, e := range result.Errors {
						if e.Field != "" {
							fmt.Fprintf(cmd.OutOrStdout(), "  • %s: %s\n", e.Field, e.Message)
						} else {
							fmt.Fprintf(cmd.OutOrStdout(), "  • %s\n", e.Message)
						}
					}
				}
			}

			if !result.Valid {
				return fmt.Errorf("validation failed")
			}
			return nil
		},
	}

	valCmd.Flags().StringVar(&registryURL, "registry", schemaregistry.DefaultNEIRSchemaURL, "schema registry URL")
	valCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json, yaml")

	return valCmd
}

func newSchemaInfoCommand() *cobra.Command {
	var registryURL string

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show schema registry information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := schemaregistry.NewNEIRClient(registryURL)
			schema, err := client.FetchSchema()
			if err != nil {
				return fmt.Errorf("fetch schema: %w", err)
			}

			title, _ := schema["title"].(string)
			desc, _ := schema["description"].(string)
			schemaID, _ := schema["$id"].(string)

			props, _ := schema["properties"].(map[string]any)
			defs, _ := schema["definitions"].(map[string]any)

			required, _ := schema["required"].([]any)
			reqFields := make([]string, 0, len(required))
			for _, r := range required {
				if s, ok := r.(string); ok {
					reqFields = append(reqFields, s)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Schema Registry\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Title:       %s\n", title)
			fmt.Fprintf(cmd.OutOrStdout(), "  Description: %s\n", desc)
			fmt.Fprintf(cmd.OutOrStdout(), "  $id:         %s\n", schemaID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Source:      %s\n", registryURL)
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Top-level sections: %d\n", len(props))
			fmt.Fprintf(cmd.OutOrStdout(), "Type definitions:   %d\n", len(defs))
			fmt.Fprintf(cmd.OutOrStdout(), "Required fields:    %s\n", strings.Join(reqFields, ", "))
			fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintln(cmd.OutOrStdout(), "Sections:")
			for name := range props {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", name)
			}

			return nil
		},
	}

	infoCmd.Flags().StringVar(&registryURL, "registry", schemaregistry.DefaultNEIRSchemaURL, "schema registry URL")

	return infoCmd
}
