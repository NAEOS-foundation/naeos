package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
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
}

type Client struct {
	config  *Config
	http    *http.Client
	authToken string
	mu       sync.RWMutex
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
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(configFilePath(), data, 0o600)
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(configFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("supabase not configured; run 'naeos supabase init'")
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

func NewClient(cfg *Config) *Client {
	return &Client{
		config: cfg,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Config() *Config {
	return c.config
}

func (c *Client) SetAuthToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authToken = token
}

func (c *Client) AuthToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authToken
}

func (c *Client) do(method, path string, headers map[string]string, body any) (*http.Response, error) {
	url := c.config.URL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
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
		return nil, fmt.Errorf("request: %w", err)
	}

	return resp, nil
}

func (c *Client) doAuth(method, path string, body any) (*http.Response, error) {
	headers := map[string]string{
		"apikey": c.config.AnonKey,
	}
	return c.do(method, path, headers, body)
}

func (c *Client) doAdmin(method, path string, body any) (*http.Response, error) {
	headers := map[string]string{
		"apikey": c.config.ServiceRoleKey,
	}
	return c.do(method, path, headers, body)
}

func decodeResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

func MaskKey(key string) string {
	if len(key) <= 8 {
		return key
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func decodeRaw(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
