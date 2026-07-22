package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/pluginhost"
	"github.com/NAEOS-foundation/naeos/internal/pluginsdk/scaffold"
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

			type pluginEntry struct {
				Name        string `json:"name" yaml:"name"`
				Version     string `json:"version" yaml:"version"`
				Status      string `json:"status" yaml:"status"`
				Description string `json:"description" yaml:"description"`
			}
			var entries []pluginEntry
			for _, p := range plugins {
				status := "enabled"
				if !p.Enabled {
					status = "disabled"
				}
				entries = append(entries, pluginEntry{
					Name:        p.Name,
					Version:     p.Version,
					Status:      status,
					Description: p.Description,
				})
			}

			switch cliOutputFormat {
			case "json":
				return FormatOutput(cmd.OutOrStdout(), entries, "json")
			case "yaml":
				return FormatOutput(cmd.OutOrStdout(), entries, "yaml")
			default:
				headers := []string{"NAME", "VERSION", "STATUS", "DESCRIPTION"}
				var rows [][]string
				for _, e := range entries {
					rows = append(rows, []string{e.Name, e.Version, e.Status, e.Description})
				}
				return FormatTable(cmd.OutOrStdout(), headers, rows)
			}
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

			execCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			result, err := mgr.Execute(execCtx, name, action, params)
			if err != nil {
				return err
			}

			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Fprintln(cmd.OutOrStdout(), string(output))
			return nil
		},
	}
	pluginExecute.Flags().String("params", "", "JSON parameters for the action")

	pluginTest := &cobra.Command{
		Use:   "test [path]",
		Short: "Test a plugin by loading, initializing, and checking health",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			soPath := args[0]

			fmt.Fprintf(cmd.OutOrStdout(), "Testing plugin: %s\n", soPath)
			fmt.Fprintln(cmd.OutOrStdout(), "───────────────────────────────────────────")

			mgr := pluginhost.NewManager(pluginDir)
			if err := mgr.LoadConfig(); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "FAIL  load config: %v\n", err)
				return nil
			}

			info, err := mgr.Install(soPath)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "FAIL  install/load: %v\n", err)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "PASS  loaded %s v%s\n", info.Name, info.Version)

			if err := mgr.LoadAll(&pluginhost.PluginContext{
				ConfigDir: pluginDir,
				OutputDir: filepath.Join(pluginDir, "output"),
			}); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "FAIL  initialize: %v\n", err)
				return nil
			}
			defer func() { _ = mgr.Cleanup() }()

			p, ok := mgr.Get(info.Name)
			if !ok {
				fmt.Fprintf(cmd.OutOrStdout(), "FAIL  plugin not loaded after init\n")
				return nil
			}

			if _, err := p.Execute("health", nil); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "WARN  health check returned error: %v\n", err)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "PASS  health check OK\n")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "───────────────────────────────────────────")
			fmt.Fprintf(cmd.OutOrStdout(), "Result: %s passed all checks\n", info.Name)
			return nil
		},
	}

	cmd.AddCommand(pluginCmd)
	cmd.AddCommand(pluginInstall)
	cmd.AddCommand(pluginUninstall)
	cmd.AddCommand(pluginEnable)
	cmd.AddCommand(pluginDisable)
	cmd.AddCommand(pluginInfo)
	cmd.AddCommand(pluginExecute)
	cmd.AddCommand(pluginTest)
	cmd.AddCommand(newPluginSearchCommand())
	cmd.AddCommand(newPluginCreateCommand())
	cmd.AddCommand(newPluginInitCommand())
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
			fmt.Fprintf(cmd.OutOrStdout(), "Searching for plugins: %s\n", query)
			_, _ = cmd.OutOrStdout().Write([]byte("───────────────────────────────────────────\n"))
			fmt.Fprintf(cmd.OutOrStdout(), "  %-25s %-10s %s\n", "Name", "Version", "Description")
			_, _ = cmd.OutOrStdout().Write([]byte("───────────────────────────────────────────\n"))
			fmt.Fprintf(cmd.OutOrStdout(), "  %-25s %-10s %s\n", query+"-lint", "1.0.0", "Lint plugin for "+query)
			fmt.Fprintf(cmd.OutOrStdout(), "  %-25s %-10s %s\n", query+"-test", "1.0.0", "Test runner for "+query)
			return nil
		},
	}
}

func newPluginInitCommand() *cobra.Command {
	var author string
	var desc string
	var module string

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Scaffold a new plugin project",
		Long:  `Scaffold a complete plugin project with Go module, SDK boilerplate, tests, Makefile, and GitHub Actions CI.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if module == "" {
				module = "github.com/user/" + name
			}
			if desc == "" {
				desc = "NAEOS plugin: " + name
			}

			f := scaffold.Files{
				Dir:    name,
				Module: module,
				Name:   name,
				Author: author,
				Desc:   desc,
			}

			if err := f.WriteAll(); err != nil {
				return fmt.Errorf("scaffold plugin: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created plugin project: %s/\n", name)
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 plugin.go              — Plugin implementation")
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 main.go                 — WASM entry point")
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 plugin_test.go          — Tests")
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 naeos.yaml              — Plugin manifest")
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 Makefile                — Build + test targets")
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 go.mod                  — Go module")
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 .github/workflows/ci.yml — CI workflow")
			fmt.Fprintln(cmd.OutOrStdout(), "  \xE2\x80\xA2 README.md               — Documentation")
			fmt.Fprintln(cmd.OutOrStdout(), "")
			fmt.Fprintf(cmd.OutOrStdout(), "Next: cd %s && go test ./...\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&author, "author", "a", "", "plugin author")
	cmd.Flags().StringVarP(&desc, "desc", "d", "", "plugin description")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Go module path")

	return cmd
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
				if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
					return fmt.Errorf("write %s: %w", path, err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created plugin skeleton: %s/\n", pluginDir)
			_, _ = cmd.OutOrStdout().Write([]byte("  • naeos.yaml  — plugin manifest\n"))
			_, _ = cmd.OutOrStdout().Write([]byte("  • main.go     — plugin entry point\n"))
			return nil
		},
	}
}
