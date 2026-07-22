package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/configschema"
	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
	}

	cmd.AddCommand(newConfigValidateCommand())
	cmd.AddCommand(newConfigShowCommand())
	cmd.AddCommand(newConfigEncryptCommand())
	cmd.AddCommand(newConfigDecryptCommand())

	return cmd
}

func newConfigEncryptCommand() *cobra.Command {
	var inputPath, outputPath, passphrase string

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt a config file with AES-256-GCM",
		Long: `Encrypt a configuration file at rest using AES-256-GCM with a passphrase.
Output is written as base64-encoded ciphertext.

Example:
  naeos config encrypt --input config.yaml --output config.enc
  naeos config encrypt --input config.yaml --passphrase "my-secret" --output config.enc`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputPath == "" {
				return fmt.Errorf("--input is required")
			}
			if passphrase == "" {
				return fmt.Errorf("--passphrase is required")
			}

			data, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("read input: %w", err)
			}

			encrypted, err := securityext.EncryptConfig(data, passphrase)
			if err != nil {
				return fmt.Errorf("encrypt: %w", err)
			}

			if outputPath != "" {
				outputPath = filepath.Clean(outputPath)
				if err := os.WriteFile(outputPath, []byte(encrypted), 0o600); err != nil { //nolint:gosec // G703: path is cleaned via filepath.Clean
					return fmt.Errorf("write output: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Encrypted config written to %s\n", outputPath)
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte(encrypted + "\n"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "path to config file (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "path to write encrypted output")
	cmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "encryption passphrase (required)")
	return cmd
}

func newConfigDecryptCommand() *cobra.Command {
	var inputPath, outputPath, passphrase string

	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt an encrypted config file",
		Long: `Decrypt a base64-encoded encrypted config file back to plaintext.

Example:
  naeos config decrypt --input config.enc --output config.yaml
  naeos config decrypt --input config.enc --passphrase "my-secret" --output config.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputPath == "" {
				return fmt.Errorf("--input is required")
			}
			if passphrase == "" {
				return fmt.Errorf("--passphrase is required")
			}

			data, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("read input: %w", err)
			}

			decrypted, err := securityext.DecryptConfig(strings.TrimSpace(string(data)), passphrase)
			if err != nil {
				return fmt.Errorf("decrypt: %w", err)
			}

			if outputPath != "" {
				outputPath = filepath.Clean(outputPath)
				if err := os.WriteFile(outputPath, decrypted, 0o600); err != nil { //nolint:gosec // G703: path is cleaned via filepath.Clean
					return fmt.Errorf("write output: %w", err)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Decrypted config written to %s\n", outputPath)
			} else {
				_, _ = cmd.OutOrStdout().Write(decrypted)
				_, _ = cmd.OutOrStdout().Write([]byte("\n"))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "path to encrypted config file (required)")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "path to write decrypted output")
	cmd.Flags().StringVarP(&passphrase, "passphrase", "p", "", "decryption passphrase (required)")
	return cmd
}

func newConfigValidateCommand() *cobra.Command {
	var inputPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a NAEOS config file against the schema",
		Long: `Validate a configuration file (YAML or JSON) against the NAEOS config schema.
Reports missing required fields and type mismatches.

Example:
  naeos config validate --input naeos.yaml
  naeos config validate --input config.json --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputPath == "" {
				return fmt.Errorf("--input is required")
			}

			data, err := os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			ext := ".yaml"
			if strings.HasSuffix(inputPath, ".json") {
				ext = ".json"
			}

			format := "yaml"
			if ext == ".json" {
				format = "json"
			}

			errs, _ := configschema.ValidateData(data, format)
			if len(errs) == 0 {
				_, _ = cmd.OutOrStdout().Write([]byte("✓ Config is valid\n"))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✗ Found %d validation error(s):\n", len(errs))
				for _, e := range errs {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", e.Field, e.Message)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "path to config file (required)")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the default config schema",
		Long:  `Display the default NAEOS configuration schema with field types and required fields.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			schema := configschema.DefaultSchema()
			_, _ = cmd.OutOrStdout().Write([]byte("NAEOS Configuration Schema\n"))
			fmt.Fprintf(cmd.OutOrStdout(), "Type: %s\n", schema.Type)
			fmt.Fprintf(cmd.OutOrStdout(), "Required: %s\n\n", strings.Join(schema.Required, ", "))
			_, _ = cmd.OutOrStdout().Write([]byte("Properties:\n"))
			for name, prop := range schema.Properties {
				req := ""
				for _, r := range schema.Required {
					if r == name {
						req = " [REQUIRED]"
						break
					}
				}
				def := ""
				if prop.Default != nil {
					def = fmt.Sprintf(" (default: %v)", prop.Default)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %-15s %-10s %s%s%s\n", name, prop.Type, prop.Description, def, req)
			}
			return nil
		},
	}
}
