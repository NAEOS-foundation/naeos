package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultTemplateRegistryURL = "https://naeos.dev/templates/registry.json"

type TemplateEntry struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	RepoURL     string   `json:"repo_url"`
	DownloadURL string   `json:"download_url"`
	Languages   []string `json:"languages"`
	Frameworks  []string `json:"frameworks"`
	Downloads   int      `json:"downloads"`
	UpdatedAt   string   `json:"updated_at"`
}

type TemplateList struct {
	Templates []TemplateEntry `json:"templates"`
}

type TemplateManifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Tags        []string `yaml:"tags"`
	Languages   []string `yaml:"languages"`
	Frameworks  []string `yaml:"frameworks"`
	License     string   `yaml:"license,omitempty"`
}

type RemoteTemplateRegistry struct {
	baseURL    string
	httpClient *http.Client
}

func NewRemoteTemplateRegistry(baseURL string) *RemoteTemplateRegistry {
	if baseURL == "" {
		baseURL = DefaultTemplateRegistryURL
	}
	return &RemoteTemplateRegistry{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *RemoteTemplateRegistry) List() ([]TemplateEntry, error) {
	var data []byte

	if strings.HasPrefix(r.baseURL, "file://") {
		path := strings.TrimPrefix(r.baseURL, "file://")
		var err error
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read local registry: %w", err)
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "GET", r.baseURL, nil)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		resp, err := r.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch template list: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
	}

	var list TemplateList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("decode template list: %w", err)
	}

	return list.Templates, nil
}

func (r *RemoteTemplateRegistry) Search(query string) ([]TemplateEntry, error) {
	templates, err := r.List()
	if err != nil {
		return nil, err
	}

	if query == "" {
		return templates, nil
	}

	query = strings.ToLower(query)
	var results []TemplateEntry
	for _, t := range templates {
		if strings.Contains(strings.ToLower(t.Name), query) ||
			strings.Contains(strings.ToLower(t.Description), query) {
			results = append(results, t)
			continue
		}
		for _, tag := range t.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, t)
				break
			}
		}
	}
	return results, nil
}

func (r *RemoteTemplateRegistry) Get(name string) (*TemplateEntry, error) {
	templates, err := r.List()
	if err != nil {
		return nil, err
	}

	for _, t := range templates {
		if t.Name == name {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("template %q not found", name)
}

func ValidateTemplateManifest(path string) (*TemplateManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template manifest: %w", err)
	}

	var m TemplateManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse template manifest: %w", err)
	}

	if m.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	if m.Version == "" {
		return nil, fmt.Errorf("template version is required")
	}
	if m.Description == "" {
		return nil, fmt.Errorf("template description is required")
	}

	return &m, nil
}

