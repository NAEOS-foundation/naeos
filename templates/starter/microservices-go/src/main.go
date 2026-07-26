package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := ":8080"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "microservices-go running")
	})
	log.Printf("Starting server on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
