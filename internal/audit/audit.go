package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEvent represents a single auditable action performed by a user.
type AuditEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	ResourceID string   `json:"resource_id,omitempty"`
	IP        string    `json:"ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Status    string    `json:"status"`
	Details   string    `json:"details,omitempty"`
}

// Auditor is the interface for logging audit events.
type Auditor interface {
	Log(event AuditEvent) error
}

// FileAuditor writes audit events to a local log file.
type FileAuditor struct {
	path string
	mu   sync.Mutex
}

// NewFileAuditor creates a file-backed auditor that writes to ~/.naeos/audit.log.
func NewFileAuditor(homeDir string) (*FileAuditor, error) {
	dir := filepath.Join(homeDir, ".naeos")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %w", err)
	}
	return &FileAuditor{
		path: filepath.Join(dir, "audit.log"),
	}, nil
}

// Log appends an audit event to the log file.
func (f *FileAuditor) Log(event AuditEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if event.ID == "" {
		event.ID = generateID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer file.Close()

	data = append(data, '\n')
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write audit event: %w", err)
	}

	return nil
}

// MemoryAuditor stores audit events in memory for testing.
type MemoryAuditor struct {
	events []AuditEvent
	mu     sync.Mutex
}

// NewMemoryAuditor creates an in-memory auditor.
func NewMemoryAuditor() *MemoryAuditor {
	return &MemoryAuditor{}
}

// Log stores an audit event in memory.
func (m *MemoryAuditor) Log(event AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event.ID == "" {
		event.ID = generateID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	m.events = append(m.events, event)
	return nil
}

// Events returns a copy of all stored audit events.
func (m *MemoryAuditor) Events() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	events := make([]AuditEvent, len(m.events))
	copy(events, m.events)
	return events
}

// Clear removes all stored audit events.
func (m *MemoryAuditor) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}

func generateID() string {
	return fmt.Sprintf("evt-%d", time.Now().UnixNano())
}
