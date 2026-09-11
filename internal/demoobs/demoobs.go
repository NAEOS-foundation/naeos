// Package demoobs wires the investor demo control plane into the NAEOS
// observability stack so that live audit events and HTTP request traces can
// be exported to external collectors (SIEM, OTLP-compatible backends).
//
// The investor demo package stays free of observability imports; this package
// bridges the two at the server/CLI boundary.
package demoobs

import (
	"github.com/NAEOS-foundation/naeos/internal/observability"
)

// Config holds the optional observability wiring for a demo server.
// An empty Config (all fields empty) disables observability entirely.
type Config struct {
	// OTLPEndpoint is the base URL of the OTLP/HTTP collector, e.g.
	// "http://localhost:4318". When set, each API request is traced.
	OTLPEndpoint string

	// SIEMEndpoint is the SIEM collector URL. When set, every audit event
	// recorded by the control plane is forwarded to the collector.
	SIEMEndpoint string

	// SIEMFormat selects the framing: "cef" (default) or "json".
	SIEMFormat string

	// TenantID is attached to exported audit events.
	TenantID string
}

// Enabled reports whether any observability component is configured.
func (c Config) Enabled() bool {
	return c.HasOTLP() || c.HasSIEM()
}

// HasOTLP reports whether OTLP/HTTP tracing is configured.
func (c Config) HasOTLP() bool {
	return c.OTLPEndpoint != ""
}

// HasSIEM reports whether a SIEM collector is configured.
func (c Config) HasSIEM() bool {
	return c.SIEMEndpoint != ""
}

// SIEMFormatName returns "cef" or "json".
func (c Config) SIEMFormatName() string {
	if c.SIEMFormat == "json" {
		return "json"
	}
	return "cef"
}

// SIEMFormatValue returns the normalized SIEM framing, defaulting to CEF.
func (c Config) SIEMFormatValue() observability.SIEMFormat {
	if c.SIEMFormat == "json" {
		return observability.SIEMJson
	}
	return observability.SIEMCEF
}
