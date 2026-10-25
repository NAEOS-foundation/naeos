// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

// Package publicadapter is the single conversion boundary from the internal
// NEIR runtime model to the public, versioned NEIR DTO.
package publicadapter

import (
	"encoding/json"
	"fmt"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
	neirv1 "github.com/NAEOS-foundation/naeos/pkg/neir/v1"
)

// ToPublic converts an internal NEIR model through its canonical JSON
// representation so the public DTO remains decoupled from internal Go types.
func ToPublic(source *model.NEIR) (*neirv1.Document, error) {
	if source == nil {
		return nil, nil
	}
	data, err := json.Marshal(source)
	if err != nil {
		return nil, fmt.Errorf("marshal internal NEIR model: %w", err)
	}
	document, err := neirv1.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("convert internal NEIR model to public v1 DTO: %w", err)
	}
	return document, nil
}
