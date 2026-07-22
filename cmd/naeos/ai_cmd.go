package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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
  naeos ai explain pipeline
  naeos ai enrich --input-file spec.yaml --stream --provider anthropic`,
	}

	svc := ai.NewService()

	llmKey := os.Getenv("NAEOS_LLM_API_KEY")
	if llmKey != "" {
		provider := ai.ProviderOpenAI
		if p := os.Getenv("NAEOS_LLM_PROVIDER"); p != "" {
			provider = ai.LLMProvider(p)
		}
		llm := ai.NewLLMService(ai.LLMConfig{
			Provider: provider,
			APIKey:   llmKey,
		})
		svc = ai.NewServiceWithLLM(llm)
	}

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

	aiEnrich := &cobra.Command{
		Use:   "enrich",
		Short: "Enrich a specification with AI-powered best practices",
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile, _ := cmd.Flags().GetString("input-file")
			if inputFile == "" {
				return fmt.Errorf("missing required --input-file")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return err
			}

			stream, _ := cmd.Flags().GetBool("stream")
			provider, _ := cmd.Flags().GetString("provider")

			apiKey := os.Getenv("NAEOS_LLM_API_KEY")
			if provider == "" {
				provider = os.Getenv("NAEOS_LLM_PROVIDER")
			}
			if provider == "" {
				provider = string(ai.ProviderOpenAI)
			}

			llm := ai.NewLLMService(ai.LLMConfig{
				Provider: ai.LLMProvider(provider),
				APIKey:   apiKey,
			})

			if stream {
				return enrichStream(context.Background(), llm, string(data), os.Stdout)
			}

			enriched, err := llm.EnrichSpec(string(data))
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), enriched)
			return nil
		},
	}

	aiEnrich.Flags().String("input-file", "", "path to specification file")
	aiEnrich.Flags().Bool("stream", false, "stream output in real-time as SSE events")
	aiEnrich.Flags().String("provider", "", "LLM provider (openai, anthropic, ollama)")

	aiCompile := &cobra.Command{
		Use:   "compile",
		Short: "Compile a specification for a target AI agent using AI",
		Long: `Compile a NAEOS specification into configuration files for a target AI agent
using AI-powered generation.

Example:
  naeos ai compile --input-file spec.yaml --target claude
  naeos ai compile --input-file spec.yaml --target opencode --provider anthropic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile, _ := cmd.Flags().GetString("input-file")
			if inputFile == "" {
				return fmt.Errorf("missing required --input-file")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return err
			}

			target, _ := cmd.Flags().GetString("target")
			providerStr, _ := cmd.Flags().GetString("provider")

			apiKey := os.Getenv("NAEOS_LLM_API_KEY")
			if providerStr == "" {
				providerStr = os.Getenv("NAEOS_LLM_PROVIDER")
			}
			if providerStr == "" {
				providerStr = string(ai.ProviderOpenAI)
			}

			llm := ai.NewLLMService(ai.LLMConfig{
				Provider: ai.LLMProvider(providerStr),
				APIKey:   apiKey,
			})

			pr, pw := io.Pipe()
			defer pr.Close()

			go func() {
				err := llm.StreamCompileSpec(context.Background(), target, string(data), pw)
				pw.CloseWithError(err)
			}()

			decoder := ai.NewSSEDecoder(pr)
			var output strings.Builder
			for {
				event, data, err := decoder.Decode()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					return err
				}

				switch event {
				case "chunk":
					var chunk struct {
						Text string `json:"text"`
					}
					if err := json.Unmarshal(data, &chunk); err == nil {
						output.WriteString(chunk.Text)
					}
				case "error":
					var errChunk struct {
						Message string `json:"message"`
					}
					if err := json.Unmarshal(data, &errChunk); err == nil {
						return fmt.Errorf("LLM error: %s", errChunk.Message)
					}
					return fmt.Errorf("LLM error: %s", string(data))
				case "done":
					fmt.Fprintln(cmd.OutOrStdout(), output.String())
					return nil
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), output.String())
			return nil
		},
	}

	aiCompile.Flags().String("input-file", "", "path to specification file")
	aiCompile.Flags().String("target", "opencode", "target AI agent (claude, copilot, cursor, gemini, codex, opencode, windsurf)")
	aiCompile.Flags().String("provider", "", "LLM provider (openai, anthropic, ollama)")

	cmd.AddCommand(aiSuggest)
	cmd.AddCommand(aiExplain)
	cmd.AddCommand(aiEnrich)
	cmd.AddCommand(aiCompile)
	return cmd
}

func enrichStream(ctx context.Context, llm *ai.LLMService, spec string, w io.Writer) error {
	pr, pw := io.Pipe()
	defer pr.Close()

	go func() {
		err := llm.StreamEnrichSpec(ctx, spec, pw)
		pw.CloseWithError(err)
	}()

	decoder := ai.NewSSEDecoder(pr)
	for {
		event, data, err := decoder.Decode()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		switch event {
		case "chunk":
			var chunk struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal(data, &chunk); err == nil {
				fmt.Fprint(w, chunk.Text)
			}
		case "error":
			var errChunk struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(data, &errChunk); err == nil {
				return fmt.Errorf("LLM error: %s", errChunk.Message)
			}
			return fmt.Errorf("LLM error: %s", string(data))
		case "done":
			fmt.Fprintln(w)
			return nil
		}
	}
}