func PublishTemplate(templateDir string, registryURL string) (*TemplateEntry, error) {
	manifestPath := filepath.Join(templateDir, "template.yaml")
	altPath := filepath.Join(templateDir, "naeos.yaml")

	var manifestFile string
	if _, err := os.Stat(manifestPath); err == nil {
		manifestFile = manifestPath
	} else if _, err := os.Stat(altPath); err == nil {
		manifestFile = altPath
	} else {
		return nil, fmt.Errorf("no template.yaml or naeos.yaml found in %s", templateDir)
	}

	m, err := ValidateTemplateManifest(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	if _, err := os.Stat(filepath.Join(templateDir, "README.md")); os.IsNotExist(err) {
		return nil, fmt.Errorf("template must have a README.md")
	}

	if _, err := os.Stat(filepath.Join(templateDir, ".naeos")); os.IsNotExist(err) {
		entries, err := os.ReadDir(templateDir)
		if err != nil {
			return nil, fmt.Errorf("read template dir: %w", err)
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("template directory is empty")
		}
	}

	entry := &TemplateEntry{
		Name:        m.Name,
		Version:     m.Version,
		Description: m.Description,
		Author:      m.Author,
		Tags:        m.Tags,
		Languages:   m.Languages,
		Frameworks:  m.Frameworks,
		UpdatedAt:   time.Now().UTC().Format("2006-01-02"),
	}

	if registryURL != "" && registryURL != DefaultTemplateRegistryURL && !strings.HasPrefix(registryURL, "file://") {
		if err := submitToRegistry(*entry, registryURL); err != nil {
			return nil, fmt.Errorf("submit to registry: %w", err)
		}
	}

	return entry, nil
}

func submitToRegistry(entry TemplateEntry, registryURL string) error {
	postURL := strings.TrimSuffix(registryURL, "/registry.json") + "/api/publish"
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", postURL, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("publish to registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func GenerateRegistryEntry(templateDir string) (*TemplateEntry, error) {
	manifestPath := filepath.Join(templateDir, "template.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifestPath = filepath.Join(templateDir, "naeos.yaml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("no template.yaml found in %s", templateDir)
		}
	}

	m, err := ValidateTemplateManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	return &TemplateEntry{
		Name:        m.Name,
		Version:     m.Version,
		Description: m.Description,
		Author:      m.Author,
		Tags:        m.Tags,
		Languages:   m.Languages,
		Frameworks:  m.Frameworks,
		UpdatedAt:   time.Now().UTC().Format("2006-01-02"),
		RepoURL:     "",
		DownloadURL: "",
	}, nil
}

func DefaultTemplates() []TemplateEntry {
	return []TemplateEntry{
		{
			Name:        "microservices-go",
			Version:     "1.0.0",
			Description: "Go microservices starter project with gRPC, REST API, and event-driven communication",
			Author:      "NAEOS Foundation",
			Tags:        []string{"go", "microservices", "grpc", "rest", "official"},
			Languages:   []string{"go"},
			Frameworks:  []string{"grpc", "rest", "nats"},
			Downloads:   320,
			UpdatedAt:   "2026-06-01",
			RepoURL:     "https://github.com/naeos-templates/microservices-go",
		},
		{
			Name:        "serverless-ts",
			Version:     "1.1.0",
			Description: "TypeScript serverless project with AWS Lambda, API Gateway, and DynamoDB",
			Author:      "NAEOS Foundation",
			Tags:        []string{"typescript", "serverless", "aws", "official"},
			Languages:   []string{"typescript"},
			Frameworks:  []string{"aws-lambda", "api-gateway", "dynamodb"},
			Downloads:   210,
			UpdatedAt:   "2026-05-28",
			RepoURL:     "https://github.com/naeos-templates/serverless-ts",
		},
		{
			Name:        "web-api-py",
			Version:     "1.0.0",
			Description: "Python FastAPI REST API starter with SQLAlchemy, PostgreSQL, and Docker",
			Author:      "NAEOS Foundation",
			Tags:        []string{"python", "fastapi", "rest", "postgres", "official"},
			Languages:   []string{"python"},
			Frameworks:  []string{"fastapi", "sqlalchemy", "postgresql"},
			Downloads:   180,
			UpdatedAt:   "2026-06-10",
			RepoURL:     "https://github.com/naeos-templates/web-api-py",
		},
		{
			Name:        "rust-cli-tool",
			Version:     "0.9.0",
			Description: "Rust CLI application starter with clap, tracing, and cross-platform builds",
			Author:      "NAEOS Foundation",
			Tags:        []string{"rust", "cli", "official"},
			Languages:   []string{"rust"},
			Frameworks:  []string{"clap"},
			Downloads:   95,
			UpdatedAt:   "2026-04-20",
			RepoURL:     "https://github.com/naeos-templates/rust-cli-tool",
		},
		{
			Name:        "fullstack-js",
			Version:     "2.0.0",
			Description: "Full-stack JavaScript/TypeScript with Next.js, Prisma, and PostgreSQL",
			Author:      "NAEOS Foundation",
			Tags:        []string{"typescript", "javascript", "nextjs", "fullstack", "postgres", "official"},
			Languages:   []string{"typescript", "javascript"},
			Frameworks:  []string{"nextjs", "prisma", "postgresql"},
			Downloads:   450,
			UpdatedAt:   "2026-06-15",
			RepoURL:     "https://github.com/naeos-templates/fullstack-js",
		},
		{
			Name:        "event-driven-java",
			Version:     "1.0.0",
			Description: "Java event-driven microservices with Kafka, Spring Boot, and Kubernetes",
			Author:      "NAEOS Foundation",
			Tags:        []string{"java", "spring", "kafka", "kubernetes", "official"},
			Languages:   []string{"java"},
			Frameworks:  []string{"spring-boot", "kafka", "kubernetes"},
			Downloads:   140,
			UpdatedAt:   "2026-05-10",
			RepoURL:     "https://github.com/naeos-templates/event-driven-java",
		},
	}
}
