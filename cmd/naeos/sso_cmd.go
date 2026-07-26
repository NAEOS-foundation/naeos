package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/auth"
)

func newSSOCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sso",
		Short: "SSO/SAML/LDAP provider management",
		Long: `Configure and manage SSO identity providers (OIDC, SAML 2.0, LDAP).

Example:
  naeos auth sso configure oidc --name azure --issuer https://login.microsoftonline.com/... --client-id ... --client-secret ...
  naeos auth sso configure saml --name okta --sso-url https://... --entity-id ...
  naeos auth sso configure ldap --name corp-ad --host ad.example.com --base-dn dc=example,dc=com
  naeos auth sso list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSSOConfigureCommand())
	cmd.AddCommand(newSSOListCommand())
	cmd.AddCommand(newSSORemoveCommand())

	return cmd
}

func newSSOConfigureCommand() *cobra.Command {
	var providerType, name, issuer, clientID, clientSecret, redirectURL string
	var scopes []string
	var ssoURL, entityID, certFile string
	var host string
	var port int
	var baseDN, bindDN, bindPassword, userFilter string

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure an SSO provider",
		Long: `Configure an SSO identity provider for authentication.

Supported provider types: oidc, saml, ldap

OIDC example:
  naeos auth sso configure oidc --name azure --issuer https://login.microsoftonline.com/tenant/v2.0 --client-id xxx --client-secret yyy --redirect-url http://localhost:8080/callback

SAML example:
  naeos auth sso configure saml --name okta --sso-url https://dev-xxxx.okta.com/app/.../sso/saml --entity-id https://example.com --cert-file ./okta-cert.pem

