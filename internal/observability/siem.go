package observability

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/audit"
)

// SIEMExporter exports audit events to a SIEM-compatible collector.
// It supports CEF (Common Event Format) framing for syslog-style ingestion
// and a JSON/NDJSON mode for modern collectors.
type SIEMExporter struct {
	endpoint string
	client   *http.Client
	format   SIEMFormat
	tenantID string
	product  string
}

type SIEMFormat int

const (
	SIEMCEF SIEMFormat = iota
	SIEMJson
)

type SIEMOption func(*SIEMExporter)

func WithSIEMFormat(f SIEMFormat) SIEMOption {
	return func(e *SIEMExporter) { e.format = f }
}

func WithSIEMClient(c *http.Client) SIEMOption {
	return func(e *SIEMExporter) { e.client = c }
}

func WithSIEMTenant(tenant string) SIEMOption {
	return func(e *SIEMExporter) { e.tenantID = tenant }
}

func WithSIEMProduct(p string) SIEMOption {
	return func(e *SIEMExporter) { e.product = p }
}

func NewSIEMExporter(endpoint string, opts ...SIEMOption) *SIEMExporter {
	e := &SIEMExporter{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		product: "naeos",
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

// FormatCEF renders an audit event in Common Event Format.
// CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
func FormatCEF(event audit.AuditEvent, product string) string {
	severity := cefSeverity(event)
	esc := func(s string) string { return strings.ReplaceAll(s, "|", "\\|") }

	ext := make([]string, 0)
	ext = append(ext, fmt.Sprintf("rt=%s", event.Timestamp.UTC().Format(time.RFC3339)))
	if event.ID != "" {
		ext = append(ext, fmt.Sprintf("eventId=%s", esc(event.ID)))
	}
	if event.UserID != "" {
		ext = append(ext, fmt.Sprintf("suid=%s", esc(event.UserID)))
	}
	if event.ResourceID != "" {
		ext = append(ext, fmt.Sprintf("resourceId=%s", esc(event.ResourceID)))
	}
	if event.IP != "" {
		ext = append(ext, fmt.Sprintf("src=%s", esc(event.IP)))
	}
	if event.Status != "" {
		ext = append(ext, fmt.Sprintf("outcome=%s", esc(event.Status)))
	}
	if event.Details != "" {
		ext = append(ext, fmt.Sprintf("msg=%s", esc(strings.ReplaceAll(event.Details, " ", "\\"))))
	}
	if event.Hash != "" {
		ext = append(ext, fmt.Sprintf("fsHash=%s", esc(event.Hash)))
	}
	if event.PreviousHash != "" {
		ext = append(ext, fmt.Sprintf("fsPrevHash=%s", esc(event.PreviousHash)))
	}
	for k, v := range event.Metadata {
		ext = append(ext, fmt.Sprintf("%s=%s", esc(k), esc(strings.ReplaceAll(v, " ", "\\"))))
	}

	return fmt.Sprintf(
		"CEF:0|NAEOS|%s|1.0|%s|%s|%d|%s",
		esc(product), esc(event.Action), esc(event.Resource), severity, strings.Join(ext, " "),
	)
}

func cefSeverity(event audit.AuditEvent) int {
	switch event.Action {
	case "login", "access":
		if event.Status == "allowed" || event.Status == "success" {
			return 3
		}
		return 8
	case "delete", "grant", "rotate", "policy.change":
		return 6
	}
	if event.Status == "failed" || event.Status == "denied" || event.Status == "blocked" {
		return 8
	}
	return 2
}

func (e *SIEMExporter) ExportEvent(event audit.AuditEvent) error {
	frame, err := e.frame(event)
	if err != nil {
		return err
	}
	return e.post(frame)
}

func (e *SIEMExporter) ExportEvents(events []audit.AuditEvent) error {
	var sb strings.Builder
	for _, ev := range events {
		frame, err := e.frame(ev)
		if err != nil {
			return err
		}
		sb.WriteString(frame)
		sb.WriteString("\n")
	}
	return e.post(sb.String())
}

func (e *SIEMExporter) frame(event audit.AuditEvent) (string, error) {
	switch e.format {
	case SIEMJson:
		data, err := json.Marshal(event)
		if err != nil {
			return "", fmt.Errorf("marshal audit event: %w", err)
		}
		return string(data), nil
	case SIEMCEF:
		return FormatCEF(event, e.product), nil
	default:
		return FormatCEF(event, e.product), nil
	}
}

func (e *SIEMExporter) post(payload string) error {
	req, err := http.NewRequest(http.MethodPost, e.endpoint, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create siem request: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain")
	if e.format == SIEMJson {
		req.Header.Set("Content-Type", "application/x-ndjson")
	}
	if e.tenantID != "" {
		req.Header.Set("X-Tenant-ID", e.tenantID)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("siem http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("siem collector returned %d: %s", resp.StatusCode, body)
	}
	return nil
}
