// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package security

type Security struct {
	Authentication *Authentication   `json:"authentication,omitempty"`
	Authorization  *Authorization    `json:"authorization,omitempty"`
	Encryption     *Encryption       `json:"encryption,omitempty"`
	Secrets        []Secret          `json:"secrets,omitempty"`
	Attributes     map[string]string `json:"attributes,omitempty"`
}

type Authentication struct {
	Method   string `json:"method,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type Authorization struct {
	Model string   `json:"model,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

type Encryption struct {
	InTransit bool   `json:"in_transit,omitempty"`
	AtRest    bool   `json:"at_rest,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
}

type Secret struct {
	Name string `json:"name"`
	Kind string `json:"kind,omitempty"`
}
