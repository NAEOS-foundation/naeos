// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package storage

type StorageType string

const (
	TypeSQL   StorageType = "sql"
	TypeNoSQL StorageType = "nosql"
	TypeFile  StorageType = "file"
	TypeCache StorageType = "cache"
	TypeQueue StorageType = "queue"
	TypeBlob  StorageType = "blob"
)

type Storage struct {
	Name        string            `json:"name"`
	Type        StorageType       `json:"type,omitempty"`
	Provider    string            `json:"provider,omitempty"`
	Connection  string            `json:"connection,omitempty"`
	Collections []Collection      `json:"collections,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type Collection struct {
	Name   string            `json:"name"`
	Schema map[string]string `json:"schema,omitempty"`
}
