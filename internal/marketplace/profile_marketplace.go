package marketplace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

type ProfileEntry struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	Industry    string         `json:"industry"`
	Tags        []string       `json:"tags"`
	Downloads   int            `json:"downloads"`
	Content     map[string]any `json:"content"`
	Readme      string         `json:"readme,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ProfileMarketplace struct {
	cacheDir string
}

func NewProfileMarketplace(cacheDir string) *ProfileMarketplace {
	return &ProfileMarketplace{cacheDir: cacheDir}
}

func (m *ProfileMarketplace) Publish(entry ProfileEntry) error {
	entries, err := m.loadProfiles()
	if err != nil {
		entries = []ProfileEntry{}
	}

	entry.UpdatedAt = time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	for i, e := range entries {
		if e.Name == entry.Name {
			entries[i] = entry
			return m.saveProfiles(entries)
		}
	}

	entries = append(entries, entry)
	return m.saveProfiles(entries)
}

func (m *ProfileMarketplace) Get(name string) (*ProfileEntry, error) {
	entries, err := m.loadProfiles()
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Name == name {
			return &entry, nil
		}
	}

	return nil, fmt.Errorf("profile %s not found", name)
}

func (m *ProfileMarketplace) Search(query string, tags []string) ([]ProfileEntry, error) {
	entries, err := m.loadProfiles()
	if err != nil {
		return nil, err
	}

	var results []ProfileEntry
	for _, entry := range entries {
		if query != "" {
			matched := false
			if containsStr(entry.Name, query) || containsStr(entry.Description, query) || containsStr(entry.Industry, query) {
				matched = true
			}
			for _, tag := range entry.Tags {
				if containsStr(tag, query) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		if len(tags) > 0 {
			hasTag := false
			for _, filterTag := range tags {
				for _, entryTag := range entry.Tags {
					if entryTag == filterTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		results = append(results, entry)
	}

	return results, nil
}

func (m *ProfileMarketplace) List() ([]ProfileEntry, error) {
	return m.loadProfiles()
}

func (m *ProfileMarketplace) Download(name, targetDir string) error {
	entry, err := m.Get(name)
	if err != nil {
		return err
	}

	profileDir := filepath.Join(targetDir, ".naeos", "profiles")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	profileName := entry.Name
	if err := securityext.ValidatePluginName(profileName); err != nil {
		return fmt.Errorf("invalid profile name %q: %w", profileName, err)
	}
	profileFile := filepath.Join(profileDir, fmt.Sprintf("%s.json", profileName))
	return os.WriteFile(profileFile, data, 0o600)
}

func (m *ProfileMarketplace) Upload(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var entry ProfileEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}

	return m.Publish(entry)
}

func (m *ProfileMarketplace) Remove(name string) error {
	entries, err := m.loadProfiles()
	if err != nil {
		return err
	}

	for i, entry := range entries {
		if entry.Name == name {
			entries = append(entries[:i], entries[i+1:]...)
			return m.saveProfiles(entries)
		}
	}

	return fmt.Errorf("profile %s not found", name)
}

func (m *ProfileMarketplace) IncrementDownloads(name string) error {
	entries, err := m.loadProfiles()
	if err != nil {
		return err
	}

	for i, entry := range entries {
		if entry.Name == name {
			entries[i].Downloads++
			entries[i].UpdatedAt = time.Now()
			return m.saveProfiles(entries)
		}
	}

	return fmt.Errorf("profile %s not found", name)
}

func (m *ProfileMarketplace) loadProfiles() ([]ProfileEntry, error) {
	path := filepath.Join(m.cacheDir, "profiles.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m.defaultProfiles(), nil
		}
		return nil, err
	}

	var entries []ProfileEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (m *ProfileMarketplace) saveProfiles(entries []ProfileEntry) error {
	if err := os.MkdirAll(m.cacheDir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.cacheDir, "profiles.json"), data, 0o600)
}

func (m *ProfileMarketplace) defaultProfiles() []ProfileEntry {
	return []ProfileEntry{
		{
			Name:        "saas-starter",
			Version:     "1.0.0",
			Description: "SaaS application starter with multi-tenant support",
			Author:      "naeos",
			Industry:    "saas",
			Tags:        []string{"saas", "multi-tenant", "billing"},
			Downloads:   250,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "auth", "path": "./auth"},
					map[string]any{"name": "billing", "path": "./billing"},
					map[string]any{"name": "analytics", "path": "./analytics"},
				},
			},
		},
		{
			Name:        "fintech-core",
			Version:     "1.0.0",
			Description: "Financial technology core with compliance and audit",
			Author:      "naeos",
			Industry:    "fintech",
			Tags:        []string{"fintech", "compliance", "audit"},
			Downloads:   180,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "ledger", "path": "./ledger"},
					map[string]any{"name": "payment", "path": "./payment"},
					map[string]any{"name": "compliance", "path": "./compliance"},
				},
			},
		},
		{
			Name:        "healthcare-hipaa",
			Version:     "1.0.0",
			Description: "Healthcare system with HIPAA compliance",
			Author:      "naeos",
			Industry:    "healthcare",
			Tags:        []string{"healthcare", "hipaa", "phi"},
			Downloads:   120,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "patient", "path": "./patient"},
					map[string]any{"name": "appointment", "path": "./appointment"},
					map[string]any{"name": "records", "path": "./records"},
				},
			},
		},
		{
			Name:        "ai-agent-base",
			Version:     "1.0.0",
			Description: "AI agent with LLM integration and tool calling",
			Author:      "naeos",
			Industry:    "ai",
			Tags:        []string{"ai", "agent", "llm"},
			Downloads:   300,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "brain", "path": "./brain"},
					map[string]any{"name": "tools", "path": "./tools"},
					map[string]any{"name": "memory", "path": "./memory"},
				},
			},
		},
		{
			Name:        "edtech-platform",
			Version:     "1.0.0",
			Description: "Education technology platform with LMS, assessment engine, and learning analytics",
			Author:      "naeos",
			Industry:    "education",
			Tags:        []string{"education", "lms", "assessment"},
			Downloads:   85,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "lms", "path": "./lms"},
					map[string]any{"name": "assessment", "path": "./assessment"},
					map[string]any{"name": "analytics", "path": "./analytics"},
				},
			},
		},
		{
			Name:        "ecommerce-engine",
			Version:     "1.0.0",
			Description: "E-commerce engine with catalog management, shopping cart, payment processing, and order fulfillment",
			Author:      "naeos",
			Industry:    "retail",
			Tags:        []string{"ecommerce", "retail", "payment"},
			Downloads:   210,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "catalog", "path": "./catalog"},
					map[string]any{"name": "cart", "path": "./cart"},
					map[string]any{"name": "payment", "path": "./payment"},
					map[string]any{"name": "order", "path": "./order"},
				},
			},
		},
		{
			Name:        "iot-backend",
			Version:     "1.0.0",
			Description: "IoT backend with device management, telemetry ingestion, and alerting",
			Author:      "naeos",
			Industry:    "iot",
			Tags:        []string{"iot", "telemetry", "devices"},
			Downloads:   140,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "devices", "path": "./devices"},
					map[string]any{"name": "telemetry", "path": "./telemetry"},
					map[string]any{"name": "alerts", "path": "./alerts"},
				},
			},
		},
		{
			Name:        "media-streaming",
			Version:     "1.0.0",
			Description: "Media streaming platform with content ingest, transcoding pipeline, and CDN delivery",
			Author:      "naeos",
			Industry:    "media",
			Tags:        []string{"media", "streaming", "cdn"},
			Downloads:   95,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "ingest", "path": "./ingest"},
					map[string]any{"name": "transcode", "path": "./transcode"},
					map[string]any{"name": "cdn", "path": "./cdn"},
				},
			},
		},
		{
			Name:        "blockchain-node",
			Version:     "1.0.0",
			Description: "Blockchain node with consensus engine, wallet, and block explorer",
			Author:      "naeos",
			Industry:    "blockchain",
			Tags:        []string{"blockchain", "web3", "crypto"},
			Downloads:   160,
			Content: map[string]any{
				"modules": []any{
					map[string]any{"name": "consensus", "path": "./consensus"},
					map[string]any{"name": "wallet", "path": "./wallet"},
					map[string]any{"name": "explorer", "path": "./explorer"},
				},
			},
		},
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}
