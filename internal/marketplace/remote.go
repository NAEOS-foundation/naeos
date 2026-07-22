package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	DefaultRegistryURL  = "https://naeos.dev"
	DefaultRegistryPath = "/api/v1/plugins"
)

type RemoteRegistry struct {
	baseURL    string
	apiPath    string
	installDir string
	httpClient *http.Client
}

func NewRemoteRegistry(baseURL, installDir string) *RemoteRegistry {
	if baseURL == "" {
		baseURL = DefaultRegistryURL
	}
	return &RemoteRegistry{
		baseURL:    baseURL,
		apiPath:    DefaultRegistryPath,
		installDir: installDir,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type RemotePlugin struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	Platform    string   `json:"platform"`
	DownloadURL string   `json:"download_url"`
	SHA256      string   `json:"sha256"`
	Size        int64    `json:"size"`
	UpdatedAt   string   `json:"updated_at"`
}

type RemotePluginList struct {
	Plugins []RemotePlugin `json:"plugins"`
}

type RemoteSearchFilter struct {
	Query    string
	Author   string
	Platform string
	Tags     []string
}

func (r *RemoteRegistry) List() ([]RemotePlugin, error) {
	url := r.baseURL + r.apiPath
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch plugin list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var list RemotePluginList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode plugin list: %w", err)
	}

	return list.Plugins, nil
}

func (r *RemoteRegistry) Search(query string) ([]RemotePlugin, error) {
	return r.SearchFilter(RemoteSearchFilter{Query: query})
}

func (r *RemoteRegistry) SearchFilter(filter RemoteSearchFilter) ([]RemotePlugin, error) {
	plugins, err := r.List()
	if err != nil {
		return nil, err
	}

	var results []RemotePlugin
	for _, p := range plugins {
		if filter.Query != "" {
			if !containsStr(p.Name, filter.Query) && !containsStr(p.Description, filter.Query) {
				matchedTag := false
				for _, tag := range p.Tags {
					if containsStr(tag, filter.Query) {
						matchedTag = true
						break
					}
				}
				if !matchedTag {
					continue
				}
			}
		}
		if filter.Author != "" && !containsStr(p.Author, filter.Author) {
			continue
		}
		if filter.Platform != "" && p.Platform != filter.Platform {
			continue
		}
		if len(filter.Tags) > 0 {
			hasTag := false
			for _, ft := range filter.Tags {
				for _, pt := range p.Tags {
					if pt == ft {
						hasTag = true
						break
					}
				}
			}
			if !hasTag {
				continue
			}
		}
		results = append(results, p)
	}
	return results, nil
}

func (r *RemoteRegistry) Install(name, version string) (string, error) {
	plugin, err := r.resolvePlugin(name, version)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(r.installDir, 0o755); err != nil {
		return "", fmt.Errorf("create install dir: %w", err)
	}

	destPath := filepath.Join(r.installDir, name+".so")
	if err := r.downloadFile(plugin.DownloadURL, destPath); err != nil {
		return "", fmt.Errorf("download plugin: %w", err)
	}

	if plugin.SHA256 != "" {
		if err := VerifyPlugin(destPath, plugin.SHA256); err != nil {
			os.Remove(destPath)
			return "", fmt.Errorf("verify plugin checksum: %w", err)
		}
	}

	metaPath := filepath.Join(r.installDir, name+".meta.json")
	meta := map[string]any{
		"name":         plugin.Name,
		"version":      plugin.Version,
		"description":  plugin.Description,
		"author":       plugin.Author,
		"checksum":     plugin.SHA256,
		"installed_at": time.Now().Format(time.RFC3339),
	}
	metaData, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(metaPath, metaData, 0o600)

	return destPath, nil
}

func (r *RemoteRegistry) Uninstall(name string) error {
	soPath := filepath.Join(r.installDir, name+".so")
	metaPath := filepath.Join(r.installDir, name+".meta.json")

	os.Remove(soPath)
	os.Remove(metaPath)
	return nil
}

func (r *RemoteRegistry) Installed() ([]map[string]any, error) {
	entries, err := os.ReadDir(r.installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Install directory does not exist — no plugins installed
		}
		return nil, err
	}

	var plugins []map[string]any
	for _, entry := range entries {
		if entry.Name() == "" || entry.IsDir() {
			continue
		}
		if len(entry.Name()) > len(".meta.json") && entry.Name()[len(entry.Name())-len(".meta.json"):] == ".meta.json" {
			data, err := os.ReadFile(filepath.Join(r.installDir, entry.Name()))
			if err != nil {
				continue
			}
			var meta map[string]any
			if json.Unmarshal(data, &meta) == nil {
				plugins = append(plugins, meta)
			}
		}
	}
	return plugins, nil
}

func (r *RemoteRegistry) resolvePlugin(name, version string) (*RemotePlugin, error) {
	plugins, err := r.Search(name)
	if err != nil {
		return nil, err
	}

	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	for _, p := range plugins {
		if p.Name == name {
			if version != "" && p.Version != version {
				continue
			}
			if p.Platform != "" && p.Platform != platform {
				continue
			}
			return &p, nil
		}
	}

	if version != "" {
		return nil, fmt.Errorf("plugin %s@%s not found", name, version)
	}
	return nil, fmt.Errorf("plugin %s not found", name)
}

func (r *RemoteRegistry) downloadFile(url, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
