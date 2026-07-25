package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/audit"
	"github.com/NAEOS-foundation/naeos/internal/compliance"
)

func newComplianceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compliance",
		Short: "Compliance reporting and audit log export",
	}

	cmd.AddCommand(newComplianceExportCommand())
	cmd.AddCommand(newComplianceCloudExportCommand())
	cmd.AddCommand(newComplianceReportCommand())
	cmd.AddCommand(newComplianceListFrameworksCommand())
	cmd.AddCommand(newComplianceVerifyCommand())
	return cmd
}

func newComplianceExportCommand() *cobra.Command {
	var format, output, auditFile string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export audit log for compliance reporting",
		Long: `Export the audit trail in JSON or CSV format for compliance purposes.

Example:
  naeos compliance export --format json --output audit-export.json
  naeos compliance export --format csv --output audit-export.csv`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if format != "json" && format != "csv" {
				return fmt.Errorf("unsupported format %q, use json or csv", format)
			}
			if output == "" {
				return fmt.Errorf("--output is required")
			}

			path := auditFile
			if path == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("cannot determine home directory: %w", err)
				}
				path = filepath.Join(homeDir, ".naeos", "audit.log")
			}

			var events []audit.AuditEvent
			data, err := os.ReadFile(path)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("reading audit file: %w", err)
				}
			} else {
				lines := strings.Split(strings.TrimSpace(string(data)), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					var evt audit.AuditEvent
					if err := json.Unmarshal([]byte(line), &evt); err != nil {
						return fmt.Errorf("parsing audit line: %w", err)
					}
					events = append(events, evt)
				}
			}

			auditor := audit.NewMemoryAuditor()
			for _, evt := range events {
				_ = auditor.Log(evt)
			}

			switch format {
			case "json":
				if err := auditor.ExportJSON(output); err != nil {
					return fmt.Errorf("export json: %w", err)
				}
			case "csv":
				if err := auditor.ExportCSV(output); err != nil {
					return fmt.Errorf("export csv: %w", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Compliance report exported to %s (format: %s)\n", output, format)
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "json", "export format: json or csv")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path (required)")
	cmd.Flags().StringVarP(&auditFile, "audit-file", "a", "", "path to audit log file (default: ~/.naeos/audit.log)")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

func newComplianceReportCommand() *cobra.Command {
	var framework, output, auditFile string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a compliance report for a framework (soc2, hipaa, gdpr)",
		Long: `Generate a compliance report by evaluating audit events against a compliance framework.

Example:
  naeos compliance report --framework soc2 --output soc2-report.json
  naeos compliance report --framework hipaa --output hipaa-report.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fw := compliance.Framework(framework)
			if _, ok := compliance.Frameworks[fw]; !ok {
				return fmt.Errorf("unsupported framework %q — supported: soc2, hipaa, gdpr", framework)
			}

			path := auditFile
			if path == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("cannot determine home directory: %w", err)
				}
				path = filepath.Join(homeDir, ".naeos", "audit.log")
			}

			var events []audit.AuditEvent
			data, err := os.ReadFile(path)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("reading audit file: %w", err)
				}
			} else {
				lines := strings.Split(strings.TrimSpace(string(data)), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					var evt audit.AuditEvent
					if err := json.Unmarshal([]byte(line), &evt); err != nil {
						return fmt.Errorf("parsing audit line: %w", err)
					}
					events = append(events, evt)
				}
			}

			report := compliance.GenerateReport(fw, events)

			if output != "" {
				data, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling report: %w", err)
				}
				if err := os.WriteFile(output, data, 0o600); err != nil {
					return fmt.Errorf("writing report: %w", err)
				}
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Compliance Report: %s\n", compliance.Frameworks[fw].Name)
			fmt.Fprintf(out, "Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
			fmt.Fprintf(out, "Controls: %d total, %d passed, %d failed\n",
				report.TotalControls, report.PassedControls, report.FailedControls)
			for _, cs := range report.ControlStatuses {
				status := "PASS"
				if !cs.Passed {
					status = "FAIL"
				}
				fmt.Fprintf(out, "  [%s] %s: %s\n", status, cs.ControlID, cs.Finding)
			}
			if output != "" {
				fmt.Fprintf(out, "\nReport saved to: %s\n", output)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&framework, "framework", "soc2", "compliance framework: soc2, hipaa, gdpr")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path")
	cmd.Flags().StringVarP(&auditFile, "audit-file", "a", "", "path to audit log file (default: ~/.naeos/audit.log)")
	return cmd
}

func newComplianceListFrameworksCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-frameworks",
		Short: "List supported compliance frameworks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-15s %-20s %s\n", "FRAMEWORK", "NAME", "DESCRIPTION")
			fmt.Fprintf(out, "%-15s %-20s %s\n", "----------", "----", "-----------")
			for name, def := range compliance.Frameworks {
				fmt.Fprintf(out, "%-15s %-20s %s\n", name, def.Name, def.Description)
			}
			return nil
		},
	}
}

func newComplianceVerifyCommand() *cobra.Command {
	var auditFile string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify audit log hash chain integrity",
		Long: `Verify the hash chain integrity of the audit log to detect tampering.

Example:
  naeos compliance verify
  naeos compliance verify --audit-file /custom/path/audit.log`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := auditFile
			if path == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("cannot determine home directory: %w", err)
				}
				path = filepath.Join(homeDir, ".naeos", "audit.log")
			}

			violations, err := audit.VerifyChainFile(path)
			if err != nil {
				return fmt.Errorf("verify chain: %w", err)
			}

			out := cmd.OutOrStdout()
			if len(violations) == 0 {
				fmt.Fprintf(out, "Audit log hash chain integrity verified: no violations found\n")
				return nil
			}

			fmt.Fprintf(out, "Found %d hash chain violation(s):\n", len(violations))
			for _, v := range violations {
				fmt.Fprintf(out, "  - %s\n", v)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&auditFile, "audit-file", "a", "", "path to audit log file (default: ~/.naeos/audit.log)")
	return cmd
}

func newComplianceCloudExportCommand() *cobra.Command {
	var provider, bucket, prefix, region, endpoint, accessKey, secretKey string
	var accountName, accountKey string
	var auditFile string

	cmd := &cobra.Command{
		Use:   "cloud-export",
		Short: "Export audit log to cloud storage (S3, GCS, Azure Blob)",
		Long: `Upload the audit log directly to cloud storage.

Examples:
  naeos compliance cloud-export --provider s3 --bucket my-bucket --access-key XXX --secret-key YYY --region us-east-1
  naeos compliance cloud-export --provider gcs --bucket my-bucket --access-key GOOG1XXX --secret-key YYY
  naeos compliance cloud-export --provider azure --account-name myaccount --account-key YYY --container my-container`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			events, err := readAuditLogEvents(auditFile)
			if err != nil {
				return err
			}

			cfg := audit.CloudConfig{
				Bucket:      bucket,
				Prefix:      prefix,
				Region:      region,
				Endpoint:    endpoint,
				AccessKey:   accessKey,
				SecretKey:   secretKey,
				AccountName: accountName,
				AccountKey:  accountKey,
			}

			switch provider {
			case "s3":
				cfg.Provider = audit.CloudS3
				if cfg.Region == "" {
					cfg.Region = "us-east-1"
				}
			case "gcs":
				cfg.Provider = audit.CloudGCS
			case "azure":
				cfg.Provider = audit.CloudAzure
				if cfg.AccountName == "" {
					return fmt.Errorf("--account-name is required for Azure")
				}
				if cfg.AccountKey == "" {
					return fmt.Errorf("--account-key is required for Azure")
				}
				if cfg.Bucket == "" {
					return fmt.Errorf("--bucket (container name) is required for Azure")
				}
			default:
				return fmt.Errorf("unsupported provider %q — supported: s3, gcs, azure", provider)
			}

			path, err := audit.ExportToCloud(cfg, events)
			if err != nil {
				return fmt.Errorf("cloud export: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Audit log exported to cloud: %s/%s\n", bucket, path)
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "s3", "cloud provider: s3, gcs, azure")
	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket or container name (required)")
	cmd.Flags().StringVar(&prefix, "prefix", "audit/", "key prefix/path")
	cmd.Flags().StringVar(&region, "region", "", "S3 region (default: us-east-1)")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "S3 custom endpoint (for MinIO)")
	cmd.Flags().StringVar(&accessKey, "access-key", "", "access key ID (S3, GCS)")
	cmd.Flags().StringVar(&secretKey, "secret-key", "", "secret access key (S3, GCS)")
	cmd.Flags().StringVar(&accountName, "account-name", "", "Azure storage account name")
	cmd.Flags().StringVar(&accountKey, "account-key", "", "Azure storage account key")
	cmd.Flags().StringVarP(&auditFile, "audit-file", "a", "", "path to audit log file (default: ~/.naeos/audit.log)")

	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("bucket")
	return cmd
}

func readAuditLogEvents(auditFile string) ([]audit.AuditEvent, error) {
	path := auditFile
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(homeDir, ".naeos", "audit.log")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading audit file: %w", err)
	}

	var events []audit.AuditEvent
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var evt audit.AuditEvent
		if err := json.Unmarshal([]byte(line), &evt); err != nil {
			return nil, fmt.Errorf("parsing audit line: %w", err)
		}
		events = append(events, evt)
	}

	return events, nil
}
