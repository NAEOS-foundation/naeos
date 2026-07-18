package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/security"
	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

func newSecurityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security",
		Short: "Security and secrets management",
		Long: `Manage encrypted secrets, sanitize input, and validate data.

Example:
  naeos security set-secret --name db-pass --value secret123
  naeos security get-secret --name db-pass
  naeos security list-secrets
  naeos security sanitize --input '<script>alert("xss")</script>'
  naeos security hash-password --password mypass
  naeos security validate --name email --value test@example.com`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSecuritySetSecretCommand())
	cmd.AddCommand(newSecurityGetSecretCommand())
	cmd.AddCommand(newSecurityListSecretsCommand())
	cmd.AddCommand(newSecuritySanitizeCommand())
	cmd.AddCommand(newSecurityHashPasswordCommand())
	cmd.AddCommand(newSecurityValidateCommand())
	cmd.AddCommand(newSecurityAuditCommand())

	return cmd
}

func newSecuritySetSecretCommand() *cobra.Command {
	var name, value, key string

	cmd := &cobra.Command{
		Use:   "set-secret",
		Short: "Store an encrypted secret",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" {
				key = "naeos-default-security-key"
			}
			sm := securityext.NewSecretManager(key)

			if err := sm.Set(name, value); err != nil {
				return fmt.Errorf("failed to store secret: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Secret '%s' stored successfully.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "secret name (required)")
	cmd.Flags().StringVar(&value, "value", "", "secret value (required)")
	cmd.Flags().StringVar(&key, "key", "", "encryption key (optional; defaults to built-in key)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("value")
	return cmd
}

func newSecurityGetSecretCommand() *cobra.Command {
	var name, key string

	cmd := &cobra.Command{
		Use:   "get-secret",
		Short: "Retrieve a secret value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" {
				key = "naeos-default-security-key"
			}
			sm := securityext.NewSecretManager(key)

			val, ok := sm.Get(name)
			if !ok {
				return fmt.Errorf("secret '%s' not found", name)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", val)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "secret name (required)")
	cmd.Flags().StringVar(&key, "key", "", "encryption key (optional; defaults to built-in key)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newSecurityListSecretsCommand() *cobra.Command {
	var key string

	cmd := &cobra.Command{
		Use:   "list-secrets",
		Short: "List all stored secrets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == "" {
				key = "naeos-default-security-key"
			}
			sm := securityext.NewSecretManager(key)

			names := sm.List()
			if len(names) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No secrets stored.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-20s\n", "SECRET NAME")
			fmt.Fprintf(out, "%-20s\n", "-----------")
			for _, name := range names {
				fmt.Fprintf(out, "%-20s\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "encryption key (optional; defaults to built-in key)")
	return cmd
}

func newSecuritySanitizeCommand() *cobra.Command {
	var input, mode string

	cmd := &cobra.Command{
		Use:   "sanitize",
		Short: "Sanitize input string",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := securityext.NewSanitizer()

			var result string
			switch mode {
			case "html":
				result = s.SanitizeHTML(input)
			case "sql":
				result = s.SanitizeSQL(input)
			case "xss":
				result = s.SanitizeXSS(input)
			case "path":
				result = s.SanitizePath(input)
			default:
				result = s.SanitizeAll(input)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Sanitized: %s\n", result)
			return nil
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "input to sanitize (required)")
	cmd.Flags().StringVar(&mode, "mode", "all", "sanitize mode (html, sql, xss, path, all)")
	cmd.MarkFlagRequired("input")
	return cmd
}

func newSecurityHashPasswordCommand() *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "hash-password",
		Short: "Hash a password with bcrypt",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			hash, err := securityext.HashPassword(password)
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", hash)
			return nil
		},
	}

	cmd.Flags().StringVar(&password, "password", "", "password to hash (required)")
	cmd.MarkFlagRequired("password")
	return cmd
}

func newSecurityValidateCommand() *cobra.Command {
	var name, value string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a value against rules",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := securityext.NewValidator()

			v.AddRule("email", securityext.RequiredRule)
			v.AddRule("name", securityext.MinLengthRule(3))

			err := v.Validate(name, value)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Validation failed: %s\n", err)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Validation passed.\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "rule name (email, name)")
	cmd.Flags().StringVar(&value, "value", "", "value to validate (required)")
	cmd.MarkFlagRequired("value")
	return cmd
}

func joinSecStrings(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	return strings.Join(ss, ", ")
}

func newSecurityAuditCommand() *cobra.Command {
	var inputFile, outputFormat string

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Run security audit on project files",
		Long: `Scan project files for common security issues:
  - Hardcoded secrets and API keys
  - SQL injection patterns
  - XSS vulnerabilities
  - Unsafe eval/deserialization
  - Debug mode enabled
  - Missing health check endpoints

Example:
  naeos security audit
  naeos security audit --input ./src
  naeos security audit --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if inputFile != "" {
				dir = inputFile
			}

			files, err := security.ScanDir(dir)
			if err != nil {
				return fmt.Errorf("scan directory: %w", err)
			}

			auditor := security.NewAuditor()
			result := auditor.AuditFiles(files)

			type auditResult struct {
				Directory string                `json:"directory" yaml:"directory"`
				Files     int                   `json:"files" yaml:"files"`
				Findings  []security.Finding    `json:"findings" yaml:"findings"`
				Summary   security.AuditSummary `json:"summary" yaml:"summary"`
			}

			data := auditResult{
				Directory: dir,
				Files:     len(files),
				Findings:  result.Finding,
				Summary:   result.Summary,
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Security Audit: %s\n", dir)
			fmt.Fprintf(out, "═══════════════════════════════════════\n")
			fmt.Fprintf(out, "Scanned %d files\n\n", len(files))

			if len(result.Finding) == 0 {
				fmt.Fprintln(out, "  ✓ No security issues found.")
			} else {
				for _, f := range result.Finding {
					icon := "✓"
					switch f.Severity {
					case security.SeverityCritical:
						icon = "✗"
					case security.SeverityHigh:
						icon = "✗"
					case security.SeverityMedium:
						icon = "⚠"
					case security.SeverityLow:
						icon = "⚠"
					}
					loc := f.File
					if f.Line > 0 {
						loc = fmt.Sprintf("%s:%d", f.File, f.Line)
					}
					fmt.Fprintf(out, "  %s [%s] %s (%s)\n", icon, strings.ToUpper(string(f.Severity)), f.Title, loc)
				}
			}

			fmt.Fprintf(out, "\nSummary: %d critical, %d high, %d medium, %d low, %d info\n",
				result.Summary.Critical, result.Summary.High, result.Summary.Medium,
				result.Summary.Low, result.Summary.Info)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "directory or file to audit (default: current directory)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}
