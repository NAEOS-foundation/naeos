package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/naeos-templates/microservices-go/internal/config"
	"github.com/naeos-templates/microservices-go/internal/order"
)

type createOrderRequest struct {
	ID    string  `json:"id"`
	Total float64 `json:"total"`
}

func main() {
	cfg := config.Load()

	store := order.NewMemoryStore()
	handler := order.NewHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		o := handler.Create(req.ID, req.Total)
		writeJSON(w, http.StatusCreated, o)
	})
	mux.HandleFunc("GET /orders/{id}", func(w http.ResponseWriter, r *http.Request) {
		o, ok := handler.Get(r.PathValue("id"))
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, o)
	})

	log.Printf("api-gateway listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
