package messagequeue

import (
	"sync"
	"time"
)

// Message

type Message struct {
	ID        string
	Topic     string
	Payload   any
	Timestamp time.Time
	Retries   int
	MaxRetries int
}

type MessageHandler func(msg *Message) error

// Queue

type Queue struct {
	name     string
	messages chan *Message
	handler  MessageHandler
	running  bool
	mu       sync.RWMutex
}

func NewQueue(name string, capacity int) *Queue {
	return &Queue{
		name:     name,
		messages: make(chan *Message, capacity),
	}
}

func (q *Queue) Publish(msg *Message) error {
	msg.Timestamp = time.Now()
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 3
	}

	select {
	case q.messages <- msg:
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

		if err := handler(msg); err != nil {
			msg.Retries++
			if msg.Retries < msg.MaxRetries {
				q.messages <- msg
			}
		}
	}
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

// Topic

type Topic struct {
	name    string
	queues  map[string]*Queue
	mu      sync.RWMutex
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

// Broker

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

// Errors

var (
	ErrQueueFull    = &QueueError{"queue is full"}
	ErrTopicNotFound = &QueueError{"topic not found"}
)

type QueueError struct {
	msg string
}

func (e *QueueError) Error() string {
	return e.msg
}

// Message Builder

func NewMessage(topic string, payload any) *Message {
	return &Message{
		ID:        generateID(),
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
		MaxRetries: 3,
	}
}

func generateID() string {
	return time.Now().Format("20060102150405.000000000")
}
