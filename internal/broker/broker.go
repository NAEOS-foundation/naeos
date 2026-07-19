package broker

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Message Broker Interface

type Broker interface {
	Name() string
	Connect(config *Config) error
	Close() error
	Publish(channel string, msg *Message) error
	Subscribe(channel string, handler MessageHandler) error
	Unsubscribe(channel string) error
	Ping() error
}

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Timeout  time.Duration
}

type Message struct {
	ID        string
	Channel   string
	Payload   []byte
	Timestamp time.Time
}

type MessageHandler func(msg *Message) error

// stubBroker is the shared base type for Redis, RabbitMQ, and Kafka stubs.
// All three adapters embed this and only override Name() and field names
// in their constructors, eliminating the previous copy-paste duplication.

type stubBroker struct {
	config    *Config
	connected bool
	channels  map[string]MessageHandler
	mu        sync.RWMutex
}

func (s *stubBroker) connect(config *Config) error {
	s.config = config
	s.connected = true
	return nil
}

func (s *stubBroker) close() error {
	s.connected = false
	return nil
}

func (s *stubBroker) ping() error {
	if !s.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (s *stubBroker) publish(channel string, msg *Message) error {
	if !s.connected {
		return fmt.Errorf("not connected")
	}

	s.mu.RLock()
	handler, ok := s.channels[channel]
	s.mu.RUnlock()

	if ok {
		go func() { _ = handler(msg) }()
	}
	return nil
}

func (s *stubBroker) subscribe(channel string, handler MessageHandler) error {
	if !s.connected {
		return fmt.Errorf("not connected")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[channel] = handler
	return nil
}

func (s *stubBroker) unsubscribe(channel string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.channels, channel)
	return nil
}

// Redis Adapter

type Redis struct {
	stubBroker
}

func NewRedis() *Redis {
	return &Redis{stubBroker{channels: make(map[string]MessageHandler)}}
}

func (r *Redis) Name() string                 { return "redis" }
func (r *Redis) Connect(config *Config) error { return r.connect(config) }
func (r *Redis) Close() error                 { return r.close() }
func (r *Redis) Ping() error                  { return r.ping() }
func (r *Redis) Publish(channel string, msg *Message) error {
	return r.publish(channel, msg)
}
func (r *Redis) Subscribe(channel string, h MessageHandler) error {
	return r.subscribe(channel, h)
}
func (r *Redis) Unsubscribe(channel string) error { return r.unsubscribe(channel) }

// RabbitMQ Adapter

type RabbitMQ struct {
	stubBroker
}

func NewRabbitMQ() *RabbitMQ {
	return &RabbitMQ{stubBroker{channels: make(map[string]MessageHandler)}}
}

func (r *RabbitMQ) Name() string                 { return "rabbitmq" }
func (r *RabbitMQ) Connect(config *Config) error { return r.connect(config) }
func (r *RabbitMQ) Close() error                 { return r.close() }
func (r *RabbitMQ) Ping() error                  { return r.ping() }
func (r *RabbitMQ) Publish(channel string, msg *Message) error {
	return r.publish(channel, msg)
}
func (r *RabbitMQ) Subscribe(channel string, h MessageHandler) error {
	return r.subscribe(channel, h)
}
func (r *RabbitMQ) Unsubscribe(channel string) error { return r.unsubscribe(channel) }

// Kafka Adapter

type Kafka struct {
	stubBroker
}

func NewKafka() *Kafka {
	return &Kafka{stubBroker{channels: make(map[string]MessageHandler)}}
}

func (k *Kafka) Name() string                 { return "kafka" }
func (k *Kafka) Connect(config *Config) error { return k.connect(config) }
func (k *Kafka) Close() error                 { return k.close() }
func (k *Kafka) Ping() error                  { return k.ping() }
func (k *Kafka) Publish(channel string, msg *Message) error {
	return k.publish(channel, msg)
}
func (k *Kafka) Subscribe(channel string, h MessageHandler) error {
	return k.subscribe(channel, h)
}
func (k *Kafka) Unsubscribe(channel string) error { return k.unsubscribe(channel) }

// Broker Manager

type Manager struct {
	brokers map[string]Broker
	mu      sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		brokers: make(map[string]Broker),
	}
}

