package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/profiles"
	"github.com/spf13/cobra"
)

func newProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage industry-specific project profiles",
		Long:  `List, search, and apply industry-specific project profiles.`,
	}

	cmd.AddCommand(newProfileListCommand())
	cmd.AddCommand(newProfileShowCommand())
	cmd.AddCommand(newProfileSearchCommand())
	cmd.AddCommand(newProfileApplyCommand())
	cmd.AddCommand(newProfileCompareCommand())
	cmd.AddCommand(newProfileCategoriesCommand())
	return cmd
}

func newProfileCompareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare [profile-a] [profile-b]",
		Short: "Compare two profiles side by side",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := profiles.NewRegistry()
			a, okA := reg.Get(args[0])
			b, okB := reg.Get(args[1])
			if !okA {
				return fmt.Errorf("profile %q not found", args[0])
			}
			if !okB {
				return fmt.Errorf("profile %q not found", args[1])
			}

			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Comparing: %s vs %s\n", a.Name, b.Name)))
			cmd.OutOrStdout().Write([]byte(strings.Repeat("─", 50) + "\n"))
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-20s %-15s %-15s\n", "Field", a.Name, b.Name)))
			cmd.OutOrStdout().Write([]byte(strings.Repeat("─", 50) + "\n"))
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-20s %-15s %-15s\n", "Industry", a.Industry, b.Industry)))
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-20s %-15s %-15s\n", "Version", a.Version, b.Version)))
			return nil
		},
	}
	return cmd
}

func newProfileCategoriesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "categories",
		Short: "List all profile categories",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := profiles.NewRegistry()
			all := reg.List()
			cats := make(map[string]int)
			for _, p := range all {
				cats[p.Industry]++
			}
			cmd.OutOrStdout().Write([]byte("Profile Categories:\n"))
			cmd.OutOrStdout().Write([]byte(strings.Repeat("─", 30) + "\n"))
			for cat, count := range cats {
				cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %-20s %d profiles\n", cat, count)))
			}
			return nil
		},
	}
	return cmd
}

func newProfileListCommand() *cobra.Command {
	var industry string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := profiles.NewRegistry()

			var list []profiles.Profile
			if industry != "" {
				list = reg.ByIndustry(industry)
			} else {
				list = reg.List()
			}

			if len(list) == 0 {
				fmt.Println("No profiles found.")
				return nil
			}

			fmt.Printf("%-20s %-35s %-20s %s\n", "ID", "NAME", "INDUSTRY", "DESCRIPTION")
			fmt.Println(strings.Repeat("-", 100))
			for _, p := range list {
				desc := p.Description
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				fmt.Printf("%-20s %-35s %-20s %s\n", p.ID, p.Name, p.Industry, desc)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&industry, "industry", "", "filter by industry")
	return cmd
}

func newProfileShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show [profile-id]",
		Short: "Show profile details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := profiles.NewRegistry()
			p, ok := reg.Get(args[0])
			if !ok {
				return fmt.Errorf("profile %q not found", args[0])
			}

			fmt.Printf("Profile: %s\n", p.Name)
			fmt.Printf("ID: %s\n", p.ID)
			fmt.Printf("Industry: %s\n", p.Industry)
			fmt.Printf("Description: %s\n\n", p.Description)

			fmt.Printf("Architecture: %s\n", p.Architecture.Pattern)
			if len(p.Architecture.Principles) > 0 {
				fmt.Println("Principles:")
				for _, pr := range p.Architecture.Principles {
					fmt.Printf("  - %s\n", pr)
				}
			}

			fmt.Printf("\nModules (%d):\n", len(p.Modules))
			for _, m := range p.Modules {
				fmt.Printf("  - %s: %s\n", m.Name, m.Description)
			}

			fmt.Printf("\nServices (%d):\n", len(p.Services))
			for _, s := range p.Services {
				fmt.Printf("  - %s (%s, port %d)\n", s.Name, s.Kind, s.Port)
			}

			fmt.Printf("\nSecurity: auth=%s, authz=%s\n", p.Security.Authentication, p.Security.Authorization)
			fmt.Printf("Deployment: %s\n", p.Deployment.Strategy)
			fmt.Printf("Testing: %s (coverage: %s)\n", p.Testing.Strategy, p.Testing.Coverage)
			return nil
		},
	}
}

func newProfileSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search profiles by name or description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := profiles.NewRegistry()
			results := reg.Search(args[0])

			if len(results) == 0 {
				fmt.Println("No profiles match your search.")
				return nil
			}

			fmt.Printf("%-20s %-35s %-20s\n", "ID", "NAME", "INDUSTRY")
			fmt.Println(strings.Repeat("-", 75))
			for _, p := range results {
				fmt.Printf("%-20s %-35s %-20s\n", p.ID, p.Name, p.Industry)
			}
			return nil
		},
	}
}

func newProfileApplyCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "apply [profile-id]",
		Short: "Generate a spec file from a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg := profiles.NewRegistry()
			p, ok := reg.Get(args[0])
			if !ok {
				return fmt.Errorf("profile %q not found", args[0])
			}

			spec := reg.ToSpecYAML(p)
			if err := os.WriteFile(output, []byte(spec), 0o644); err != nil {
				return fmt.Errorf("write spec: %w", err)
			}

			fmt.Printf("Profile %q applied to %s\n", p.Name, output)
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "naeos.spec.yaml", "output file path")
	return cmd
}
