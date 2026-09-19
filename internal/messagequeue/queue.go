// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package messagequeue

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// IdempotencyStore provides deduplication for message processing.
type IdempotencyStore interface {
	// Dedup checks if a message with the given idempotency key has already been processed.
	// If the key exists, the message is a duplicate and should be skipped.
	// The key is registered for the specified duration.
	Dedup(key string, ttl time.Duration) (isDuplicate bool, err error)
	// Cleanup removes expired entries from the store.
	Cleanup()
}

// InMemoryIdempotencyStore is a thread-safe in-memory implementation of IdempotencyStore.
type InMemoryIdempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]time.Time
	ttl     time.Duration
}

type idempotencyEntry struct {
	key       string
	processedAt time.Time
}

// NewInMemoryIdempotencyStore creates a new in-memory idempotency store with the given default TTL.
func NewInMemoryIdempotencyStore(defaultTTL time.Duration) *InMemoryIdempotencyStore {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute
	}
	return &InMemoryIdempotencyStore{
		entries: make(map[string]time.Time),
		ttl:     defaultTTL,
	}
}

// Dedup checks if the key has already been processed. If not, registers it.
func (s *InMemoryIdempotencyStore) Dedup(key string, ttl time.Duration) (isDuplicate bool, err error) {
	if key == "" {
		return false, nil
	}
	if ttl <= 0 {
		ttl = s.ttl
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[key]; exists {
		return true, nil
	}

	s.entries[key] = time.Now().Add(ttl)
	return false, nil
}

// Cleanup removes expired entries from the store.
func (s *InMemoryIdempotencyStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, expiry := range s.entries {
		if now.After(expiry) {
			delete(s.entries, key)
		}
	}
}

// IdempotencyStats tracks idempotency-related metrics.
type IdempotencyStats struct {
	DedupHits   int64
	DedupMisses int64
	TotalKeys   int64
}

type Message struct {
	ID               string
	IdempotencyKey   string
	Topic            string
	Payload          any
	Timestamp        time.Time
	Retries          int
	MaxRetries       int
	Processed        bool
	ProcessedAt      time.Time
}

type MessageHandler func(msg *Message) error

type Queue struct {
	name            string
	messages        chan *Message
	handler         MessageHandler
	running         bool
	mu              sync.RWMutex
	stats           QueueStats
	dead            []*Message
	maxDead         int
	metrics         *QueueMetrics
	idempotencyStore IdempotencyStore
	idempotencyTTL  time.Duration
	idempotencyStats IdempotencyStats
}

type QueueStats struct {
	Published       int64
	Consumed        int64
	Failed          int64
	DeadLettered    int64
	DedupSkipped    int64
	IdempotencyHits int64
	IdempotencyMisses int64
}

type QueueMetrics struct {
	QueueDepth   int64
	ProcessRate  float64
	AvgLatencyMs float64
	TotalLatency int64
	LatencyCount int64
}

func NewQueue(name string, capacity int) *Queue {
	return &Queue{
		name:            name,
		messages:        make(chan *Message, capacity),
		maxDead:        100,
		metrics:        &QueueMetrics{},
		idempotencyTTL:  5 * time.Minute,
	}
}

// WithIdempotencyStore configures the queue with an idempotency store for deduplication.
func (q *Queue) WithIdempotencyStore(store IdempotencyStore, ttl time.Duration) *Queue {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.idempotencyStore = store
	if ttl > 0 {
		q.idempotencyTTL = ttl
	}
	return q
}

func (q *Queue) Publish(msg *Message) error {
	msg.Timestamp = time.Now()
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 3
	}

	// Check idempotency before publishing
	if msg.IdempotencyKey != "" && q.idempotencyStore != nil {
		isDuplicate, err := q.idempotencyStore.Dedup(msg.IdempotencyKey, q.idempotencyTTL)
		if err != nil {
			return err
		}
		if isDuplicate {
			q.mu.Lock()
			q.stats.DedupSkipped++
			q.stats.IdempotencyHits++
			q.mu.Unlock()
			return ErrDuplicateMessage
		}
		q.mu.Lock()
		q.stats.IdempotencyMisses++
		q.mu.Unlock()
	}

	select {
	case q.messages <- msg:
		atomic.AddInt64(&q.stats.Published, 1)
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Subscribe(handler MessageHandler) {
	q.mu.Lock()
	q.handler = handler
	q.running = true
	q.mu.Unlock()

	go q.consume()
}

func (q *Queue) consume() {
	// Start periodic cleanup for idempotency store
	var cleanupDone chan struct{}
	if q.idempotencyStore != nil {
		cleanupDone = make(chan struct{})
		go func() {
			ticker := time.NewTicker(q.idempotencyTTL)
			defer ticker.Stop()
			for {
				select {
				case <-cleanupDone:
					return
				case <-ticker.C:
					q.idempotencyStore.Cleanup()
				}
			}
		}()
	}
	defer func() {
		if cleanupDone != nil {
			close(cleanupDone)
		}
	}()

	for msg := range q.messages {
		q.mu.RLock()
		running := q.running
		handler := q.handler
		q.mu.RUnlock()

		if !running {
			return
		}

		// Check if message was already processed (idempotency check on consume)
		if msg.Processed {
			atomic.AddInt64(&q.stats.DedupSkipped, 1)
			atomic.AddInt64(&q.stats.IdempotencyHits, 1)
			continue
		}

		start := time.Now()
		if err := handler(msg); err != nil {
			msg.Retries++
			if msg.Retries < msg.MaxRetries {
				q.messages <- msg
			} else {
				q.addToDead(msg)
				atomic.AddInt64(&q.stats.DeadLettered, 1)
			}
			atomic.AddInt64(&q.stats.Failed, 1)
		} else {
			msg.Processed = true
			msg.ProcessedAt = time.Now()
			atomic.AddInt64(&q.stats.Consumed, 1)
			atomic.AddInt64(&q.stats.IdempotencyMisses, 1)
		}

		elapsed := time.Since(start).Milliseconds()
		q.updateLatency(elapsed)
	}
}

func (q *Queue) updateLatency(ms int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.metrics.TotalLatency += ms
	q.metrics.LatencyCount++
	if q.metrics.LatencyCount > 0 {
		q.metrics.AvgLatencyMs = float64(q.metrics.TotalLatency) / float64(q.metrics.LatencyCount)
	}
}

func (q *Queue) addToDead(msg *Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.dead) >= q.maxDead {
		q.dead = q.dead[1:]
	}
	q.dead = append(q.dead, msg)
}

func (q *Queue) DeadLetters() []*Message {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]*Message, len(q.dead))
	copy(out, q.dead)
	return out
}

