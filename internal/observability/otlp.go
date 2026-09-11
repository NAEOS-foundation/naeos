package observability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OTLPHTTPExporter exports spans to an OTLP-compatible collector via HTTP/JSON.
// Endpoint should be the base URL of the OTLP HTTP collector (e.g. "http://localhost:4318").
type OTLPHTTPExporter struct {
	endpoint string
	client   *http.Client
	headers  map[string]string
}

type otlpExportRequest struct {
	ResourceSpans []otlpResourceSpans `json:"resourceSpans"`
}

type otlpResourceSpans struct {
	Resource   otlpResource     `json:"resource"`
	ScopeSpans []otlpScopeSpans `json:"scopeSpans"`
}

type otlpResource struct {
	Attributes []otlpKV `json:"attributes"`
}

type otlpScopeSpans struct {
	Scope otlpScope  `json:"scope"`
	Spans []otlpSpan `json:"spans"`
}

type otlpScope struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type otlpSpan struct {
	TraceID           string      `json:"traceId"`
	SpanID            string      `json:"spanId"`
	ParentSpanID      string      `json:"parentSpanId,omitempty"`
	Name              string      `json:"name"`
	StartTimeUnixNano int64       `json:"startTimeUnixNano"`
	EndTimeUnixNano   int64       `json:"endTimeUnixNano"`
	Attributes        []otlpKV    `json:"attributes"`
	Events            []otlpEvent `json:"events,omitempty"`
	Status            otlpStatus  `json:"status"`
}

type otlpKV struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type otlpEvent struct {
	Name       string   `json:"name"`
	Timestamp  int64    `json:"timeUnixNano"`
	Attributes []otlpKV `json:"attributes,omitempty"`
}

type otlpStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func NewOTLPHTTPExporter(endpoint string, opts ...OTLPOption) *OTLPHTTPExporter {
	e := &OTLPHTTPExporter{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

type OTLPOption func(*OTLPHTTPExporter)

func WithOTLPClient(c *http.Client) OTLPOption {
	return func(e *OTLPHTTPExporter) { e.client = c }
}

func WithOTLPHeaders(h map[string]string) OTLPOption {
	return func(e *OTLPHTTPExporter) { e.headers = h }
}

func (e *OTLPHTTPExporter) ExportSpans(spans []*Span) error {
	req := otlpExportRequest{
		ResourceSpans: []otlpResourceSpans{
			{
				Resource: otlpResource{
					Attributes: []otlpKV{
						{Key: "service.name", Value: "naeos"},
					},
				},
				ScopeSpans: []otlpScopeSpans{
					{
						Scope: otlpScope{
							Name:    "naeos-tracer",
							Version: "1",
						},
						Spans: e.convertSpans(spans),
					},
				},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal otlp request: %w", err)
	}

	url := e.endpoint + "/v1/traces"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range e.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("otlp http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("otlp collector returned %d: %s", resp.StatusCode, body)
	}
	return nil
}

func (e *OTLPHTTPExporter) ExportMetrics(metrics []*Metric) error {
	return nil
}

func (e *OTLPHTTPExporter) ExportLogs(entries []LogEntry) error {
	return nil
}

func (e *OTLPHTTPExporter) convertSpans(spans []*Span) []otlpSpan {
	out := make([]otlpSpan, 0, len(spans))
	for _, s := range spans {
		os := otlpSpan{
			TraceID:           s.TraceID,
			SpanID:            s.SpanID,
			ParentSpanID:      s.ParentID,
			Name:              s.Name,
			StartTimeUnixNano: s.StartTime.UnixNano(),
			EndTimeUnixNano:   s.EndTime.UnixNano(),
			Attributes:        convertAttrs(s.Attributes),
			Status: otlpStatus{
				Code:    int(s.Status.Code),
				Message: s.Status.Message,
			},
		}
		for _, ev := range s.Events {
			os.Events = append(os.Events, otlpEvent{
				Name:       ev.Name,
				Timestamp:  ev.Timestamp.UnixNano(),
				Attributes: convertAttrs(ev.Attributes),
			})
		}
		out = append(out, os)
	}
	return out
}

func convertAttrs(attrs map[string]any) []otlpKV {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]otlpKV, 0, len(attrs))
	for k, v := range attrs {
		out = append(out, otlpKV{Key: k, Value: v})
	}
	return out
}
