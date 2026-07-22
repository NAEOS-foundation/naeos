package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type AuditEvent struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	UserID     string            `json:"user_id"`
	Action     string            `json:"action"`
	Resource   string            `json:"resource"`
	ResourceID string            `json:"resource_id,omitempty"`
	IP         string            `json:"ip,omitempty"`
	UserAgent  string            `json:"user_agent,omitempty"`
	Status     string            `json:"status"`
	Details    string            `json:"details,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type Auditor interface {
	Log(event AuditEvent) error
}

type FileAuditor struct {
	path string
	mu   sync.Mutex
}

func NewFileAuditor(homeDir string) (*FileAuditor, error) {
	dir := filepath.Join(homeDir, ".naeos")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create audit directory: %w", err)
	}
	return &FileAuditor{
		path: filepath.Join(dir, "audit.log"),
	}, nil
}

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

type MemoryAuditor struct {
	events []AuditEvent
	mu     sync.Mutex
}

func NewMemoryAuditor() *MemoryAuditor {
	return &MemoryAuditor{}
}

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

func (m *MemoryAuditor) Events() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	events := make([]AuditEvent, len(m.events))
	copy(events, m.events)
	return events
}

func (m *MemoryAuditor) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}

type Query struct {
	From     time.Time
	To       time.Time
	UserID   string
	Action   string
	Resource string
	Status   string
	Limit    int
	Offset   int
}

func (m *MemoryAuditor) Query(q Query) []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []AuditEvent
	for _, e := range m.events {
		if !q.From.IsZero() && e.Timestamp.Before(q.From) {
			continue
		}
		if !q.To.IsZero() && e.Timestamp.After(q.To) {
			continue
		}
		if q.UserID != "" && e.UserID != q.UserID {
			continue
		}
		if q.Action != "" && e.Action != q.Action {
			continue
		}
		if q.Resource != "" && e.Resource != q.Resource {
			continue
		}
		if q.Status != "" && e.Status != q.Status {
			continue
		}
		result = append(result, e)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	if q.Offset > 0 {
		if q.Offset >= len(result) {
			return nil
		}
		result = result[q.Offset:]
	}
	if q.Limit > 0 && len(result) > q.Limit {
		result = result[:q.Limit]
	}

	return result
}

type Aggregation struct {
	ByAction   map[string]int
	ByUser     map[string]int
	ByResource map[string]int
	ByStatus   map[string]int
	ByDay      map[string]int
	Total      int
}

func (m *MemoryAuditor) Aggregate() Aggregation {
	m.mu.Lock()
	defer m.mu.Unlock()

	agg := Aggregation{
		ByAction:   make(map[string]int),
		ByUser:     make(map[string]int),
		ByResource: make(map[string]int),
		ByStatus:   make(map[string]int),
		ByDay:      make(map[string]int),
	}

	for _, e := range m.events {
		agg.Total++
		agg.ByAction[e.Action]++
		agg.ByUser[e.UserID]++
		agg.ByResource[e.Resource]++
		agg.ByStatus[e.Status]++
		day := e.Timestamp.Format("2006-01-02")
		agg.ByDay[day]++
	}

	return agg
}

type RetentionPolicy struct {
	MaxAge   time.Duration
	MaxCount int
	KeepDays int
}

func (m *MemoryAuditor) ApplyRetention(policy RetentionPolicy) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-policy.MaxAge)
	original := len(m.events)

	var kept []AuditEvent
	for _, e := range m.events {
		if policy.MaxAge != 0 && e.Timestamp.Before(cutoff) {
			continue
		}
		kept = append(kept, e)
	}

	if policy.MaxCount > 0 && len(kept) > policy.MaxCount {
		kept = kept[len(kept)-policy.MaxCount:]
	}

	m.events = kept
	return original - len(m.events)
}

func (m *MemoryAuditor) ExportCSV(path string) error {
	m.mu.Lock()
	events := make([]AuditEvent, len(m.events))
	copy(events, m.events)
	m.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("id,timestamp,user_id,action,resource,resource_id,status,ip,user_agent,details\n")
	for _, e := range events {
		fmt.Fprintf(&sb, "%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			escapeCSV(e.ID),
			e.Timestamp.Format(time.RFC3339),
			escapeCSV(e.UserID),
			escapeCSV(e.Action),
			escapeCSV(e.Resource),
			escapeCSV(e.ResourceID),
			escapeCSV(e.Status),
			escapeCSV(e.IP),
			escapeCSV(e.UserAgent),
			escapeCSV(e.Details),
		)
	}
	return os.WriteFile(path, []byte(sb.String()), 0o600)
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func (m *MemoryAuditor) ExportJSON(path string) error {
	m.mu.Lock()
	events := make([]AuditEvent, len(m.events))
	copy(events, m.events)
	m.mu.Unlock()

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal events: %w", err)
	}
	return os.WriteFile(path, data, 0o600)
}

func (m *MemoryAuditor) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func (m *MemoryAuditor) Latest() *AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.events) == 0 {
		return nil
	}
	return &m.events[len(m.events)-1]
}

func (m *MemoryAuditor) Oldest() *AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.events) == 0 {
		return nil
	}
	return &m.events[0]
}

func (m *MemoryAuditor) ByID(id string) *AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.events {
		if e.ID == id {
			return &e
		}
	}
	return nil
}

func (m *MemoryAuditor) UserActions(userID string) []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []AuditEvent
	for _, e := range m.events {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result
}

func (m *MemoryAuditor) FailedEvents() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []AuditEvent
	for _, e := range m.events {
		if e.Status == "failed" || e.Status == "error" {
			result = append(result, e)
		}
	}
	return result
}

func generateID() string {
	return fmt.Sprintf("evt-%d", time.Now().UnixNano())
}