func (m *Manager) Register(name string, broker Broker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.brokers[name] = broker
}

func (m *Manager) Get(name string) (Broker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	broker, ok := m.brokers[name]
	return broker, ok
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.brokers, name)
}

func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.brokers))
	for name := range m.brokers {
		names = append(names, name)
	}
	return names
}

func (m *Manager) ConnectAll(configs map[string]*Config) error {
	for name, config := range configs {
		broker, ok := m.Get(name)
		if !ok {
			continue
		}
		if err := broker.Connect(config); err != nil {
			return fmt.Errorf("failed to connect to %s: %w", name, err)
		}
	}
	return nil
}

func (m *Manager) CloseAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, broker := range m.brokers {
		if err := broker.Close(); err != nil {
			return fmt.Errorf("failed to close %s: %w", name, err)
		}
	}
	return nil
}

// Message Builder

func NewMessage(channel string, payload []byte) *Message {
	return &Message{
		ID:        generateID(),
		Channel:   channel,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}

// InMemoryBroker is a full in-memory broker with fan-out delivery to multiple
// subscribers per channel, message ordering guarantees, and publish confirmation.

type InMemoryBroker struct {
	config         *Config
	connected      bool
	subscribers    map[string][]MessageHandler
	mu             sync.RWMutex
	published      chan *Message
	publishConfirm chan struct{}
	deadLetter     chan *Message
	deadLetterH    MessageHandler
}

func NewInMemoryBroker() *InMemoryBroker {
	return &InMemoryBroker{
		subscribers:    make(map[string][]MessageHandler),
		published:      make(chan *Message, 256),
		publishConfirm: make(chan struct{}, 256),
		deadLetter:     make(chan *Message, 256),
	}
}

func (b *InMemoryBroker) Name() string { return "memory" }

func (b *InMemoryBroker) Connect(config *Config) error {
	b.config = config
	b.connected = true
	return nil
}

func (b *InMemoryBroker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.connected = false
	b.subscribers = make(map[string][]MessageHandler)
	return nil
}

func (b *InMemoryBroker) Ping() error {
	if !b.connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (b *InMemoryBroker) Publish(channel string, msg *Message) error {
	if !b.connected {
		return fmt.Errorf("not connected")
	}
	if msg == nil {
		return fmt.Errorf("message is nil")
	}
	msg.Channel = channel

	b.mu.RLock()
	handlers := make([]MessageHandler, len(b.subscribers[channel]))
	copy(handlers, b.subscribers[channel])
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(msg); err != nil {
			if b.deadLetterH != nil {
				_ = b.deadLetterH(msg)
			}
		}
	}

	select {
	case b.published <- msg:
	default:
	}
	select {
	case b.publishConfirm <- struct{}{}:
	default:
	}
	return nil
}

func (b *InMemoryBroker) Subscribe(channel string, handler MessageHandler) error {
	if !b.connected {
		return fmt.Errorf("not connected")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[channel] = append(b.subscribers[channel], handler)
	return nil
}

func (b *InMemoryBroker) Unsubscribe(channel string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, channel)
	return nil
}

func (b *InMemoryBroker) SubscriberCount(channel string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers[channel])
}

func (b *InMemoryBroker) SetDeadLetterHandler(handler MessageHandler) {
	b.deadLetterH = handler
}

func (b *InMemoryBroker) DeadLetterChan() <-chan *Message {
	return b.deadLetter
}

func (b *InMemoryBroker) PublishedChan() <-chan *Message {
	return b.published
}

func (b *InMemoryBroker) PublishConfirmChan() <-chan struct{} {
	return b.publishConfirm
}

