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
			mgr := database.NewManager()

			var db database.Database
			switch dbType {
			case "sqlite":
				db = database.NewSQLite()
			case "postgresql":
				db = database.NewPostgreSQL()
			case "mysql":
				db = database.NewMySQL()
			default:
				return fmt.Errorf("unsupported database type: %s", dbType)
			}

			mgr.Register(name, db)

			cfg := &database.Config{
				Host:     host,
				Port:     port,
				User:     user,
				Password: pass,
				Database: dbname,
				SSLMode:  sslmode,
			}

			if err := db.Connect(cfg); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
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
	return &cobra.Command{
		Use:   "list",
		Short: "List all database connections",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := database.NewManager()

			names := mgr.List()
			if len(names) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No database connections.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-15s\n", "NAME")
			fmt.Fprintf(out, "%-15s\n", "----")
			for _, name := range names {
				fmt.Fprintf(out, "%-15s\n", name)
			}
			return nil
		},
	}
}

func newDBMigrateCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := database.NewManager()

			db, ok := mgr.Get(name)
			if !ok {
				return fmt.Errorf("database '%s' not found", name)
			}

			if err := db.Migrate(nil); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Migrations applied to '%s' successfully.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "connection name (required)")
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
			mgr := database.NewManager()

			db, ok := mgr.Get(name)
			if !ok {
				return fmt.Errorf("database '%s' not found", name)
			}

			if err := db.Close(); err != nil {
				return fmt.Errorf("failed to disconnect: %w", err)
			}

			mgr.Remove(name)
			fmt.Fprintf(cmd.OutOrStdout(), "Disconnected from '%s'.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "connection name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}
