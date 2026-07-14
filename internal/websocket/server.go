package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var defaultUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Server struct {
	clients         map[*Client]bool
	broadcast       chan []byte
	register        chan *Client
	unregister      chan *Client
	mu              sync.RWMutex
	allowedOrigins  []string
	upgrader        websocket.Upgrader
}

type Client struct {
	conn    *websocket.Conn
	server  *Server
	send    chan []byte
	id      string
	created time.Time
	writeMu sync.Mutex
}

func (c *Client) writeMessage(msgType int, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteMessage(msgType, data)
}

type Message struct {
	Type    string      `json:"type"`
	Payload any `json:"payload"`
	Time    time.Time   `json:"time"`
}

func NewServer() *Server {
	s := &Server{
		clients:   make(map[*Client]bool),
		broadcast: make(chan []byte, 256),
		register:  make(chan *Client),
		unregister: make(chan *Client),
	}
	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     s.checkOrigin,
	}
	return s
}

func (s *Server) SetAllowedOrigins(origins []string) {
	s.allowedOrigins = origins
}

func (s *Server) checkOrigin(r *http.Request) bool {
	if len(s.allowedOrigins) == 0 {
		return true
	}
	origin := r.Header.Get("Origin")
	for _, allowed := range s.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			s.sendSystemMessage("client connected")

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mu.Unlock()

		case message := <-s.broadcast:
			s.mu.RLock()
			var full []*Client
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					full = append(full, client)
				}
			}
			s.mu.RUnlock()
			if len(full) > 0 {
				s.mu.Lock()
				for _, client := range full {
					if _, ok := s.clients[client]; ok {
						close(client.send)
						delete(s.clients, client)
					}
				}
				s.mu.Unlock()
			}
		}
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	client := &Client{
		conn:    conn,
		server:  s,
		send:    make(chan []byte, 256),
		id:      generateID(),
		created: time.Now(),
	}

	s.register <- client
	go client.writePump()
	go client.readPump()
}

func (s *Server) Broadcast(eventType string, payload any) {
	msg := Message{
		Type:    eventType,
		Payload: payload,
		Time:    time.Now(),
	}
	data, _ := json.Marshal(msg)
	s.broadcast <- data
}

func (s *Server) sendSystemMessage(text string) {
	s.Broadcast("system", map[string]string{"message": text})
}

func (s *Server) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *Server) Stop() {
	s.mu.Lock()
	for client := range s.clients {
		client.writeMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server shutting down"))
		close(client.send)
		delete(s.clients, client)
	}
	s.mu.Unlock()
}

func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(65536)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var incoming Message
		if err := json.Unmarshal(msg, &incoming); err != nil {
			continue
		}

		switch incoming.Type {
		case "ping":
			pong, _ := json.Marshal(Message{Type: "pong", Time: time.Now()})
			c.writeMessage(websocket.TextMessage, pong)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.writeMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			c.writeMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.writeMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func generateID() string {
	return time.Now().Format("client-20060102150405.000000000")
}

// EventBroadcaster sends events to all connected clients
type EventBroadcaster struct {
	server *Server
}

func NewEventBroadcaster(server *Server) *EventBroadcaster {
	return &EventBroadcaster{server: server}
}

func (b *EventBroadcaster) PipelineStarted(pipelineID string) {
	b.server.Broadcast("pipeline.started", map[string]string{"pipeline_id": pipelineID})
}

func (b *EventBroadcaster) PipelineCompleted(pipelineID string, duration string) {
	b.server.Broadcast("pipeline.completed", map[string]string{"pipeline_id": pipelineID, "duration": duration})
}

func (b *EventBroadcaster) PipelineFailed(pipelineID string, errMsg string) {
	b.server.Broadcast("pipeline.failed", map[string]string{"pipeline_id": pipelineID, "error": errMsg})
}

func (b *EventBroadcaster) SpecValidated(valid bool, errors []string) {
	b.server.Broadcast("spec.validated", map[string]any{"valid": valid, "errors": errors})
}

func (b *EventBroadcaster) ArtifactGenerated(name string, path string) {
	b.server.Broadcast("artifact.generated", map[string]string{"name": name, "path": path})
}

func (b *EventBroadcaster) LogMessage(level string, message string) {
	b.server.Broadcast("log", map[string]string{"level": level, "message": message})
}
