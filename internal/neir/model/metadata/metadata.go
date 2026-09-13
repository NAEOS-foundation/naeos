// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package metadata

import "time"

type Metadata struct {
	NEIRVersion    string            `json:"neir_version,omitempty"`
	SchemaVersion  string            `json:"schema_version,omitempty"`
	ProjectVersion string            `json:"project_version,omitempty"`
	CreatedAt      *time.Time        `json:"created_at,omitempty"`
	ModifiedAt     *time.Time        `json:"modified_at,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Source         *SourceRef        `json:"source,omitempty"`
}

type SourceRef struct {
	Kind string `json:"kind,omitempty"`
	Ref  string `json:"ref,omitempty"`
}
