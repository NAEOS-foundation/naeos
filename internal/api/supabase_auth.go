// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/NAEOS-foundation/naeos/internal/auth"
)

// Supabase role claim values for access tokens.
const (
	supabaseRoleAuthenticated = "authenticated"
	supabaseRoleAnon          = "anon"
)

// supabaseAppMetadataRoleKey is the app_metadata key operators set in Supabase
// to grant NAEOS RBAC roles to a user. app_metadata is only writable through
// the Supabase admin API (service role), never by end users, so it is the only
// metadata source trusted for role mapping.
const supabaseAppMetadataRoleKey = "naeos_roles"

// defaultSupabaseRBACRole is granted when a Supabase user presents no usable
// NAEOS role mapping.
const defaultSupabaseRBACRole = "viewer"

// SupabaseClaims carries the identity information NAEOS reads from a
// Supabase access token payload.
type SupabaseClaims struct {
	Sub   string
	Email string
	Name  string
	Role  string
	// AppRoles holds the raw app_metadata.naeos_roles values before they are
	// filtered against the NAEOS RBAC role registry.
	AppRoles []string
}

// SupabaseAuth validates Supabase access tokens (RS256, asymmetric signing
// keys) against the project JWKS endpoint and exposes the identity claims
// needed for RBAC provisioning.
type SupabaseAuth struct {
	validator *JWKSValidator
}

// NewSupabaseAuth builds a Supabase token validator. The issuer must be the
// Supabase auth issuer (for example https://<ref>.supabase.co/auth/v1), the
// audience the expected token aud value ("authenticated" by default), and the
// JWKS endpoint the project's /.well-known/jwks.json URL.
func NewSupabaseAuth(issuer, audience, jwksURL string) (*SupabaseAuth, error) {
	validator, err := NewJWKSValidator(issuer, audience, jwksURL)
	if err != nil {
		return nil, err
	}
	return &SupabaseAuth{validator: validator}, nil
}

// Validate verifies signature and claims of a Supabase access token and
// returns both the generic JWT claims and the Supabase identity claims.
// Anonymous (anon) tokens are rejected.
func (s *SupabaseAuth) Validate(ctx context.Context, token string) (*JWTClaims, *SupabaseClaims, error) {
	baseClaims, err := s.validator.Validate(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	claims, err := ParseSupabaseClaims(token)
	if err != nil {
		return nil, nil, err
	}
	if claims.Role != supabaseRoleAuthenticated {
		return nil, nil, fmt.Errorf("supabase token role %q is not authenticated", claims.Role)
	}
	if claims.Sub == "" || claims.Sub != baseClaims.Sub {
		return nil, nil, fmt.Errorf("supabase token subject mismatch")
	}
	return baseClaims, claims, nil
}

// ParseSupabaseClaims decodes the payload segment of a JWT and extracts the
// Supabase identity claims. Signature verification is performed separately by
// SupabaseAuth.Validate.
func ParseSupabaseClaims(token string) (*SupabaseClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed supabase token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("malformed supabase token claims: %w", err)
	}
	var raw struct {
		Sub         string         `json:"sub"`
		Email       string         `json:"email"`
		Role        string         `json:"role"`
		UserMeta    map[string]any `json:"user_metadata"`
		AppMetadata struct {
			NAEOSRoles []any  `json:"naeos_roles"`
			Raw        string `json:"-"`
		} `json:"app_metadata"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("malformed supabase token claims: %w", err)
	}
	claims := &SupabaseClaims{
		Sub:   raw.Sub,
		Email: raw.Email,
		Role:  raw.Role,
	}
	if name, ok := raw.UserMeta["name"].(string); ok {
		claims.Name = name
	}
	if claims.Name == "" && claims.Email != "" {
		claims.Name = claims.Email
	}
	roles := make([]string, 0, len(raw.AppMetadata.NAEOSRoles))
	for _, value := range raw.AppMetadata.NAEOSRoles {
		if role, ok := value.(string); ok {
			if trimmed := strings.TrimSpace(role); trimmed != "" {
				roles = append(roles, trimmed)
			}
		}
	}
	claims.AppRoles = roles
	return claims, nil
}

// ensureSupabaseUser provisions the Supabase identity in the RBAC manager on
// first use and refreshes the granted roles on every validated token so that
// app_metadata.naeos_roles changes take effect without a server restart.
func (s *Server) ensureSupabaseUser(claims *SupabaseClaims) {
	if s.authManager == nil || claims == nil || claims.Sub == "" {
		return
	}
	roles := MapSupabaseRoles(claims, s.authManager.RBAC())
	user := &auth.User{
		ID:    claims.Sub,
		Email: claims.Email,
		Name:  claims.Name,
		Roles: roles,
	}
	s.authManager.CreateUser(user)
}

// MapSupabaseRoles filters raw app_metadata.naeos_roles values against the
// RBAC role registry so only known roles are granted, falling back to the
// default role for authenticated users with no usable mapping.
func MapSupabaseRoles(claims *SupabaseClaims, rbac *auth.RBAC) []string {
	if rbac == nil {
		return []string{defaultSupabaseRBACRole}
	}
	known := make(map[string]bool)
	for _, name := range rbac.ListRoles() {
		known[name] = true
	}
	roles := make([]string, 0, len(claims.AppRoles))
	seen := make(map[string]bool)
	for _, role := range claims.AppRoles {
		if !known[role] || seen[role] {
			continue
		}
		seen[role] = true
		roles = append(roles, role)
	}
	if len(roles) == 0 && known[defaultSupabaseRBACRole] {
		roles = append(roles, defaultSupabaseRBACRole)
	}
	return roles
}
