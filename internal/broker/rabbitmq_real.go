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
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		"guest", config.Password, config.Host, config.Port)
	if config.Password == "" {
		url = fmt.Sprintf("amqp://guest:guest@%s:%d/", config.Host, config.Port)
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		slog.Error("rabbitmq connect failed", "host", config.Host, "port", config.Port, "error", err)
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "connect to rabbitmq")
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		slog.Error("rabbitmq channel open failed", "error", err)
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "open channel")
	}

	r.mu.Lock()
	r.conn = conn
	r.channel = ch
	r.config = config
	r.mu.Unlock()

	slog.Info("rabbitmq connected", "host", config.Host, "port", config.Port)
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
	r.mu.RLock()
	conn := r.conn
	r.mu.RUnlock()
	if conn == nil {
		return naeoserr.ErrNotConnected
	}
	if conn.IsClosed() {
		return naeoserr.Wrap(naeoserr.ErrNetwork, "connection closed", nil)
	}
	return nil
}

func (r *RealRabbitMQ) Publish(channel string, msg *Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel == nil {
		return naeoserr.ErrNotConnected
	}
	queue, ok := r.queues[channel]
	if !ok {
		var err error
		queue, err = r.channel.QueueDeclare(
			channel, true, false, false, false, nil,
		)
		if err != nil {
			slog.Error("rabbitmq declare queue failed", "channel", channel, "error", err)
			return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "declare queue %s", channel)
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
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel == nil {
		return naeoserr.ErrNotConnected
	}

	queue, ok := r.queues[channel]
	if !ok {
		var err error
		queue, err = r.channel.QueueDeclare(
			channel, true, false, false, false, nil,
		)
		if err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "declare queue %s", channel)
		}
		r.queues[channel] = queue
	}

	deliveries, err := r.channel.Consume(
		queue.Name, "", false, false, false, false, nil,
	)
	if err != nil {
		slog.Error("rabbitmq consume failed", "channel", channel, "error", err)
		return naeoserr.Wrapf(err, naeoserr.ErrNetwork, "consume from %s", channel)
	}

	r.consumers[channel] = deliveries

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
