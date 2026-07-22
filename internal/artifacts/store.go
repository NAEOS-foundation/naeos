package artifacts

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

// ArtifactKind identifies the type of a build artifact.
type ArtifactKind string

const (
	// KindCode represents source code artifacts.
	KindCode ArtifactKind = "code"
	// KindConfig represents configuration file artifacts.
	KindConfig ArtifactKind = "config"
	// KindDocs represents documentation artifacts.
	KindDocs ArtifactKind = "docs"
	// KindDocker represents Dockerfile and compose artifacts.
	KindDocker ArtifactKind = "docker"
	// KindCI represents CI/CD pipeline artifacts.
	KindCI ArtifactKind = "ci"
	// KindAI represents AI agent configuration artifacts.
	KindAI ArtifactKind = "ai"
	// KindTest represents test file artifacts.
	KindTest ArtifactKind = "test"
	// KindMigration represents database migration artifacts.
	KindMigration ArtifactKind = "migration"
	// KindProfile represents profile artifacts.
	KindProfile ArtifactKind = "profile"
	// KindOther represents uncategorized artifacts.
	KindOther ArtifactKind = "other"
)

// Artifact represents a single build artifact with metadata and content.
type Artifact struct {
	ID          string            `json:"id"`
	Path        string            `json:"path"`
	Content     []byte            `json:"-"`
	ContentHash string            `json:"content_hash"`
	Kind        ArtifactKind      `json:"kind"`
	Language    string            `json:"language,omitempty"`
	NEIRVersion string            `json:"neir_version,omitempty"`
	Source      string            `json:"source,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	ModifiedAt  time.Time         `json:"modified_at"`
	Size        int64             `json:"size"`
}

// StoreManifest is the on-disk manifest that tracks all artifacts in a store.
type StoreManifest struct {
	Version   string     `json:"version"`
	Project   string     `json:"project"`
	Artifacts []Artifact `json:"artifacts"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Store manages a collection of artifacts with deduplication and persistence.
type Store struct {
	root     string
	manifest StoreManifest
	byPath   map[string]*Artifact
	byKind   map[ArtifactKind][]*Artifact
	byHash   map[string]*Artifact
}

// NewStore creates a new artifact store rooted at the given directory.
func NewStore(root string) *Store {
	return &Store{
		root:   root,
		byPath: make(map[string]*Artifact),
		byKind: make(map[ArtifactKind][]*Artifact),
		byHash: make(map[string]*Artifact),
		manifest: StoreManifest{
			Version:   "1.0",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// Add inserts an artifact into the store, deduplicating by content hash.
func (s *Store) Add(path string, content []byte, kind ArtifactKind, opts ...Option) (*Artifact, error) {
	if _, err := securityext.ValidateFilePath(filepath.Join(s.root, path), s.root); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	hash := computeHash(content)
	if existing, ok := s.byHash[hash]; ok {
		return existing, nil
	}

	artifact := &Artifact{
		ID:          generateID(path),
		Path:        path,
		Content:     content,
		ContentHash: hash,
		Kind:        kind,
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		Size:        int64(len(content)),
	}

	for _, opt := range opts {
		opt(artifact)
	}

	s.byPath[path] = artifact
	s.byKind[kind] = append(s.byKind[kind], artifact)
	s.byHash[hash] = artifact
	s.manifest.Artifacts = append(s.manifest.Artifacts, *artifact)
	s.manifest.UpdatedAt = time.Now()

	return artifact, nil
}

// Get retrieves an artifact by its path.
func (s *Store) Get(path string) (*Artifact, bool) {
	a, ok := s.byPath[path]
	return a, ok
}

// GetByKind returns all artifacts of the given kind.
func (s *Store) GetByKind(kind ArtifactKind) []*Artifact {
	return s.byKind[kind]
}

// GetByHash retrieves an artifact by its content hash.
func (s *Store) GetByHash(hash string) (*Artifact, bool) {
	a, ok := s.byHash[hash]
	return a, ok
}

// List returns all artifacts sorted by path.
func (s *Store) List() []Artifact {
	result := make([]Artifact, 0, len(s.manifest.Artifacts))
	result = append(result, s.manifest.Artifacts...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})
	return result
}

// Remove deletes an artifact from the store by path.
func (s *Store) Remove(path string) bool {
	artifact, ok := s.byPath[path]
	if !ok {
		return false
	}

	delete(s.byPath, path)
	delete(s.byHash, artifact.ContentHash)

	kindArtifacts := s.byKind[artifact.Kind]
	for i, a := range kindArtifacts {
		if a.Path == path {
			s.byKind[artifact.Kind] = append(kindArtifacts[:i], kindArtifacts[i+1:]...)
			break
		}
	}

	for i, a := range s.manifest.Artifacts {
		if a.Path == path {
			s.manifest.Artifacts = append(s.manifest.Artifacts[:i], s.manifest.Artifacts[i+1:]...)
			break
		}
	}

	s.manifest.UpdatedAt = time.Now()
	return true
}

// WriteToDisk persists the artifact manifest and all artifact contents to disk.
func (s *Store) WriteToDisk() error {
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return fmt.Errorf("create store dir: %w", err)
	}

	for _, a := range s.manifest.Artifacts {
		filePath, err := securityext.ValidateFilePath(filepath.Join(s.root, a.Path), s.root)
		if err != nil {
			return fmt.Errorf("invalid artifact path %s: %w", a.Path, err)
		}
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", dir, err)
		}
		if err := os.WriteFile(filePath, a.Content, 0o600); err != nil {
			return fmt.Errorf("write %s: %w", a.Path, err)
		}
	}

	manifestPath := filepath.Join(s.root, ".artifacts.json")
	data, err := json.MarshalIndent(s.manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	return os.WriteFile(manifestPath, data, 0o600)
}

// LoadFromDisk reads the artifact manifest and contents from disk into the store.
func (s *Store) LoadFromDisk() error {
	manifestPath := filepath.Join(s.root, ".artifacts.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read manifest: %w", err)
	}

	var manifest StoreManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("unmarshal manifest: %w", err)
	}

	s.manifest = manifest
	s.byPath = make(map[string]*Artifact)
	s.byKind = make(map[ArtifactKind][]*Artifact)
	s.byHash = make(map[string]*Artifact)

	for i := range s.manifest.Artifacts {
		a := &s.manifest.Artifacts[i]
		filePath, err := securityext.ValidateFilePath(filepath.Join(s.root, a.Path), s.root)
		if err != nil {
			continue
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		a.Content = content
		s.byPath[a.Path] = a
		s.byKind[a.Kind] = append(s.byKind[a.Kind], a)
		s.byHash[a.ContentHash] = a
	}

	return nil
}

// Deduplicate removes duplicate artifacts that share the same content hash.
func (s *Store) Deduplicate() int {
	seen := make(map[string]bool)
	removed := 0

	// Collect paths to remove first (avoid modifying map during iteration)
	var toRemove []string
	for path, a := range s.byPath {
		if seen[a.ContentHash] {
			toRemove = append(toRemove, path)
		} else {
			seen[a.ContentHash] = true
		}
	}

	for _, path := range toRemove {
		s.Remove(path)
		removed++
	}

	return removed
}

// Summary returns a count of artifacts grouped by kind.
func (s *Store) Summary() map[string]int {
	summary := make(map[string]int)
	for _, a := range s.manifest.Artifacts {
		summary[string(a.Kind)]++
	}
	summary["total"] = len(s.manifest.Artifacts)
	return summary
}

// SetProject sets the project name in the store manifest.
func (s *Store) SetProject(name string) {
	s.manifest.Project = name
}

// SetNEIRVersion stamps the NEIR version on all artifacts in the store.
func (s *Store) SetNEIRVersion(v string) {
	for i := range s.manifest.Artifacts {
		s.manifest.Artifacts[i].NEIRVersion = v
	}
}

// Option configures optional artifact properties during creation.
type Option func(*Artifact)

// WithLanguage sets the language metadata on an artifact.
func WithLanguage(lang string) Option {
	return func(a *Artifact) {
		a.Language = lang
	}
}

// WithNEIRVersion sets the NEIR version metadata on an artifact.
func WithNEIRVersion(v string) Option {
	return func(a *Artifact) {
		a.NEIRVersion = v
	}
}

// WithSource sets the source metadata on an artifact.
func WithSource(src string) Option {
	return func(a *Artifact) {
		a.Source = src
	}
}

// WithMetadata adds a key-value metadata pair to an artifact.
func WithMetadata(key, value string) Option {
	return func(a *Artifact) {
		if a.Metadata == nil {
			a.Metadata = make(map[string]string)
		}
		a.Metadata[key] = value
	}
}

func computeHash(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h)
}

func generateID(path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", h[:8])
}

// DetectKind infers the ArtifactKind from a file path based on extension and name.
func DetectKind(path string) ArtifactKind {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(filepath.Base(path))

	switch {
	case base == "dockerfile" || base == "docker-compose.yml" || base == "docker-compose.yaml":
		return KindDocker
	case base == "agents.md" || base == "claude.md" || strings.Contains(path, ".cursor") || strings.Contains(path, ".gemini") || strings.Contains(path, ".opencode") || strings.Contains(path, "copilot-instructions"):
		return KindAI
	case strings.HasPrefix(path, ".github/"):
		return KindCI
	case ext == ".go" || ext == ".ts" || ext == ".js" || ext == ".py" || ext == ".java" || ext == ".rs":
		return KindCode
	case ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".toml":
		return KindConfig
	case ext == ".md" || ext == ".rst" || ext == ".txt":
		return KindDocs
	case strings.Contains(path, "migration"):
		return KindMigration
	case strings.Contains(path, "profile"):
		return KindProfile
	default:
		return KindOther
	}
}