LDAP example:
  naeos auth sso configure ldap --name corp-ad --host ad.example.com --port 389 --base-dn dc=example,dc=com --bind-dn cn=admin,dc=example,dc=com --bind-password secret`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			providerType := args[0]

			if name == "" {
				return fmt.Errorf("--name is required")
			}

			var provider auth.SSOProvider
			cfg := auth.SSOConfig{
				Name:        name,
				RedirectURL: redirectURL,
				Scopes:      scopes,
				CertFile:    certFile,
				Enabled:     true,
			}

			switch providerType {
			case "oidc":
				if issuer == "" {
					return fmt.Errorf("--issuer is required for OIDC")
				}
				if clientID == "" {
					return fmt.Errorf("--client-id is required for OIDC")
				}
				if clientSecret == "" {
					return fmt.Errorf("--client-secret is required for OIDC")
				}
				cfg.Type = auth.ProviderOIDC
				cfg.Issuer = issuer
				cfg.ClientID = clientID
				cfg.ClientSecret = clientSecret
				if len(cfg.Scopes) == 0 {
					cfg.Scopes = []string{"openid", "email", "profile"}
				}
				p := auth.NewOIDCProvider(cfg)
				if err := p.Discover(); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: discovery failed (%v), saving config anyway\n", err)
				}
				provider = p

			case "saml":
				if ssoURL == "" {
					return fmt.Errorf("--sso-url is required for SAML")
				}
				if entityID == "" {
					return fmt.Errorf("--entity-id is required for SAML")
				}
				cfg.Type = auth.ProviderSAML
				cfg.SSOURL = ssoURL
				cfg.EntityID = entityID
				provider = auth.NewSAMLProvider(cfg)

			case "ldap":
				if host == "" {
					return fmt.Errorf("--host is required for LDAP")
				}
				if baseDN == "" {
					return fmt.Errorf("--base-dn is required for LDAP")
				}
				cfg.Type = auth.ProviderLDAP
				cfg.Host = host
				cfg.Port = port
				cfg.BaseDN = baseDN
				cfg.BindDN = bindDN
				cfg.BindPassword = bindPassword
				cfg.UserFilter = userFilter
				provider = auth.NewLDAPProvider(cfg)

			default:
				return fmt.Errorf("unsupported provider type %q — supported: oidc, saml, ldap", providerType)
			}

			if err := provider.Validate(); err != nil {
				return fmt.Errorf("validate: %w", err)
			}

			mgr := auth.NewManager()
			if err := mgr.SSO().Register(provider); err != nil {
				return fmt.Errorf("register: %w", err)
			}

			if err := saveSSOConfig(cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "SSO provider %q (%s) configured successfully\n", name, providerType)
			return nil
		},
	}

	cmd.Flags().StringVar(&providerType, "provider-type", "", "")
	cmd.Flags().StringVar(&name, "name", "", "provider name (required)")
	cmd.Flags().StringVar(&issuer, "issuer", "", "OIDC issuer URL")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth2 client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth2 client secret")
	cmd.Flags().StringVar(&redirectURL, "redirect-url", "http://localhost:8080/callback", "OAuth2 redirect URL")
	cmd.Flags().StringArrayVar(&scopes, "scope", nil, "OAuth2 scopes")
	cmd.Flags().StringVar(&ssoURL, "sso-url", "", "SAML SSO URL")
	cmd.Flags().StringVar(&entityID, "entity-id", "", "SAML entity ID")
	cmd.Flags().StringVar(&certFile, "cert-file", "", "SAML x509 certificate file path")
	cmd.Flags().StringVar(&host, "host", "", "LDAP host")
	cmd.Flags().IntVar(&port, "port", 389, "LDAP port (636 for LDAPS)")
	cmd.Flags().StringVar(&baseDN, "base-dn", "", "LDAP base DN")
	cmd.Flags().StringVar(&bindDN, "bind-dn", "", "LDAP bind DN")
	cmd.Flags().StringVar(&bindPassword, "bind-password", "", "LDAP bind password")
	cmd.Flags().StringVar(&userFilter, "user-filter", "", "LDAP user filter (default: (uid=%s))")

	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newSSOListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured SSO providers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configs, err := loadSSOConfigs()
			if err != nil {
				return fmt.Errorf("load configs: %w", err)
			}

			out := cmd.OutOrStdout()
			if len(configs) == 0 {
				fmt.Fprintf(out, "No SSO providers configured.\n")
				fmt.Fprintf(out, "Use: naeos auth sso configure oidc|saml|ldap --name <name> ...\n")
				return nil
			}

			fmt.Fprintf(out, "%-20s %-10s %-40s %s\n", "NAME", "TYPE", "ISSUER/HOST", "ENABLED")
			fmt.Fprintf(out, "%-20s %-10s %-40s %s\n", "--------------------", "----------", "----------------------------------------", "-------")
			for _, c := range configs {
				host := c.Issuer
				if host == "" {
					host = fmt.Sprintf("%s:%d", c.Host, c.Port)
				}
				if host == "" {
					host = c.SSOURL
				}
				enabled := "yes"
				if !c.Enabled {
					enabled = "no"
				}
				fmt.Fprintf(out, "%-20s %-10s %-40s %s\n", c.Name, c.Type, host, enabled)
			}
			return nil
		},
	}
}

func newSSORemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an SSO provider configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			configs, err := loadSSOConfigs()
			if err != nil {
				return fmt.Errorf("load configs: %w", err)
			}

			var found bool
			var kept []auth.SSOConfig
			for _, c := range configs {
				if c.Name == name {
					found = true
					continue
				}
				kept = append(kept, c)
			}

			if !found {
				return fmt.Errorf("SSO provider %q not found", name)
			}

			if err := saveSSOConfigs(kept); err != nil {
				return fmt.Errorf("save configs: %w", err)
			}

			mgr := auth.NewManager()
			mgr.SSO().Remove(name)

			fmt.Fprintf(cmd.OutOrStdout(), "Removed SSO provider %q\n", name)
			return nil
		},
	}
}

func ssoConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(homeDir, ".config", "naeos")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "sso.json"), nil
}

func saveSSOConfig(cfg auth.SSOConfig) error {
	configs, err := loadSSOConfigs()
	if err != nil {
		configs = nil
	}

	var found bool
	for i, c := range configs {
		if c.Name == cfg.Name {
			configs[i] = cfg
			found = true
			break
		}
	}
	if !found {
		configs = append(configs, cfg)
	}

	return saveSSOConfigs(configs)
}

func saveSSOConfigs(configs []auth.SSOConfig) error {
	path, err := ssoConfigPath()
	if err != nil {
		return err
	}

	// Redact secrets
	safe := make([]auth.SSOConfig, len(configs))
	for i, c := range configs {
		safe[i] = c
		safe[i].ClientSecret = ""
		safe[i].BindPassword = ""
	}

	data, err := json.MarshalIndent(safe, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func loadSSOConfigs() ([]auth.SSOConfig, error) {
	path, err := ssoConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var configs []auth.SSOConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}
