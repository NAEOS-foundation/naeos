package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/supabase"
)

func newSupabaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supabase",
		Short: "Supabase backend management",
		Long: `Manage Supabase projects: authentication, storage, and SQL queries.

Examples:
  naeos supabase init
  naeos supabase auth signup --email user@example.com --password secret
  naeos supabase auth signin --email user@example.com --password secret
  naeos supabase auth user
  naeos supabase storage list-buckets
  naeos supabase sql "SELECT * FROM users"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSupabaseInitCommand())
	cmd.AddCommand(newSupabaseAuthCommand())
	cmd.AddCommand(newSupabaseStorageCommand())
	cmd.AddCommand(newSupabaseSQLCommand())
	cmd.AddCommand(newSupabaseStatusCommand())

	return cmd
}

func getSupabaseClient() (*supabase.Client, error) {
	cfg, err := supabase.LoadConfig()
	if err != nil {
		return nil, err
	}
	return supabase.NewClient(cfg), nil
}

// --- init ---

func newSupabaseInitCommand() *cobra.Command {
	var projectRef, url, anonKey, serviceRoleKey, jwksURL string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Supabase project configuration",
		Long: `Configure a Supabase project. Reads from environment variables if flags are not set.

Env vars:
  SUPABASE_PROJECT_REF, SUPABASE_URL, SUPABASE_PUBLISHABEL_KEY (anon),
  SUPABASE_SECRET_KEY (service role), SUPABASE_JWKS_URL

Examples:
  naeos supabase init --project-ref abcdefg
  naeos supabase init --project-ref abcdefg --anon-key "eyJ..." --service-role-key "eyJ..."
  SUPABASE_URL=https://abc.supabase.co SUPABASE_PUBLISHABEL_KEY=eyJ... naeos supabase init --project-ref abc`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectRef == "" {
				projectRef = os.Getenv("SUPABASE_PROJECT_REF")
			}
			if projectRef == "" {
				return fmt.Errorf("--project-ref is required (or set SUPABASE_PROJECT_REF)")
			}
			if url == "" {
				url = os.Getenv("SUPABASE_URL")
			}
			if url == "" {
				url = "https://" + projectRef + ".supabase.co"
			}
			if anonKey == "" {
				anonKey = os.Getenv("SUPABASE_ANON_KEY")
			}
			if anonKey == "" {
				anonKey = os.Getenv("SUPABASE_PUBLISHABEL_KEY")
			}
			if anonKey == "" {
				anonKey = os.Getenv("SUPABASE_PUBLISHABLE_KEY")
			}
			if serviceRoleKey == "" {
				serviceRoleKey = os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
			}
			if serviceRoleKey == "" {
				serviceRoleKey = os.Getenv("SUPABASE_SECRET_KEY")
			}
			if jwksURL == "" {
				jwksURL = os.Getenv("SUPABASE_JWKS_URL")
			}

			cfg := &supabase.Config{
				ProjectRef:     projectRef,
				URL:            url,
				AnonKey:        anonKey,
				ServiceRoleKey: serviceRoleKey,
				JWKSURL:        jwksURL,
			}

			if err := supabase.SaveConfig(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Supabase project '%s' configured.\n", projectRef)
			fmt.Fprintf(cmd.OutOrStdout(), "  URL: %s\n", url)
			return nil
		},
	}

	cmd.Flags().StringVar(&projectRef, "project-ref", "", "Supabase project reference (required)")
	cmd.Flags().StringVar(&url, "url", "", "Supabase project URL (default: https://<ref>.supabase.co)")
	cmd.Flags().StringVar(&anonKey, "anon-key", "", "Supabase anon/public key")
	cmd.Flags().StringVar(&serviceRoleKey, "service-role-key", "", "Supabase service role key")
	cmd.Flags().StringVar(&jwksURL, "jwks-url", "", "Supabase JWKS URL for JWT verification")
	_ = cmd.MarkFlagRequired("project-ref")
	return cmd
}

// --- auth subcommand group ---

func newSupabaseAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Supabase authentication management",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSupabaseAuthSignupCommand())
	cmd.AddCommand(newSupabaseAuthSigninCommand())
	cmd.AddCommand(newSupabaseAuthSignoutCommand())
	cmd.AddCommand(newSupabaseAuthUserCommand())
	cmd.AddCommand(newSupabaseAuthAdminCommand())

	return cmd
}

