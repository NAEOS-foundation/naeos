package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed templates/*
var templatesFS embed.FS

type Dashboard struct {
	templates *template.Template
}

func New() (*Dashboard, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		templates: tmpl,
	}, nil
}

func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = d.templates.ExecuteTemplate(w, "index.html", nil)
}

type Stats struct {
	Projects  int    `json:"projects"`
	Artifacts int    `json:"artifacts"`
	Pipelines int    `json:"pipelines"`
	LastRun   string `json:"last_run"`
}

var (
	globalStats Stats
	statsMu     sync.RWMutex
	statsFile   string
)

func GetStats() *Stats {
	statsMu.RLock()
	defer statsMu.RUnlock()
	s := globalStats
	return &s
}

func RecordPipelineRun() {
	statsMu.Lock()
	defer statsMu.Unlock()
	globalStats.Pipelines++
	globalStats.LastRun = time.Now().Format(time.RFC3339)
	persistStats()
}

func SetProjects(n int) {
	statsMu.Lock()
	defer statsMu.Unlock()
	globalStats.Projects = n
	persistStats()
}

func SetArtifacts(n int) {
	statsMu.Lock()
	defer statsMu.Unlock()
	globalStats.Artifacts = n
	persistStats()
}

func SetStatsFile(path string) {
	statsMu.Lock()
	defer statsMu.Unlock()
	statsFile = path
	loadStats()
}

func persistStats() {
	if statsFile == "" {
		return
	}
	data, err := json.MarshalIndent(globalStats, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(statsFile, data, 0o600)
}

func loadStats() {
	if statsFile == "" {
		return
	}
	data, err := os.ReadFile(statsFile)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &globalStats)
}

// --- DashboardConfig ---

type DashboardConfig struct {
	RefreshInterval time.Duration `json:"refresh_interval"`
	MaxLogEntries   int           `json:"max_log_entries"`
	MaxSubscribers  int           `json:"max_subscribers"`
	StatsFilePath   string        `json:"stats_file_path"`
}

func DefaultConfig() DashboardConfig {
	return DashboardConfig{
		RefreshInterval: 5 * time.Second,
		MaxLogEntries:   500,
		MaxSubscribers:  64,
		StatsFilePath:   "",
	}
}

// --- ActivityLog ---

type LogLevel string

const (
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
}

type LogFilter struct {
	Level     LogLevel
	Since     time.Time
	Before    time.Time
	Substring string
	Limit     int
}

type ActivityLog struct {
	mu      sync.RWMutex
	entries []LogEntry
	nextID  int64
	maxLen  int
	onLog   func(LogEntry)
}

func (al *ActivityLog) SetLogCallback(fn func(LogEntry)) {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.onLog = fn
}

func NewActivityLog(maxEntries int) *ActivityLog {
	if maxEntries <= 0 {
		maxEntries = 500
	}
	return &ActivityLog{
		entries: make([]LogEntry, 0, maxEntries),
		nextID:  1,
		maxLen:  maxEntries,
	}
}

func (al *ActivityLog) Add(level LogLevel, message string) LogEntry {
	al.mu.Lock()

	entry := LogEntry{
		ID:        al.nextID,
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
	}
	al.nextID++
	al.entries = append(al.entries, entry)

	if len(al.entries) > al.maxLen {
		al.entries = al.entries[len(al.entries)-al.maxLen:]
	}

	cb := al.onLog
	al.mu.Unlock()

	if cb != nil {
		cb(entry)
	}

	return entry
}

func (al *ActivityLog) Entries() []LogEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()
	out := make([]LogEntry, len(al.entries))
	copy(out, al.entries)
	return out
}

func (al *ActivityLog) Filter(f LogFilter) []LogEntry {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var out []LogEntry
	for _, e := range al.entries {
		if f.Level != "" && e.Level != f.Level {
			continue
		}
		if !f.Since.IsZero() && e.Timestamp.Before(f.Since) {
			continue
		}
		if !f.Before.IsZero() && e.Timestamp.After(f.Before) {
			continue
		}
		if f.Substring != "" && !strings.Contains(e.Message, f.Substring) {
			continue
		}
		out = append(out, e)
	}

	if f.Limit > 0 && len(out) > f.Limit {
		out = out[len(out)-f.Limit:]
	}
	return out
}

func (al *ActivityLog) Len() int {
	al.mu.RLock()
	defer al.mu.RUnlock()
	return len(al.entries)
}

func (al *ActivityLog) Clear() {
	al.mu.Lock()
	defer al.mu.Unlock()
	al.entries = al.entries[:0]
	al.nextID = 1
}

// --- ComponentHealth ---

type HealthStatus string

const (
	Healthy   HealthStatus = "healthy"
	Degraded  HealthStatus = "degraded"
	Unhealthy HealthStatus = "unhealthy"
)

type ComponentInfo struct {
	Name       string       `json:"name"`
	Status     HealthStatus `json:"status"`
	LastCheck  time.Time    `json:"last_check"`
	Message    string       `json:"message"`
	CheckCount int64        `json:"check_count"`
}

type ComponentHealth struct {
	mu         sync.RWMutex
	components map[string]*ComponentInfo
}

func NewComponentHealth() *ComponentHealth {
	return &ComponentHealth{
		components: make(map[string]*ComponentInfo),
	}
}

func (ch *ComponentHealth) Set(name string, status HealthStatus, msg string) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	info, ok := ch.components[name]
	if !ok {
		info = &ComponentInfo{Name: name}
		ch.components[name] = info
	}
	info.Status = status
	info.LastCheck = time.Now()
	info.Message = msg
	info.CheckCount++
}

func (ch *ComponentHealth) Get(name string) (*ComponentInfo, bool) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	info, ok := ch.components[name]
	if !ok {
		return nil, false
	}
	cp := *info
	return &cp, true
}

func (ch *ComponentHealth) All() []ComponentInfo {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	out := make([]ComponentInfo, 0, len(ch.components))
	for _, info := range ch.components {
		out = append(out, *info)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func (ch *ComponentHealth) Remove(name string) bool {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if _, ok := ch.components[name]; !ok {
		return false
	}
	delete(ch.components, name)
	return true
}

func (ch *ComponentHealth) Summary() (total int, healthyCount int, degradedCount int, unhealthyCount int) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	for _, info := range ch.components {
		total++
		switch info.Status {
		case Healthy:
			healthyCount++
		case Degraded:
			degradedCount++
		case Unhealthy:
			unhealthyCount++
		}
	}
	return
}

// --- EventNotifier ---

type EventType string

const (
	EventStatsUpdate  EventType = "stats_update"
	EventLogEntry     EventType = "log_entry"
	EventHealthChange EventType = "health_change"
)

type DashboardEvent struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type subscriber struct {
	ch     chan DashboardEvent
	done   chan struct{}
	closed bool
}

type EventNotifier struct {
	mu          sync.RWMutex
	subscribers map[string]*subscriber
	nextID      int64
	bufferSize  int
}

func NewEventNotifier(bufferSize int) *EventNotifier {
	if bufferSize <= 0 {
		bufferSize = 32
	}
	return &EventNotifier{
		subscribers: make(map[string]*subscriber),
		bufferSize:  bufferSize,
	}
}

func (en *EventNotifier) Subscribe() (string, <-chan DashboardEvent) {
	en.mu.Lock()
	defer en.mu.Unlock()

	id := fmt.Sprintf("sub_%d", en.nextID)
	en.nextID++
	sub := &subscriber{
		ch:   make(chan DashboardEvent, en.bufferSize),
		done: make(chan struct{}),
	}
	en.subscribers[id] = sub
	return id, sub.ch
}

func (en *EventNotifier) Unsubscribe(id string) {
	en.mu.Lock()
	defer en.mu.Unlock()

	sub, ok := en.subscribers[id]
	if !ok {
		return
	}
	if !sub.closed {
		close(sub.done)
		close(sub.ch)
		sub.closed = true
	}
	delete(en.subscribers, id)
}

func (en *EventNotifier) Broadcast(evt DashboardEvent) {
	en.mu.RLock()
	defer en.mu.RUnlock()

	for _, sub := range en.subscribers {
		select {
		case sub.ch <- evt:
		default:
		}
	}
}

func (en *EventNotifier) SubscriberCount() int {
	en.mu.RLock()
	defer en.mu.RUnlock()
	return len(en.subscribers)
}

func (en *EventNotifier) Close() {
	en.mu.Lock()
	defer en.mu.Unlock()

	for id, sub := range en.subscribers {
		if !sub.closed {
			close(sub.done)
			close(sub.ch)
			sub.closed = true
		}
		delete(en.subscribers, id)
	}
}

// --- JSON API Handler ---

type APIResponse struct {
	OK        bool        `json:"ok"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type APIHandler struct {
	dashboard *Dashboard
	log       *ActivityLog
	health    *ComponentHealth
	config    DashboardConfig
}

func NewAPIHandler(d *Dashboard, al *ActivityLog, ch *ComponentHealth, cfg DashboardConfig) *APIHandler {
	return &APIHandler{
		dashboard: d,
		log:       al,
		health:    ch,
		config:    cfg,
	}
}

func (ah *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/api/stats":
		ah.handleStats(w, r)
	case "/api/activity":
		ah.handleActivity(w, r)
	case "/api/health":
		ah.handleHealth(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(APIResponse{
			OK:        false,
			Error:     "not found",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}
}

func (ah *APIHandler) handleStats(w http.ResponseWriter, _ *http.Request) {
	stats := GetStats()
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(APIResponse{
		OK:        true,
		Data:      stats,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (ah *APIHandler) handleActivity(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := LogFilter{
		Limit: ah.config.MaxLogEntries,
	}

	if lvl := q.Get("level"); lvl != "" {
		filter.Level = LogLevel(lvl)
	}
	if sub := q.Get("contains"); sub != "" {
		filter.Substring = sub
	}
	if lim := q.Get("limit"); lim != "" {
		var n int
		_, _ = fmt.Sscanf(lim, "%d", &n)
		if n > 0 {
			filter.Limit = n
		}
	}

	entries := ah.log.Filter(filter)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(APIResponse{
		OK:        true,
		Data:      entries,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (ah *APIHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	all := ah.health.All()
	total, healthyCount, degradedCount, unhealthyCount := ah.health.Summary()

	payload := map[string]interface{}{
		"components": all,
		"summary": map[string]int{
			"total":     total,
			"healthy":   healthyCount,
			"degraded":  degradedCount,
			"unhealthy": unhealthyCount,
		},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(APIResponse{
		OK:        true,
		Data:      payload,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// --- Broadcast helpers ---

func BroadcastStatsUpdate(en *EventNotifier) {
	stats := GetStats()
	en.Broadcast(DashboardEvent{
		Type:      EventStatsUpdate,
		Payload:   stats,
		Timestamp: time.Now(),
	})
}

func BroadcastLogEntry(en *EventNotifier, entry LogEntry) {
	en.Broadcast(DashboardEvent{
		Type:      EventLogEntry,
		Payload:   entry,
		Timestamp: time.Now(),
	})
}

func BroadcastHealthChange(en *EventNotifier, info ComponentInfo) {
	en.Broadcast(DashboardEvent{
		Type:      EventHealthChange,
		Payload:   info,
		Timestamp: time.Now(),
	})
}
