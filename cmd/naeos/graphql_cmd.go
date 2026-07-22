package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/graphql"
	"github.com/NAEOS-foundation/naeos/internal/version"
)

var (
	graphqlPort string
)

func newGraphQLCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graphql",
		Short: "Start GraphQL API server",
		Long:  `Start GraphQL API server for flexible querying.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			schema := &graphql.Schema{
				Types: map[string]*graphql.TypeDef{
					"Health": {
						Name: "Health",
						Fields: map[string]*graphql.FieldDef{
							"status":  {Name: "status", Type: "String"},
							"version": {Name: "version", Type: "String"},
						},
					},
				},
				Queries: &graphql.OperationDef{
					Fields: map[string]*graphql.FieldDef{
						"health": {
							Name: "health",
							Type: "String",
							Resolve: func(ctx *graphql.Context, args map[string]any) (any, error) {
								return map[string]string{
									"status":  "healthy",
									"version": version.String(),
								}, nil
							},
						},
						"version": {
							Name: "version",
							Type: "String",
							Resolve: func(ctx *graphql.Context, args map[string]any) (any, error) {
								return version.String(), nil
							},
						},
					},
				},
			}

			handler := graphql.Handler(schema)
			http.Handle("/graphql", handler)

			fmt.Printf("GraphQL server starting on http://localhost%s/graphql\n", graphqlPort)
			srv := &http.Server{
				Addr:              graphqlPort,
				ReadHeaderTimeout: 10 * time.Second,
				ReadTimeout:       15 * time.Second,
				WriteTimeout:      15 * time.Second,
				IdleTimeout:       60 * time.Second,
			}
			return srv.ListenAndServe()
		},
	}

	cmd.Flags().StringVarP(&graphqlPort, "port", "p", ":8082", "GraphQL server port")

	return cmd
}
