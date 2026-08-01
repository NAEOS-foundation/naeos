package order

import (
	"errors"
	"sync"
	"time"
)

type Order struct {
	ID        string    `json:"id"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Store interface {
	Save(o Order) error
	Find(id string) (Order, bool)
}

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Create(id string, total float64) Order {
	o := Order{ID: id, Total: total, Status: "created", CreatedAt: time.Now().UTC()}
	_ = h.store.Save(o)
	return o
}

func (h *Handler) Get(id string) (Order, bool) {
	return h.store.Find(id)
}

type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Order
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Order)}
}

func (s *MemoryStore) Save(o Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[o.ID] = o
	return nil
}

func (s *MemoryStore) Find(id string) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.items[id]
	return o, ok
}

var ErrNotFound = errors.New("order not found")
