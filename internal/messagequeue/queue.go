// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package messagequeue

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Message struct {
	ID         string
	Topic      string
	Payload    any
	Timestamp  time.Time
	Retries    int
	MaxRetries int
}

type MessageHandler func(msg *Message) error

type Queue struct {
	name     string
	messages chan *Message
	handler  MessageHandler
	running  bool
	mu       sync.RWMutex
	stats    QueueStats
	dead     []*Message
	maxDead  int
	metrics  *QueueMetrics
}

type QueueStats struct {
	Published    int64
	Consumed     int64
	Failed       int64
	DeadLettered int64
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
		name:     name,
		messages: make(chan *Message, capacity),
		maxDead:  100,
		metrics:  &QueueMetrics{},
	}
}

func (q *Queue) Publish(msg *Message) error {
	msg.Timestamp = time.Now()
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 3
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
	for msg := range q.messages {
		q.mu.RLock()
		running := q.running
		handler := q.handler
		q.mu.RUnlock()

		if !running {
			return
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
			atomic.AddInt64(&q.stats.Consumed, 1)
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
	return QueueStats{
		Published:    atomic.LoadInt64(&q.stats.Published),
		Consumed:     atomic.LoadInt64(&q.stats.Consumed),
		Failed:       atomic.LoadInt64(&q.stats.Failed),
		DeadLettered: atomic.LoadInt64(&q.stats.DeadLettered),
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
	ErrQueueFull     = &QueueError{"queue is full"}
	ErrTopicNotFound = &QueueError{"topic not found"}
)

type QueueError struct {
	msg string
}

func (e *QueueError) Error() string {
	return e.msg
}

func NewMessage(topic string, payload any) *Message {
	return &Message{
		ID:         generateID(),
		Topic:      topic,
		Payload:    payload,
		Timestamp:  time.Now(),
		MaxRetries: 3,
	}
}

func generateID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}
