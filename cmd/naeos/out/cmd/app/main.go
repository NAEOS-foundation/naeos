// Copyright 2026 NAEOS Foundation
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/example/sample-specification/internal/core"
	coreconfig "github.com/example/sample-specification/internal/core/config"
	corehttp "github.com/example/sample-specification/internal/core/http"
	coremiddleware "github.com/example/sample-specification/internal/core/middleware"
)

func main() {
	cfg := coreconfig.Load("config.yaml")
	handler := core.NewHandler(nil)
	_ = handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "hello from sample-specification on port %d", cfg.Port)
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