// Middleware is a function that wraps a MessageHandler, forming a chain of
// interceptors that execute before the final handler.

type Middleware func(next MessageHandler) MessageHandler

// Chain applies a slice of Middleware to a handler, composing them left to right.
// The first middleware in the slice is the outermost (executes first on publish).

func Chain(handler MessageHandler, middlewares ...Middleware) MessageHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// MessageFilter filters messages by channel pattern, payload substring, or
// a custom predicate function.

type MessageFilter struct {
	ChannelPattern string
	PayloadMatch   string
	Predicate      func(msg *Message) bool
}

func NewMessageFilter() *MessageFilter {
	return &MessageFilter{}
}

func (f *MessageFilter) SetChannelPattern(pattern string) {
	f.ChannelPattern = pattern
}

func (f *MessageFilter) SetPayloadMatch(match string) {
	f.PayloadMatch = match
}

func (f *MessageFilter) SetPredicate(fn func(msg *Message) bool) {
	f.Predicate = fn
}

func (f *MessageFilter) Match(msg *Message) bool {
	if msg == nil {
		return false
	}
	if f.ChannelPattern != "" {
		if matched := matchGlob(f.ChannelPattern, msg.Channel); !matched {
			return false
		}
	}
	if f.PayloadMatch != "" {
		if !strings.Contains(string(msg.Payload), f.PayloadMatch) {
			return false
		}
	}
	if f.Predicate != nil && !f.Predicate(msg) {
		return false
	}
	return true
}

func (f *MessageFilter) WrapHandler(handler MessageHandler) MessageHandler {
	return func(msg *Message) error {
		if f.Match(msg) {
			return handler(msg)
		}
		return nil
	}
}

// matchGlob does simple glob matching where * matches any sequence of characters.
func matchGlob(pattern, s string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == s
	}
	parts := strings.Split(pattern, "*")
	if len(parts) == 2 {
		prefix, suffix := parts[0], parts[1]
		if !strings.HasPrefix(s, prefix) {
			return false
		}
		rest := s[len(prefix):]
		return strings.HasSuffix(rest, suffix)
	}
	return strings.Contains(s, strings.ReplaceAll(pattern, "*", ""))
}

// RetryConfig holds retry behavior for publish failures.

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}
}

