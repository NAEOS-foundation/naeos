//go:build !nobroker

package broker

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
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
		slog.Error("nats connect failed", "host", config.Host, "port", config.Port, "error", err)
		return fmt.Errorf("connect to NATS: %w", err)
	}

	slog.Info("nats connected", "host", config.Host, "port", config.Port)
	n.conn = conn
	return nil
}

func (n *RealNATS) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for channel, sub := range n.subscribers {
		_ = sub.Unsubscribe()
		delete(n.subscribers, channel)
	}

	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

func (n *RealNATS) Ping() error {
	if n.conn == nil {
		return naeoserr.ErrNotConnected
	}
	if n.conn.IsClosed() {
		return naeoserr.Wrap(naeoserr.ErrNetwork, "connection closed", nil)
	}
	return nil
}

func (n *RealNATS) Publish(channel string, msg *Message) error {
	if n.conn == nil {
		return naeoserr.ErrNotConnected
	}

	data := msg.Payload
	if data == nil {
		data = []byte{}
	}

	return n.conn.Publish(channel, data)
}

func (n *RealNATS) Subscribe(channel string, handler MessageHandler) error {
	if n.conn == nil {
		return naeoserr.ErrNotConnected
	}

	sub, err := n.conn.Subscribe(channel, func(m *nats.Msg) {
		msg := &Message{
			ID:        generateID(),
			Channel:   m.Subject,
			Payload:   m.Data,
			Timestamp: time.Now(),
		}
		_ = handler(msg)
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
		_ = sub.Unsubscribe()
		delete(n.subscribers, channel)
	}
	return nil
}
