package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/workspace"
)

func newWorkspaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage multi-module workspaces",
		Long: `Manage multi-module workspaces for NAEOS projects.

Example:
  naeos workspace init my-workspace
  naeos workspace add my-module ./modules/my-module
  naeos workspace list
  naeos workspace remove my-module`,
	}

	var rootDir string

	wsInit := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := workspace.NewManager(rootDir)
			ws, err := mgr.Init(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized workspace %s at %s\n", ws.Name, ws.Root)
			return nil
		},
	}

	wsAdd := &cobra.Command{
		Use:   "add [name] [path]",
		Short: "Add a module to the workspace",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := workspace.NewManager(rootDir)
			if err := mgr.AddModule(args[0], args[1], ""); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added module %s\n", args[0])
			return nil
		},
	}

	wsRemove := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a module from the workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := workspace.NewManager(rootDir)
			if err := mgr.RemoveModule(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed module %s\n", args[0])
			return nil
		},
	}

	wsList := &cobra.Command{
		Use:   "list",
		Short: "List workspace modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := workspace.NewManager(rootDir)
			modules, err := mgr.ListModules()
			if err != nil {
				return err
			}
			if len(modules) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No modules in workspace")
				return nil
			}
			for _, m := range modules {
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %s\n", m.Name, m.Path)
			}
			return nil
		},
	}

	cmd.AddCommand(wsInit)
	cmd.AddCommand(wsAdd)
	cmd.AddCommand(wsRemove)
	cmd.AddCommand(wsList)
	cmd.PersistentFlags().StringVar(&rootDir, "root", ".", "workspace root directory")
	return cmd
}