func (q *Queue) Stop() {
	q.mu.Lock()
	q.running = false
	q.mu.Unlock()
	close(q.messages)
}

func (q *Queue) Len() int {
	return len(q.messages)
}

func (q *Queue) Name() string {
	return q.name
}

func (q *Queue) Stats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return QueueStats{
		Published:         atomic.LoadInt64(&q.stats.Published),
		Consumed:          atomic.LoadInt64(&q.stats.Consumed),
		Failed:            atomic.LoadInt64(&q.stats.Failed),
		DeadLettered:      atomic.LoadInt64(&q.stats.DeadLettered),
		DedupSkipped:      atomic.LoadInt64(&q.stats.DedupSkipped),
		IdempotencyHits:   atomic.LoadInt64(&q.stats.IdempotencyHits),
		IdempotencyMisses: atomic.LoadInt64(&q.stats.IdempotencyMisses),
	}
}

func (q *Queue) Metrics() QueueMetrics {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return QueueMetrics{
		QueueDepth:   int64(len(q.messages)),
		AvgLatencyMs: q.metrics.AvgLatencyMs,
		TotalLatency: q.metrics.TotalLatency,
		LatencyCount: q.metrics.LatencyCount,
	}
}

func (q *Queue) IdempotencyStats() IdempotencyStats {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return IdempotencyStats{
		DedupHits:   atomic.LoadInt64(&q.stats.IdempotencyHits),
		DedupMisses: atomic.LoadInt64(&q.stats.IdempotencyMisses),
		TotalKeys:   atomic.LoadInt64(&q.stats.DedupSkipped),
	}
}

