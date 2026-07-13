package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Span struct {
	Name      string            `json:"name"`
	ID        string            `json:"id"`
	TraceID   string            `json:"trace_id"`
	ParentID  string            `json:"parent_id,omitempty"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
	Status    string            `json:"status"`
}

type Exporter interface {
	ExportSpans(spans []Span) error
	Flush() error
}

type Config struct {
	Endpoint string
	Timeout  time.Duration
	BatchSize int
}

type Service struct {
	config   Config
	exporter Exporter
	spans    []Span
	mu       sync.Mutex
}

func NewService(config Config, exporter Exporter) *Service {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	return &Service{
		config:   config,
		exporter: exporter,
		spans:    make([]Span, 0, config.BatchSize),
	}
}

func (s *Service) StartSpan(name string) *Span {
	return &Span{
		Name:      name,
		ID:        generateID(),
		TraceID:   generateID(),
		StartTime: time.Now(),
		Status:    "ok",
	}
}

func (s *Service) StartSpanWithParent(name, parentID string) *Span {
	return &Span{
		Name:      name,
		ID:        generateID(),
		TraceID:   generateID(),
		ParentID:  parentID,
		StartTime: time.Now(),
		Status:    "ok",
	}
}

func (s *Service) EndSpan(span *Span) {
	span.EndTime = time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spans = append(s.spans, *span)
	if len(s.spans) >= s.config.BatchSize {
		s.flushUnsafe()
	}
}

func (s *Service) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.flushUnsafe()
}

func (s *Service) flushUnsafe() error {
	if len(s.spans) == 0 {
		return nil
	}
	batch := make([]Span, len(s.spans))
	copy(batch, s.spans)
	s.spans = s.spans[:0]
	return s.exporter.ExportSpans(batch)
}

func (s *Service) SpanCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.spans)
}

type HTTPExporter struct {
	endpoint string
	client   *http.Client
	spans    []Span
	mu       sync.Mutex
}

func NewHTTPExporter(endpoint string, timeout time.Duration) *HTTPExporter {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &HTTPExporter{
		endpoint: endpoint,
		client:   &http.Client{Timeout: timeout},
		spans:    make([]Span, 0),
	}
}

func (e *HTTPExporter) ExportSpans(spans []Span) error {
	e.mu.Lock()
	e.spans = append(e.spans, spans...)
	e.mu.Unlock()

	return e.Flush()
}

func (e *HTTPExporter) Flush() error {
	e.mu.Lock()
	if len(e.spans) == 0 {
		e.mu.Unlock()
		return nil
	}
	batch := make([]Span, len(e.spans))
	copy(batch, e.spans)
	e.spans = e.spans[:0]
	e.mu.Unlock()

	data, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("marshal spans: %w", err)
	}

	resp, err := e.client.Post(e.endpoint+"/v1/traces", "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("export spans: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("export failed with status %d", resp.StatusCode)
	}
	return nil
}

var idCounter uint64

func generateID() string {
	idCounter++
	return fmt.Sprintf("span-%d", idCounter)
}
