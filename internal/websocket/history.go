// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type HistoryEntry struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	Room      string    `json:"room,omitempty"`
}

type History struct {
	entries  []HistoryEntry
	maxSize  int
	mu       sync.RWMutex
	entrySeq int64
}

func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &History{
		entries: make([]HistoryEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (h *History) Add(entryType string, payload any, room string) HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entrySeq++
	e := HistoryEntry{
		ID:        fmt.Sprintf("hist-%d", h.entrySeq),
		Type:      entryType,
		Payload:   payload,
		Timestamp: time.Now(),
		Room:      room,
	}

	if len(h.entries) >= h.maxSize {
		h.entries = h.entries[1:]
	}
	h.entries = append(h.entries, e)
	return e
}

func (h *History) Recent(n int) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || n > len(h.entries) {
		n = len(h.entries)
	}

	result := make([]HistoryEntry, n)
	copy(result, h.entries[len(h.entries)-n:])
	return result
}

func (h *History) Since(t time.Time) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []HistoryEntry
	for _, e := range h.entries {
		if e.Timestamp.After(t) || e.Timestamp.Equal(t) {
			result = append(result, e)
		}
	}
	return result
}

func (h *History) FilterByRoom(room string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []HistoryEntry
	for _, e := range h.entries {
		if e.Room == room {
			result = append(result, e)
		}
	}
	return result
}

func (h *History) FilterByType(eventType string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []HistoryEntry
	for _, e := range h.entries {
		if e.Type == eventType {
			result = append(result, e)
		}
	}
	return result
}

func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = h.entries[:0]
}

func (h *History) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

func (h *History) ReplayToClient(c *Client, n int) {
	h.mu.RLock()
	entries := make([]HistoryEntry, len(h.entries))
	copy(entries, h.entries)
	h.mu.RUnlock()

	if n <= 0 || n > len(entries) {
		n = len(entries)
	}

	for _, e := range entries[len(entries)-n:] {
		msg := Message{
			Type:    "history." + e.Type,
			Payload: e,
			Time:    e.Timestamp,
		}
		data, _ := json.Marshal(msg)
		c.send <- data
	}
}

func (h *History) ReplayRoomToClient(c *Client, room string, n int) {
	h.mu.RLock()
	var filtered []HistoryEntry
	for _, e := range h.entries {
		if e.Room == room {
			filtered = append(filtered, e)
		}
	}
	h.mu.RUnlock()

	if n <= 0 || n > len(filtered) {
		n = len(filtered)
	}

	for _, e := range filtered[len(filtered)-n:] {
		msg := Message{
			Type:    "history." + e.Type,
			Payload: e,
			Time:    e.Timestamp,
		}
		data, _ := json.Marshal(msg)
		c.send <- data
	}
}
