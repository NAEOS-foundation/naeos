package demoobs

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/NAEOS-foundation/naeos/internal/observability"
)

// TracingMiddleware wraps an http.Handler so that every request is recorded
// as an OTLP span and exported to the configured collector. The span name,
// HTTP method, path, request ID (X-Request-ID), and tenant (X-Tenant-ID) are
// attached as attributes, and the span status reflects the HTTP result.
func TracingMiddleware(tracer *observability.Tracer, exporter observability.Exporter, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			span := tracer.StartSpan(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			span.Attributes["http.request_id"] = r.Header.Get("X-Request-ID")
			span.Attributes["tenant_id"] = r.Header.Get("X-Tenant-ID")
			span.Attributes["http.client_ip"] = r.RemoteAddr
			if r.URL.RawQuery != "" {
				span.Attributes["http.query"] = r.URL.RawQuery
			}

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			tracer.EndSpan(span)
			tracer.SetStatus(span, httpStatus(sw.status), http.StatusText(sw.status))

			if exporter == nil {
				return
			}
			spans := tracer.GetSpansByTrace(span.TraceID)
			go func() {
				if err := exporter.ExportSpans(spans); err != nil && logger != nil {
					logger.Error("otlp export failed",
						"trace_id", span.TraceID, "error", err)
				}
			}()
		})
	}
}

// statusWriter captures the response status for tracing.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func httpStatus(code int) observability.SpanStatusCode {
	if code >= 400 {
		return observability.SpanStatusError
	}
	return observability.SpanStatusOK
}

// NewDemoTracer builds a tracer used only for demo request tracing.
func NewDemoTracer() *observability.Tracer {
	return observability.NewTracer("naeos-investor-demo")
}
