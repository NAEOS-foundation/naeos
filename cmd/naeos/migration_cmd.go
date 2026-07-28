package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/database"
)

func newMigrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migration",
		Short: "Database migration management",
	}

	cmd.AddCommand(newMigrationStatusCommand())

	return cmd
}

type migrationStatusRow struct {
	Name      string `json:"name" yaml:"name"`
	Driver    string `json:"driver" yaml:"driver"`
	Database  string `json:"database" yaml:"database"`
	Version   int    `json:"version" yaml:"version"`
	AppliedAt string `json:"applied_at" yaml:"applied_at"`
	Status    string `json:"status" yaml:"status"`
}

func newMigrationStatusCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status for all configured databases",
		Long: `Display the current migration status of all configured databases.

Reads saved connections and queries each database's _migrations table.

Example:
  naeos migration status
  naeos migration status --output json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := database.NewConnectionStore()
			conns, err := store.List()
			if err != nil {
				return fmt.Errorf("list connections: %w", err)
			}

			if len(conns) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No database connections configured.")
				fmt.Fprintln(cmd.OutOrStdout(), "Use 'naeos db connect' to add one.")
				return nil
			}

			var rows []migrationStatusRow
			for _, c := range conns {
				row := migrationStatusRow{
					Name:     c.Name,
					Driver:   c.Driver,
					Database: c.Config.Database,
				}

				db := database.New(c.Driver)
				if db == nil {
					row.Status = "unsupported driver"
					rows = append(rows, row)
					continue
				}

				if err := db.Connect(c.Config); err != nil {
					row.Status = fmt.Sprintf("disconnected (%v)", err)
					rows = append(rows, row)
					continue
				}

				q := "SELECT version, COALESCE(name,''), COALESCE(applied_at::text,'') FROM _migrations ORDER BY version DESC LIMIT 1"
				result, err := db.Query(q)
				if err != nil {
					row.Version = 0
					row.Status = "no migrations table"
					rows = append(rows, row)
					db.Close()
					continue
				}

				if len(result) == 0 {
					row.Version = 0
					row.Status = "no migrations applied"
				} else {
					if v, ok := result[0]["version"]; ok {
						switch vv := v.(type) {
						case int64:
							row.Version = int(vv)
						case float64:
							row.Version = int(vv)
						}
					}
					if a, ok := result[0]["applied_at"]; ok {
						row.AppliedAt = fmt.Sprintf("%v", a)
					}
					row.Status = "ok"
				}
				db.Close()
				rows = append(rows, row)
			}

			sort.Slice(rows, func(i, j int) bool {
				return rows[i].Name < rows[j].Name
			})

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), rows, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-20s %-12s %-8s %-20s %s\n", "CONNECTION", "DRIVER", "VERSION", "APPLIED_AT", "STATUS")
			fmt.Fprintf(out, "%-20s %-12s %-8s %-20s %s\n", strings.Repeat("-", 20), strings.Repeat("-", 12), strings.Repeat("-", 8), strings.Repeat("-", 20), strings.Repeat("-", 20))
			for _, r := range rows {
				fmt.Fprintf(out, "%-20s %-12s %-8d %-20s %s\n", r.Name, r.Driver, r.Version, r.AppliedAt, r.Status)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}
