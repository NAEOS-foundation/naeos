package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cliVerbose      bool
	cliDryRun       bool
	cliOutputFormat string
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
		Long:          "NAEOS is a declarative engineering runtime for specification-driven project delivery.\n\nSpecify Once. Build Anywhere.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cliVerbose {
				fmt.Fprintln(os.Stderr, "[naeos] verbose mode enabled")
			}
			if cliDryRun {
				fmt.Fprintln(os.Stderr, "[naeos] dry-run mode enabled")
			}
		},
	}

	root.PersistentFlags().BoolVar(&cliVerbose, "verbose", false, "enable verbose logging")
	root.PersistentFlags().BoolVar(&cliDryRun, "dry-run", false, "global dry-run mode: preview without writing to disk")
	root.PersistentFlags().StringVar(&cliOutputFormat, "output-format", "table", "output format: json, yaml, table")

	root.AddCommand(newInitCommand())
	root.AddCommand(newRunCommand())
	root.AddCommand(newValidateCommand())
	root.AddCommand(newInspectCommand())
	root.AddCommand(newDoctorCommand())
	root.AddCommand(newRepairCommand())
	root.AddCommand(newScaffoldCommand())
	root.AddCommand(newExportCommand())
	root.AddCommand(newPreviewCommand())
	root.AddCommand(newKernelCommand())
	root.AddCommand(newVersionCommand())
	root.AddCommand(newCompletionCommand())
	root.AddCommand(newDiffCommand())
	root.AddCommand(newLintCommand())
	root.AddCommand(newLockCommand())
	root.AddCommand(newRollbackCommand())
	root.AddCommand(newCreateCommand())
	root.AddCommand(newPluginCommand())
	root.AddCommand(newTemplateCommand())
	root.AddCommand(newWorkspaceCommand())
	root.AddCommand(newAICommand())
	root.AddCommand(newDocsCommand())
	root.AddCommand(newAuditCommand())
	root.AddCommand(newMigrateCommand())
	root.AddCommand(newMarketplaceCommand())
	root.AddCommand(newProfileCommand())
	root.AddCommand(newWatchCommand())
	root.AddCommand(newStatusCommand())
	root.AddCommand(newContextCommand())
	root.AddCommand(newMCPCommand())
	root.AddCommand(newTestCommand())
	root.AddCommand(newDocsGenCommand())
	root.AddCommand(newAPICommand())
	root.AddCommand(newDashboardCommand())
	root.AddCommand(newCloudCommand())
	root.AddCommand(newCICDCommand())
	root.AddCommand(newWebSocketCommand())
	root.AddCommand(newGraphQLCommand())
	root.AddCommand(newMonitorCommand())
	root.AddCommand(newAuthCommand())
	root.AddCommand(newDBCommand())
	root.AddCommand(newBrokerCommand())
	root.AddCommand(newSearchCommand())
	root.AddCommand(newWorkflowCommand())
	root.AddCommand(newGatewayCommand())
	root.AddCommand(newObservabilityCommand())
	root.AddCommand(newSecurityCommand())
	root.AddCommand(newPerfCommand())
	root.AddCommand(newDXCommand())
	root.AddCommand(newDistributedCommand())
	root.AddCommand(newEventsCommand())
	root.AddCommand(newImportCommand())
	root.AddCommand(newHealthCommand())
	root.AddCommand(newBenchmarkCommand())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newHistoryCommand())
	root.AddCommand(newMigrationCmd())
	root.AddCommand(newDeployCommand())
	root.AddCommand(newTUICommand())
	root.AddCommand(newArtifactsCommand())
	root.AddCommand(newBuildCommand())
	root.AddCommand(newComplianceCommand())
	root.AddCommand(newSchemaCommand())
	root.AddCommand(newCLIDocsGenCommand())
	return root
}
