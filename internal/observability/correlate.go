package observability

import (
	"context"
	"net/http"
	"sync"
)

type contextKey string

const (
	requestIDContextKey contextKey = "request_id"
	tenantIDContextKey  contextKey = "tenant_id"
)

// CorrelationStore maps request_id to (trace_id, tenant_id).
type CorrelationStore struct {
	mu    sync.RWMutex
	index map[string]CorrelationEntry
	byReq map[string]string   // request_id → trace_id
	byTnt map[string][]string // tenant_id → []trace_id
}

type CorrelationEntry struct {
	RequestID string
	TraceID   string
	TenantID  string
}

func NewCorrelationStore() *CorrelationStore {
	return &CorrelationStore{
		index: make(map[string]CorrelationEntry),
		byReq: make(map[string]string),
		byTnt: make(map[string][]string),
	}
}

func (cs *CorrelationStore) Record(requestID, traceID, tenantID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.index[requestID] = CorrelationEntry{
		RequestID: requestID,
		TraceID:   traceID,
		TenantID:  tenantID,
	}
	cs.byReq[requestID] = traceID
	if tenantID != "" {
		cs.byTnt[tenantID] = append(cs.byTnt[tenantID], traceID)
	}
}

func (cs *CorrelationStore) ByRequestID(requestID string) (CorrelationEntry, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	entry, ok := cs.index[requestID]
	return entry, ok
}

func (cs *CorrelationStore) TraceIDsByTenant(tenantID string) []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.byTnt[tenantID]
}

func (cs *CorrelationStore) TraceIDForRequest(requestID string) (string, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	tid, ok := cs.byReq[requestID]
	return tid, ok
}

func (cs *CorrelationStore) Len() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.index)
}

// CorrelationMiddleware extracts X-Request-ID and X-Tenant-ID headers
// and stores them in the request context.
func CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		tntID := r.Header.Get("X-Tenant-ID")

		ctx := r.Context()
		if reqID != "" {
			ctx = setContextValue(ctx, requestIDContextKey, reqID)
		}
		if tntID != "" {
			ctx = setContextValue(ctx, tenantIDContextKey, tntID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(r *http.Request) string {
	v, _ := r.Context().Value(requestIDContextKey).(string)
	return v
}

func TenantIDFromContext(r *http.Request) string {
	v, _ := r.Context().Value(tenantIDContextKey).(string)
	return v
}

func setContextValue(ctx context.Context, key contextKey, val string) context.Context {
	return context.WithValue(ctx, key, val)
}

// SpanContext carrier for propagation
type SpanCarrier struct {
	RequestID string
	TenantID  string
	TraceID   string
}

// LinkFromCorrelation produces a SpanCarrier linking request + tenant to a trace.
func LinkFromCorrelation(cs *CorrelationStore, requestID, tenantID, traceID string) SpanCarrier {
	cs.Record(requestID, traceID, tenantID)
	return SpanCarrier{
		RequestID: requestID,
		TenantID:  tenantID,
		TraceID:   traceID,
	}
}
