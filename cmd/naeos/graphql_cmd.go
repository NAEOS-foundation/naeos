package main

import (
	"fmt"
	"net/http"
	"github.com/NAEOS-foundation/naeos/internal/graphql"
	"github.com/NAEOS-foundation/naeos/internal/version"
	"github.com/spf13/cobra"
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
							Resolve: func(ctx *graphql.Context, args map[string]interface{}) (interface{}, error) {
								return map[string]string{
									"status":  "healthy",
									"version": version.String(),
								}, nil
							},
						},
						"version": {
							Name: "version",
							Type: "String",
							Resolve: func(ctx *graphql.Context, args map[string]interface{}) (interface{}, error) {
								return version.String(), nil
							},
						},
					},
				},
			}

			handler := graphql.Handler(schema)
			http.Handle("/graphql", handler)

			fmt.Printf("GraphQL server starting on http://localhost%s/graphql\n", graphqlPort)
			return http.ListenAndServe(graphqlPort, nil)
		},
	}

	cmd.Flags().StringVarP(&graphqlPort, "port", "p", ":8082", "GraphQL server port")

	return cmd
}
