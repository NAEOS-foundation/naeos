package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/marketplace"
)

func newMarketplaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "marketplace",
		Short: "Browse and install specification templates from the marketplace",
		Long: `Browse and install specification templates from the NAEOS marketplace.

Example:
  naeos marketplace search "web api"
  naeos marketplace install web-api-template`,
	}

	var cacheDir string

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for specification templates",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := marketplace.NewClient(cacheDir)
			query := strings.Join(args, " ")
			results, err := client.Search(marketplace.SearchFilter{Query: query, Limit: 10})
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No results found")
				return nil
			}
			for _, r := range results {
				fmt.Fprintf(cmd.OutOrStdout(), "%-25s %-10s %s\n", r.Name, r.Version, r.Description)
			}
			return nil
		},
	}

	installCmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Install a template from the marketplace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := marketplace.NewClient(cacheDir)
			if err := client.Install(args[0], "."); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed template %s\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(searchCmd)
	cmd.AddCommand(installCmd)
	cmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", filepath.Join(".", ".naeos", "cache"), "cache directory")
	return cmd
}
