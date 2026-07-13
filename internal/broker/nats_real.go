package broker

import (
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type RealNATS struct {
	conn        *nats.Conn
	config      *Config
	subscribers map[string]*nats.Subscription
	mu          sync.RWMutex
}

func NewRealNATS() *RealNATS {
	return &RealNATS{
		subscribers: make(map[string]*nats.Subscription),
	}
}

func (n *RealNATS) Name() string {
	return "nats"
}

func (n *RealNATS) Connect(config *Config) error {
	n.config = config
	url := fmt.Sprintf("nats://%s:%d", config.Host, config.Port)
	if config.Password != "" {
		url = fmt.Sprintf("nats://:%s@%s:%d", config.Password, config.Host, config.Port)
	}

	opts := []nats.Option{}
	if config.Timeout > 0 {
		opts = append(opts, nats.Timeout(config.Timeout))
	}

	conn, err := nats.Connect(url, opts...)
	if err != nil {
		return fmt.Errorf("connect to NATS: %w", err)
	}

	n.conn = conn
	return nil
}

func (n *RealNATS) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for channel, sub := range n.subscribers {
		sub.Unsubscribe()
		delete(n.subscribers, channel)
	}

	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

func (n *RealNATS) Ping() error {
	if n.conn == nil {
		return fmt.Errorf("not connected")
	}
	if n.conn.IsClosed() {
		return fmt.Errorf("connection closed")
	}
	return nil
}

func (n *RealNATS) Publish(channel string, msg *Message) error {
	if n.conn == nil {
		return fmt.Errorf("not connected")
	}

	data := msg.Payload
	if data == nil {
		data = []byte{}
	}

	return n.conn.Publish(channel, data)
}

func (n *RealNATS) Subscribe(channel string, handler MessageHandler) error {
	if n.conn == nil {
		return fmt.Errorf("not connected")
	}

	sub, err := n.conn.Subscribe(channel, func(m *nats.Msg) {
		msg := &Message{
			ID:        generateID(),
			Channel:   m.Subject,
			Payload:   m.Data,
			Timestamp: time.Now(),
		}
		handler(msg)
	})
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", channel, err)
	}

	n.mu.Lock()
	n.subscribers[channel] = sub
	n.mu.Unlock()

	return nil
}

func (n *RealNATS) Unsubscribe(channel string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if sub, ok := n.subscribers[channel]; ok {
		sub.Unsubscribe()
		delete(n.subscribers, channel)
	}
	return nil
}
