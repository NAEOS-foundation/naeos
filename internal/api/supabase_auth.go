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
}
