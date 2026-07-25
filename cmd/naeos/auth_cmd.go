package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/auth"
)

func newAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and authorization management",
		Long: `Manage users, roles, API keys, and OAuth2 providers.

Example:
  naeos auth whoami --api-key <key>
  naeos auth create-user --name john --email john@example.com --role admin
  naeos auth create-key --user-id u1 --name my-api-key
  naeos auth list-users
  naeos auth list-roles`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newAuthWhoamiCommand())
	cmd.AddCommand(newAuthCreateUserCommand())
	cmd.AddCommand(newAuthCreateKeyCommand())
	cmd.AddCommand(newAuthListUsersCommand())
	cmd.AddCommand(newAuthListRolesCommand())
	cmd.AddCommand(newAuthLoginCommand())
	cmd.AddCommand(newAuthLogoutCommand())
	cmd.AddCommand(newAuthCreateRoleCommand())
	cmd.AddCommand(newAuthDeleteRoleCommand())
	cmd.AddCommand(newAuthAssignRoleCommand())
	cmd.AddCommand(newAuthCreateRoleFromTemplateCommand())
	cmd.AddCommand(newAuthListRoleTemplatesCommand())
	cmd.AddCommand(newSSOCommand())

	return cmd
}

func newAuthWhoamiCommand() *cobra.Command {
	var apiKey string

	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show current authenticated identity",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := auth.NewManager()

			if apiKey == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "No API key provided. Use --api-key flag.")
				return nil
			}

			user, ok := mgr.AuthenticateAPIKey(apiKey)
			if !ok {
				fmt.Fprintln(cmd.OutOrStdout(), "Authentication failed: invalid API key.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "User:  %s\n", user.Name)
			fmt.Fprintf(out, "ID:    %s\n", user.ID)
			fmt.Fprintf(out, "Email: %s\n", user.Email)
			fmt.Fprintf(out, "Roles: %s\n", strings.Join(user.Roles, ", "))
			return nil
		},
	}

	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key to authenticate with")
	return cmd
}

func newAuthCreateUserCommand() *cobra.Command {
	var name, email string
	var roles []string

	cmd := &cobra.Command{
		Use:   "create-user",
		Short: "Create a new user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := auth.NewUserStore("")
			mgr := auth.NewManager()

			user := &auth.User{
				ID:        generateSimpleID(),
				Name:      name,
				Email:     email,
				Roles:     roles,
				CreatedAt: time.Now(),
			}
			mgr.CreateUser(user)

			if err := store.Add(user); err != nil {
				return fmt.Errorf("failed to persist user: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Created user: %s (ID: %s)\n", user.Name, user.ID)
			fmt.Fprintf(out, "Roles: %s\n", strings.Join(user.Roles, ", "))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "user name (required)")
	cmd.Flags().StringVar(&email, "email", "", "user email")
	cmd.Flags().StringArrayVar(&roles, "role", nil, "user roles")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newAuthCreateKeyCommand() *cobra.Command {
	var userID, name string

	cmd := &cobra.Command{
		Use:   "create-key",
		Short: "Create a new API key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := auth.NewManager()

			key, err := mgr.APIKeys().Generate(userID, name, []string{"read", "write"}, time.Now().Add(24*time.Hour))
			if err != nil {
				return fmt.Errorf("failed to create API key: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Created API key: %s\n", key)
			fmt.Fprintf(out, "Name: %s | User: %s\n", name, userID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "key name (required)")
	cmd.Flags().StringVar(&userID, "user-id", "", "associated user ID (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("user-id")
	return cmd
}

func newAuthListUsersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-users",
		Short: "List all users",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := auth.NewUserStore("")

			users, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list users: %w", err)
			}

			if len(users) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No users found. Create users with 'naeos auth create-user'.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-10s %-15s %-25s %-20s\n", "ID", "NAME", "EMAIL", "ROLES")
			fmt.Fprintf(out, "%-10s %-15s %-25s %-20s\n", "----", "----", "-----", "-----")
			for _, u := range users {
				roles := strings.Join(u.Roles, ", ")
				fmt.Fprintf(out, "%-10s %-15s %-25s %-20s\n", u.ID, u.Name, u.Email, roles)
			}
			return nil
		},
	}
}

func newAuthListRolesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-roles",
		Short: "List all roles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := auth.NewManager()
			roles := mgr.RBAC().ListRoles()

			if len(roles) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No roles defined. Add roles via RBAC.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-25s %s\n", "ROLE", "PERMISSIONS")
			fmt.Fprintf(out, "%-25s %s\n", "-------------------------", "-----------")
			for _, name := range roles {
				role, _ := mgr.RBAC().GetRole(name)
				perms := ""
				if role != nil {
					perms = strings.Join(role.Permissions, ", ")
				}
				fmt.Fprintf(out, "%-25s %s\n", name, perms)
			}
			return nil
		},
	}
}

