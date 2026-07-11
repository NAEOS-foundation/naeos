package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/lock"
)

func newLockCommand() *cobra.Command {
	var outputFile, verifyFile string

	lockGen := &cobra.Command{
		Use:   "generate",
		Short: "Generate a lock file from current artifacts",
		Long: `Generate a SHA-256 based lock file for reproducible builds.

Example:
  naeos lock generate file1.go file2.go
  naeos lock generate -o naeos.lock *.go`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var artifactInfos []lock.ArtifactInfo
			for _, arg := range args {
				data, err := os.ReadFile(arg)
				if err != nil {
					return fmt.Errorf("read %s: %w", arg, err)
				}
				artifactInfos = append(artifactInfos, lock.ArtifactInfo{
					Path:    arg,
					Content: data,
				})
			}

			lockFile, err := lock.Generate(artifactInfos)
			if err != nil {
				return err
			}

			if outputFile == "" {
				outputFile = "naeos.lock"
			}

			if err := lock.WriteToFile(lockFile, outputFile); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Generated %s with %d artifacts\n", outputFile, len(lockFile.Artifacts))
			return nil
		},
	}

	lockVerify := &cobra.Command{
		Use:   "verify",
		Short: "Verify current artifacts against lock file",
		Long: `Verify that current files match the lock file checksums.

Example:
  naeos lock verify file1.go file2.go
  naeos lock verify --lock-file naeos.lock *.go`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if verifyFile == "" {
				verifyFile = "naeos.lock"
			}

			lockFile, err := lock.ReadFromFile(verifyFile)
			if err != nil {
				return err
			}

			var current []lock.ArtifactInfo
			for _, arg := range args {
				data, err := os.ReadFile(arg)
				if err != nil {
					continue
				}
				current = append(current, lock.ArtifactInfo{
					Path:    arg,
					Content: data,
				})
			}

			changes, err := lock.Verify(lockFile, current)
			if err != nil {
				return err
			}

			if len(changes) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Lock file verified: no changes detected")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Changes detected:\n")
				for _, c := range changes {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", c)
				}
			}
			return nil
		},
	}

	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Manage lock files for reproducible builds",
		Long: `Generate and verify lock files for reproducible builds using SHA-256 checksums.

Example:
  naeos lock generate file1.go file2.go
  naeos lock verify file1.go file2.go`,
	}

	cmd.AddCommand(lockGen)
	cmd.AddCommand(lockVerify)

	cmd.Flags().StringVarP(&outputFile, "output", "o", "naeos.lock", "path for the lock file")
	lockVerify.Flags().StringVarP(&verifyFile, "lock-file", "l", "naeos.lock", "path to lock file to verify against")
	return cmd
}
