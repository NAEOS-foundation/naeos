// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package infrastructure

type Provider string

const (
	ProviderAWS   Provider = "aws"
	ProviderGCP   Provider = "gcp"
	ProviderAzure Provider = "azure"
	ProviderLocal Provider = "local"
)

type Infrastructure struct {
	Provider    Provider          `json:"provider,omitempty"`
	Region      string            `json:"region,omitempty"`
	Project     string            `json:"project,omitempty"`
	Environment string            `json:"environment,omitempty"`
	Resources   []Resource        `json:"resources,omitempty"`
	Networking  []Network         `json:"networking,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type Resource struct {
	Name string            `json:"name"`
	Kind string            `json:"kind,omitempty"`
	Type string            `json:"type,omitempty"`
	Spec map[string]string `json:"spec,omitempty"`
}

type Network struct {
	Name  string `json:"name"`
	Kind  string `json:"kind,omitempty"`
	Ports []int  `json:"ports,omitempty"`
}