func newAuthLoginCommand() *cobra.Command {
	var provider string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login via OAuth2 provider",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := auth.NewManager()

			oauth, ok := mgr.GetOAuth2(provider)
			if !ok {
				fmt.Fprintf(cmd.OutOrStdout(), "OAuth2 provider '%s' not registered.\n", provider)
				return nil
			}

			url := oauth.GetAuthorizationURL("naeos-callback")

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Open this URL to authenticate:\n%s\n", url)
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "google", "OAuth2 provider (google, github)")
	return cmd
}

func newAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout current session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Logged out successfully.")
			return nil
		},
	}
}

func newAuthCreateRoleCommand() *cobra.Command {
	var permissions, deny []string
	var parents []string

	cmd := &cobra.Command{
		Use:   "create-role <name>",
		Short: "Create a custom RBAC role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			mgr := auth.NewManager()
			rbac := mgr.RBAC()

			ra := make(map[string][]string)
			for _, p := range permissions {
				parts := strings.SplitN(p, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid permission format %q, use resource:action (e.g., spec:read)", p)
				}
				ra[parts[0]] = append(ra[parts[0]], parts[1])
			}

			var denyMap map[string][]string
			if len(deny) > 0 {
				denyMap = make(map[string][]string)
				for _, d := range deny {
					parts := strings.SplitN(d, ":", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid deny format %q, use resource:action", d)
					}
					denyMap[parts[0]] = append(denyMap[parts[0]], parts[1])
				}
			}

			rbac.AddRole(&auth.Role{
				Name:            name,
				ResourceActions: ra,
				Deny:            denyMap,
				Parents:         parents,
			})

			fmt.Fprintf(cmd.OutOrStdout(), "Created role: %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&permissions, "permission", nil, "permissions (resource:action, e.g., spec:read)")
	cmd.Flags().StringArrayVar(&deny, "deny", nil, "denied permissions (resource:action)")
	cmd.Flags().StringArrayVar(&parents, "parent", nil, "parent roles to inherit from")
	return cmd
}

func newAuthDeleteRoleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-role <name>",
		Short: "Delete an RBAC role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			mgr := auth.NewManager()
			rbac := mgr.RBAC()

			_, ok := rbac.GetRole(name)
			if !ok {
				return fmt.Errorf("role %q not found", name)
			}

			rbac.RemoveRole(name)
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted role: %s\n", name)
			return nil
		},
	}
}

func newAuthAssignRoleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "assign-role <user-id> <role-name>",
		Short: "Assign a role to a user",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID, roleName := args[0], args[1]
			mgr := auth.NewManager()

			user, ok := mgr.GetUser(userID)
			if !ok {
				return fmt.Errorf("user %q not found", userID)
			}

			_, ok = mgr.RBAC().GetRole(roleName)
			if !ok {
				return fmt.Errorf("role %q not found", roleName)
			}

			mgr.RBAC().AssignRole(user, roleName)
			fmt.Fprintf(cmd.OutOrStdout(), "Assigned role %q to user %q\n", roleName, userID)
			return nil
		},
	}
}

func newAuthCreateRoleFromTemplateCommand() *cobra.Command {
	var roleName string
	var parents []string

	cmd := &cobra.Command{
		Use:   "create-role-from-template <template-name>",
		Short: "Create a role from a compliance template (auditor, soc2_auditor, gdpr_admin, hipaa_admin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateName := args[0]
			if roleName == "" {
				roleName = templateName
			}

			mgr := auth.NewManager()
			rbac := mgr.RBAC()

			auth.SetupRoleTemplate(rbac, templateName, roleName, parents)

			fmt.Fprintf(cmd.OutOrStdout(), "Created role %q from template %q\n", roleName, templateName)
			return nil
		},
	}

	cmd.Flags().StringVar(&roleName, "role-name", "", "custom role name (defaults to template name)")
	cmd.Flags().StringArrayVar(&parents, "parent", nil, "parent roles to inherit from")
	return cmd
}

func newAuthListRoleTemplatesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-role-templates",
		Short: "List available role templates for compliance frameworks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-25s %s\n", "TEMPLATE", "DESCRIPTION")
			fmt.Fprintf(out, "%-25s %s\n", "-------------------------", "-----------")
			fmt.Fprintf(out, "%-25s %s\n", "auditor", "Read-only audit, spec, pipeline access")
			fmt.Fprintf(out, "%-25s %s\n", "soc2_auditor", "SOC 2 auditor — read access to audit, spec, pipeline, cloud")
			fmt.Fprintf(out, "%-25s %s\n", "gdpr_admin", "GDPR admin — user data management with audit trail")
			fmt.Fprintf(out, "%-25s %s\n", "hipaa_admin", "HIPAA admin — healthcare compliance with audit & config")
			fmt.Fprintf(out, "\nUse: naeos auth create-role-from-template <template-name> --role-name <name>\n")
			return nil
		},
	}
}

func generateSimpleID() string {
	return time.Now().Format("20060102150405")
}
