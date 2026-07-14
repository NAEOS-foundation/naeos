package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	naeoserrors "github.com/NAEOS-foundation/naeos/internal/errors"
)

// CommandRunner is the interface for executing external commands.
type CommandRunner interface {
	Run(name string, args []string, dir string) ([]byte, error)
}

// ExecCommandRunner executes commands using os/exec.
type ExecCommandRunner struct{}

func (r *ExecCommandRunner) Run(name string, args []string, dir string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%s %s: %w\nstderr: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// TerraformRunner manages terraform init/plan/apply in a working directory.
type TerraformRunner struct {
	WorkDir string
	Runner  CommandRunner
}

// NewTerraformRunner creates a runner that executes terraform in the given directory.
func NewTerraformRunner(workDir string) *TerraformRunner {
	return &TerraformRunner{
		WorkDir: workDir,
		Runner:  &ExecCommandRunner{},
	}
}

// NewTerraformRunnerWithRunner creates a runner with a custom command runner for testing.
func NewTerraformRunnerWithRunner(workDir string, runner CommandRunner) *TerraformRunner {
	return &TerraformRunner{
		WorkDir: workDir,
		Runner:  runner,
	}
}

func (t *TerraformRunner) writeHCL(hcl string) error {
	if err := os.MkdirAll(t.WorkDir, 0o755); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to create terraform working directory", err)
	}
	path := filepath.Join(t.WorkDir, "main.tf")
	if err := os.WriteFile(path, []byte(hcl), 0o644); err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to write HCL file", err)
	}
	return nil
}

// Init runs terraform init in the working directory.
func (t *TerraformRunner) Init() error {
	_, err := t.Runner.Run("terraform", []string{"init", "-input=false"}, t.WorkDir)
	if err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "terraform init failed", err)
	}
	return nil
}

// PlanOutput summarizes the changes detected by terraform plan.
type PlanOutput struct {
	Changes struct {
		Add    int `json:"add"`
		Change int `json:"change"`
		Destroy int `json:"destroy"`
	} `json:"changes"`
}

// Plan runs terraform plan and returns a summary of the planned changes.
func (t *TerraformRunner) Plan() (*PlanOutput, error) {
	raw, err := t.Runner.Run("terraform", []string{"plan", "-input=false", "-json"}, t.WorkDir)
	if err != nil {
		return nil, naeoserrors.Wrap(naeoserrors.ErrCloud, "terraform plan failed", err)
	}
	return parsePlanJSON(raw)
}

func parsePlanOutput(raw []byte) (*PlanOutput, error) {
	return parsePlanJSON(raw)
}

