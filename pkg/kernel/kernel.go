// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package kernel

import (
	"fmt"
	"sort"
	"sync"

	"github.com/NAEOS-foundation/naeos/internal/events"
)

type Kernel struct {
	mutex    sync.RWMutex
	services map[string]any
	bus      *events.Bus
	metrics  Metrics
	started  bool
}

func NewKernel() *Kernel {
	return &Kernel{
		services: map[string]any{},
		bus:      events.NewBus(),
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
	return k.bus.Subscribe(topic, func(event events.Event) error {
		handler(event.Payload)
		return nil
	})
}

func (k *Kernel) Publish(topic string, payload any) {
	_ = k.bus.Publish(topic, payload)
}

func (k *Kernel) Topics() []string {
	return k.bus.Topics()
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
