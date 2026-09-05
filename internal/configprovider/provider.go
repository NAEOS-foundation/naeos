package configprovider

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Scheme identifies a config value source.
type Scheme string

const (
	// SchemeEnv reads values from environment variables.
	SchemeEnv Scheme = "env"
	// SchemeFile reads values from file contents.
	SchemeFile Scheme = "file"
	// SchemeSecret reads values from Kubernetes secrets.
	SchemeSecret Scheme = "secret"
	// SchemeVault reads values from HashiCorp Vault KV stores.
	SchemeVault Scheme = "vault"
)

// Reference is a parsed config value reference (e.g. "env:API_KEY").
type Reference struct {
	Scheme Scheme
	Value  string
}

// String renders the reference back to its text form.
func (r Reference) String() string {
	if r.Scheme == "" {
		return r.Value
	}
	return string(r.Scheme) + ":" + r.Value
}

// ParseReference parses a config reference string. Values whose prefix is not
// a known scheme are returned verbatim as plain values.
func ParseReference(s string) (Reference, error) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return Reference{Value: s}, nil
	}
	scheme := Scheme(s[:idx])
	switch scheme {
	case SchemeEnv, SchemeFile, SchemeSecret, SchemeVault:
		return Reference{Scheme: scheme, Value: s[idx+1:]}, nil
	default:
		return Reference{Value: s}, nil
	}
}

// Provider resolves config references for a specific source type.
type Provider interface {
	// Name returns the provider identifier.
	Name() string
	// Scheme returns the reference scheme this provider handles.
	Scheme() Scheme
	// Resolve returns the value for the given reference.
	Resolve(ref Reference) (string, error)
	// Connected reports whether the provider backend is reachable.
	Connected() bool
}

// EnvProvider resolves references from environment variables.
type EnvProvider struct {
	getenv func(string) string
}

// NewEnvProvider creates an environment variable provider.
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{getenv: os.Getenv}
}

// Name returns "env".
func (p *EnvProvider) Name() string { return "env" }

// Scheme returns env.
func (p *EnvProvider) Scheme() Scheme { return SchemeEnv }

// Resolve reads the named environment variable.
func (p *EnvProvider) Resolve(ref Reference) (string, error) {
	val := p.getenv(ref.Value)
	if val == "" {
		return "", naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("environment variable %q not set", ref.Value))
	}
	return val, nil
}

// Connected reports true (environment is always available).
func (p *EnvProvider) Connected() bool { return true }

// FileProvider resolves references from file contents.
type FileProvider struct {
	readfile func(string) ([]byte, error)
}

// NewFileProvider creates a file content provider.
func NewFileProvider() *FileProvider {
	return &FileProvider{readfile: os.ReadFile}
}

// Name returns "file".
func (p *FileProvider) Name() string { return "file" }

// Scheme returns file.
func (p *FileProvider) Scheme() Scheme { return SchemeFile }

// Resolve reads the named file and returns trimmed contents.
func (p *FileProvider) Resolve(ref Reference) (string, error) {
	data, err := p.readfile(ref.Value)
	if err != nil {
		return "", naeoserr.Wrapf(err, naeoserr.ErrNotFound, "read config file %q", ref.Value)
	}
	return strings.TrimSpace(string(data)), nil
}

// Connected reports true.
func (p *FileProvider) Connected() bool { return true }

// K8sSecretProvider resolves references from in-memory Kubernetes secrets.
// The reference format is "namespace/name/key".
type K8sSecretProvider struct {
	mu      sync.RWMutex
	secrets map[string]string
}

// NewK8sSecretProvider creates a Kubernetes secret provider.
func NewK8sSecretProvider() *K8sSecretProvider {
	return &K8sSecretProvider{secrets: make(map[string]string)}
}

// Name returns "secret".
func (p *K8sSecretProvider) Name() string { return "secret" }

// Scheme returns secret.
func (p *K8sSecretProvider) Scheme() Scheme { return SchemeSecret }

// AddSecret registers a secret value under "namespace/name/key".
func (p *K8sSecretProvider) AddSecret(namespace, name, key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.secrets[namespace+"/"+name+"/"+key] = value
}

// RemoveSecret deregisters a secret value.
func (p *K8sSecretProvider) RemoveSecret(namespace, name, key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.secrets, namespace+"/"+name+"/"+key)
}

