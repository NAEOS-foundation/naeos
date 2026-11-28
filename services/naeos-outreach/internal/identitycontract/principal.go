// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

// Package identitycontract defines the identity contract for the trusted NAEOS
// Outreach adapter boundary. It does not authenticate credentials or authorize.
package identitycontract

import (
	"errors"
	"strings"
)

const ContractVersion = "v1"

var ErrInvalidPrincipal = errors.New("invalid trusted principal")

// Principal is issuer-qualified. Authorization attributes are intentionally absent.
type Principal struct {
	Issuer  string `json:"issuer"`
	Subject string `json:"subject"`
}

// Validate requires non-empty exact issuer and subject strings. The values are
// not normalized because normalization could collapse distinct identities.
func (p Principal) Validate() error {
	if strings.TrimSpace(p.Issuer) == "" || strings.TrimSpace(p.Subject) == "" {
		return ErrInvalidPrincipal
	}
	if strings.TrimSpace(p.Issuer) != p.Issuer || strings.TrimSpace(p.Subject) != p.Subject {
		return ErrInvalidPrincipal
	}
	return nil
}

// Adapter is a future trusted identity adapter contract. Implementors must
// verify credentials and derive claims only from verified claims/trusted lookup.
// It is intentionally not wired into HTTP handlers in Phase 2A.
type Adapter interface {
	ResolvePrincipal(credential string) (Principal, error)
}
