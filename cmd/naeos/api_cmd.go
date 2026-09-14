// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/api"
)

var (
	apiPort   string
	apiAuth   bool
	apiSecret string
)

func newAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Start NAEOS REST API server",
		Long:  `Start the NAEOS REST API server for external integrations and web dashboard.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			auth := &api.AuthConfig{
				Enabled:   apiAuth,
				JWTSecret: apiSecret,
			}

			server := api.NewServer(":"+apiPort, auth)
			return server.Start()
		},
	}

	cmd.Flags().StringVarP(&apiPort, "port", "p", "8080", "API server port")
	cmd.Flags().BoolVarP(&apiAuth, "auth", "a", false, "Enable JWT authentication")
	cmd.Flags().StringVarP(&apiSecret, "secret", "s", "", "JWT secret key")

	return cmd
}
