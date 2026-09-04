package serve

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// Listener describes a single network listener the daemon exposes. A listener
// may serve plain HTTP or TLS (HTTPS) depending on whether TLCert/TLSKey are set.
type Listener struct {
	// Addr is the listen address, e.g. ":8080" or "127.0.0.1:9090".
	Addr string `yaml:"addr" json:"addr"`
	// Name is an optional human-friendly label for observability.
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	// TLSCert is an optional path to a PEM certificate.
	TLSCert string `yaml:"tls_cert,omitempty" json:"tls_cert,omitempty"`
	// TLSKey is an optional path to a PEM private key.
	TLSKey string `yaml:"tls_key,omitempty" json:"tls_key,omitempty"`
	// API enables the NAEOS REST API on this listener.
	API bool `yaml:"api,omitempty" json:"api,omitempty"`
	// DashboardOnly restricts the listener to static dashboard assets when true.
	DashboardOnly bool `yaml:"dashboard_only,omitempty" json:"dashboard_only,omitempty"`
}

// IsTLS reports whether the listener should be served over HTTPS.
func (l Listener) IsTLS() bool {
	return l.TLSCert != "" || l.TLSKey != ""
}

// Auth holds authentication settings applied to API listeners.
type Auth struct {
	Enabled   bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	JWTSecret string `yaml:"jwt_secret,omitempty" json:"jwt_secret,omitempty"`
}

// Config is the server daemon configuration. It is parsed from a YAML file via
// LoadConfig and validated by Validate.
type Config struct {
	// Listeners are the network endpoints to open. When empty, a default
	// ":8080" API listener is used.
	Listeners []Listener `yaml:"listeners" json:"listeners"`
	// Auth applies to every API listener unless overridden per-listener.
	Auth Auth `yaml:"auth,omitempty" json:"auth,omitempty"`
	// ShutdownTimeout is the graceful shutdown grace period (e.g. "30s").
	ShutdownTimeout string `yaml:"shutdown_timeout,omitempty" json:"shutdown_timeout,omitempty"`
	// ReadTimeout / WriteTimeout / IdleTimeout tune the underlying HTTP server.
	ReadTimeout  string `yaml:"read_timeout,omitempty" json:"read_timeout,omitempty"`
	WriteTimeout string `yaml:"write_timeout,omitempty" json:"write_timeout,omitempty"`
	IdleTimeout  string `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
	// APIKeys maps a client API key to a requests-per-second limit.
	APIKeys map[string]int `yaml:"api_keys,omitempty" json:"api_keys,omitempty"`
	// LogLevel is "debug", "info", "warn" or "error".
	LogLevel string `yaml:"log_level,omitempty" json:"log_level,omitempty"`
}

// DefaultConfig returns a Config with sensible production defaults.
func DefaultConfig() *Config {
	return &Config{
		Listeners: []Listener{
			{Addr: ":8080", Name: "api", API: true},
		},
		ShutdownTimeout: "30s",
		ReadTimeout:     "15s",
		WriteTimeout:    "15s",
		IdleTimeout:     "60s",
		LogLevel:        "info",
	}
}

// LoadConfig reads and decodes a YAML server configuration file, applying
// defaults for any field not present, then validates the result.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, naeoserr.Wrap(naeoserr.ErrConfig, "read server config", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, naeoserr.Wrap(naeoserr.ErrConfig, "parse server config", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Validate checks the configuration for consistency and returns a descriptive
// error on the first problem found.
func (c *Config) Validate() error {
	if len(c.Listeners) == 0 {
		c.Listeners = DefaultConfig().Listeners
	}
	for _, l := range c.Listeners {
		if strings.TrimSpace(l.Addr) == "" {
			return naeoserr.New(naeoserr.ErrValidation, "server config: listener addr is required")
		}
		if l.IsTLS() && (l.TLSCert == "" || l.TLSKey == "") {
			return naeoserr.New(naeoserr.ErrValidation,
				fmt.Sprintf("server config: listener %q requires both tls_cert and tls_key", l.Addr))
		}
	}
	switch strings.ToLower(c.LogLevel) {
	case "", "debug", "info", "warn", "error":
	default:
		return naeoserr.New(naeoserr.ErrValidation,
			fmt.Sprintf("server config: invalid log_level %q", c.LogLevel))
	}
	return nil
}
