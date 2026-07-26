package auth

import (
	"fmt"
	"sync"
)

type ProviderType string

const (
	ProviderOIDC ProviderType = "oidc"
	ProviderSAML ProviderType = "saml"
	ProviderLDAP ProviderType = "ldap"
)

type SSOConfig struct {
	Name         string            `json:"name"`
	Type         ProviderType      `json:"type"`
	Issuer       string            `json:"issuer,omitempty"`
	ClientID     string            `json:"client_id,omitempty"`
	ClientSecret string            `json:"-"` // never serialized
	RedirectURL  string            `json:"redirect_url,omitempty"`
	Scopes       []string          `json:"scopes,omitempty"`
	MetadataURL  string            `json:"metadata_url,omitempty"`
	CertFile     string            `json:"cert_file,omitempty"`
	EntityID     string            `json:"entity_id,omitempty"`
	SSOURL       string            `json:"sso_url,omitempty"`
	Host         string            `json:"host,omitempty"`
	Port         int               `json:"port,omitempty"`
	BaseDN       string            `json:"base_dn,omitempty"`
	BindDN       string            `json:"bind_dn,omitempty"`
	BindPassword string            `json:"-"` // never serialized
	UserFilter   string            `json:"user_filter,omitempty"`
	AttrMap      map[string]string `json:"attr_map,omitempty"`
	Enabled      bool              `json:"enabled"`
}

type SSOProvider interface {
	Type() ProviderType
	Name() string
	Config() SSOConfig
	Validate() error
}

type SSORegistry struct {
	providers map[string]SSOProvider
	mu        sync.RWMutex
}

func NewSSORegistry() *SSORegistry {
	return &SSORegistry{
		providers: make(map[string]SSOProvider),
	}
}

func (r *SSORegistry) Register(p SSOProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := p.Validate(); err != nil {
		return fmt.Errorf("validate %s provider %q: %w", p.Type(), p.Name(), err)
	}

	r.providers[p.Name()] = p
	return nil
}

func (r *SSORegistry) Get(name string) (SSOProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	return p, ok
}

func (r *SSORegistry) List() []SSOProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]SSOProvider, 0, len(r.providers))
	for _, p := range r.providers {
		out = append(out, p)
	}
	return out
}

func (r *SSORegistry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
}
