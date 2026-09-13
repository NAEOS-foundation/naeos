// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package docs

type DocKind string

const (
	KindGuide     DocKind = "guide"
	KindReference DocKind = "reference"
	KindADR       DocKind = "adr"
	KindRFC       DocKind = "rfc"
	KindChangelog DocKind = "changelog"
)

type Documentation struct {
	Guides     []Doc             `json:"guides,omitempty"`
	References []Doc             `json:"references,omitempty"`
	ADRs       []Doc             `json:"adrs,omitempty"`
	RFCs       []Doc             `json:"rfcs,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Doc struct {
	Title   string  `json:"title"`
	Path    string  `json:"path,omitempty"`
	Kind    DocKind `json:"kind,omitempty"`
	Summary string  `json:"summary,omitempty"`
}
