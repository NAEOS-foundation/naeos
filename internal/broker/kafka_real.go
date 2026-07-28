//go:build !nobroker

package broker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type RealKafka struct {
	reader *kafka.Reader
	writer *kafka.Writer
	config *Config
	mu     sync.RWMutex
	subCtx map[string]context.CancelFunc
	subMu  sync.Mutex
}

func NewRealKafka() *RealKafka {
	return &RealKafka{
		subCtx: make(map[string]context.CancelFunc),
	}
}

func (k *RealKafka) Name() string {
	return "kafka"
}

func (k *RealKafka) Connect(config *Config) error {
	broker := fmt.Sprintf("%s:%d", config.Host, config.Port)

	k.mu.Lock()
	k.config = config
	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		Compression:  compress.Snappy,
	}
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    "default",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	k.mu.Unlock()

	slog.Info("kafka connected", "host", config.Host, "port", config.Port)
	return nil
}

func (k *RealKafka) Close() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.reader != nil {
		k.reader.Close()
	}
	if k.writer != nil {
		return k.writer.Close()
	}
	return nil
}

func (k *RealKafka) Ping() error {
	k.mu.RLock()
	writer := k.writer
	k.mu.RUnlock()
	if writer == nil {
		return naeoserr.ErrNotConnected
	}
	return nil
}

func (k *RealKafka) Publish(channel string, msg *Message) error {
	k.mu.RLock()
	writer := k.writer
	config := k.config
	k.mu.RUnlock()
	if writer == nil {
		return naeoserr.ErrNotConnected
	}

	data := msg.Payload
	if data == nil {
		data = []byte{}
	}

	timeout := 30 * time.Second
	if config != nil && config.Timeout > 0 {
		timeout = config.Timeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return writer.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(channel),
			Value: data,
			Time:  time.Now(),
		},
	)
}

func (k *RealKafka) Subscribe(channel string, handler MessageHandler) error {
	k.mu.RLock()
	config := k.config
	k.mu.RUnlock()
	if config == nil {
		return naeoserr.ErrNotConnected
	}

	broker := fmt.Sprintf("%s:%d", config.Host, config.Port)

	k.subMu.Lock()
	if _, ok := k.subCtx[channel]; ok {
		k.subMu.Unlock()
		return naeoserr.New(naeoserr.ErrInternal, fmt.Sprintf("already subscribed to %s", channel))
	}
	ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec // cancel stored in subCtx, called by Unsubscribe
	k.subCtx[channel] = cancel
	k.subMu.Unlock()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    channel,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	go func() {
		defer reader.Close()
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				if strings.Contains(err.Error(), "reader is closed") || ctx.Err() != nil {
					return
				}
				continue
			}
			msg := &Message{
				ID:        generateID(),
				Channel:   channel,
				Payload:   m.Value,
				Timestamp: m.Time,
			}
			_ = handler(msg)
		}
	}()

	return nil
}

func (k *RealKafka) Unsubscribe(channel string) error {
	k.subMu.Lock()
	cancel, ok := k.subCtx[channel]
	if ok {
		delete(k.subCtx, channel)
	}
	k.subMu.Unlock()
	if ok {
		cancel()
	}
	return nil
}
