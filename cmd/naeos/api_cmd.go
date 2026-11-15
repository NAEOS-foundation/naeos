// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/api"
	"github.com/NAEOS-foundation/naeos/internal/auth"
	"github.com/NAEOS-foundation/naeos/internal/database"
)

var (
	apiPort     string
	apiAuth     bool
	apiSecret   string
	apiAuthMode string
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

func apiDatabaseConfigFromEnv() (string, *database.Config, error) {
	driver := strings.TrimSpace(os.Getenv("NAEOS_DB_DRIVER"))
	dsn := strings.TrimSpace(os.Getenv("NAEOS_DB_DATABASE"))
	if driver == "" && dsn == "" {
		return "", nil, nil
	}
	if driver == "" {
		driver = "sqlite"
	}

	dbConfig := &database.Config{Database: dsn}
	switch driver {
	case "postgres", "postgresql":
		dbConfig.Host = strings.TrimSpace(os.Getenv("NAEOS_DB_HOST"))
		dbConfig.User = os.Getenv("NAEOS_DB_USER")
		password, err := readSecretEnv("NAEOS_DB_PASSWORD", "NAEOS_DB_PASSWORD_FILE")
		if err != nil {
			return "", nil, err
		}
		dbConfig.Password = password
		dbConfig.Port = 5432
		dbConfig.SSLMode = strings.TrimSpace(os.Getenv("NAEOS_DB_SSLMODE"))
		if dbConfig.SSLMode == "" {
			dbConfig.SSLMode = "verify-full"
		}
	case "mysql", "mariadb":
		dbConfig.Host = strings.TrimSpace(os.Getenv("NAEOS_DB_HOST"))
		dbConfig.User = os.Getenv("NAEOS_DB_USER")
		password, err := readSecretEnv("NAEOS_DB_PASSWORD", "NAEOS_DB_PASSWORD_FILE")
		if err != nil {
			return "", nil, err
		}
		dbConfig.Password = password
		dbConfig.Port = 3306
	}
	if port := strings.TrimSpace(os.Getenv("NAEOS_DB_PORT")); port != "" {
		dbPort, err := strconv.Atoi(port)
		if err != nil || dbPort < 1 || dbPort > 65535 {
			return "", nil, fmt.Errorf("invalid NAEOS_DB_PORT %q: must be an integer from 1 to 65535", port)
		}
		dbConfig.Port = dbPort
	}
	if driver == "postgres" || driver == "postgresql" || driver == "mysql" || driver == "mariadb" {
		if dbConfig.Host == "" || dbConfig.User == "" || dbConfig.Password == "" || dbConfig.Database == "" {
			return "", nil, fmt.Errorf("NAEOS_DB_HOST, NAEOS_DB_USER, NAEOS_DB_PASSWORD, and NAEOS_DB_DATABASE are required for %s", driver)
		}
	}
	return driver, dbConfig, nil
}

// attachDatabase connects the API server to a database when the
// NAEOS_DB_DRIVER and NAEOS_DB_DATABASE environment variables are set.
func attachDatabase(server *api.Server) error {
	driver, dbConfig, err := apiDatabaseConfigFromEnv()
	if err != nil {
		return err
	}
	if dbConfig == nil {
		return nil
	}

	db := database.New(driver)
	if db == nil {
		return fmt.Errorf("unknown NAEOS_DB_DRIVER %q", driver)
	}

	if err := db.Connect(dbConfig); err != nil {
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
const apiAuthModeEnv = "NAEOS_API_AUTH_MODE"

// registerAPIAuth attaches an auth manager seeded with an admin user so that a
// legacy HS256 token whose "sub" matches the configured admin user can use protected routes.
func registerAPIAuth(server *api.Server, encryptionKey string) error {
	m := auth.NewManager(encryptionKey)
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

func validateAPIAuthSecrets(jwtSecret, encryptionKey string) error {
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 bytes; set NAEOS_API_JWT_SECRET or pass --secret")
	}
	if len(encryptionKey) < 32 {
		return fmt.Errorf("NAEOS_ENCRYPTION_KEY must be at least 32 bytes")
	}
	if jwtSecret == encryptionKey {
		return fmt.Errorf("NAEOS_API_JWT_SECRET and NAEOS_ENCRYPTION_KEY must be different")
	}
	return nil
}

func resolveAPIJWTSecret(cmd *cobra.Command) (string, error) {
	if cmd.Flags().Changed("secret") {
		return apiSecret, nil
	}
	if apiSecret != "" {
		return apiSecret, nil
	}
	return readSecretEnv("NAEOS_API_JWT_SECRET", "NAEOS_API_JWT_SECRET_FILE")
}

func readSecretEnv(valueEnv, fileEnv string) (string, error) {
	value := os.Getenv(valueEnv)
	file := strings.TrimSpace(os.Getenv(fileEnv))
	if value != "" && file != "" {
		return "", fmt.Errorf("%s and %s must not both be set", valueEnv, fileEnv)
	}
	if value != "" {
		return value, nil
	}
	if file == "" {
		return "", nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", fileEnv, err)
	}
	secret := strings.TrimSuffix(string(data), "\n")
	secret = strings.TrimSuffix(secret, "\r")
	if secret == "" {
		return "", fmt.Errorf("%s points to an empty secret file", fileEnv)
	}
	return secret, nil
}

func resolveAPIAuthMode(cmd *cobra.Command) string {
	if cmd.Flags().Changed("auth-mode") {
		return strings.ToLower(strings.TrimSpace(apiAuthMode))
	}
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(apiAuthModeEnv)))
	if mode == "" {
		return "hs256"
	}
	return mode
}

func newAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Start NAEOS REST API server",
		Long:  `Start the NAEOS REST API server for external integrations and web dashboard.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jwtSecret, err := resolveAPIJWTSecret(cmd)
			if err != nil {
				return err
			}
			encryptionKey, err := readSecretEnv("NAEOS_ENCRYPTION_KEY", "NAEOS_ENCRYPTION_KEY_FILE")
			if err != nil {
				return err
			}
			authMode := resolveAPIAuthMode(cmd)
			var supabaseAuth *api.SupabaseAuthVerifier
			if apiAuth {
				switch authMode {
				case "hs256":
					if err := validateAPIAuthSecrets(jwtSecret, encryptionKey); err != nil {
						return err
					}
				case "supabase":
					if jwtSecret != "" {
						return fmt.Errorf("NAEOS_API_JWT_SECRET and --secret must not be set when %s=supabase", apiAuthModeEnv)
					}
					var err error
					supabaseAuth, err = api.NewSupabaseAuthVerifier(api.SupabaseAuthConfig{
						URL:      os.Getenv("SUPABASE_URL"),
						Audience: os.Getenv("SUPABASE_JWT_AUDIENCE"),
						JWKSURL:  os.Getenv("SUPABASE_JWKS_URL"),
					})
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("unsupported %s %q: use hs256 or supabase", apiAuthModeEnv, authMode)
				}
			}
			if supabaseAuth != nil {
				defer supabaseAuth.Close()
			}

			auth := &api.AuthConfig{
				Enabled:       apiAuth,
				JWTSecret:     jwtSecret,
				EncryptionKey: encryptionKey,
				SupabaseAuth:  supabaseAuth,
			}

			server := api.NewServer(":"+apiPort, auth)
			if err := attachDatabase(server); err != nil {
				return err
			}
			if apiAuth && authMode == "hs256" {
				if err := registerAPIAuth(server, encryptionKey); err != nil {
					return err
				}
			}
			return server.Start()
		},
	}

	cmd.Flags().StringVarP(&apiPort, "port", "p", "8080", "API server port")
	cmd.Flags().BoolVarP(&apiAuth, "auth", "a", false, "Enable JWT authentication")
	cmd.Flags().StringVarP(&apiSecret, "secret", "s", "", "JWT secret key (defaults to NAEOS_API_JWT_SECRET)")
	cmd.Flags().StringVar(&apiAuthMode, "auth-mode", "", "Authentication mode: hs256 or supabase (defaults to NAEOS_API_AUTH_MODE or hs256)")

	return cmd
}
