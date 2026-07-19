package broker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

type RealKafka struct {
	reader *kafka.Reader
	writer *kafka.Writer
	config *Config
	mu     sync.RWMutex
}

func NewRealKafka() *RealKafka {
	return &RealKafka{}
}

func (k *RealKafka) Name() string {
	return "kafka"
}

func (k *RealKafka) Connect(config *Config) error {
	k.config = config
	broker := fmt.Sprintf("%s:%d", config.Host, config.Port)

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
	if k.writer == nil {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (k *RealKafka) Publish(channel string, msg *Message) error {
	if k.writer == nil {
		return fmt.Errorf("not connected")
	}

	data := msg.Payload
	if data == nil {
		data = []byte{}
	}

	return k.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(channel),
			Value: data,
			Time:  time.Now(),
		},
	)
}

func (k *RealKafka) Subscribe(channel string, handler MessageHandler) error {
	if k.config == nil {
		return fmt.Errorf("not connected")
	}

	broker := fmt.Sprintf("%s:%d", k.config.Host, k.config.Port)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    channel,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	go func() {
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				if strings.Contains(err.Error(), "reader is closed") {
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
	return nil
}
