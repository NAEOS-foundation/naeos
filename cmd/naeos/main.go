// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	clihelpers "github.com/NAEOS-foundation/naeos/cmd/naeos/internal/clihelpers"

	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/ai"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/api"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/artifacts"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/benchmark"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/cloud"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/control"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/core"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/cicd"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/create"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/db"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/demo"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/docs"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/dx"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/evidence"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/generation"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/history"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/importer"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/kernel"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/lsp"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/neir"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/observability"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/plugin"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/profile"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/runtime"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/security"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/serve"
	"github.com/NAEOS-foundation/naeos/cmd/naeos/commands/test"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exitErr *clihelpers.ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	root := NewRootCommand()
	root.SetArgs(args)
	return root.Execute()
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "naeos",
		Short:         "NAEOS CLI - Declarative Engineering Runtime",
		Long:          "NAEOS is a declarative engineering runtime for specification-driven project delivery.\n\nArchitecture Drives Engineering.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if clihelpers.CLIVerbose {
				fmt.Fprintln(os.Stderr, "[naeos] verbose mode enabled")
			}
			if clihelpers.CLIDryRun {
				fmt.Fprintln(os.Stderr, "[naeos] dry-run mode enabled")
			}
		},
	}

	root.PersistentFlags().BoolVar(&clihelpers.CLIVerbose, "verbose", false, "enable verbose logging")
	root.PersistentFlags().BoolVar(&clihelpers.CLIDryRun, "dry-run", false, "global dry-run mode: preview without writing to disk")
	root.PersistentFlags().StringVar(&clihelpers.CLIOutputFormat, "output-format", "table", "output format: json, yaml, table")

	root.AddCommand(core.NewInitCommand())
	root.AddCommand(core.NewRunCommand())
	root.AddCommand(core.NewValidateCommand())
	root.AddCommand(core.NewInspectCommand())
	root.AddCommand(dx.NewDoctorCommand())
	root.AddCommand(dx.NewRepairCommand())
	root.AddCommand(generation.NewScaffoldCommand())
	root.AddCommand(generation.NewExportCommand())
	root.AddCommand(serve.NewPreviewCommand())
	root.AddCommand(kernel.NewKernelCommand())
	root.AddCommand(core.NewVersionCommand())
	root.AddCommand(core.NewCompletionCommand())
	root.AddCommand(history.NewDiffCommand())
	root.AddCommand(observability.NewLintCommand())
	root.AddCommand(control.NewLockCommand())
	root.AddCommand(cicd.NewRollbackCommand())
	root.AddCommand(create.NewCreateCommand())
	root.AddCommand(plugin.NewPluginCommand())
	root.AddCommand(generation.NewTemplateCommand())
	root.AddCommand(core.NewWorkspaceCommand())
	root.AddCommand(ai.NewAICommand())
	root.AddCommand(docs.NewDocsCommand())
	root.AddCommand(evidence.NewAuditCommand())
	root.AddCommand(cicd.NewMigrateCommand())
	root.AddCommand(plugin.NewMarketplaceCommand())
	root.AddCommand(profile.NewProfileCommand())
	root.AddCommand(observability.NewWatchCommand())
	root.AddCommand(core.NewStatusCommand())
	root.AddCommand(core.NewContextCommand())
	root.AddCommand(ai.NewMCPCommand())
	root.AddCommand(test.NewTestCommand())
	root.AddCommand(docs.NewDocsGenCommand())
	root.AddCommand(api.NewAPICommand())
	root.AddCommand(observability.NewDashboardCommand())
	root.AddCommand(cloud.NewCloudCommand())
	root.AddCommand(cicd.NewCICDCommand())
	root.AddCommand(runtime.NewWebSocketCommand())
	root.AddCommand(api.NewGraphQLCommand())
	root.AddCommand(observability.NewMonitorCommand())
	root.AddCommand(security.NewAuthCommand())
	root.AddCommand(db.NewDBCommand())
	root.AddCommand(lsp.NewLSPCommand())
	root.AddCommand(runtime.NewBrokerCommand())
	root.AddCommand(ai.NewSearchCommand())
	root.AddCommand(cicd.NewWorkflowCommand())
	root.AddCommand(kernel.NewAgentCommand())
	root.AddCommand(runtime.NewGatewayCommand())
	root.AddCommand(observability.NewObservabilityCommand())
	root.AddCommand(security.NewSecurityCommand())
	root.AddCommand(observability.NewPerfCommand())
	root.AddCommand(dx.NewDXCommand())
	root.AddCommand(runtime.NewDistributedCommand())
	root.AddCommand(runtime.NewEventsCommand())
	root.AddCommand(importer.NewImportCommand())
	root.AddCommand(core.NewHealthCommand())
	root.AddCommand(benchmark.NewBenchmarkCommand())
	root.AddCommand(core.NewConfigCommand())
	root.AddCommand(history.NewHistoryCommand())
	root.AddCommand(cicd.NewMigrationCmd())
	root.AddCommand(cicd.NewDeployCommand())
	root.AddCommand(dx.NewTUICommand())
	root.AddCommand(artifacts.NewArtifactsCommand())
	root.AddCommand(cicd.NewBuildCommand())
	root.AddCommand(control.NewComplianceCommand())
	root.AddCommand(neir.NewSchemaCommand())
	root.AddCommand(cloud.NewSupabaseCommand())
	root.AddCommand(docs.NewCLIDocsGenCommand())
	root.AddCommand(control.NewPolicyCommand())
	root.AddCommand(control.NewControlCommand())
	root.AddCommand(kernel.NewRuntimeCommand())
	root.AddCommand(evidence.NewEvidenceCommand())
	root.AddCommand(evidence.NewVerifyCommand())
	root.AddCommand(security.NewSBomCommand())
	root.AddCommand(security.NewSignCommand())
	root.AddCommand(cloud.NewHelmCommand())
	root.AddCommand(security.NewAirgapCommand())
	root.AddCommand(serve.NewServeCommand())
	root.AddCommand(demo.NewDemoCommand())
	return root
}
