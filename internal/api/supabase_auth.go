// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

const defaultSupabaseAudience = "authenticated"

// SupabaseAuthConfig configures verification of Supabase access tokens.
type SupabaseAuthConfig struct {
	URL      string
	Audience string
	JWKSURL  string
}

// SupabaseAuthVerifier validates Supabase JWTs using the project's asymmetric JWKS.
type SupabaseAuthVerifier struct {
	issuer   string
	audience string
	jwks     keyfunc.Keyfunc
	cancel   context.CancelFunc
}

type supabaseClaims struct {
	jwt.RegisteredClaims
	Role        string                     `json:"role"`
	AppMetadata map[string]json.RawMessage `json:"app_metadata"`
}

// SupabaseUser contains the trusted identity and authorization claims used by the API.
type SupabaseUser struct {
	ID       string
	Roles    []string
	TenantID string
}

// NewSupabaseAuthVerifier creates a verifier for Supabase's asymmetric signing keys.
func NewSupabaseAuthVerifier(config SupabaseAuthConfig) (*SupabaseAuthVerifier, error) {
	return newSupabaseAuthVerifier(config, &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})
}

func newSupabaseAuthVerifier(config SupabaseAuthConfig, client *http.Client) (*SupabaseAuthVerifier, error) {
	issuer, err := supabaseIssuer(config.URL)
	if err != nil {
		return nil, err
	}
	audience := strings.TrimSpace(config.Audience)
	if audience == "" {
		audience = defaultSupabaseAudience
	}

	jwksURL := strings.TrimSpace(config.JWKSURL)
	if jwksURL == "" {
		jwksURL = issuer + "/.well-known/jwks.json"
	}
	if err := validateHTTPSURL(jwksURL); err != nil {
		return nil, fmt.Errorf("invalid Supabase JWKS URL: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	noErrorReturnFirstRequest := false
	jwks, err := keyfunc.NewDefaultOverrideCtx(ctx, []string{jwksURL}, keyfunc.Override{
		Client:                    client,
		HTTPTimeout:               5 * time.Second,
		NoErrorReturnFirstHTTPReq: &noErrorReturnFirstRequest,
		RefreshInterval:           5 * time.Minute,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("initialize Supabase JWKS verifier: %w", err)
	}
	return &SupabaseAuthVerifier{
		issuer:   issuer,
		audience: audience,
		jwks:     jwks,
		cancel:   cancel,
	}, nil
}

func supabaseIssuer(raw string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(raw), "/")
	if err := validateHTTPSURL(base); err != nil {
		return "", fmt.Errorf("invalid Supabase URL: %w", err)
	}
	parsed, _ := url.Parse(base)
	switch parsed.Path {
	case "":
		return base + "/auth/v1", nil
	case "/auth/v1":
		return base, nil
	default:
		return "", errors.New("Supabase URL path must be empty or /auth/v1")
	}
}

func validateHTTPSURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("must be an HTTPS URL without user info, query, or fragment")
	}
	return nil
}

// Validate verifies signature, issuer, audience, timing, and Supabase role claims.
func (v *SupabaseAuthVerifier) Validate(tokenString string) (*SupabaseUser, error) {
	if v == nil || v.jwks == nil {
		return nil, errors.New("Supabase authentication is not configured")
	}
	claims := &supabaseClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		v.jwks.Keyfunc,
		jwt.WithValidMethods([]string{"ES256", "RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf("validate Supabase access token: %w", err)
	}
	if !token.Valid || claims.Subject == "" || claims.IssuedAt == nil {
		return nil, errors.New("Supabase access token is missing required claims")
	}
	if claims.Role != "authenticated" {
		return nil, errors.New("Supabase token role is not authenticated")
	}

	roles, err := supabaseNAEOSRoles(claims.AppMetadata)
	if err != nil {
		return nil, err
	}
	tenantID, err := supabaseTenantID(claims.AppMetadata)
	if err != nil {
		return nil, err
	}
	return &SupabaseUser{ID: claims.Subject, Roles: roles, TenantID: tenantID}, nil
}

func supabaseNAEOSRoles(metadata map[string]json.RawMessage) ([]string, error) {
	var roles []string
	if raw, ok := metadata["naeos_roles"]; ok {
		if err := json.Unmarshal(raw, &roles); err != nil {
			return nil, errors.New("Supabase app_metadata.naeos_roles must be an array of strings")
		}
	}
	allowed := map[string]bool{"admin": true, "developer": true, "viewer": true}
	result := make([]string, 0, len(roles))
	seen := make(map[string]bool, len(roles))
	for _, role := range roles {
		if allowed[role] && !seen[role] {
			result = append(result, role)
			seen[role] = true
		}
	}
	if len(result) == 0 {
		result = []string{"viewer"}
	}
	return result, nil
}

func supabaseTenantID(metadata map[string]json.RawMessage) (string, error) {
	raw, ok := metadata["naeos_tenant_id"]
	if !ok {
		return "", nil
	}
	var tenantID string
	if err := json.Unmarshal(raw, &tenantID); err != nil {
		return "", errors.New("Supabase app_metadata.naeos_tenant_id must be a string")
	}
	return tenantID, nil
}

// Close stops the background JWKS refresh goroutine.
func (v *SupabaseAuthVerifier) Close() {
	if v != nil && v.cancel != nil {
		v.cancel()
	}
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

// SupabaseAuth validates Supabase access tokens (RS256 or ES256 asymmetric signing
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
