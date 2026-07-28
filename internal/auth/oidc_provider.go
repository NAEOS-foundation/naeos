package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type OIDCProvider struct {
	config     SSOConfig
	discovery  *OIDCDiscoveryDocument
	jwks       *JWKSResponse
	httpClient *http.Client
}

type OIDCDiscoveryDocument struct {
	Issuer                           string   `json:"issuer"`
	JWKSUri                          string   `json:"jwks_uri"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserinfoEndpoint                 string   `json:"userinfo_endpoint"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string   `json:"kty"`
	Alg string   `json:"alg,omitempty"`
	Kid string   `json:"kid,omitempty"`
	Use string   `json:"use,omitempty"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c,omitempty"`
}

type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

type OIDCUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name,omitempty"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Picture       string `json:"picture,omitempty"`
}

func NewOIDCProvider(cfg SSOConfig) *OIDCProvider {
	return &OIDCProvider{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *OIDCProvider) Type() ProviderType { return ProviderOIDC }

func (p *OIDCProvider) Name() string { return p.config.Name }

func (p *OIDCProvider) Config() SSOConfig { return p.config }

func (p *OIDCProvider) Validate() error {
	if p.config.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.config.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	if p.config.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if p.config.ClientSecret == "" {
		return fmt.Errorf("client_secret is required")
	}
	return nil
}

func (p *OIDCProvider) Discover() error {
	doc, err := fetchOIDCDiscovery(p.config.Issuer, p.httpClient)
	if err != nil {
		return fmt.Errorf("discovery: %w", err)
	}
	p.discovery = doc

	jwks, err := fetchJWKS(doc.JWKSUri, p.httpClient)
	if err != nil {
		return fmt.Errorf("jwks: %w", err)
	}
	p.jwks = jwks

	return nil
}

func (p *OIDCProvider) GetAuthorizationURL(state string) string {
	return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		p.discovery.AuthorizationEndpoint,
		p.config.ClientID,
		p.config.RedirectURL,
		strings.Join(p.config.Scopes, " "),
		state,
	)
}

func (p *OIDCProvider) ExchangeCode(code string) (*OAuth2Token, error) {
	if p.discovery == nil {
		if err := p.Discover(); err != nil {
			return nil, err
		}
	}

	body := fmt.Sprintf(
		"grant_type=authorization_code&code=%s&redirect_uri=%s&client_id=%s&client_secret=%s",
		code, p.config.RedirectURL, p.config.ClientID, p.config.ClientSecret,
	)

	req, err := http.NewRequestWithContext(context.Background(), "POST", p.discovery.TokenEndpoint, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp OIDCTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	token := &OAuth2Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	if tokenResp.IDToken != "" {
		claims, err := p.verifyIDToken(tokenResp.IDToken)
		if err == nil {
			token.IDToken = tokenResp.IDToken
			token.UserID = claims.Sub
		}
	}

	return token, nil
}

func (p *OIDCProvider) GetUser(token *OAuth2Token) (*OAuth2User, error) {
	if p.discovery == nil || p.discovery.UserinfoEndpoint == "" {
		return p.extractUserFromIDToken(token)
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", p.discovery.UserinfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request: %w", err)
	}
	defer resp.Body.Close()

	var ui OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&ui); err != nil {
		return nil, fmt.Errorf("decode userinfo: %w", err)
	}

	return &OAuth2User{
		ID:    ui.Sub,
		Email: ui.Email,
		Name:  ui.Name,
	}, nil
}

func (p *OIDCProvider) extractUserFromIDToken(token *OAuth2Token) (*OAuth2User, error) {
	if token.IDToken == "" {
		return nil, fmt.Errorf("no id_token available")
	}

	parts := strings.Split(token.IDToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode id_token payload: %w", err)
	}

	var claims struct {
		Sub   string `json:"sub"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parse id_token claims: %w", err)
	}

	return &OAuth2User{
		ID:    claims.Sub,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}

func (p *OIDCProvider) verifyIDToken(idToken string) (*OIDCUserInfo, error) {
	if p.jwks == nil {
		if err := p.Discover(); err != nil {
			return nil, err
		}
	}

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token")
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}

	payload := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(payload))

	var valid bool
	for _, key := range p.jwks.Keys {
		if key.Kty == "RSA" {
			pub, err := jwkToPublicKey(key)
			if err != nil {
				continue
			}
			if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], sigBytes); err == nil {
				valid = true
				break
			}
		}
	}

	if !valid {
		return nil, fmt.Errorf("id_token signature verification failed")
	}

	var ui OIDCUserInfo
	payloadBytes, _ := base64.RawURLEncoding.DecodeString(parts[1])
	_ = json.Unmarshal(payloadBytes, &ui)

	return &ui, nil
}

func fetchOIDCDiscovery(issuer string, client *http.Client) (*OIDCDiscoveryDocument, error) {
	issuer = strings.TrimRight(issuer, "/")
	req, _ := http.NewRequestWithContext(context.Background(), "GET", issuer+"/.well-known/openid-configuration", nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch discovery: %w", err)
	}
	defer resp.Body.Close()

	var doc OIDCDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode discovery: %w", err)
	}
	return &doc, nil
}

func fetchJWKS(uri string, client *http.Client) (*JWKSResponse, error) {
	req, _ := http.NewRequestWithContext(context.Background(), "GET", uri, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}
	return &jwks, nil
}

func jwkToPublicKey(key JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}

	pub := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}

	return pub, nil
}

func ParseCertificatePEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found")
	}
	return x509.ParseCertificate(block.Bytes)
}