type Topic struct {
	name   string
	queues map[string]*Queue
	mu     sync.RWMutex
}

func NewTopic(name string) *Topic {
	return &Topic{
		name:   name,
		queues: make(map[string]*Queue),
	}
}

func (t *Topic) Subscribe(name string, handler MessageHandler) *Queue {
	t.mu.Lock()
	defer t.mu.Unlock()

	queue := NewQueue(name, 100)
	queue.Subscribe(handler)
	t.queues[name] = queue
	return queue
}

func (t *Topic) Publish(msg *Message) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, queue := range t.queues {
		if err := queue.Publish(msg); err != nil {
			return err
		}
	}
	return nil
}

func (t *Topic) Unsubscribe(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if queue, ok := t.queues[name]; ok {
		queue.Stop()
		delete(t.queues, name)
	}
}

func (t *Topic) Subscribers() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.queues)
}

type Broker struct {
	topics map[string]*Topic
	mu     sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		topics: make(map[string]*Topic),
	}
}

func (b *Broker) CreateTopic(name string) *Topic {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.topics[name]; exists {
		return b.topics[name]
	}

	topic := NewTopic(name)
	b.topics[name] = topic
	return topic
}

func (b *Broker) GetTopic(name string) (*Topic, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	topic, ok := b.topics[name]
	return topic, ok
}

func (b *Broker) DeleteTopic(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topic, ok := b.topics[name]; ok {
		topic.mu.Lock()
		for _, queue := range topic.queues {
			queue.Stop()
		}
		topic.mu.Unlock()
		delete(b.topics, name)
	}
}

func (b *Broker) Publish(topic string, msg *Message) error {
	t, ok := b.GetTopic(topic)
	if !ok {
		return ErrTopicNotFound
	}
	return t.Publish(msg)
}

func (b *Broker) Subscribe(topic, queue string, handler MessageHandler) error {
	t, ok := b.GetTopic(topic)
	if !ok {
		t = b.CreateTopic(topic)
	}
	t.Subscribe(queue, handler)
	return nil
}

func (b *Broker) ListTopics() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	names := make([]string, 0, len(b.topics))
	for name := range b.topics {
		names = append(names, name)
	}
	return names
}

func (b *Broker) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, topic := range b.topics {
		topic.mu.Lock()
		for _, queue := range topic.queues {
			queue.Stop()
		}
		topic.mu.Unlock()
	}
}

var (
	ErrQueueFull          = &QueueError{"queue is full"}
	ErrTopicNotFound      = &QueueError{"topic not found"}
	ErrDuplicateMessage   = &QueueError{"duplicate message: idempotency key already processed"}
)

type QueueError struct {
	msg string
}

func (e *QueueError) Error() string {
	return e.msg
}

func NewMessage(topic string, payload any) *Message {
	return &Message{
		ID:           generateID(),
		Topic:        topic,
		Payload:      payload,
		Timestamp:    time.Now(),
		MaxRetries:   3,
	}
}

// NewIdempotentMessage creates a new message with an idempotency key for deduplication.
func NewIdempotentMessage(topic string, payload any, idempotencyKey string) *Message {
	return &Message{
		ID:               generateID(),
		IdempotencyKey:   idempotencyKey,
		Topic:            topic,
		Payload:          payload,
		Timestamp:        time.Now(),
		MaxRetries:       3,
	}
}

func generateID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}
