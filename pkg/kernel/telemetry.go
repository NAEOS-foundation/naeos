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
