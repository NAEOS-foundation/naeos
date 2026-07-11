package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/plugin"
)

func newPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage NAEOS plugins",
		Long: `Manage NAEOS plugins (install, uninstall, list).

Example:
  naeos plugin list
  naeos plugin install ./my-plugin.so
  naeos plugin uninstall my-plugin`,
	}

	var pluginDir string

	pluginList := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := plugin.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				return err
			}
			plugins := mgr.List()
			if len(plugins) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No plugins installed")
				return nil
			}
			for _, p := range plugins {
				status := "enabled"
				if !p.Enabled {
					status = "disabled"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-10s %s\n", p.Name, p.Version, status)
			}
			return nil
		},
	}

	pluginInstall := &cobra.Command{
		Use:   "install [path]",
		Short: "Install a plugin from a .so file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := plugin.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				return err
			}
			info, err := mgr.Install(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed plugin %s v%s\n", info.Name, info.Version)
			return nil
		},
	}

	pluginUninstall := &cobra.Command{
		Use:   "uninstall [name]",
		Short: "Uninstall a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := plugin.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				return err
			}
			if err := mgr.Uninstall(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled plugin %s\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(pluginList)
	cmd.AddCommand(pluginInstall)
	cmd.AddCommand(pluginUninstall)
	cmd.PersistentFlags().StringVar(&pluginDir, "plugin-dir", filepath.Join(os.Getenv("HOME"), ".naeos", "plugins"), "plugin directory")
	return cmd
}
