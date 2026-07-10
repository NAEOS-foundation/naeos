package kernel

import (
	"fmt"
	"sort"
	"sync"
)

type Kernel struct {
	mutex       sync.RWMutex
	services    map[string]any
	subscribers map[string][]func(any)
	metrics     Metrics
	started     bool
}

func NewKernel() *Kernel {
	return &Kernel{
		services:    map[string]any{},
		subscribers: map[string][]func(any){},
	}
}

func (k *Kernel) Register(name string, service any) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	k.mutex.Lock()
	defer k.mutex.Unlock()

	if _, exists := k.services[name]; exists {
		return fmt.Errorf("service %q already registered", name)
	}
	k.services[name] = service
	return nil
}

func (k *Kernel) Resolve(name string) (any, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	service, exists := k.services[name]
	if !exists {
		return nil, fmt.Errorf("service %q not found", name)
	}
	return service, nil
}

func (k *Kernel) RegisteredServices() []string {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	names := make([]string, 0, len(k.services))
	for name := range k.services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (k *Kernel) Start() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if k.started {
		return fmt.Errorf("kernel already started")
	}

	for _, service := range k.services {
		if lifecycle, ok := service.(Lifecycle); ok {
			if err := lifecycle.Initialize(); err != nil {
				return fmt.Errorf("initialize service: %w", err)
			}
			if err := lifecycle.Start(); err != nil {
				return fmt.Errorf("start service: %w", err)
			}
		}
	}

	k.started = true
	return nil
}

func (k *Kernel) Stop() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	if !k.started {
		return fmt.Errorf("kernel is not running")
	}

	for _, service := range k.services {
		if lifecycle, ok := service.(Lifecycle); ok {
			if err := lifecycle.Stop(); err != nil {
				return fmt.Errorf("stop service: %w", err)
			}
		}
	}

	k.started = false
	return nil
}

func (k *Kernel) Subscribe(topic string, handler func(any)) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	k.mutex.Lock()
	defer k.mutex.Unlock()

	k.subscribers[topic] = append(k.subscribers[topic], handler)
	return nil
}

func (k *Kernel) Publish(topic string, payload any) {
	k.mutex.RLock()
	handlers := append([]func(any){}, k.subscribers[topic]...)
	k.mutex.RUnlock()

	for _, handler := range handlers {
		handler(payload)
	}
}

func (k *Kernel) Topics() []string {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	topics := make([]string, 0, len(k.subscribers))
	for topic := range k.subscribers {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	return topics
}

func (k *Kernel) EmitTelemetry(event TelemetryEvent) error {
	if event.Name == "" {
		return fmt.Errorf("telemetry event name cannot be empty")
	}

	k.mutex.Lock()
	defer k.mutex.Unlock()

	k.metrics.Events++
	k.metrics.LastEvent = event
	return nil
}

func (k *Kernel) Metrics() Metrics {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	return k.metrics
}