func (rc *RetryConfig) delay(attempt int) time.Duration {
	delay := time.Duration(float64(rc.BaseDelay) * pow(rc.Multiplier, float64(attempt)))
	if delay > rc.MaxDelay {
		delay = rc.MaxDelay
	}
	return delay
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// PublishWithRetry attempts to publish a message, retrying on failure with
// exponential backoff.

func PublishWithRetry(b Broker, channel string, msg *Message, rc *RetryConfig) error {
	if rc == nil {
		rc = DefaultRetryConfig()
	}
	var lastErr error
	for attempt := 0; attempt < rc.MaxAttempts; attempt++ {
		lastErr = b.Publish(channel, msg)
		if lastErr == nil {
			return nil
		}
		if attempt < rc.MaxAttempts-1 {
			time.Sleep(rc.delay(attempt))
		}
	}
	return fmt.Errorf("publish failed after %d attempts: %w", rc.MaxAttempts, lastErr)
}

// DeadLetterChannel provides a channel-based dead letter queue for messages
// that failed processing.

type DeadLetterChannel struct {
	messages chan *Message
	handler  MessageHandler
	mu       sync.Mutex
}

func NewDeadLetterChannel(bufferSize int) *DeadLetterChannel {
	return &DeadLetterChannel{
		messages: make(chan *Message, bufferSize),
	}
}

func (dlc *DeadLetterChannel) Handler() MessageHandler {
	return func(msg *Message) error {
		dlc.mu.Lock()
		defer dlc.mu.Unlock()
		select {
		case dlc.messages <- msg:
			return nil
		default:
			return fmt.Errorf("dead letter queue full")
		}
	}
}

func (dlc *DeadLetterChannel) Drain(handler MessageHandler) {
	dlc.mu.Lock()
	dlc.handler = handler
	dlc.mu.Unlock()

	go func() {
		for msg := range dlc.messages {
			dlc.mu.Lock()
			h := dlc.handler
			dlc.mu.Unlock()
			if h != nil {
				_ = h(msg)
			}
		}
	}()
}

func (dlc *DeadLetterChannel) Messages() <-chan *Message {
	return dlc.messages
}

func (dlc *DeadLetterChannel) Len() int {
	return len(dlc.messages)
}

func (dlc *DeadLetterChannel) Close() {
	close(dlc.messages)
}

// ConnectionPool manages a pool of Broker connections with round-robin
// selection, health checking, and pool-level metrics.

type PoolMetrics struct {
	Total        int64
	Healthy      int64
	Unhealthy    int64
	NextCalls    int64
	HealthChecks int64
}

type ConnectionPool struct {
	brokers     []Broker
	current     uint64
	healthy     []bool
	createdAt   []time.Time
	maxLifetime time.Duration
	mu          sync.RWMutex
	checkFn     func(Broker) bool
	metrics     PoolMetrics
	stopCh      chan struct{}
	running     bool
}

func NewConnectionPool(brokers ...Broker) *ConnectionPool {
	healthy := make([]bool, len(brokers))
	createdAt := make([]time.Time, len(brokers))
	now := time.Now()
	for i := range healthy {
		healthy[i] = true
		createdAt[i] = now
	}
	return &ConnectionPool{
		brokers:   brokers,
		healthy:   healthy,
		createdAt: createdAt,
		checkFn:   func(b Broker) bool { return b.Ping() == nil },
		metrics:   PoolMetrics{Total: int64(len(brokers)), Healthy: int64(len(brokers))},
		stopCh:    make(chan struct{}),
	}
}

func (cp *ConnectionPool) SetHealthCheck(fn func(Broker) bool) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.checkFn = fn
}

func (cp *ConnectionPool) SetMaxLifetime(d time.Duration) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.maxLifetime = d
	now := time.Now()
	for i := range cp.brokers {
		if d <= 0 {
			cp.healthy[i] = true
		} else if now.Sub(cp.createdAt[i]) > d {
			cp.healthy[i] = false
		}
	}
	cp.updateMetrics()
}

func (cp *ConnectionPool) StartHealthCheck(interval time.Duration) {
	cp.mu.Lock()
	if cp.running {
		cp.mu.Unlock()
		return
	}
	cp.running = true
	stopCh := cp.stopCh
	cp.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cp.CheckHealth()
			case <-stopCh:
				return
			}
		}
	}()
}

func (cp *ConnectionPool) StopHealthCheck() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if cp.running {
		close(cp.stopCh)
		cp.running = false
		cp.stopCh = make(chan struct{})
	}
}

func (cp *ConnectionPool) Next() Broker {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	cp.metrics.NextCalls++

	if len(cp.brokers) == 0 {
		return nil
	}
	start := atomic.AddUint64(&cp.current, 1) - 1
	for i := uint64(0); i < uint64(len(cp.brokers)); i++ {
		idx := (start + i) % uint64(len(cp.brokers))
		if cp.healthy[idx] {
			return cp.brokers[idx]
		}
	}
	return nil
}

func (cp *ConnectionPool) Len() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.brokers)
}

func (cp *ConnectionPool) HealthyCount() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	count := 0
	for _, h := range cp.healthy {
		if h {
			count++
		}
	}
	return count
}

func (cp *ConnectionPool) CheckHealth() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.metrics.HealthChecks++

	for i, b := range cp.brokers {
		healthy := cp.checkFn(b)
		if healthy && cp.maxLifetime > 0 && time.Since(cp.createdAt[i]) > cp.maxLifetime {
			healthy = false
		}
		cp.healthy[i] = healthy
	}
	cp.updateMetrics()
}

