// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

var configFile = "config.json"

// configDirOverride is set by SetConfigDir for testing.
var configDirOverride string

func SetConfigDir(dir string) {
	configDirOverride = dir
}

func configDir() string {
	if configDirOverride != "" {
		return configDirOverride
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".naeos/supabase"
	}
	return filepath.Join(home, ".naeos/supabase")
}

type Config struct {
	ProjectRef     string `json:"project_ref"`
	URL            string `json:"url"`
	AnonKey        string `json:"anon_key"`
	ServiceRoleKey string `json:"service_role_key"`
	JWKSURL        string `json:"jwks_url"`
	DBHost         string `json:"db_host"`
	DBPort         int    `json:"db_port"`
	DBPassword     string `json:"db_password"`
	ManagementURL  string `json:"management_url"`
	AccessToken    string `json:"access_token"`
}

type Client struct {
	config     *Config
	http       *http.Client
	authToken  string
	mu         sync.RWMutex
	sessionMu  sync.Mutex
	session    *Session
}

func DefaultConfigPath() string {
	return configDir()
}

func configFilePath() string {
	return filepath.Join(configDir(), configFile)
}

func SaveConfig(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrConfig, "create config dir")
	}
	data, err := json.MarshalIndent(cfg, "", "  ") //nolint:gosec // intentional storage of session token
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrParse, "marshal config")
	}
	return os.WriteFile(configFilePath(), data, 0o600)
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(configFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, naeoserr.New(naeoserr.ErrConfig, "supabase not configured; run 'naeos supabase init'")
		}
		return nil, naeoserr.Wrapf(err, naeoserr.ErrConfig, "read config")
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "parse config")
	}
	return &cfg, nil
}

func NewClient(cfg *Config) *Client {
	c := &Client{
		config: cfg,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	return c
}

func (c *Client) Config() *Config {
	return c.config
}

func (c *Client) SetAuthToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authToken = token
}

// SetSession stores the current Supabase user session in memory.
func (c *Client) SetSession(session *Session) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if session == nil {
		c.session = nil
		c.SetAuthToken("")
		return
	}
	copy := *session
	c.session = &copy
	c.SetAuthToken(copy.AccessToken)
}

// AuthSession returns a copy of the current Supabase user session.
func (c *Client) AuthSession() *Session {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.session == nil {
		return nil
	}
	copy := *c.session
	return &copy
}

func (c *Client) refreshSessionIfNeeded() error {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.session == nil || c.session.RefreshToken == "" ||
		(!c.session.ExpiresAt.IsZero() && time.Until(c.session.ExpiresAt) > time.Minute) {
		return nil
	}
	body, err := c.do(http.MethodPost, "/auth/v1/token?grant_type=refresh_token",
		map[string]string{"apikey": c.config.AnonKey},
		map[string]string{"refresh_token": c.session.RefreshToken})
	if err != nil {
		return fmt.Errorf("refresh Supabase session: %w", err)
	}
	session, err := jsonUnmarshal[Session](body)
	if err != nil {
		return err
	}
	session.ExpiresAt = time.Now().Add(time.Duration(session.ExpiresIn) * time.Second)
	c.session = session
	c.SetAuthToken(session.AccessToken)
	if err := SaveSession(session); err != nil {
		return fmt.Errorf("save refreshed Supabase session: %w", err)
	}
	return nil
}

func (c *Client) AuthToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authToken
}

func (c *Client) userAccessToken() string {
	if token := c.AuthToken(); token != "" {
		return token
	}
	return c.config.AnonKey
}

func (c *Client) do(method, path string, headers map[string]string, body any) ([]byte, error) {
	url := c.config.URL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "marshal request")
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, reqBody)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if token := c.AuthToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "request")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "read response")
	}
	if resp.StatusCode >= 400 {
		return nil, naeoserr.New(naeoserr.ErrCloud, fmt.Sprintf("API error %d: %s", resp.StatusCode, string(data)))
	}

	return data, nil
}

func (c *Client) doAuth(method, path string, body any) ([]byte, error) {
	if path != "/auth/v1/token?grant_type=refresh_token" {
		if err := c.refreshSessionIfNeeded(); err != nil {
			return nil, err
		}
	}
	headers := map[string]string{
		"apikey":        c.config.AnonKey,
		"Authorization": "Bearer " + c.userAccessToken(),
	}
	return c.do(method, path, headers, body)
}

func (c *Client) doAdmin(method, path string, body any) ([]byte, error) {
	headers := map[string]string{
		"apikey":        c.config.ServiceRoleKey,
		"Authorization": "Bearer " + c.config.ServiceRoleKey,
	}
	return c.do(method, path, headers, body)
}

func (c *Client) doManagement(method, path string, headers map[string]string, body any) ([]byte, error) {
	baseURL := c.managementURL()
	fullURL := baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "marshal request")
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, fullURL, reqBody)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.config.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "request")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrNetwork, "read response")
	}
	if resp.StatusCode >= 400 {
		return nil, naeoserr.New(naeoserr.ErrCloud, fmt.Sprintf("API error %d: %s", resp.StatusCode, string(data)))
	}

	return data, nil
}

func jsonUnmarshal[T any](data []byte) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrParse, "decode response")
	}
	return &result, nil
}

func (c *Client) managementURL() string {
	if c.config.ManagementURL != "" {
		return c.config.ManagementURL
	}
	return "https://api.supabase.com"
}

func MaskKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:4] + "..." + key[len(key)-4:]
}
