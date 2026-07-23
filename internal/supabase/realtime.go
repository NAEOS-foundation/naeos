package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type RealtimeChannel struct {
	Topic      string
	Event      string
	Callback   func(msg RealtimeMessage)
	unsub      chan struct{}
}

type RealtimeMessage struct {
	Topic    string         `json:"topic"`
	Event    string         `json:"event"`
	Payload  map[string]any `json:"payload"`
	Ref      string         `json:"ref"`
}

type RealtimeClient struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	channels map[string][]*RealtimeChannel
	done     chan struct{}
}

func (c *Client) NewRealtimeClient() (*RealtimeClient, error) {
	u := c.config.URL + "/realtime/v1/websocket?vsn=1.0.0&apikey=" + c.config.AnonKey

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(u, http.Header{
		"Authorization": []string{"Bearer " + c.AuthToken()},
		"apikey":        []string{c.config.AnonKey},
	})
	if err != nil {
		return nil, fmt.Errorf("dial realtime: %w", err)
	}

	rc := &RealtimeClient{
		conn:     conn,
		channels: make(map[string][]*RealtimeChannel),
		done:     make(chan struct{}),
	}

	go rc.readLoop()
	return rc, nil
}

func (rc *RealtimeClient) readLoop() {
	defer close(rc.done)
	for {
		_, data, err := rc.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg RealtimeMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		rc.mu.Lock()
		channels := rc.channels[msg.Topic]
		globalChannels := rc.channels["*"]
		allChans := append([]*RealtimeChannel{}, channels...)
		allChans = append(allChans, globalChannels...)
		rc.mu.Unlock()

		for _, ch := range allChans {
			if ch.Event == "*" || ch.Event == msg.Event {
				select {
				case <-ch.unsub:
				default:
					ch.Callback(msg)
				}
			}
		}
	}
}

func (rc *RealtimeClient) Subscribe(topic, event string, callback func(msg RealtimeMessage)) (*RealtimeChannel, error) {
	ch := &RealtimeChannel{
		Topic:    topic,
		Event:    event,
		Callback: callback,
		unsub:    make(chan struct{}),
	}

	rc.mu.Lock()
	rc.channels[topic] = append(rc.channels[topic], ch)
	rc.mu.Unlock()

	msg := map[string]any{
		"topic":    topic,
		"event":    "phx_join",
		"payload":  map[string]any{},
		"ref":      fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	if err := rc.send(msg); err != nil {
		return nil, err
	}

	return ch, nil
}

func (rc *RealtimeClient) Unsubscribe(ch *RealtimeChannel) {
	close(ch.unsub)

	rc.mu.Lock()
	channels := rc.channels[ch.Topic]
	for i, c := range channels {
		if c == ch {
			rc.channels[ch.Topic] = append(channels[:i], channels[i+1:]...)
			break
		}
	}
	if len(rc.channels[ch.Topic]) == 0 {
		delete(rc.channels, ch.Topic)
	}
	rc.mu.Unlock()

	rc.send(map[string]any{
		"topic":   ch.Topic,
		"event":   "phx_leave",
		"payload": map[string]any{},
	})
}

func (rc *RealtimeClient) send(msg any) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.conn.WriteJSON(msg)
}

func (rc *RealtimeClient) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.conn.Close()
}

func (rc *RealtimeClient) Done() <-chan struct{} {
	return rc.done
}
