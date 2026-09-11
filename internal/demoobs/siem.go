package demoobs

import (
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/NAEOS-foundation/naeos/internal/audit"
	"github.com/NAEOS-foundation/naeos/internal/investordemo"
	"github.com/NAEOS-foundation/naeos/internal/observability"
)

// SIEMForwarder implements investordemo.AuditObserver by exporting every
// recorded control-plane audit event to a SIEM collector. Forwarding is
// asynchronous: a worker goroutine drains a bounded channel, and events are
// dropped (and counted) if the collector cannot keep up, so a slow SIEM can
// never block the control plane.
type SIEMForwarder struct {
	exporter *observability.SIEMExporter
	logger   *slog.Logger
	tenantID string

	ch   chan *investordemo.AuditEvent
	done chan struct{}
	wg   sync.WaitGroup

	dropped atomic.Uint64
}

// NewSIEMForwarder builds a forwarder targeting the given collector URL and
// starts the worker goroutine.
func NewSIEMForwarder(endpoint string, format observability.SIEMFormat, opts ...Option) *SIEMForwarder {
	logger := slog.Default()
	tenantID := ""
	for _, o := range opts {
		if o.logger != nil {
			logger = o.logger
		}
		if o.tenantID != "" {
			tenantID = o.tenantID
		}
	}

	expOpts := []observability.SIEMOption{
		observability.WithSIEMFormat(format),
		observability.WithSIEMProduct("naeos-investor-demo"),
	}
	if tenantID != "" {
		expOpts = append(expOpts, observability.WithSIEMTenant(tenantID))
	}

	f := &SIEMForwarder{
		exporter: observability.NewSIEMExporter(endpoint, expOpts...),
		logger:   logger,
		tenantID: tenantID,
		ch:       make(chan *investordemo.AuditEvent, 256),
		done:     make(chan struct{}),
	}
	f.wg.Add(1)
	go f.worker()
	return f
}

// Option configures a SIEMForwarder.
type Option struct {
	tenantID string
	logger   *slog.Logger
}

// WithTenant attaches a tenant identifier to exported events.
func WithTenant(tenantID string) Option {
	return Option{tenantID: tenantID}
}

// WithLogger sets the logger used for forwarding errors.
func WithLogger(logger *slog.Logger) Option {
	return Option{logger: logger}
}

// OnRecorded implements investordemo.AuditObserver. It never blocks; if the
// buffer is full, the event is dropped and the drop is counted.
func (f *SIEMForwarder) OnRecorded(event *investordemo.AuditEvent) {
	select {
	case f.ch <- event:
	default:
		f.dropped.Add(1)
	}
}

// DroppedCount returns the number of events dropped due to a slow collector.
func (f *SIEMForwarder) DroppedCount() uint64 {
	return f.dropped.Load()
}

// Close stops the worker and waits for in-flight exports to finish.
func (f *SIEMForwarder) Close() {
	close(f.done)
	f.wg.Wait()
}

func (f *SIEMForwarder) worker() {
	defer f.wg.Done()
	for {
		select {
		case <-f.done:
			return
		case event := <-f.ch:
			if err := f.exporter.ExportEvent(toAuditEvent(event)); err != nil {
				f.logger.Error("siem export failed",
					"event_id", event.EventID, "error", err)
			}
		}
	}
}

// toAuditEvent maps a control-plane audit event onto the generic audit event
// consumed by the SIEM exporter and the tamper-evident audit chain.
func toAuditEvent(e *investordemo.AuditEvent) audit.AuditEvent {
	metadata := map[string]string{}
	if e.AgentID != "" {
		metadata["agent_id"] = e.AgentID
	}
	if e.PolicyID != "" {
		metadata["policy_id"] = e.PolicyID
	}
	if e.PolicyVersion != 0 {
		metadata["policy_version"] = strconv.Itoa(e.PolicyVersion)
	}
	for k, v := range e.Details {
		if sv, ok := v.(string); ok {
			metadata["detail_"+k] = sv
		}
	}

	return audit.AuditEvent{
		ID:        e.EventID,
		Timestamp: e.Timestamp,
		UserID:    e.AgentID,
		Action:    e.EventType,
		Resource:  string(e.RequestedCapability),
		Status:    e.Decision,
		Details:   e.Reason,
		Metadata:  metadata,
	}
}