func (cp *ConnectionPool) SetHealthy(index int, healthy bool) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	if index >= 0 && index < len(cp.healthy) {
		cp.healthy[index] = healthy
		cp.updateMetrics()
	}
}

func (cp *ConnectionPool) PoolMetrics() PoolMetrics {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.metrics
}

func (cp *ConnectionPool) CloseAll() error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	for _, b := range cp.brokers {
		if err := b.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (cp *ConnectionPool) updateMetrics() {
	healthy := int64(0)
	for _, h := range cp.healthy {
		if h {
			healthy++
		}
	}
	cp.metrics.Total = int64(len(cp.brokers))
	cp.metrics.Healthy = healthy
	cp.metrics.Unhealthy = int64(len(cp.brokers)) - healthy
}

// Metrics tracks publish/receive/error counts and per-channel subscriber counts.

type Metrics struct {
	published   int64
	received    int64
	errors      int64
	subscribers map[string]int64
	mu          sync.RWMutex
}

func NewMetrics() *Metrics {
	return &Metrics{
		subscribers: make(map[string]int64),
	}
}

func (m *Metrics) IncPublished() { atomic.AddInt64(&m.published, 1) }
func (m *Metrics) IncReceived()  { atomic.AddInt64(&m.received, 1) }
func (m *Metrics) IncErrors()    { atomic.AddInt64(&m.errors, 1) }

func (m *Metrics) PublishedCount() int64 { return atomic.LoadInt64(&m.published) }
func (m *Metrics) ReceivedCount() int64  { return atomic.LoadInt64(&m.received) }
func (m *Metrics) ErrorsCount() int64    { return atomic.LoadInt64(&m.errors) }

func (m *Metrics) SetSubscriberCount(channel string, count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers[channel] = count
}

func (m *Metrics) SubscriberCount(channel string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.subscribers[channel]
}

func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.published, 0)
	atomic.StoreInt64(&m.received, 0)
	atomic.StoreInt64(&m.errors, 0)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers = make(map[string]int64)
}

// MetricsBroker wraps a Broker and records metrics on every operation.

type MetricsBroker struct {
	broker  Broker
	metrics *Metrics
}

func NewMetricsBroker(broker Broker, metrics *Metrics) *MetricsBroker {
	return &MetricsBroker{broker: broker, metrics: metrics}
}

func (mb *MetricsBroker) Name() string { return mb.broker.Name() }

func (mb *MetricsBroker) Connect(config *Config) error { return mb.broker.Connect(config) }
func (mb *MetricsBroker) Close() error                 { return mb.broker.Close() }
func (mb *MetricsBroker) Ping() error                  { return mb.broker.Ping() }

func (mb *MetricsBroker) Publish(channel string, msg *Message) error {
	mb.metrics.IncPublished()
	err := mb.broker.Publish(channel, msg)
	if err != nil {
		mb.metrics.IncErrors()
	}
	return err
}

func (mb *MetricsBroker) Subscribe(channel string, handler MessageHandler) error {
	wrapped := func(msg *Message) error {
		mb.metrics.IncReceived()
		return handler(msg)
	}
	err := mb.broker.Subscribe(channel, wrapped)
	if err != nil {
		mb.metrics.IncErrors()
		return err
	}
	mb.metrics.SetSubscriberCount(channel, mb.metrics.SubscriberCount(channel)+1)
	return nil
}

func (mb *MetricsBroker) Unsubscribe(channel string) error {
	mb.metrics.SetSubscriberCount(channel, 0)
	return mb.broker.Unsubscribe(channel)
}

func (mb *MetricsBroker) Metrics() *Metrics { return mb.metrics }

var (
	ErrNotConnected   = errors.New("not connected")
	ErrPoolEmpty      = errors.New("broker pool is empty")
	ErrMessageNil     = errors.New("message is nil")
	ErrDeadLetterFull = errors.New("dead letter queue full")
)