func parsePlanJSON(raw []byte) (*PlanOutput, error) {
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var event struct {
			Type        string      `json:"@message"`
			Changes     *struct {
				Add    int `json:"add"`
				Change int `json:"change"`
				Destroy int `json:"destroy"`
			} `json:"changes"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Changes != nil {
			return &PlanOutput{
				Changes: struct {
					Add    int `json:"add"`
					Change int `json:"change"`
					Destroy int `json:"destroy"`
				}{
					Add:     event.Changes.Add,
					Change:  event.Changes.Change,
					Destroy: event.Changes.Destroy,
				},
			}, nil
		}
	}
	return &PlanOutput{}, nil
}

// Apply runs terraform apply with auto-approve.
func (t *TerraformRunner) Apply() error {
	_, err := t.Runner.Run("terraform", []string{"apply", "-auto-approve", "-input=false"}, t.WorkDir)
	if err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "terraform apply failed", err)
	}
	return nil
}

// ApplyDestroy runs terraform destroy with auto-approve.
func (t *TerraformRunner) ApplyDestroy() error {
	_, err := t.Runner.Run("terraform", []string{"destroy", "-auto-approve", "-input=false"}, t.WorkDir)
	if err != nil {
		return naeoserrors.Wrap(naeoserrors.ErrCloud, "terraform destroy failed", err)
	}
	return nil
}

// OutputValue represents a single terraform output value.
type OutputValue struct {
	Value     any `json:"value"`
	Sensitive bool        `json:"sensitive"`
	Type      any `json:"type"`
}

// Output retrieves all terraform output values.
func (t *TerraformRunner) Output() (map[string]OutputValue, error) {
	raw, err := t.Runner.Run("terraform", []string{"output", "-json"}, t.WorkDir)
	if err != nil {
		return nil, naeoserrors.Wrap(naeoserrors.ErrCloud, "terraform output failed", err)
	}
	var outputs map[string]OutputValue
	if err := json.Unmarshal(raw, &outputs); err != nil {
		return nil, naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to parse terraform output", err)
	}
	return outputs, nil
}

// Deploy writes HCL, inits, and applies terraform in one step.
func (t *TerraformRunner) Deploy(hcl string) error {
	if err := t.writeHCL(hcl); err != nil {
		return err
	}
	if err := t.Init(); err != nil {
		return err
	}
	if err := t.Apply(); err != nil {
		return err
	}
	return nil
}

// DestroyAll inits and destroys all terraform-managed resources.
func (t *TerraformRunner) DestroyAll() error {
	if err := t.Init(); err != nil {
		return err
	}
	if err := t.ApplyDestroy(); err != nil {
		return err
	}
	return nil
}

// TempWorkDir creates a temporary directory for terraform operations.
func TempWorkDir(prefix string) (string, error) {
	dir, err := os.MkdirTemp("", prefix+"-terraform-*")
	if err != nil {
		return "", naeoserrors.Wrap(naeoserrors.ErrCloud, "failed to create temp directory", err)
	}
	return dir, nil
}

type poolEntry struct {
	runner    *TerraformRunner
	lastUsed  time.Time
	initDone  bool
}

// RunnerPool manages a pool of reusable TerraformRunner instances.
type RunnerPool struct {
	mu       sync.RWMutex
	entries  map[string]*poolEntry
	maxSize  int
	idleTTL  time.Duration
}

// NewRunnerPool creates a pool with the given max size and idle TTL.
func NewRunnerPool(maxSize int, idleTTL time.Duration) *RunnerPool {
	if maxSize <= 0 {
		maxSize = 16
	}
	if idleTTL <= 0 {
		idleTTL = 30 * time.Minute
	}
	p := &RunnerPool{
		entries: make(map[string]*poolEntry),
		maxSize: maxSize,
		idleTTL: idleTTL,
	}
	go p.cleanupLoop()
	return p
}

func poolKey(project string, provider CloudProvider) string {
	return fmt.Sprintf("%s:%s", project, provider)
}

// Get retrieves a cached runner for the given project and provider.
func (p *RunnerPool) Get(project string, provider CloudProvider) (*TerraformRunner, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := poolKey(project, provider)
	entry, ok := p.entries[key]
	if !ok {
		return nil, false
	}
	entry.lastUsed = time.Now()
	return entry.runner, true
}

// Put adds a runner to the pool, evicting the oldest entry if at capacity.
func (p *RunnerPool) Put(project string, provider CloudProvider, runner *TerraformRunner, initDone bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := poolKey(project, provider)
	if len(p.entries) >= p.maxSize {
		p.evictOldest()
	}

	p.entries[key] = &poolEntry{
		runner:   runner,
		lastUsed: time.Now(),
		initDone: initDone,
	}
}

// Remove deletes a runner from the pool.
func (p *RunnerPool) Remove(project string, provider CloudProvider) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := poolKey(project, provider)
	delete(p.entries, key)
}

// Size returns the number of runners currently in the pool.
func (p *RunnerPool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.entries)
}

func (p *RunnerPool) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range p.entries {
		if oldestKey == "" || entry.lastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastUsed
		}
	}
	if oldestKey != "" {
		delete(p.entries, oldestKey)
	}
}

func (p *RunnerPool) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		p.cleanup()
	}
}

func (p *RunnerPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for key, entry := range p.entries {
		if now.Sub(entry.lastUsed) > p.idleTTL {
			delete(p.entries, key)
		}
	}
}

var defaultPool *RunnerPool
var defaultPoolOnce sync.Once

// GetDefaultPool returns the package-level singleton RunnerPool.
func GetDefaultPool() *RunnerPool {
	defaultPoolOnce.Do(func() {
		defaultPool = NewRunnerPool(16, 30*time.Minute)
	})
	return defaultPool
}
