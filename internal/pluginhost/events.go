package pluginhost

import (
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// PipelineEvent represents a lifecycle event in the pipeline.
type PipelineEvent string

const (
	EventBeforeParse        PipelineEvent = "before_parse"
	EventAfterParse         PipelineEvent = "after_parse"
	EventBeforeGenerate     PipelineEvent = "before_generate"
	EventAfterGenerate      PipelineEvent = "after_generate"
	EventOnPipelineComplete PipelineEvent = "on_pipeline_complete"
	EventOnPipelineFailed   PipelineEvent = "on_pipeline_failed"
)

// EventData holds contextual data passed with pipeline events.
type EventData struct {
	PipelineID string
	Stage      string
	Artifacts  int
	Duration   string
	Error      string
	Extra      map[string]any
}

// EventHandler is a function that handles a pipeline event for a specific plugin.
type EventHandler func(pluginName string, data *EventData) error

// EventBus routes pipeline lifecycle events to subscribed plugins.
type EventBus struct {
	subscriptions map[PipelineEvent]map[string]EventHandler
	mu            sync.RWMutex
}

// NewEventBus creates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscriptions: make(map[PipelineEvent]map[string]EventHandler),
	}
}

// Subscribe registers a plugin to receive a specific pipeline event.
func (eb *EventBus) Subscribe(event PipelineEvent, pluginName string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.subscriptions[event] == nil {
		eb.subscriptions[event] = make(map[string]EventHandler)
	}
	eb.subscriptions[event][pluginName] = handler
}

// Unsubscribe removes a plugin's subscription for a specific event.
func (eb *EventBus) Unsubscribe(event PipelineEvent, pluginName string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if subs, ok := eb.subscriptions[event]; ok {
		delete(subs, pluginName)
	}
}

// Emit dispatches an event to all subscribed plugins. Returns errors from failed handlers.
func (eb *EventBus) Emit(event PipelineEvent, data *EventData) []error {
	eb.mu.RLock()
	subs := make(map[string]EventHandler, len(eb.subscriptions[event]))
	for name, handler := range eb.subscriptions[event] {
		subs[name] = handler
	}
	eb.mu.RUnlock()

	var errs []error
	for name, handler := range subs {
		if err := handler(name, data); err != nil {
			errs = append(errs, naeoserr.Wrapf(err, naeoserr.ErrPlugin, "plugin %s", name))
		}
	}
	return errs
}

// Subscribers returns the list of plugin names subscribed to an event.
func (eb *EventBus) Subscribers(event PipelineEvent) []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var names []string
	for name := range eb.subscriptions[event] {
		names = append(names, name)
	}
	return names
}

// HasSubscribers returns true if any plugins are subscribed to the given event.
func (eb *EventBus) HasSubscribers(event PipelineEvent) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscriptions[event]) > 0
}

// PluginEventBus wraps EventBus to implement the pipeline.PipelineObserver interface,
// bridging pipeline lifecycle events to plugin subscriptions.
type PluginEventBus struct {
	bus *EventBus
}

// NewPluginEventBus creates a PluginEventBus backed by the given EventBus.
func NewPluginEventBus(bus *EventBus) *PluginEventBus {
	return &PluginEventBus{bus: bus}
}

// OnPipelineStart implements PipelineObserver. Emits EventOnPipelineComplete is not
// applicable; this is a no-op for start events.
func (p *PluginEventBus) OnPipelineStart(pipelineID string) {
	// Pipeline start is not in the plugin event set; no-op.
}

// OnPipelineComplete implements PipelineObserver.
func (p *PluginEventBus) OnPipelineComplete(pipelineID string, artifacts int, duration string) {
	p.bus.Emit(EventOnPipelineComplete, &EventData{
		PipelineID: pipelineID,
		Artifacts:  artifacts,
		Duration:   duration,
	})
}

// OnPipelineFailed implements PipelineObserver.
func (p *PluginEventBus) OnPipelineFailed(pipelineID string, errMsg string) {
	p.bus.Emit(EventOnPipelineFailed, &EventData{
		PipelineID: pipelineID,
		Error:      errMsg,
	})
}

// OnArtifactGenerated implements PipelineObserver.
func (p *PluginEventBus) OnArtifactGenerated(name string, path string) {
	// Artifact generation events are not in the plugin event set; no-op.
}

// EmitBeforeParse emits a BeforeParse event.
func (p *PluginEventBus) EmitBeforeParse(pipelineID string) []error {
	return p.bus.Emit(EventBeforeParse, &EventData{PipelineID: pipelineID, Stage: "parse"})
}

// EmitAfterParse emits an AfterParse event.
func (p *PluginEventBus) EmitAfterParse(pipelineID string) []error {
	return p.bus.Emit(EventAfterParse, &EventData{PipelineID: pipelineID, Stage: "parse"})
}

// EmitBeforeGenerate emits a BeforeGenerate event.
func (p *PluginEventBus) EmitBeforeGenerate(pipelineID string) []error {
	return p.bus.Emit(EventBeforeGenerate, &EventData{PipelineID: pipelineID, Stage: "generate"})
}

// EmitAfterGenerate emits an AfterGenerate event.
func (p *PluginEventBus) EmitAfterGenerate(pipelineID string) []error {
	return p.bus.Emit(EventAfterGenerate, &EventData{PipelineID: pipelineID, Stage: "generate"})
}
