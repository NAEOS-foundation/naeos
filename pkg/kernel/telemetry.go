// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package kernel

type TelemetryEvent struct {
	Name      string
	Timestamp int64
	Payload   map[string]any
}

type Metrics struct {
	Events    int
	LastEvent TelemetryEvent
}