func newSupabaseAuthSignupCommand() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "signup",
		Short: "Sign up a new user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			result, err := client.SignUp(supabase.SignUpParams{
				Email:    email,
				Password: password,
			})
			if err != nil {
				return fmt.Errorf("signup failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "User created:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  ID:    %s\n", result.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Email: %s\n", result.Email)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "user email (required)")
	cmd.Flags().StringVar(&password, "password", "", "user password (required)")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func newSupabaseAuthSigninCommand() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "signin",
		Short: "Sign in with email and password",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			session, err := client.SignInWithEmail(email, password)
			if err != nil {
				return fmt.Errorf("signin failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Signed in as: %s\n", session.User.Email)
			fmt.Fprintf(cmd.OutOrStdout(), "Access token: %s\n", session.AccessToken[:20]+"...")
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "user email (required)")
	cmd.Flags().StringVar(&password, "password", "", "user password (required)")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func newSupabaseAuthSignoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "signout",
		Short: "Sign out current user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			if err := client.SignOut(); err != nil {
				return fmt.Errorf("signout failed: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Signed out.")
			return nil
		},
	}
}

func newSupabaseAuthUserCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "user",
		Short: "Get current authenticated user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			user, err := client.GetUser()
			if err != nil {
				return fmt.Errorf("get user failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "ID:    %s\n", user.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Email: %s\n", user.Email)
			fmt.Fprintf(cmd.OutOrStdout(), "Role:  %s\n", user.Role)
			if user.EmailConfirmedAt != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Email confirmed: %s\n", user.EmailConfirmedAt)
			}
			return nil
		},
	}
}

func newSupabaseAuthAdminCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Admin auth operations (requires service_role_key)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSupabaseAuthAdminListCommand())
	cmd.AddCommand(newSupabaseAuthAdminCreateCommand())
	cmd.AddCommand(newSupabaseAuthAdminDeleteCommand())

	return cmd
}

func newSupabaseAuthAdminListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all users (admin)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			users, err := client.AdminListUsers()
			if err != nil {
				return fmt.Errorf("list users failed: %w", err)
			}

			if len(users) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No users found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-36s %-25s %s\n", "ID", "EMAIL", "CREATED")
			fmt.Fprintf(cmd.OutOrStdout(), "%-36s %-25s %s\n", strings.Repeat("-", 36), strings.Repeat("-", 25), strings.Repeat("-", 20))
			for _, u := range users {
				fmt.Fprintf(cmd.OutOrStdout(), "%-36s %-25s %s\n", u.ID, u.Email, u.CreatedAt)
			}
			return nil
		},
	}
}

func newSupabaseAuthAdminCreateCommand() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user (admin)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			user, err := client.AdminCreateUser(email, password, nil)
			if err != nil {
				return fmt.Errorf("create user failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created user: %s (%s)\n", user.Email, user.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "user email (required)")
	cmd.Flags().StringVar(&password, "password", "", "user password (required)")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func newSupabaseAuthAdminDeleteCommand() *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a user (admin)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			if err := client.AdminDeleteUser(userID); err != nil {
				return fmt.Errorf("delete user failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted user: %s\n", userID)
			return nil
		},
	}

	cmd.Flags().StringVar(&userID, "user-id", "", "user ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	return cmd
}

// --- storage subcommand group ---

func newSupabaseStorageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Supabase storage management",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSupabaseStorageListBucketsCommand())
	cmd.AddCommand(newSupabaseStorageCreateBucketCommand())
	cmd.AddCommand(newSupabaseStorageListFilesCommand())
	cmd.AddCommand(newSupabaseStorageUploadCommand())
	cmd.AddCommand(newSupabaseStorageDownloadCommand())
	cmd.AddCommand(newSupabaseStorageDeleteCommand())

	return cmd
}

func newSupabaseStorageListBucketsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-buckets",
		Short: "List all storage buckets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			buckets, err := client.ListBuckets()
			if err != nil {
				return fmt.Errorf("list buckets failed: %w", err)
			}

			if len(buckets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No buckets found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-25s %-25s %-5s %s\n", "ID", "NAME", "PUBLIC", "CREATED")
			fmt.Fprintf(cmd.OutOrStdout(), "%-25s %-25s %-5s %s\n", strings.Repeat("-", 25), strings.Repeat("-", 25), strings.Repeat("-", 5), strings.Repeat("-", 20))
			for _, b := range buckets {
				public := "no"
				if b.Public {
					public = "yes"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-25s %-25s %-5s %s\n", b.ID, b.Name, public, b.CreatedAt)
			}
			return nil
		},
	}
}

