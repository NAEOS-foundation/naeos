package main

import (
	"fmt"
	"os"

	"github.com/NAEOS-foundation/naeos/internal/migration"
	"github.com/spf13/cobra"
)

func newMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Manage spec schema migrations",
		Long:  `Migrate NAEOS specification files to newer schema versions.`,
	}

	cmd.AddCommand(newMigrateRunCommand())
	cmd.AddCommand(newMigratePlanCommand())
	cmd.AddCommand(newMigrateVersionsCommand())
	return cmd
}

func newMigrateRunCommand() *cobra.Command {
	var dryRun bool
	var outputPath string

	cmd := &cobra.Command{
		Use:   "run [spec-file]",
		Short: "Migrate spec to the latest schema version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specFile := args[0]
			content, err := os.ReadFile(specFile)
			if err != nil {
				return fmt.Errorf("read spec: %w", err)
			}

			planner := migration.NewPlanner()
			plan, err := planner.Plan(migration.CurrentVersion, migration.TargetVersion)
			if err != nil {
				return fmt.Errorf("plan migration: %w", err)
			}

			if len(plan) == 0 {
				fmt.Println("Spec is already at the latest version.")
				return nil
			}

			fmt.Printf("Migration plan: %s -> %s\n", migration.CurrentVersion, migration.TargetVersion)
			for _, step := range plan {
				fmt.Printf("  [%s -> %s] %s\n", step.FromVersion, step.ToVersion, step.Description)
			}

			result, err := planner.Migrate(content, migration.CurrentVersion, migration.TargetVersion)
			if err != nil {
				return fmt.Errorf("migrate: %w", err)
			}

			if dryRun {
				fmt.Println("\n--- DRY RUN OUTPUT ---")
				fmt.Println(string(result))
				return nil
			}

			outPath := outputPath
			if outPath == "" {
				outPath = specFile
			}

			if err := os.WriteFile(outPath, result, 0o644); err != nil {
				return fmt.Errorf("write migrated spec: %w", err)
			}

			fmt.Printf("\nMigrated %s -> %s\n", specFile, outPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing")
	cmd.Flags().StringVar(&outputPath, "output", "", "output file (empty = overwrite in place)")
	return cmd
}

func newMigratePlanCommand() *cobra.Command {
	var fromVer string
	var toVer string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show migration plan without applying",
		RunE: func(cmd *cobra.Command, args []string) error {
			planner := migration.NewPlanner()
			plan, err := planner.Plan(fromVer, toVer)
			if err != nil {
				return fmt.Errorf("plan migration: %w", err)
			}

			if len(plan) == 0 {
				fmt.Println("No migrations needed. Spec is at the latest version.")
				return nil
			}

			fmt.Printf("Migration plan: %s -> %s\n\n", fromVer, toVer)
			for i, step := range plan {
				fmt.Printf("%d. [%s -> %s] %s\n", i+1, step.FromVersion, step.ToVersion, step.Description)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromVer, "from", migration.CurrentVersion, "source version")
	cmd.Flags().StringVar(&toVer, "to", migration.TargetVersion, "target version")
	return cmd
}

func newMigrateVersionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "versions",
		Short: "List supported migration versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Supported versions:")
			fmt.Printf("  Current: %s\n", migration.CurrentVersion)
			fmt.Printf("  Target:  %s\n", migration.TargetVersion)
			fmt.Println("\nMigration steps:")
			fmt.Println("  0.1.0 -> 0.2.0: Add generation section")
			fmt.Println("  0.2.0 -> 0.3.0: Add testing section")
			return nil
		},
	}
}
