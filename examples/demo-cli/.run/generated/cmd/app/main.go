// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/example/demo-app/internal/core"
	coreconfig "github.com/example/demo-app/internal/core/config"
	corehttp "github.com/example/demo-app/internal/core/http"
	coremiddleware "github.com/example/demo-app/internal/core/middleware"
)

func main() {
	cfg := coreconfig.Load("config.yaml")
	handler := core.NewHandler(nil)
	_ = handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "hello from demo-app on port %d", cfg.Port)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "api v1 ready")
	})
	mux.HandleFunc("/api/v1/resources", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "resources endpoint")
	})
	_ = corehttp.Handler{}
	wrapped := coremiddleware.LoggingMiddleware{}.Wrap(mux)
	log.Printf("listening on :%d", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), wrapped); err != nil {
		log.Fatal(err)
	}
}
