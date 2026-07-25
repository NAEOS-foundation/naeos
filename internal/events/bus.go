package events

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Event struct {
	Topic     string
	Payload   any
	Timestamp time.Time
	ID        string
}

type Handler func(event Event) error

type AsyncHandler struct {
	Handler   Handler
	Buffer    int
	workers   int
	ch        chan Event
	dead      []Event
	maxDead   int
	running   bool
	mu        sync.Mutex
	stopCh    chan struct{}
	deadCount int64
}

func NewAsyncHandler(handler Handler, workers, buffer int) *AsyncHandler {
	if workers <= 0 {
		workers = 1
	}
	if buffer <= 0 {
		buffer = 100
	}
	return &AsyncHandler{
		Handler: handler,
		Buffer:  buffer,
		workers: workers,
		maxDead: 100,
		stopCh:  make(chan struct{}),
	}
}

func (ah *AsyncHandler) Start() {
	ah.mu.Lock()
	if ah.running {
		ah.mu.Unlock()
		return
	}
	ah.running = true
	ah.ch = make(chan Event, ah.Buffer)
	ah.mu.Unlock()

	for i := 0; i < ah.workers; i++ {
		go ah.consume()
	}
}

func (ah *AsyncHandler) consume() {
	for {
		select {
		case event, ok := <-ah.ch:
			if !ok {
				return
			}
			if err := ah.Handler(event); err != nil {
				ah.mu.Lock()
				if len(ah.dead) >= ah.maxDead {
					ah.dead = ah.dead[1:]
				}
				ah.dead = append(ah.dead, event)
				ah.mu.Unlock()
				atomic.AddInt64(&ah.deadCount, 1)
			}
		case <-ah.stopCh:
			return
		}
	}
}

func (ah *AsyncHandler) Stop() {
	ah.mu.Lock()
	if !ah.running {
		ah.mu.Unlock()
		return
	}
	ah.running = false
	ah.mu.Unlock()
	close(ah.stopCh)
}

func (ah *AsyncHandler) DeadLetters() []Event {
	ah.mu.Lock()
	defer ah.mu.Unlock()
	out := make([]Event, len(ah.dead))
	copy(out, ah.dead)
	return out
}

func (ah *AsyncHandler) DeadLetterCount() int64 {
	return atomic.LoadInt64(&ah.deadCount)
}

type Middleware func(event Event, next Handler) error

type EventBus interface {
	Publish(topic string, payload any) error
	Subscribe(topic string, handler Handler) error
	Unsubscribe(topic string) error
	Topics() []string
	SubscriberCount(topic string) int
}

type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
	middleware  []Middleware
	history     []Event
	maxHistory  int
	eventID     int64
}

func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[string][]Handler),
		maxHistory:  1000,
	}
}

func (b *Bus) Use(mw Middleware) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.middleware = append(b.middleware, mw)
}

func (b *Bus) Publish(topic string, payload any) error {
	b.mu.RLock()
	handlers := b.subscribers[topic]
	wildcards := b.subscribers["*"]
	allHandlers := make([]Handler, 0, len(handlers)+len(wildcards))
	allHandlers = append(allHandlers, handlers...)
	allHandlers = append(allHandlers, wildcards...)
	maxHistory := b.maxHistory
	b.mu.RUnlock()

	event := Event{
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
		ID:        fmt.Sprintf("evt-%d", atomic.AddInt64(&b.eventID, 1)),
	}

	b.mu.Lock()
	if maxHistory > 0 {
		b.history = append(b.history, event)
		if len(b.history) > maxHistory {
			b.history = b.history[len(b.history)-maxHistory:]
		}
	}
	b.mu.Unlock()

	handler := func(event Event) error {
		for _, h := range allHandlers {
			if err := h(event); err != nil {
				return err
			}
		}
		return nil
	}

	b.mu.RLock()
	mws := make([]Middleware, len(b.middleware))
	copy(mws, b.middleware)
	b.mu.RUnlock()

	for i := len(mws) - 1; i >= 0; i-- {
		mw := mws[i]
		next := handler
		handler = func(event Event) error {
			return mw(event, next)
		}
	}

	return handler(event)
}

func (b *Bus) Subscribe(topic string, handler Handler) error {
	if handler == nil {
		return naeoserr.New(naeoserr.ErrInternal, "handler must not be nil")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], handler)
	return nil
}

func (b *Bus) Unsubscribe(topic string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.subscribers[topic]; !exists {
		return naeoserr.New(naeoserr.ErrInternal, fmt.Sprintf("no subscribers for topic %s", topic))
	}
	delete(b.subscribers, topic)
	return nil
}

func (b *Bus) Topics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]string, 0, len(b.subscribers))
	for topic := range b.subscribers {
		result = append(result, topic)
	}
	return result
}

func (b *Bus) SubscriberCount(topic string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers[topic])
}

func (b *Bus) History() []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]Event, len(b.history))
	copy(out, b.history)
	return out
}

func (b *Bus) HistorySince(since time.Time) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []Event
	for _, e := range b.history {
		if e.Timestamp.After(since) || e.Timestamp.Equal(since) {
			out = append(out, e)
		}
	}
	return out
}

func (b *Bus) HistoryByTopic(topic string) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []Event
	for _, e := range b.history {
		if e.Topic == topic {
			out = append(out, e)
		}
	}
	return out
}

func (b *Bus) HistoryFilter(topicPattern string, since time.Time, limit int) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var out []Event
	for _, e := range b.history {
		if e.Timestamp.Before(since) {
			continue
		}
		if topicPattern != "" && !matchTopic(e.Topic, topicPattern) {
			continue
		}
		out = append(out, e)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func (b *Bus) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history = nil
}

func matchTopic(topic, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(topic, prefix)
	}
	return topic == pattern
}
