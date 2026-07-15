package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/database"
)

func newDBCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database connection and migration management",
		Long: `Manage database connections, run migrations, and inspect schemas.

Example:
  naeos db connect --type sqlite --name mydb
  naeos db list
  naeos db migrate --name mydb
  naeos db disconnect --name mydb`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newDBConnectCommand())
	cmd.AddCommand(newDBListCommand())
	cmd.AddCommand(newDBMigrateCommand())
	cmd.AddCommand(newDBDisconnectCommand())
	cmd.AddCommand(newDBStatusCommand())

	return cmd
}

func newDBConnectCommand() *cobra.Command {
	var dbType, name, host, user, pass, dbname, sslmode string
	var port int

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &database.Config{
				Host:     host,
				Port:     port,
				User:     user,
				Password: pass,
				Database: dbname,
				SSLMode:  sslmode,
			}

			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("invalid config: %w", err)
			}

			db := database.New(dbType)
			if db == nil {
				return fmt.Errorf("unsupported database type: %s", dbType)
			}

			if err := db.Connect(cfg); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}

			store := database.NewConnectionStore()
			if err := store.Add(name, dbType, cfg); err != nil {
				db.Close()
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Connected to %s database '%s'\n", dbType, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&dbType, "type", "sqlite", "database type (sqlite, postgresql, mysql)")
	cmd.Flags().StringVar(&name, "name", "", "connection name (required)")
	cmd.Flags().StringVar(&host, "host", "localhost", "database host")
	cmd.Flags().IntVar(&port, "port", 5432, "database port")
	cmd.Flags().StringVar(&user, "user", "", "database username")
	cmd.Flags().StringVar(&pass, "pass", "", "database password")
	cmd.Flags().StringVar(&dbname, "database", "", "database name")
	cmd.Flags().StringVar(&sslmode, "sslmode", "disable", "SSL mode")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newDBListCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all database connections",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := database.NewConnectionStore()
			conns, err := store.List()
			if err != nil {
				return err
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), conns, outputFormat)
			}

			if len(conns) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No database connections.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-15s %-12s %-20s\n", "NAME", "TYPE", "DATABASE")
			fmt.Fprintf(out, "%-15s %-12s %-20s\n", "----", "----", "--------")
			for _, c := range conns {
				dbName := ""
				if c.Config != nil {
					dbName = c.Config.Database
				}
				fmt.Fprintf(out, "%-15s %-12s %-20s\n", c.Name, c.Driver, dbName)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newDBMigrateCommand() *cobra.Command {
	var name, migrationsDir string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := database.NewConnectionStore()
			saved, err := store.Get(name)
			if err != nil {
				return err
			}

			db := database.New(saved.Driver)
			if db == nil {
				return fmt.Errorf("unsupported driver: %s", saved.Driver)
			}

			if err := db.Connect(saved.Config); err != nil {
				return fmt.Errorf("connect: %w", err)
			}
			defer db.Close()

			var migrations []database.Migration
			if migrationsDir != "" {
				var err error
				migrations, err = database.LoadMigrations(migrationsDir)
				if err != nil {
					return fmt.Errorf("load migrations: %w", err)
				}
			}

			if err := db.Migrate(migrations); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Applied %d migration(s) to '%s'.\n", len(migrations), name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "connection name (required)")
	cmd.Flags().StringVar(&migrationsDir, "dir", "", "migration directory (optional)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newDBDisconnectCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Disconnect from a database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := database.NewConnectionStore()
			if err := store.Remove(name); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Disconnected from '%s'.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "connection name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newDBStatusCommand() *cobra.Command {
	var name string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show database connection status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := database.NewConnectionStore()
			saved, err := store.Get(name)
			if err != nil {
				return err
			}

			db := database.New(saved.Driver)
			if db == nil {
				return fmt.Errorf("unsupported driver: %s", saved.Driver)
			}

			type statusResult struct {
				Connection string `json:"connection" yaml:"connection"`
				Driver     string `json:"driver" yaml:"driver"`
				Database   string `json:"database" yaml:"database"`
				Host       string `json:"host" yaml:"host"`
				Status     string `json:"status" yaml:"status"`
				Error      string `json:"error,omitempty" yaml:"error,omitempty"`
			}

			data := statusResult{
				Connection: name,
				Driver:     saved.Driver,
				Database:   saved.Config.Database,
				Host:       fmt.Sprintf("%s:%d", saved.Config.Host, saved.Config.Port),
			}

			if err := db.Connect(saved.Config); err != nil {
				data.Status = "DISCONNECTED"
				data.Error = err.Error()
				if outputFormat != "" && outputFormat != "text" {
					return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Status: DISCONNECTED (%s)\n", err)
				return nil
			}
			defer db.Close()

			if err := db.HealthCheck(); err != nil {
				data.Status = "UNHEALTHY"
				data.Error = err.Error()
				if outputFormat != "" && outputFormat != "text" {
					return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Status: UNHEALTHY (%s)\n", err)
				return nil
			}

			data.Status = "HEALTHY"

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), data, outputFormat)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Connection: %s\n", name)
			fmt.Fprintf(out, "Driver:     %s\n", saved.Driver)
			fmt.Fprintf(out, "Database:   %s\n", saved.Config.Database)
			fmt.Fprintf(out, "Host:       %s:%d\n", saved.Config.Host, saved.Config.Port)
			fmt.Fprintf(out, "Status:     HEALTHY\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "connection name (required)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	cmd.MarkFlagRequired("name")
	return cmd
}
