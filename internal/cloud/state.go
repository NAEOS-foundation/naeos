package cloud

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	naeoserrors "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/securityext"
)

// DeploymentRecord stores the state of a completed cloud deployment.
type DeploymentRecord struct {
	Project      string             `json:"project"`
	Provider     CloudProvider      `json:"provider"`
	Environment  string             `json:"environment"`
	Region       string             `json:"region"`
	Resources    []DeployedResource `json:"resources"`
	TerraformDir string             `json:"terraform_dir"`
	Timestamp    time.Time          `json:"timestamp"`
	Status       string             `json:"status"`
}

// StateManager persists deployment records to disk.
type StateManager struct {
	mu      sync.RWMutex
	baseDir string
}

// NewStateManager creates a state manager using ~/.naeos/cloud as the base directory.
func NewStateManager() *StateManager {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".naeos", "cloud")
	return &StateManager{baseDir: base}
}

// NewStateManagerWithDir creates a state manager with a custom base directory.
func NewStateManagerWithDir(baseDir string) *StateManager {
	return &StateManager{baseDir: baseDir}
}

func (s *StateManager) deploymentPath(project, provider string) string {
	return filepath.Join(s.baseDir, project, provider, "deployment.json")
}

// Save persists a deployment record to disk.
func (s *StateManager) Save(record *DeploymentRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.Project == "" {
		return naeoserrors.New(naeoserrors.ErrValidation, "project name is required")
	}
	if record.Provider == "" {
		return naeoserrors.New(naeoserrors.ErrValidation, "provider is required")
	}

	if err := securityext.ValidatePluginName(record.Project); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrValidation, "invalid project name", err)
	}
	if err := securityext.ValidatePluginName(string(record.Provider)); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrValidation, "invalid provider name", err)
	}

	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}

	dir := filepath.Dir(s.deploymentPath(record.Project, string(record.Provider)))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to create state directory", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to marshal deployment record", err)
	}

	path := s.deploymentPath(record.Project, string(record.Provider))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to write deployment record", err)
	}
	return nil
}

// Load retrieves a deployment record by project and provider.
func (s *StateManager) Load(project string, provider CloudProvider) (*DeploymentRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := securityext.ValidatePluginName(project); err != nil {
		return nil, naeoserrors.Wrap(naeoserrors.ErrValidation, "invalid project name", err)
	}
	if err := securityext.ValidatePluginName(string(provider)); err != nil {
		return nil, naeoserrors.Wrap(naeoserrors.ErrValidation, "invalid provider name", err)
	}

	path := s.deploymentPath(project, string(provider))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, naeoserrors.Wrap(naeoserrors.ErrNotFound, fmt.Sprintf("deployment not found: %s/%s", project, provider), err)
		}
		return nil, naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to read deployment record", err)
	}

	var record DeploymentRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to parse deployment record", err)
	}
	return &record, nil
}

// List returns all deployment records sorted by timestamp descending.
func (s *StateManager) List() ([]DeploymentRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var records []DeploymentRecord

	projects, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return records, nil
		}
		return nil, naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to read state directory", err)
	}

	for _, projectDir := range projects {
		if !projectDir.IsDir() {
			continue
		}
		providers, err := os.ReadDir(filepath.Join(s.baseDir, projectDir.Name()))
		if err != nil {
			continue
		}
		for _, providerDir := range providers {
			if !providerDir.IsDir() {
				continue
			}
			recordPath := filepath.Join(s.baseDir, projectDir.Name(), providerDir.Name(), "deployment.json")
			data, err := os.ReadFile(recordPath)
			if err != nil {
				continue
			}
			var record DeploymentRecord
			if err := json.Unmarshal(data, &record); err != nil {
				continue
			}
			records = append(records, record)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.After(records[j].Timestamp)
	})

	return records, nil
}

// Delete removes a deployment record from disk.
func (s *StateManager) Delete(project string, provider CloudProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := securityext.ValidatePluginName(project); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrValidation, "invalid project name", err)
	}
	if err := securityext.ValidatePluginName(string(provider)); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrValidation, "invalid provider name", err)
	}

	dir := filepath.Join(s.baseDir, project, string(provider))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return naeoserrors.Wrap(naeoserrors.ErrNotFound, fmt.Sprintf("deployment not found: %s/%s", project, provider), err)
	}
	if err := os.RemoveAll(dir); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to delete deployment state", err)
	}

	parent := filepath.Join(s.baseDir, project)
	entries, _ := os.ReadDir(parent)
	if len(entries) == 0 {
		os.Remove(parent)
	}

	return nil
}
