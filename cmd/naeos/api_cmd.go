// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/api"
	"github.com/NAEOS-foundation/naeos/internal/auth"
	"github.com/NAEOS-foundation/naeos/internal/database"
)

var (
	apiPort   string
	apiAuth   bool
	apiSecret string
)

// apiPipelineRunsSchema mirrors the columns written by the API server's
// run persistence (internal/api/server.go).
const apiPipelineRunsSchema = `CREATE TABLE IF NOT EXISTS pipeline_runs (
	id         TEXT PRIMARY KEY,
	status     TEXT NOT NULL DEFAULT '',
	project    TEXT NOT NULL DEFAULT '',
	modules    INTEGER NOT NULL DEFAULT 0,
	services   INTEGER NOT NULL DEFAULT 0,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

// attachDatabase connects the API server to a database when the
// NAEOS_DB_DRIVER and NAEOS_DB_DATABASE environment variables are set.
func attachDatabase(server *api.Server) error {
	driver := os.Getenv("NAEOS_DB_DRIVER")
	dsn := os.Getenv("NAEOS_DB_DATABASE")
	if driver == "" && dsn == "" {
		return nil
	}
	if driver == "" {
		driver = "sqlite"
	}

	db := database.New(driver)
	if db == nil {
		return fmt.Errorf("unknown NAEOS_DB_DRIVER %q", driver)
	}

	if err := db.Connect(&database.Config{Database: dsn}); err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	switch driver {
	case "sqlite", "postgres", "postgresql", "mysql", "mariadb":
		if _, err := db.Exec(apiPipelineRunsSchema); err != nil {
			return fmt.Errorf("initialize database schema: %w", err)
		}
	}

	server.SetDatabase(db)
	return nil
}

// apiAdminUserEnv overrides the RBAC user ID seeded when JWT auth is enabled.
const apiAdminUserEnv = "NAEOS_API_ADMIN_USER"

// registerAPIAuth attaches an auth manager seeded with an admin user so that a
// JWT whose "sub" matches the configured admin user can use protected routes.
func registerAPIAuth(server *api.Server) error {
	m := auth.NewManager(os.Getenv("NAEOS_ENCRYPTION_KEY"))
	auth.SetupDefaultRoles(m.RBAC())

	userID := os.Getenv(apiAdminUserEnv)
	if userID == "" {
		userID = "admin"
	}

	m.CreateUser(&auth.User{
		ID:    userID,
		Name:  "NAEOS API Admin",
		Email: "admin@localhost",
		Roles: []string{"admin"},
	})
	server.SetAuthManager(m)
	return nil
}

func newAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Start NAEOS REST API server",
		Long:  `Start the NAEOS REST API server for external integrations and web dashboard.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			auth := &api.AuthConfig{
				Enabled:   apiAuth,
				JWTSecret: apiSecret,
			}

			server := api.NewServer(":"+apiPort, auth)
			if err := attachDatabase(server); err != nil {
				return err
			}
			if apiAuth && apiSecret != "" {
				if err := registerAPIAuth(server); err != nil {
					return err
				}
			}
			return server.Start()
		},
	}

	cmd.Flags().StringVarP(&apiPort, "port", "p", "8080", "API server port")
	cmd.Flags().BoolVarP(&apiAuth, "auth", "a", false, "Enable JWT authentication")
	cmd.Flags().StringVarP(&apiSecret, "secret", "s", "", "JWT secret key")

	return cmd
}
