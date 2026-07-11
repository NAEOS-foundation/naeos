package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/templates"
)

func newTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage generation templates",
		Long: `Manage NAEOS generation templates.

Example:
  naeos template list
  naeos template add my-template "template content"
  naeos template remove my-template`,
	}

	var templatesDir string

	templateList := &cobra.Command{
		Use:   "list",
		Short: "List available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := templates.NewManager(templatesDir)
			tmpls, err := mgr.List()
			if err != nil {
				return err
			}
			for _, t := range tmpls {
				custom := ""
				if t.IsCustom {
					custom = " (custom)"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %s%s\n", t.Name, t.Path, custom)
			}
			return nil
		},
	}

	templateAdd := &cobra.Command{
		Use:   "add [name] [content]",
		Short: "Add a custom template",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := templates.NewManager(templatesDir)
			if err := mgr.AddCustom(args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added template %s\n", args[0])
			return nil
		},
	}

	templateRemove := &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a custom template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := templates.NewManager(templatesDir)
			if err := mgr.RemoveCustom(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed template %s\n", args[0])
			return nil
		},
	}

	cmd.AddCommand(templateList)
	cmd.AddCommand(templateAdd)
	cmd.AddCommand(templateRemove)
	cmd.PersistentFlags().StringVar(&templatesDir, "templates-dir", filepath.Join(".", ".naeos", "templates"), "templates directory")
	return cmd
}
