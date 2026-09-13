// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package api

type Protocol string

const (
	ProtocolHTTP    Protocol = "http"
	ProtocolGRPC    Protocol = "grpc"
	ProtocolGraphQL Protocol = "graphql"
	ProtocolWS      Protocol = "websocket"
)

type API struct {
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	Protocol    Protocol          `json:"protocol,omitempty"`
	Description string            `json:"description,omitempty"`
	Endpoints   []APIEndpoint     `json:"endpoints,omitempty"`
	Schemas     []Schema          `json:"schemas,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type APIEndpoint struct {
	Method      string `json:"method,omitempty"`
	Path        string `json:"path,omitempty"`
	Summary     string `json:"summary,omitempty"`
	RequestRef  string `json:"request_ref,omitempty"`
	ResponseRef string `json:"response_ref,omitempty"`
}

type Schema struct {
	Name   string            `json:"name"`
	Fields map[string]string `json:"fields,omitempty"`
}
