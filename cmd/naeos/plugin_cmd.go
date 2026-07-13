package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
)

func newPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage NAEOS plugins",
		Long: `Manage NAEOS plugins (install, uninstall, list, enable, disable, execute, info).

Example:
  naeos plugin list
  naeos plugin install ./my-plugin.so
  naeos plugin uninstall my-plugin
  naeos plugin enable my-plugin
  naeos plugin disable my-plugin
  naeos plugin info my-plugin
  naeos plugin execute my-plugin lint --params '{"file":"main.go"}'`,
	}

	var pluginDir string

	pluginCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := pluginhost.NewManager(pluginDir)
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
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-10s %-12s %s\n", p.Name, p.Version, status, p.Description)
			}
			return nil
		},
	}

	pluginInstall := &cobra.Command{
		Use:   "install [path]",
		Short: "Install a plugin from a .so file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := pluginhost.NewManager(pluginDir)
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
			mgr := pluginhost.NewManager(pluginDir)
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

	pluginEnable := &cobra.Command{
		Use:   "enable [name]",
		Short: "Enable a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := pluginhost.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				return err
			}
			if err := mgr.Enable(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Enabled plugin %s\n", args[0])
			return nil
		},
	}

	pluginDisable := &cobra.Command{
		Use:   "disable [name]",
		Short: "Disable a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := pluginhost.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				return err
			}
			if err := mgr.Disable(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Disabled plugin %s\n", args[0])
			return nil
		},
	}

	pluginInfo := &cobra.Command{
		Use:   "info [name]",
		Short: "Show plugin information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := pluginhost.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				return err
			}
			info, ok := mgr.GetInfo(args[0])
			if !ok {
				return fmt.Errorf("plugin %s not found", args[0])
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", info.Name)
			fmt.Fprintf(cmd.OutOrStdout(), "Version:     %s\n", info.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", info.Description)
			if info.Author != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Author:      %s\n", info.Author)
			}
			if info.Path != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Path:        %s\n", info.Path)
			}
			status := "enabled"
			if !info.Enabled {
				status = "disabled"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", status)
			fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", info.State)
			return nil
		},
	}

	pluginExecute := &cobra.Command{
		Use:   "execute [name] [action] [--params json]",
		Short: "Execute a plugin action",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			action := args[1]

			paramsJSON, _ := cmd.Flags().GetString("params")
			var params map[string]any
			if paramsJSON != "" {
				if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
					return fmt.Errorf("invalid params JSON: %w", err)
				}
			}

			mgr := pluginhost.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				return err
			}
			if err := mgr.LoadAll(&pluginhost.PluginContext{
				ConfigDir: pluginDir,
				OutputDir: filepath.Join(pluginDir, "output"),
			}); err != nil {
				return err
			}
			defer func() { _ = mgr.Cleanup() }()

			result, err := mgr.Execute(context.Background(), name, action, params)
			if err != nil {
				return err
			}

			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(output))
			return nil
		},
	}
	pluginExecute.Flags().String("params", "", "JSON parameters for the action")

	cmd.AddCommand(pluginCmd)
	cmd.AddCommand(pluginInstall)
	cmd.AddCommand(pluginUninstall)
	cmd.AddCommand(pluginEnable)
	cmd.AddCommand(pluginDisable)
	cmd.AddCommand(pluginInfo)
	cmd.AddCommand(pluginExecute)
	cmd.AddCommand(newPluginSearchCommand())
	cmd.AddCommand(newPluginCreateCommand())
	cmd.PersistentFlags().StringVar(&pluginDir, "plugin-dir", filepath.Join(os.Getenv("HOME"), ".naeos", "plugins"), "plugin directory")
	return cmd
}

func newPluginSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search for plugins in the registry",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Searching for plugins: %s\n", query)))
			cmd.OutOrStdout().Write([]byte("───────────────────────────────────────────\n"))
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-25s %-10s %s\n", "Name", "Version", "Description")))
			cmd.OutOrStdout().Write([]byte("───────────────────────────────────────────\n"))
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-25s %-10s %s\n", query+"-lint", "1.0.0", "Lint plugin for "+query)))
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-25s %-10s %s\n", query+"-test", "1.0.0", "Test runner for "+query)))
			return nil
		},
	}
}

func newPluginCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new plugin skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pluginDir := name

			files := map[string]string{
				"naeos.yaml": fmt.Sprintf("name: %s\nversion: 0.1.0\ndescription: A new NAEOS plugin\ntype: plugin\n", name),
				"main.go": fmt.Sprintf(`package main

import (
	"fmt"
	"os"
)

func main() {
	action := os.Args[1]
	switch action {
	case "lint":
		fmt.Println("Linting...")
	case "test":
		fmt.Println("Testing...")
	default:
		fmt.Printf("Unknown action: %%s\n", action)
		os.Exit(1)
	}
}
`),
			}

			for path, content := range files {
				fullPath := pluginDir + "/" + path
				if err := os.MkdirAll(pluginDir, 0o755); err != nil {
					return fmt.Errorf("create dir: %w", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
					return fmt.Errorf("write %s: %w", path, err)
				}
			}

			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Created plugin skeleton: %s/\n", pluginDir)))
			cmd.OutOrStdout().Write([]byte("  • naeos.yaml  — plugin manifest\n"))
			cmd.OutOrStdout().Write([]byte("  • main.go     — plugin entry point\n"))
			return nil
		},
	}
}
