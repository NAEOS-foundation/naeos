package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	port := ":8080"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "microservices-go running")
	})
	log.Printf("Starting server on %s", port)
	srv := &http.Server{
		Addr:              port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
