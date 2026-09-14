// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"encoding/json"
	"sync"
	"time"
)

type Room struct {
	Name    string
	clients map[*Client]bool
	mu      sync.RWMutex
}

type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) GetOrCreate(name string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if r, ok := rm.rooms[name]; ok {
		return r
	}
	r := &Room{Name: name, clients: make(map[*Client]bool)}
	rm.rooms[name] = r
	return r
}

func (rm *RoomManager) Get(name string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	r, ok := rm.rooms[name]
	return r, ok
}

func (rm *RoomManager) Delete(name string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	delete(rm.rooms, name)
}

func (rm *RoomManager) List() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	var names []string
	for name := range rm.rooms {
		names = append(names, name)
	}
	return names
}

func (rm *RoomManager) ClientCount(roomName string) int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if r, ok := rm.rooms[roomName]; ok {
		r.mu.RLock()
		defer r.mu.RUnlock()
		return len(r.clients)
	}
	return 0
}

func (r *Room) Join(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[c] = true
}

func (r *Room) Leave(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, c)
}

func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *Room) Clients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var clients []*Client
	for c := range r.clients {
		clients = append(clients, c)
	}
	return clients
}

func (r *Room) Broadcast(eventType string, payload any) {
	msg := Message{
		Type:    eventType,
		Payload: payload,
		Time:    time.Now(),
	}
	data, _ := json.Marshal(msg)

	r.mu.RLock()
	var full []*Client
	for client := range r.clients {
		select {
		case client.send <- data:
		default:
			full = append(full, client)
		}
	}
	r.mu.RUnlock()

	if len(full) > 0 {
		r.mu.Lock()
		for _, client := range full {
			delete(r.clients, client)
		}
		r.mu.Unlock()
	}
}

func (r *Room) BroadcastExcept(eventType string, payload any, exclude *Client) {
	msg := Message{
		Type:    eventType,
		Payload: payload,
		Time:    time.Now(),
	}
	data, _ := json.Marshal(msg)

	r.mu.RLock()
	for client := range r.clients {
		if client == exclude {
			continue
		}
		select {
		case client.send <- data:
		default:
		}
	}
	r.mu.RUnlock()
}

func (s *Server) JoinRoom(roomName string, c *Client) {
	room := s.rooms.GetOrCreate(roomName)
	room.Join(c)

	s.mu.Lock()
	if s.clientRooms == nil {
		s.clientRooms = make(map[*Client]map[string]bool)
	}
	if s.clientRooms[c] == nil {
		s.clientRooms[c] = make(map[string]bool)
	}
	s.clientRooms[c][roomName] = true
	s.mu.Unlock()
}

func (s *Server) LeaveRoom(roomName string, c *Client) {
	if room, ok := s.rooms.Get(roomName); ok {
		room.Leave(c)
	}
	s.mu.Lock()
	if s.clientRooms != nil {
		delete(s.clientRooms[c], roomName)
	}
	s.mu.Unlock()
}

func (s *Server) BroadcastToRoom(roomName string, eventType string, payload any) {
	if room, ok := s.rooms.Get(roomName); ok {
		room.Broadcast(eventType, payload)
	}
}

func (s *Server) RoomManager() *RoomManager {
	return s.rooms
}
