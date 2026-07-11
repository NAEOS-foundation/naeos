package main

import (
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/artifacts"
	"github.com/spf13/cobra"
)

func newArtifactsCommand() *cobra.Command {
	var storeDir string

	cmd := &cobra.Command{
		Use:   "artifacts",
		Short: "Manage generated project artifacts",
		Long:  `Track, deduplicate, and manage generated artifacts with metadata and checksums.`,
	}

	cmd.PersistentFlags().StringVar(&storeDir, "dir", ".naeos/artifacts", "artifact store directory")

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all tracked artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := artifacts.NewStore(storeDir)
			if err := store.LoadFromDisk(); err != nil {
				return fmt.Errorf("load store: %w", err)
			}

			list := store.List()
			if len(list) == 0 {
				fmt.Println("No artifacts tracked.")
				return nil
			}

			fmt.Printf("%-40s %-10s %-8s %s\n", "PATH", "KIND", "SIZE", "HASH")
			fmt.Println(strings.Repeat("-", 100))
			for _, a := range list {
				hash := a.ContentHash
				if len(hash) > 8 {
					hash = hash[:8]
				}
				fmt.Printf("%-40s %-10s %-8d %s\n", a.Path, a.Kind, a.Size, hash)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "info [path]",
		Short: "Show artifact details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := artifacts.NewStore(storeDir)
			if err := store.LoadFromDisk(); err != nil {
				return fmt.Errorf("load store: %w", err)
			}

			a, ok := store.Get(args[0])
			if !ok {
				return fmt.Errorf("artifact %q not found", args[0])
			}

			fmt.Printf("Path: %s\n", a.Path)
			fmt.Printf("Kind: %s\n", a.Kind)
			fmt.Printf("Language: %s\n", a.Language)
			fmt.Printf("Size: %d bytes\n", a.Size)
			fmt.Printf("Hash: %s\n", a.ContentHash)
			fmt.Printf("Created: %s\n", a.CreatedAt.Format("2006-01-02 15:04:05"))
			if len(a.Metadata) > 0 {
				fmt.Println("Metadata:")
				for k, v := range a.Metadata {
					fmt.Printf("  %s: %s\n", k, v)
				}
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "dedup",
		Short: "Remove duplicate artifacts by content hash",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := artifacts.NewStore(storeDir)
			if err := store.LoadFromDisk(); err != nil {
				return fmt.Errorf("load store: %w", err)
			}

			removed := store.Deduplicate()
			if err := store.WriteToDisk(); err != nil {
				return fmt.Errorf("save store: %w", err)
			}

			fmt.Printf("Removed %d duplicate artifacts.\n", removed)
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "summary",
		Short: "Show artifact count by kind",
		RunE: func(cmd *cobra.Command, args []string) error {
			store := artifacts.NewStore(storeDir)
			if err := store.LoadFromDisk(); err != nil {
				return fmt.Errorf("load store: %w", err)
			}

			summary := store.Summary()
			fmt.Println("Artifact Summary:")
			fmt.Println(strings.Repeat("-", 30))
			for kind, count := range summary {
				fmt.Printf("  %-15s %d\n", kind, count)
			}
			return nil
		},
	})

	return cmd
}