// Resolve looks up the secret by its reference value.
func (p *K8sSecretProvider) Resolve(ref Reference) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(strings.Split(ref.Value, "/")) != 3 {
		return "", naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("invalid secret reference %q (want namespace/name/key)", ref.Value))
	}
	val, ok := p.secrets[ref.Value]
	if !ok {
		return "", naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("secret %q not found", ref.Value))
	}
	return val, nil
}

// Connected reports whether any secrets are registered.
func (p *K8sSecretProvider) Connected() bool {
	return len(p.secrets) > 0
}

// VaultProvider resolves references from an in-memory Vault KV store.
// The reference format is "path#key".
type VaultProvider struct {
	mu     sync.RWMutex
	secrets map[string]string
}

// NewVaultProvider creates a Vault KV provider.
func NewVaultProvider() *VaultProvider {
	return &VaultProvider{secrets: make(map[string]string)}
}

// Name returns "vault".
func (p *VaultProvider) Name() string { return "vault" }

// Scheme returns vault.
func (p *VaultProvider) Scheme() Scheme { return SchemeVault }

// AddSecret registers a secret at "path#key".
func (p *VaultProvider) AddSecret(path, key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.secrets[path+"#"+key] = value
}

// Resolve looks up the secret by its reference value.
func (p *VaultProvider) Resolve(ref Reference) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !strings.Contains(ref.Value, "#") {
		return "", naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("invalid vault reference %q (want path#key)", ref.Value))
	}
	val, ok := p.secrets[ref.Value]
	if !ok {
		return "", naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("vault secret %q not found", ref.Value))
	}
	return val, nil
}

// Connected reports whether any secrets are registered.
func (p *VaultProvider) Connected() bool {
	return len(p.secrets) > 0
}

// Chain resolves config values by trying providers in registration order.
type Chain struct {
	mu        sync.RWMutex
	providers []Provider
}

// NewChain creates a config resolution chain.
func NewChain(providers ...Provider) *Chain {
	return &Chain{providers: providers}
}

// Add appends a provider to the chain.
func (c *Chain) Add(p Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.providers = append(c.providers, p)
}

// Providers returns the registered provider names in order.
func (c *Chain) Providers() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, 0, len(c.providers))
	for _, p := range c.providers {
		names = append(names, p.Name())
	}
	return names
}

// Resolve resolves a config value for the given reference.
func (c *Chain) Resolve(ref Reference) (ResolveResult, error) {
	c.mu.RLock()
	providers := make([]Provider, len(c.providers))
	copy(providers, c.providers)
	c.mu.RUnlock()

	var lastErr error
	for _, p := range providers {
		if p.Scheme() != ref.Scheme {
			continue
		}
		val, err := p.Resolve(ref)
		if err == nil {
			return ResolveResult{
				Value:    val,
				Provider: p.Name(),
				Scheme:   ref.Scheme,
			}, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return ResolveResult{}, lastErr
	}
	return ResolveResult{}, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("no provider registered for scheme %q", ref.Scheme))
}

// Resolver resolves an entire config map containing references.
type Resolver struct {
	chain *Chain
}

// NewResolver creates a config map resolver backed by a chain.
func NewResolver(chain *Chain) *Resolver {
	return &Resolver{chain: chain}
}

// Config is a resolved configuration value.
type Config struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Provider string `json:"provider,omitempty"`
	Scheme   Scheme `json:"scheme,omitempty"`
}

// ResolveMap resolves every value in a config map. Values without a known
// scheme are passed through verbatim.
func (r *Resolver) ResolveMap(input map[string]string) ([]Config, error) {
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out []Config
	for _, k := range keys {
		cfg, err := r.ResolveKey(k, input[k])
		if err != nil {
			return out, err
		}
		out = append(out, cfg)
	}
	return out, nil
}

// ResolveKey resolves a single config key.
func (r *Resolver) ResolveKey(key, raw string) (Config, error) {
	ref, err := ParseReference(raw)
	if err != nil {
		return Config{}, err
	}
	// A plain value (no scheme) passes through.
	if ref.Scheme == "" {
		return Config{Key: key, Value: ref.Value}, nil
	}
	result, err := r.chain.Resolve(ref)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Key:      key,
		Value:    result.Value,
		Provider: result.Provider,
		Scheme:   result.Scheme,
	}, nil
}

// ResolveResult captures the outcome of resolving a reference.
type ResolveResult struct {
	Value    string
	Provider string
	Scheme   Scheme
}