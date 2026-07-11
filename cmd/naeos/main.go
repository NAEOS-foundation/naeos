package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	cliVerbose bool
	cliDryRun  bool
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	root := newRootCommand()
	root.SetArgs(args)
	return root.Execute()
}

func newRootCommand() *cobra.Command {
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
	return root
}
