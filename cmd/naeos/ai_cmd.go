package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/ai"
)

func newAICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "AI-powered assistance commands",
		Long: `AI-powered commands for specification improvement and concept explanation.

Example:
  naeos ai suggest --input-file spec.yaml
  naeos ai explain pipeline`,
	}

	svc := ai.NewService()

	aiSuggest := &cobra.Command{
		Use:   "suggest",
		Short: "Get AI suggestions for improving a specification",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile, _ := cmd.Flags().GetString("input-file")
			if inputFile == "" {
				return fmt.Errorf("missing required --input-file")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return err
			}

			suggestions, err := svc.Suggest(string(data))
			if err != nil {
				return err
			}

			for _, s := range suggestions {
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s (%s priority)\n  %s\n\n",
					s.Category, s.Title, s.Priority, s.Description)
			}
			return nil
		},
	}

	aiExplain := &cobra.Command{
		Use:   "explain [topic]",
		Short: "Explain a NAEOS concept",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exp, err := svc.Explain(args[0], "")
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n", exp.Content)
			for _, d := range exp.Details {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", d)
			}
			return nil
		},
	}

	aiSuggest.Flags().String("input-file", "", "path to specification file")

	cmd.AddCommand(aiSuggest)
	cmd.AddCommand(aiExplain)
	return cmd
}
