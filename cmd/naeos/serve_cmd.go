package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/internal/serve"
)

func newServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run NAEOS as a production daemon",
		Long: `Run the NAEOS server as a production daemon with TLS, graceful
shutdown, multiple listeners, and systemd integration.

Example:
  naeos serve                          # run with defaults on :8080
  naeos serve --config server.yaml     # run from a config file
  naeos serve run --port 9443 --tls-cert cert.pem --tls-key key.pem
  naeos serve config > server.yaml     # print a starter config
  naeos serve install --config server.yaml   # install a systemd unit`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd)
		},
	}

	cmd.AddCommand(newServeRunCommand())
	cmd.AddCommand(newServeInstallCommand())
	cmd.AddCommand(newServeUninstallCommand())
	cmd.AddCommand(newServeConfigCommand())

	cmd.Flags().String("config", "", "path to server config file (YAML)")
	cmd.Flags().StringP("port", "p", "8080", "API listener port")
	cmd.Flags().String("tls-cert", "", "path to TLS certificate (enables HTTPS)")
	cmd.Flags().String("tls-key", "", "path to TLS private key")
	cmd.Flags().String("jwt-secret", "", "JWT secret for API auth")
	cmd.Flags().Bool("auth", false, "enable JWT authentication")

	return cmd
}

func newServeRunCommand() *cobra.Command {
	var config, port, tlsCert, tlsKey, jwtSecret string
	var auth bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the NAEOS daemon in the foreground",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := buildServeConfig(config, port, tlsCert, tlsKey, jwtSecret, auth)
			if err != nil {
				return err
			}
			srv, err := serve.New(cfg)
			if err != nil {
				return err
			}
			return srv.Start()
		},
	}

	cmd.Flags().StringVar(&config, "config", "", "path to server config file (YAML)")
	cmd.Flags().StringVarP(&port, "port", "p", "8080", "API listener port")
	cmd.Flags().StringVar(&tlsCert, "tls-cert", "", "path to TLS certificate")
	cmd.Flags().StringVar(&tlsKey, "tls-key", "", "path to TLS private key")
	cmd.Flags().StringVar(&jwtSecret, "jwt-secret", "", "JWT secret for API auth")
	cmd.Flags().BoolVar(&auth, "auth", false, "enable JWT authentication")
	return cmd
}

func newServeConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print a starter server configuration (YAML)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := yaml.Marshal(serve.DefaultConfig())
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
}

func newServeInstallCommand() *cobra.Command {
	var config, binary, user, group string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a systemd unit for the NAEOS daemon",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config == "" {
				return fmt.Errorf("--config is required to install the systemd unit")
			}
			absConfig, err := filepath.Abs(config)
			if err != nil {
				return err
			}
			bin := binary
			if bin == "" {
				bin, err = os.Executable()
				if err != nil {
					return err
				}
			}
			unit := serve.SystemdUnit(serve.SystemdConfig{
				BinaryPath: bin,
				ConfigPath: absConfig,
				WorkingDir: ".",
				User:       user,
				Group:      group,
			})
			unitPath := "/etc/systemd/system/naeos.service"
			if user != "" || os.Geteuid() != 0 {
				home, _ := os.UserHomeDir()
				unitDir := filepath.Join(home, ".config", "systemd", "user")
				if err := os.MkdirAll(unitDir, 0755); err != nil {
					return err
				}
				unitPath = filepath.Join(unitDir, "naeos.service")
			}
			if err := os.WriteFile(unitPath, []byte(unit), 0644); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "installed systemd unit: %s\n", unitPath)
			if os.Geteuid() == 0 && user == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "run: systemctl daemon-reload && systemctl enable --now naeos")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&config, "config", "", "path to server config file (required)")
	cmd.Flags().StringVar(&binary, "binary", "", "absolute path to the naeos binary (default: current executable)")
	cmd.Flags().StringVar(&user, "user", "", "system user to run the service as")
	cmd.Flags().StringVar(&group, "group", "", "system group to run the service as")
	_ = cmd.MarkFlagRequired("config")
	return cmd
}

func newServeUninstallCommand() *cobra.Command {
	var user bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the installed systemd unit",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			unitPath := "/etc/systemd/system/naeos.service"
			if user || os.Geteuid() != 0 {
				home, _ := os.UserHomeDir()
				unitPath = filepath.Join(home, ".config", "systemd", "user", "naeos.service")
			}
			if _, err := os.Stat(unitPath); os.IsNotExist(err) {
				return fmt.Errorf("systemd unit not found: %s", unitPath)
			}
			if err := os.Remove(unitPath); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed systemd unit: %s\n", unitPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&user, "user", false, "operate on the user systemd unit instead of the system unit")
	return cmd
}

func runServe(cmd *cobra.Command) error {
	config, _ := cmd.Flags().GetString("config")
	port, _ := cmd.Flags().GetString("port")
	tlsCert, _ := cmd.Flags().GetString("tls-cert")
	tlsKey, _ := cmd.Flags().GetString("tls-key")
	jwtSecret, _ := cmd.Flags().GetString("jwt-secret")
	auth, _ := cmd.Flags().GetBool("auth")

	cfg, err := buildServeConfig(config, port, tlsCert, tlsKey, jwtSecret, auth)
	if err != nil {
		return err
	}
	srv, err := serve.New(cfg)
	if err != nil {
		return err
	}
	return srv.Start()
}

func buildServeConfig(config, port, tlsCert, tlsKey, jwtSecret string, auth bool) (*serve.Config, error) {
	if config != "" {
		return serve.LoadConfig(config)
	}
	cfg := serve.DefaultConfig()
	cfg.Listeners = []serve.Listener{{
		Addr:    ":" + port,
		Name:    "api",
		API:     true,
		TLSCert: tlsCert,
		TLSKey:  tlsKey,
	}}
	cfg.Auth = serve.Auth{Enabled: auth, JWTSecret: jwtSecret}
	return cfg, nil
}