func newSupabaseStorageCreateBucketCommand() *cobra.Command {
	var name string
	var public bool

	cmd := &cobra.Command{
		Use:   "create-bucket",
		Short: "Create a new storage bucket",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			bucket, err := client.CreateBucket(name, public)
			if err != nil {
				return fmt.Errorf("create bucket failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Bucket created: %s\n", bucket.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "bucket name (required)")
	cmd.Flags().BoolVar(&public, "public", false, "make bucket public")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newSupabaseStorageListFilesCommand() *cobra.Command {
	var bucket, prefix string

	cmd := &cobra.Command{
		Use:   "list-files",
		Short: "List files in a bucket",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			files, err := client.ListFiles(bucket, prefix)
			if err != nil {
				return fmt.Errorf("list files failed: %w", err)
			}

			if len(files) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No files in bucket '%s'.\n", bucket)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %s\n", "NAME", "SIZE", "UPDATED")
			fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10s %s\n", strings.Repeat("-", 40), strings.Repeat("-", 10), strings.Repeat("-", 20))
			for _, f := range files {
				fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-10d %s\n", f.Name, f.Metadata.Size, f.UpdatedAt)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket name (required)")
	cmd.Flags().StringVar(&prefix, "prefix", "", "file prefix filter")
	_ = cmd.MarkFlagRequired("bucket")
	return cmd
}

func newSupabaseStorageUploadCommand() *cobra.Command {
	var bucket, source, dest string

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file to storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			if err := client.UploadFile(bucket, source, dest); err != nil {
				return fmt.Errorf("upload failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Uploaded %s -> %s/%s\n", source, bucket, dest)
			return nil
		},
	}

	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket name (required)")
	cmd.Flags().StringVar(&source, "source", "", "local file path (required)")
	cmd.Flags().StringVar(&dest, "dest", "", "remote path (default: basename of source)")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("source")
	return cmd
}

func newSupabaseStorageDownloadCommand() *cobra.Command {
	var bucket, source, dest string

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download a file from storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			if err := client.DownloadFile(bucket, source, dest); err != nil {
				return fmt.Errorf("download failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Downloaded %s/%s -> %s\n", bucket, source, dest)
			return nil
		},
	}

	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket name (required)")
	cmd.Flags().StringVar(&source, "source", "", "remote file path (required)")
	cmd.Flags().StringVar(&dest, "dest", "", "local destination path (required)")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("dest")
	return cmd
}

func newSupabaseStorageDeleteCommand() *cobra.Command {
	var bucket, path string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a file from storage",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			if err := client.DeleteFile(bucket, path); err != nil {
				return fmt.Errorf("delete failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s/%s\n", bucket, path)
			return nil
		},
	}

	cmd.Flags().StringVar(&bucket, "bucket", "", "bucket name (required)")
	cmd.Flags().StringVar(&path, "path", "", "file path (required)")
	_ = cmd.MarkFlagRequired("bucket")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

// --- SQL ---

func newSupabaseSQLCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sql",
		Short: "Execute SQL query via Supabase Management API",
		Long: `Execute a SQL query against the Supabase database using the service role key.

Examples:
  naeos supabase sql "SELECT * FROM users"
  naeos supabase sql "CREATE TABLE test (id SERIAL PRIMARY KEY, name TEXT)"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getSupabaseClient()
			if err != nil {
				return err
			}

			result, err := client.ExecuteSQL(args[0])
			if err != nil {
				return fmt.Errorf("SQL execution failed: %w", err)
			}

			if result.Error != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "SQL error: %s\n", result.Error)
				return nil
			}

			if len(result.Rows) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "Query executed successfully (0 rows).")
				return nil
			}

			var keys []string
			for i, row := range result.Rows {
				if i == 0 {
					keys = make([]string, 0, len(row))
					for k := range row {
						keys = append(keys, k)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%-20s", "")
					for _, k := range keys {
						fmt.Fprintf(cmd.OutOrStdout(), "%-20s", k)
					}
					fmt.Fprintln(cmd.OutOrStdout())
					fmt.Fprintf(cmd.OutOrStdout(), "%-20s", strings.Repeat("-", 20))
					for range keys {
						fmt.Fprintf(cmd.OutOrStdout(), "%-20s", strings.Repeat("-", 20))
					}
					fmt.Fprintln(cmd.OutOrStdout())
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Row %-16d", i+1)
				for _, k := range keys {
					fmt.Fprintf(cmd.OutOrStdout(), "%-20v", row[k])
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}
}

// --- status ---

func newSupabaseStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check Supabase connection status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := supabase.LoadConfig()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Not configured: %v\n", err)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Project ref: %s\n", cfg.ProjectRef)
			fmt.Fprintf(cmd.OutOrStdout(), "URL:         %s\n", cfg.URL)
			fmt.Fprintf(cmd.OutOrStdout(), "Anon key:    %s\n", supabase.MaskKey(cfg.AnonKey))
			fmt.Fprintf(cmd.OutOrStdout(), "Service key: %s\n", supabase.MaskKey(cfg.ServiceRoleKey))
			if cfg.JWKSURL != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "JWKS URL:    %s\n", cfg.JWKSURL)
			}

			if cfg.AnonKey == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Status: PARTIAL (anon key not set)\n")
				return nil
			}

			client := supabase.NewClient(cfg)
			user, err := client.GetUser()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Status: CONNECTED (auth check: %v)\n", err)
				return nil
			}

			if user != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Status: CONNECTED\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Status: CONNECTED\n")
			}
			return nil
		},
	}
}




