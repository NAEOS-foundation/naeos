package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/search"
)

func searchDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".naeos", "search")
}

func openSearchIndex(name string) (*search.Persistent, error) {
	dir := filepath.Join(searchDir(), name)
	return search.NewPersistent(dir)
}

func newSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Full-text search engine management",
		Long: `Manage search indexes, query documents, and perform full-text search.

Example:
  naeos search index --id doc1 --title "Hello World" --content "This is a test"
  naeos search query --term "hello"
  naeos search count
  naeos search delete --id doc1
  naeos search list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newSearchIndexCommand())
	cmd.AddCommand(newSearchQueryCommand())
	cmd.AddCommand(newSearchCountCommand())
	cmd.AddCommand(newSearchDeleteCommand())
	cmd.AddCommand(newSearchListCommand())

	return cmd
}

func newSearchIndexCommand() *cobra.Command {
	var name, id, title, content string
	var tags []string

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index a document",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openSearchIndex(name)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}

			doc := &search.Document{
				ID:      id,
				Index:   name,
				Title:   title,
				Content: content,
				Tags:    tags,
			}

			if err := eng.Index(doc); err != nil {
				return fmt.Errorf("failed to index: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Indexed document '%s' in '%s'\n", id, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "default", "search index name")
	cmd.Flags().StringVar(&id, "id", "", "document ID (required)")
	cmd.Flags().StringVar(&title, "title", "", "document title")
	cmd.Flags().StringVar(&content, "content", "", "document content")
	cmd.Flags().StringArrayVar(&tags, "tag", nil, "document tags")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newSearchQueryCommand() *cobra.Command {
	var name, term string
	var limit int
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "query",
		Short: "Search for documents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openSearchIndex(name)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}

			q := &search.Query{
				Text:  term,
				Limit: limit,
			}

			result, err := eng.Search(q)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			type searchHit struct {
				ID      string   `json:"id"`
				Title   string   `json:"title"`
				Content string   `json:"content"`
				Tags    []string `json:"tags,omitempty"`
			}

			type searchOutput struct {
				Index string     `json:"index"`
				Term  string     `json:"term"`
				Total int        `json:"total"`
				Hits  []searchHit `json:"hits"`
			}

			var hits []searchHit
			for _, hit := range result.Hits {
				hits = append(hits, searchHit{
					ID:      hit.Document.ID,
					Title:   hit.Document.Title,
					Content: hit.Document.Content,
					Tags:    hit.Document.Tags,
				})
			}

			output := searchOutput{
				Index: name,
				Term:  term,
				Total: result.Total,
				Hits:  hits,
			}

			if outputFormat == "json" {
				data, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal search result: %w", err)
				}
				cmd.OutOrStdout().Write(data)
				cmd.OutOrStdout().Write([]byte("\n"))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Found %d results:\n\n", result.Total)
			for i, hit := range result.Hits {
				fmt.Fprintf(out, "%d. [%s] %s\n", i+1, hit.Document.ID, hit.Document.Title)
				fmt.Fprintf(out, "   %s\n\n", hit.Document.Content)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "default", "search index name")
	cmd.Flags().StringVar(&term, "term", "", "search term (required)")
	cmd.Flags().IntVar(&limit, "limit", 10, "max results")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json")
	cmd.MarkFlagRequired("term")
	return cmd
}

func newSearchCountCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "count",
		Short: "Count documents in index",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openSearchIndex(name)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Documents in '%s': %d\n", name, eng.Count())
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "default", "search index name")
	return cmd
}

func newSearchDeleteCommand() *cobra.Command {
	var name, id string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a document from index",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, err := openSearchIndex(name)
			if err != nil {
				return fmt.Errorf("failed to open index: %w", err)
			}

			if err := eng.Delete(id); err != nil {
				return fmt.Errorf("failed to delete: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted document '%s' from '%s'\n", id, name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "default", "search index name")
	cmd.Flags().StringVar(&id, "id", "", "document ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newSearchListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all search indexes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := searchDir()
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Search indexes: none\n")
				return nil
			}
			var names []string
			for _, e := range entries {
				if e.IsDir() {
					names = append(names, e.Name())
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Search indexes: %s\n", strings.Join(names, ", "))
			return nil
		},
	}
}
