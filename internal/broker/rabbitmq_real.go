//go:build !nobroker

package broker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type RealRabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	config    *Config
	queues    map[string]amqp.Queue
	consumers map[string]<-chan amqp.Delivery
	mu        sync.RWMutex
}

func NewRealRabbitMQ() *RealRabbitMQ {
	return &RealRabbitMQ{
		queues:    make(map[string]amqp.Queue),
		consumers: make(map[string]<-chan amqp.Delivery),
	}
}

func (r *RealRabbitMQ) Name() string {
	return "rabbitmq"
}

func (r *RealRabbitMQ) Connect(config *Config) error {
	r.config = config
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		"guest", config.Password, config.Host, config.Port)
	if config.Password == "" {
		url = fmt.Sprintf("amqp://guest:guest@%s:%d/", config.Host, config.Port)
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		slog.Error("rabbitmq connect failed", "host", config.Host, "port", config.Port, "error", err)
		return fmt.Errorf("connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		slog.Error("rabbitmq channel open failed", "error", err)
		return fmt.Errorf("open channel: %w", err)
	}

	slog.Info("rabbitmq connected", "host", config.Host, "port", config.Port)
	r.conn = conn
	r.channel = ch
	return nil
}

func (r *RealRabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for queue := range r.consumers {
		delete(r.consumers, queue)
	}

	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RealRabbitMQ) Ping() error {
	if r.conn == nil {
		return naeoserr.ErrNotConnected
	}
	if r.conn.IsClosed() {
		return naeoserr.Wrap(naeoserr.ErrNetwork, "connection closed", nil)
	}
	return nil
}

func (r *RealRabbitMQ) Publish(channel string, msg *Message) error {
	if r.channel == nil {
		return naeoserr.ErrNotConnected
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	queue, ok := r.queues[channel]
	if !ok {
		var err error
		queue, err = r.channel.QueueDeclare(
			channel, true, false, false, false, nil,
		)
		if err != nil {
			slog.Error("rabbitmq declare queue failed", "channel", channel, "error", err)
			return fmt.Errorf("declare queue %s: %w", channel, err)
		}
		r.queues[channel] = queue
	}

	data := msg.Payload
	if data == nil {
		data = []byte{}
	}

	return r.channel.PublishWithContext(
		context.Background(),
		"", queue.Name, false, false,
		amqp.Publishing{
			ContentType: "application/octet-stream",
			Body:        data,
			Timestamp:   time.Now(),
		},
	)
}

func (r *RealRabbitMQ) Subscribe(channel string, handler MessageHandler) error {
	if r.channel == nil {
		return naeoserr.ErrNotConnected
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	queue, ok := r.queues[channel]
	if !ok {
		var err error
		queue, err = r.channel.QueueDeclare(
			channel, true, false, false, false, nil,
		)
		if err != nil {
			return fmt.Errorf("declare queue %s: %w", channel, err)
		}
		r.queues[channel] = queue
	}

	deliveries, err := r.channel.Consume(
		queue.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		slog.Error("rabbitmq consume failed", "channel", channel, "error", err)
		return fmt.Errorf("consume from %s: %w", channel, err)
	}

	r.mu.Lock()
	r.consumers[channel] = deliveries
	r.mu.Unlock()

	go func() {
		for d := range deliveries {
			msg := &Message{
				ID:        generateID(),
				Channel:   channel,
				Payload:   d.Body,
				Timestamp: d.Timestamp,
			}
			_ = handler(msg)
			_ = d.Ack(false)
		}
	}()

	return nil
}

func (r *RealRabbitMQ) Unsubscribe(channel string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.consumers, channel)
	return nil
}
