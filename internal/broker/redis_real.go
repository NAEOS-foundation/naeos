//go:build !nobroker

package broker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type RealRedis struct {
	client      *redis.Client
	config      *Config
	subscribers map[string]*redis.PubSub
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

func NewRealRedis() *RealRedis {
	return &RealRedis{
		subscribers: make(map[string]*redis.PubSub),
	}
}

func (r *RealRedis) Name() string {
	return "redis"
}

func (r *RealRedis) Connect(config *Config) error {
	r.config = config
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  config.Timeout,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		slog.Error("redis connect failed", "host", config.Host, "port", config.Port, "error", err)
		return fmt.Errorf("connect to redis: %w", err)
	}

	slog.Info("redis connected", "host", config.Host, "port", config.Port)
	r.client = rdb
	_, r.cancel = context.WithCancel(context.Background())
	return nil
}

func (r *RealRedis) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for channel, sub := range r.subscribers {
		_ = sub.Close()
		delete(r.subscribers, channel)
	}

	if r.cancel != nil {
		r.cancel()
	}
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *RealRedis) Ping() error {
	if r.client == nil {
		return naeoserr.ErrNotConnected
	}
	return r.client.Ping(context.Background()).Err()
}

func (r *RealRedis) Publish(channel string, msg *Message) error {
	if r.client == nil {
		return naeoserr.ErrNotConnected
	}

	data := msg.Payload
	if data == nil {
		data = []byte{}
	}

	return r.client.Publish(context.Background(), channel, data).Err()
}

func (r *RealRedis) Subscribe(channel string, handler MessageHandler) error {
	if r.client == nil {
		return naeoserr.ErrNotConnected
	}

	sub := r.client.Subscribe(context.Background(), channel)

	if err := sub.Ping(context.Background()); err != nil {
		_ = sub.Close()
		return fmt.Errorf("subscribe to %s: %w", channel, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		ch := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-ch:
				if !ok {
					return
				}
				msg := &Message{
					ID:        generateID(),
					Channel:   m.Channel,
					Payload:   []byte(m.Payload),
					Timestamp: time.Now(),
				}
				_ = handler(msg)
			}
		}
	}()

	r.mu.Lock()
	r.subscribers[channel] = sub
	r.mu.Unlock()

	return nil
}

func (r *RealRedis) Unsubscribe(channel string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sub, ok := r.subscribers[channel]; ok {
		_ = sub.Close()
		delete(r.subscribers, channel)
	}
	return nil
}
